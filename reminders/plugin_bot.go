package reminders

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/scheduledevents2"
	seventsmodels "github.com/mrbentarikau/pagst/common/scheduledevents2/models"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/jinzhu/gorm"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
)

var logger = common.GetPluginLogger(&Plugin{})

var _ bot.BotInitHandler = (*Plugin)(nil)
var _ commands.CommandProvider = (*Plugin)(nil)

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p, cmds...)
}

func (p *Plugin) BotInit() {
	// scheduledevents.RegisterEventHandler("reminders_check_user", checkUserEvtHandlerLegacy)
	scheduledevents2.RegisterHandler("reminders_check_user", int64(0), checkUserScheduledEvent)
	scheduledevents2.RegisterLegacyMigrater("reminders_check_user", migrateLegacyScheduledEvents)
}

// Reminder management commands
var cmds = []*commands.YAGCommand{
	{
		CmdCategory:  commands.CategoryTool,
		Name:         "Remindme",
		Description:  "Schedules a reminder, example: 'remindme 1h30min are you alive still?'\nSwitch -repeat will repeat the reminder starting with min duration of 1 hour.",
		Aliases:      []string{"remind", "reminder"},
		RequiredArgs: 2,
		Arguments: []*dcmd.ArgDef{
			{Name: "Time", Type: &commands.DurationArg{}},
			{Name: "Message", Type: dcmd.String},
			//{Name: "date", Help: "Enables specific date", Type: dcmd.String},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "channel", Type: dcmd.Channel},
			{Name: "repeat", Help: "Repeat the reminder at set duration"},
		},
		SlashCommandEnabled: true,
		DefaultEnabled:      true,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			currentReminders, _ := GetUserReminders(parsed.Author.ID)
			if len(currentReminders) >= 25 {
				return "You can have a maximum of 25 active reminders, list your reminders with the `reminders` command", nil
			}

			fromNow := parsed.Args[0].Value.(time.Duration)

			repeatDuration := time.Duration(0)
			if parsed.Switch("repeat").Value != nil && parsed.Switch("repeat").Value.(bool) && fromNow.Hours() >= 1 {
				repeatDuration = fromNow
			}

			durString := common.HumanizeDuration(common.DurationPrecisionSeconds, fromNow)
			when := time.Now().Add(fromNow)
			tUnix := fmt.Sprint(when.Unix())

			if when.After(time.Now().Add(time.Hour * 24 * 366)) {
				return "Can be max 365 days from now...", nil
			}

			id := parsed.ChannelID
			if c := parsed.Switch("channel"); c.Value != nil {
				id = c.Value.(*dstate.ChannelState).ID

				hasPerms, err := bot.AdminOrPermMS(parsed.GuildData.GS.ID, id, parsed.GuildData.MS, discordgo.PermissionSendMessages|discordgo.PermissionReadMessages)
				if err != nil {
					return "Failed checking permissions, please try again or join the support server.", err
				}

				if !hasPerms {
					return "You do not have permissions to send messages there", nil
				}
			}

			_, err := NewReminder(parsed.Author.ID, parsed.GuildData.GS.ID, id, parsed.Args[1].Str(), when, repeatDuration)
			if err != nil {
				return nil, err
			}

			return "Set a reminder in " + durString + " from now (<t:" + tUnix + ":f>)\nView reminders with the reminders command", nil
		},
	},
	{
		CmdCategory:         commands.CategoryTool,
		Name:                "Reminders",
		Description:         "Lists your active reminders",
		SlashCommandEnabled: true,
		DefaultEnabled:      true,
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "raw", Help: "Raw, legacy output"},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			pagination := true
			maxLength := 5

			if parsed.Switches["raw"].Value != nil && parsed.Switches["raw"].Value.(bool) {
				pagination = false
			}

			currentReminders, err := GetUserReminders(parsed.Author.ID)
			if err != nil {
				return nil, err
			}

			if !pagination {
				out := "Your reminders:\n"
				out += stringReminders(currentReminders, false)
				out += "```Remove a reminder with 'delreminder/rmreminder (id)' where id is the first number for each reminder above.\nTo clear all reminders, use 'delreminder' with the '-a' switch.```"
				return out, nil
			}

			pm, err := paginatedmessages.CreatePaginatedMessage(
				parsed.GuildData.GS.ID, parsed.ChannelID, 1, int(math.Ceil(float64(len(currentReminders))/float64(maxLength))), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					paginatedEmbed := embedCreator(currentReminders, i, maxLength, 0, parsed)
					return paginatedEmbed, nil
				})
			if err != nil {
				return fmt.Sprintf("Something went wrong: %s", err), nil
			}

			return pm, nil
		},
	},
	{
		CmdCategory:         commands.CategoryTool,
		Name:                "CReminders",
		Aliases:             []string{"channelreminders"},
		Description:         "Lists reminders in channel, only users with 'manage channel' permissions can use this.",
		SlashCommandEnabled: true,
		DefaultEnabled:      true,
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "raw", Help: "Raw, legacy output"},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			pagination := true
			maxLength := 5

			if parsed.Switches["raw"].Value != nil && parsed.Switches["raw"].Value.(bool) {
				pagination = false
			}

			ok, err := bot.AdminOrPermMS(parsed.GuildData.GS.ID, parsed.ChannelID, parsed.GuildData.MS, discordgo.PermissionManageChannels)
			if err != nil {
				return nil, err
			}
			if !ok {
				return "You do not have access to this command (requires manage channel permission)", nil
			}

			currentReminders, err := GetChannelReminders(parsed.ChannelID)
			if err != nil {
				return nil, err
			}

			if !pagination {
				out := "Reminders in this channel:\n"
				out += stringReminders(currentReminders, true)
				out += "```Remove a reminder with 'delreminder/rmreminder (id)' where id is the first number for each reminder above```"
				return out, nil
			}

			pm, err := paginatedmessages.CreatePaginatedMessage(
				parsed.GuildData.GS.ID, parsed.ChannelID, 1, int(math.Ceil(float64(len(currentReminders))/float64(maxLength))), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					paginatedEmbed := embedCreator(currentReminders, i, maxLength, 1, parsed)
					return paginatedEmbed, nil
				})
			if err != nil {
				return fmt.Sprintf("Something went wrong: %s", err), nil
			}

			return pm, nil
		},
	},
	{
		CmdCategory:  commands.CategoryTool,
		Name:         "DelReminder",
		Aliases:      []string{"rmreminder"},
		Description:  "Deletes a reminder. You can delete reminders from other users provided you are running this command in the same guild the reminder was created in and have the Manage Channel permission in the channel the reminder was created in.\n\n Switch -a combined with -userid flag give proper roles the possibility to delete all user's reminders in that server.",
		RequiredArgs: 0,
		Arguments: []*dcmd.ArgDef{
			{Name: "ID", Type: dcmd.Int},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "a", Help: "All"},
			{Name: "userid", Type: dcmd.Int, Default: 0, Help: "userID"},
		},
		SlashCommandEnabled: true,
		DefaultEnabled:      true,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			var reminder Reminder

			clearAll := parsed.Switch("a").Value != nil && parsed.Switch("a").Value.(bool)
			if clearAll {
				var userID int64
				userIDArg := parsed.Switch("userid").Int64()
				if userIDArg > 0 && parsed.Author.ID != userIDArg {
					ok, err := bot.AdminOrPermMS(parsed.GuildData.GS.ID, parsed.GuildData.CS.ID, parsed.GuildData.MS, discordgo.PermissionManageChannels)

					if err != nil {
						return nil, err
					}

					if !ok {
						return "You need manage server permission in the server the reminder is in to delete reminders that are not your own", nil
					}

					userID = userIDArg
				} else {
					userID = parsed.Author.ID
				}

				db := common.GORM.Where("user_id = ?", userID).Delete(&reminder)

				err := db.Error
				if err != nil {
					return "Error clearing reminders", err
				}

				count := db.RowsAffected
				if count == 0 {
					return "No reminders to clear", nil
				}
				return fmt.Sprintf("Cleared %d reminders", count), nil
			}

			if len(parsed.Args) == 0 || parsed.Args[0].Value == nil {
				return "No reminder ID provided", nil
			}

			err := common.GORM.Where(parsed.Args[0].Int()).First(&reminder).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return "No reminder by that id found", nil
				}
				return "Error retrieving reminder", err
			}

			// Check perms
			if reminder.UserID != discordgo.StrID(parsed.Author.ID) {
				if reminder.GuildID != parsed.GuildData.GS.ID {
					return "You can only delete reminders that are not your own in the guild the reminder was originally created", nil
				}
				ok, err := bot.AdminOrPermMS(reminder.GuildID, reminder.ChannelIDInt(), parsed.GuildData.MS, discordgo.PermissionManageChannels)
				if err != nil {
					return nil, err
				}
				if !ok {
					return "You need manage channel permission in the channel the reminder is in to delete reminders that are not your own", nil
				}
			}

			// Do the actual deletion
			err = common.GORM.Delete(reminder).Error
			if err != nil {
				return nil, err
			}

			// Check if we should remove the scheduled event
			currentReminders, err := GetUserReminders(reminder.UserIDInt())
			if err != nil {
				return nil, err
			}

			delMsg := fmt.Sprintf("Deleted reminder **#%d**: '%s'", reminder.ID, limitString(reminder.Message))

			// If there is another reminder with the same timestamp, do not remove the scheduled event
			for _, v := range currentReminders {
				if v.When == reminder.When {
					return delMsg, nil
				}
			}

			return delMsg, nil
		},
	},
}

