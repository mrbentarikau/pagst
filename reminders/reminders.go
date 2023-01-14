package reminders

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/common/scheduledevents2"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type ModalData struct {
	AuthorID  int64
	ChannelID int64
	Content   string
	GuildID   int64
	MessageID int64
}

type Plugin struct {
	RemindmeData ModalData
}

func RegisterPlugin() {
	err := common.GORM.AutoMigrate(&Reminder{}).Error
	if err != nil {
		panic(err)
	}

	p := &Plugin{}
	common.RegisterPlugin(p)
}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Reminders",
		SysName:  "reminders",
		Category: common.PluginCategoryMisc,
	}
}

type Reminder struct {
	gorm.Model
	UserID     string
	ChannelID  string
	GuildID    int64
	Message    string
	When       int64
	Repeat     int64
	AppCommand bool
}

func (r *Reminder) UserIDInt() (i int64) {
	i, _ = strconv.ParseInt(r.UserID, 10, 64)
	return
}

func (r *Reminder) ChannelIDInt() (i int64) {
	i, _ = strconv.ParseInt(r.ChannelID, 10, 64)
	return
}

func (r *Reminder) Trigger() error {
	reminderRepeat := "**Reminder** for"
	if r.Repeat > 0 {
		reminderRepeat = "**Repeated reminder** for"
		userID := r.UserIDInt()
		repeatDuration := time.Duration(r.Repeat) * time.Nanosecond
		when := time.Now().Add(repeatDuration)
		r.When = when.Unix()

		err := common.GORM.Save(r).Error
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		err = scheduledevents2.ScheduleEvent("reminders_check_user", r.GuildID, when, userID)
		if err != nil {
			logger.Info("Reminders repeat scheduler:", err)
		}
	} else {
		// remove the actual reminder
		rows := common.GORM.Delete(r).RowsAffected
		if rows < 1 {
			logger.Info("Tried to execute multiple reminders at once")
		}
	}

	logger.WithFields(logrus.Fields{"channel": r.ChannelID, "user": r.UserID, "message": r.Message, "id": r.ID}).Info("Triggered reminder")
	member, _ := bot.GetMember(r.GuildID, r.UserIDInt())
	if member != nil {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Reminder #%d from %s", r.Model.ID, common.ConfBotName.GetString()),
			Description: common.ReplaceServerInvites(r.Message, r.GuildID, "(removed-invite)"),
			Color:       int(rand.Int63n(16777215)),
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		var sendDM bool
		channelID := r.ChannelIDInt()
		if r.AppCommand {
			sendDM = true
			channel, err := common.BotSession.UserChannelCreate(r.UserIDInt())
			if err != nil {
				return err
			}

			msgSend := &discordgo.MessageSend{
				AllowedMentions: discordgo.AllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
				},
			}

			gs := bot.State.GetGuild(r.GuildID)
			gIcon := discordgo.EndpointGuildIcon(gs.ID, gs.Icon)
			embedInfo := fmt.Sprintf("DM from the server %s", gs.Name)

			embed.Footer = &discordgo.MessageEmbedFooter{
				Text:    embedInfo,
				IconURL: gIcon,
			}

			msgSend.Embeds = []*discordgo.MessageEmbed{embed}

			_, err = common.BotSession.ChannelMessageSendComplex(channel.ID, msgSend)
			if err != nil {
				embed.Footer = &discordgo.MessageEmbedFooter{}
				sendDM = false
			}
		}
		if !sendDM {
			mqueue.QueueMessage(&mqueue.QueuedElement{
				Source:       "reminder",
				SourceItemID: "",

				GuildID:      r.GuildID,
				ChannelID:    channelID,
				MessageEmbed: embed,
				MessageStr:   reminderRepeat + " <@" + r.UserID + ">", // : " + common.ReplaceServerInvites(r.Message, r.GuildID, "(removed-invite)"),
				AllowedMentions: discordgo.AllowedMentions{
					Users: []int64{r.UserIDInt()},
				},

				Priority: 10, // above all feeds
			})
		}
	}
	return nil
}

func GetUserReminders(userID int64) (results []*Reminder, err error) {
	err = common.GORM.Where(&Reminder{UserID: discordgo.StrID(userID)}).Find(&results).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func GetChannelReminders(channel int64) (results []*Reminder, err error) {
	err = common.GORM.Where(&Reminder{ChannelID: discordgo.StrID(channel)}).Find(&results).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func NewReminder(userID int64, guildID int64, channelID int64, message string, when time.Time, repeatDuration time.Duration, isAppCmd ...bool) (*Reminder, error) {
	var appCmd bool
	whenUnix := when.Unix()

	if len(isAppCmd) > 0 {
		appCmd = true
	}

	reminder := &Reminder{
		UserID:     discordgo.StrID(userID),
		ChannelID:  discordgo.StrID(channelID),
		Message:    message,
		When:       whenUnix,
		GuildID:    guildID,
		Repeat:     repeatDuration.Nanoseconds(),
		AppCommand: appCmd,
	}

	err := common.GORM.Create(reminder).Error
	if err != nil {
		return nil, err
	}

	err = scheduledevents2.ScheduleEvent("reminders_check_user", guildID, when, userID)
	// err = scheduledevents.ScheduleEvent("reminders_check_user:"+strconv.FormatInt(whenUnix, 10), discordgo.StrID(userID), when)
	return reminder, err
}

/*
func checkUserEvtHandlerLegacy(evt string) error {
	split := strings.Split(evt, ":")
	if len(split) < 2 {
		logger.Error("Handled invalid check user scheduled event: ", evt)
		return nil
	}

	parsed, _ := strconv.ParseInt(split[1], 10, 64)
	reminders, err := GetUserReminders(parsed)
	if err != nil {
		return err
	}

	now := time.Now()
	nowUnix := now.Unix()
	for _, v := range reminders {
		if v.When <= nowUnix {
			err := v.Trigger()
			if err != nil {
				// Try again
				return err
			}
		}
	}

	return nil
}
*/
