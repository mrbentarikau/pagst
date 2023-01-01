package reminders

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/rsvp"
	"github.com/mrbentarikau/pagst/timezonecompanion"
	zws "github.com/trubitsyn/go-zero-width"
)

var (
	AddReminderCustomID = "Add Reminder"
	RemindmeExtraNotes  = "remindme_modal_extra_notes"
	RemindmeWhenDate    = "remindme_modal_when_date"
)

func (p *Plugin) createRemindMeModal() *commands.YAGCommand {
	return &commands.YAGCommand{
		CmdCategory:          commands.CategoryFun,
		Name:                 AddReminderCustomID,
		CmdNoRun:             true,
		RunInDM:              false,
		HideFromHelp:         true,
		HideFromCommandsPage: false,
		DefaultEnabled:       false,

		ApplicationCommandEnabled: true,
		ApplicationCommandType:    3,
		IsResponseModal:           true,
		IsResponseEphemeral:       true,
		NameLocalizations:         remindmeLocalizations["AppName"],

		Arguments: []*dcmd.ArgDef{
			{Name: "Message", Type: dcmd.String},
		},

		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			content := parsed.Args[0].Value.(string)
			cID := parsed.ChannelID
			mID := parsed.SlashCommandTriggerData.Interaction.DataCommand.TargetID

			message, _ := parsed.Session.ChannelMessage(cID, mID)
			if content == "" {
				if message != nil {
					if len(message.Embeds) > 0 {
						content = "Reminder for Embeds..."
					} else if len(message.Attachments) > 0 {
						content = "Reminder for Attachments..."
					} else {
						content = message.Content
					}
				}
			}

			p.RemindmeData = ModalData{
				AuthorID:  parsed.Author.ID,
				ChannelID: cID,
				GuildID:   parsed.GuildData.GS.ID,
				MessageID: mID,
				Content:   content,
			}

			ic := parsed.SlashCommandTriggerData.Interaction
			locale := ic.Locale

			params := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID:   AddReminderCustomID,
					Title:      getTranslation("ModalTitle", locale),
					Components: getModalComponents(locale),
					Flags:      64,
				},
			}

			err := parsed.Session.CreateInteractionResponse(ic.ID, ic.Token, params)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	}
}

func getTranslation(key string, locale discordgo.Locale) string {
	if localeValue, ok := (*remindmeLocalizations[key])[locale]; !ok {
		return (*remindmeLocalizations[key])["en-GB"]
	} else {
		return localeValue
	}
}

func getModalComponents(locale discordgo.Locale) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		/*
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:  "remindme_modal_content",
						Label:     "Reminder Content:",
						Style:     discordgo.TextInputShort,
						Required:  true,
						MaxLength: len(content),
						Value:     content,
					},
				},
			},
		*/
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:  RemindmeExtraNotes,
					Label:     getTranslation("ExtraNotes", locale),
					Style:     discordgo.TextInputParagraph,
					Required:  false,
					MaxLength: 280,
				},
			},
		},

		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    RemindmeWhenDate,
					Label:       getTranslation("WhenDate", locale),
					Style:       discordgo.TextInputShort,
					Required:    false,
					MaxLength:   32,
					Placeholder: fmt.Sprintf("e.g 5m, %d-12-31 12:34, tomorrow, etc", time.Now().UTC().Year()),
				},
			},
		},
	}
}

func (p *Plugin) handleRemindmeInteractionCreate(ic *discordgo.InteractionCreate) {
	if ic.Type != discordgo.InteractionModalSubmit {
		// Not a modal interaction
		return
	}

	if ic.DataCommand == nil && ic.DataModal == nil && (p.RemindmeData == ModalData{}) {
		// Modal interaction had no data
		return
	}

	// default 1 hour reminder delay
	duration := time.Minute * 20
	when := time.Now().UTC().Add(duration)

	if ic.Type == discordgo.InteractionModalSubmit && ic.ModalSubmitData().CustomID == AddReminderCustomID {
		var err error
		reminderContent := p.RemindmeData.Content
		icDataModelComps := ic.DataModal.Components

		extraNotes := icDataModelComps[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		if zws.RemoveZeroWidthCharacters(strings.TrimSpace(extraNotes)) != "" {
			reminderContent += "\n\n```Extra Notes:\n" + extraNotes + "```"
		}

		reminderContent += fmt.Sprintf("\n\nJump to message: [jump](%schannels/%d/%d/%d)", discordgo.EndpointDiscord, p.RemindmeData.GuildID, p.RemindmeData.ChannelID, p.RemindmeData.MessageID)

		var now time.Time
		durationStr := icDataModelComps[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		if durationStr != "" {
			duration, err = common.ParseDuration(durationStr)
			if err != nil || duration == 0 {
				dateParser := createDateParser()

				registeredTimezone := timezonecompanion.GetUserTimezone(p.RemindmeData.AuthorID)
				if registeredTimezone == nil || rsvp.UTCRegex.MatchString(durationStr) {
					registeredTimezone = time.UTC
				}

				now = time.Now().In(registeredTimezone)
				t, err := dateParser.Parse(durationStr, now)
				if err == nil && t != nil {
					when = t.Time
				}
			} else {
				when = time.Now().UTC().Add(duration)
			}
		}

		err = common.BotSession.CreateInteractionResponse(ic.ID, ic.Token, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: 64},
		})
		if err != nil {
			logger.WithError(err).Error("Failed creating AddRemindme Deferred Response")
			return
		}

		repeatDuration := time.Duration(0)
		channelID := p.RemindmeData.ChannelID
		durString := common.HumanizeDuration(common.DurationPrecisionSeconds, time.Until(when))
		tUnix := fmt.Sprint(when.Unix())
		responseContent := "Set a reminder in " + durString + " from now (<t:" + tUnix + ":f>)\nView reminders with the reminders command..."

		var properTime = true
		if when.After(time.Now().Add(time.Hour * 24 * 366)) {
			responseContent = "Reminder can only be set max 365 days from now..."
			properTime = false
		}

		if when.Before(time.Now().Add(time.Minute*4 + time.Second*59)) {
			responseContent = "Reminder must be set at least **5 minutes** from now..."
			responseContent += fmt.Sprintf("```\n\nwhen: %s\nvs\nnow: %s```\nUse `settimezone` to adjust your timezone better.", when.Format(time.RFC822), now.Format(time.RFC822))
			properTime = false
		}

		if properTime {
			_, err = NewReminder(p.RemindmeData.AuthorID, p.RemindmeData.GuildID, channelID, reminderContent, when, repeatDuration, true)
			if err != nil {
				return
			}
		}

		colour := int(rand.Int63n(16777215))
		if !properTime {
			colour = 0xd2322d
		}

		responseEmbed := &discordgo.MessageEmbed{
			Color:       colour,
			Description: responseContent,
		}

		params := &discordgo.WebhookParams{
			Embeds:          []*discordgo.MessageEmbed{responseEmbed},
			AllowedMentions: &discordgo.AllowedMentions{},
			Flags:           64,
		}

		_, err = common.BotSession.CreateFollowupMessage(common.BotApplication.ID, ic.Token, params)
		if err != nil {
			logger.WithError(err).Error("Failed creating AddRemindme FollowupMessage")
			return
		}

	}
}