func stringReminders(reminders []*Reminder, displayUsernames bool) string {
	out := ""
	var repeatedReminder string
	for _, v := range reminders {
		parsedCID, _ := strconv.ParseInt(v.ChannelID, 10, 64)

		t := time.Unix(v.When, 0)
		tUnix := t.Unix()
		timeFromNow := common.HumanizeTime(common.DurationPrecisionMinutes, t)

		if v.Repeat > 0 {
			repeatedReminder = "- repeated reminder"
		} else {
			repeatedReminder = ""
		}
		if !displayUsernames {
			channel := "<#" + discordgo.StrID(parsedCID) + ">"
			out += fmt.Sprintf("**%d**: %s: '%s' - %s from now (<t:%d:f>) %s\n", v.ID, channel, limitString(v.Message), timeFromNow, tUnix, repeatedReminder)
		} else {
			member, _ := bot.GetMember(v.GuildID, v.UserIDInt())
			username := "Unknown user"
			if member != nil {
				username = member.User.Username
			}
			out += fmt.Sprintf("**%d**: %s: '%s' - %s from now (<t:%d:f>) %s\n", v.ID, username, limitString(v.Message), timeFromNow, tUnix, repeatedReminder)
		}
	}
	return out
}

func checkUserScheduledEvent(evt *seventsmodels.ScheduledEvent, data interface{}) (retry bool, err error) {
	// !important! the evt.GuildID can be 1 in cases where it was migrated from the legacy scheduled event system

	userID := *data.(*int64)

	reminders, err := GetUserReminders(userID)
	if err != nil {
		return true, err
	}

	now := time.Now()
	nowUnix := now.Unix()
	for _, v := range reminders {
		if v.When <= nowUnix {
			err := v.Trigger()
			if err != nil {
				// possibly try again
				return scheduledevents2.CheckDiscordErrRetry(err), err
			}
		}
	}

	return false, nil
}

func migrateLegacyScheduledEvents(t time.Time, data string) error {
	split := strings.Split(data, ":")
	if len(split) < 2 {
		logger.Error("invalid check user scheduled event: ", data)
		return nil
	}

	parsed, _ := strconv.ParseInt(split[1], 10, 64)

	return scheduledevents2.ScheduleEvent("reminders_check_user", 1, t, parsed)
}

func limitString(s string) string {
	if utf8.RuneCountInString(s) < 50 {
		return s
	}

	runes := []rune(s)
	return string(runes[:47]) + "..."
}

func embedCreator(currentReminders []*Reminder, i, ml, flag int, parsed *dcmd.Data) *discordgo.MessageEmbed {
	//var username string
	member := parsed.GuildData.MS
	embedTitle := []string{"Your reminders:", "Reminders in this channel:"}
	embedDescription := []string{"Remove a reminder with `delreminder/rmreminder (id)` where id is the first number for each reminder above.\nTo clear all reminders, use `delreminder` with the `-a` switch.", "Remove a reminder with `delreminder/rmreminder (id)` where id is the first number for each reminder above."}

	embed := &discordgo.MessageEmbed{
		Title:       embedTitle[flag],
		Color:       int(rand.Int63n(16777215)),
		Description: embedDescription[flag],
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: discordgo.EndpointUserAvatar(member.User.ID, member.User.Avatar),
		},
	}
	for k, v := range currentReminders[i*ml:] {
		var username string
		if k < ml {
			if flag > 0 {
				remindGuy, _ := bot.GetMember(v.GuildID, v.UserIDInt())
				username = "Unknown user"
				if remindGuy != nil {
					username = remindGuy.User.Username
				}
			}
			repeat := ""
			if v.Repeat > 0 {
				repeat = "Repeated"
			}
			t := time.Unix(v.When, 0)
			tStr := t.Format(time.RFC822)
			timeFromNow := common.HumanizeTime(common.DurationPrecisionMinutes, t)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("**#%d** %s", v.ID, username), Value: fmt.Sprintf("<#%s>\n%s\n%s from now (%s)\n%s", v.ChannelID, v.Message, timeFromNow, tStr, repeat)})
		}
	}
	return embed
}
