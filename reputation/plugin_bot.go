package reputation

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/mrbentarikau/pagst/reputation/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/web"
)

var _ bot.BotInitHandler = (*Plugin)(nil)
var _ commands.CommandProvider = (*Plugin)(nil)

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p, cmds...)
}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLastLegacy(p, handleMessageCreate, eventsystem.EventMessageCreate)
}

var thanksRegex = common.ThanksRegex

var repDisabledError = "**Rep command is disabled on this server. Enable it from the control panel.**"

func handleMessageCreate(evt *eventsystem.EventData) {
	var inbuiltThanks, customThanks bool

	msg := evt.MessageCreate()

	conf, err := GetConfig(evt.Context(), msg.GuildID)
	if err != nil || !conf.Enabled || (conf.DisableThanksDetection && conf.DisableCustomThanksDetection) {
		return
	}

	if !bot.IsNormalUserMessage(msg.Message) {
		return
	}

	if len(msg.Mentions) < 1 || msg.GuildID == 0 || msg.Author.Bot {
		return
	}

	if !evt.HasFeatureFlag(featureFlagThanksEnabled) && !evt.HasFeatureFlag(featureFlagCustomThanksEnabled) {
		return
	}

	cState := evt.CSOrThread()
	if cState == nil {
		return // No channel state, ignore
	}

	channelID := msg.ChannelID
	// Check if thanks detection is allowed in the parent channel
	if cState.Type.IsThread() {
		channelID = cState.ParentID
	}
	if !isThanksDetectionAllowedInChannel(conf, channelID) {
		return
	}

	who := msg.Mentions[0]
	if who.ID == msg.Author.ID {
		return
	}

	if !conf.DisableThanksDetection {
		if thanksRegex.MatchString(msg.Content) {
			inbuiltThanks = true
		}
	}

	if !conf.DisableCustomThanksDetection && len(conf.CustomThanksRegex) > 0 {
		customThanksRegex, err := regexp.Compile(conf.CustomThanksRegex)
		if err == nil && customThanksRegex.MatchString(msg.Content) {
			customThanks = true
		}
	}

	if inbuiltThanks || customThanks {
		target, err := bot.GetMember(msg.GuildID, who.ID)
		sender := dstate.MemberStateFromMember(msg.Member)
		if err != nil {
			logger.WithError(err).Error("Failed retrieving target member")
			return
		}

		if err = CanModifyRep(conf, sender, target); err != nil {
			return
		}

		err = ModifyRep(evt.Context(), conf, msg.GuildID, sender, target, 1)
		if err != nil {
			if err == ErrCooldown {
				// Ignore this error silently
				return
			}
			logger.WithError(err).Error("Failed giving rep")
			return
		}

		go analytics.RecordActiveUnit(msg.GuildID, &Plugin{}, "auto_add_rep")

		newScore, newRank, err := GetUserStats(msg.GuildID, who.ID)
		if err != nil {
			newScore = -1
			newRank = -1
			logger.WithError(err).Error("Failed retrieving target stats")
			return
		}

		content := fmt.Sprintf("Gave +1 %s to **%s** (current: `#%d` - `%d`)", conf.PointsName, who.Mention(), newRank, newScore)
		common.BotSession.ChannelMessageSend(msg.ChannelID, content)
	}
}

var cmds = []*commands.YAGCommand{
	{
		CmdCategory:  commands.CategoryFun,
		Name:         "TakeRep",
		Aliases:      []string{"-", "tr", "trep", "-rep"},
		Description:  "Takes away absolute value of rep from someone.",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.User},
			{Name: "Num", Type: dcmd.Int, Default: 1},
		},
		ApplicationCommandEnabled: true,
		DefaultEnabled:            false,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			parsed.Args[1].Value = -int(math.Abs(parsed.Args[1].Float64()))
			return CmdGiveRep(parsed)
		},
	},
	{
		CmdCategory:               commands.CategoryFun,
		Name:                      "GiveRep",
		Aliases:                   []string{"+", "gr", "grep", "+rep"},
		Description:               "Gives absolute value of rep to someone.",
		RequiredArgs:              1,
		ApplicationCommandEnabled: true,
		DefaultEnabled:            false,
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.User},
			{Name: "Num", Type: dcmd.Int, Default: 1},
		},

		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			parsed.Args[1].Value = int(math.Abs(parsed.Args[1].Float64()))
			return CmdGiveRep(parsed)
		},
	},
	{
		CmdCategory:               commands.CategoryFun,
		Name:                      "SetRep",
		Aliases:                   []string{"SetRepID"}, // alias for legacy reasons, used to be a standalone command
		Description:               "Sets someones rep, this is an admin command and bypasses cooldowns and other restrictions.",
		RequiredArgs:              2,
		ApplicationCommandEnabled: true,
		DefaultEnabled:            false,
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.UserID},
			{Name: "Num", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			conf, err := GetConfig(parsed.Context(), parsed.GuildData.GS.ID)
			if err != nil {
				return "An error occurred while finding the server config", err
			}

			if !conf.Enabled {
				return repDisabledError, nil
			}

			if !IsAdmin(parsed.GuildData.GS, parsed.GuildData.MS, conf) {
				return "You're not a reputation admin. (no manage server perms and no rep admin role)", nil
			}

			targetID := parsed.Args[0].Int64()
			targetUsername := strconv.FormatInt(targetID, 10)
			targetMember, _ := bot.GetMember(parsed.GuildData.GS.ID, targetID)
			if targetMember != nil {
				targetUsername = targetMember.User.Username
			} else {
				prevMember, err := userPresentInRepLog(targetID, parsed.GuildData.GS.ID, parsed)
				if err != nil {
					return nil, err
				}
				if !prevMember {
					return "Invalid User. This user never received/gave rep in this server", nil
				}
			}

			err = SetRep(parsed.Context(), parsed.GuildData.GS.ID, parsed.GuildData.MS.User.ID, targetID, int64(parsed.Args[1].Int()))
			if err != nil {
				return nil, err
			}

			return fmt.Sprintf("Set **%s** %s to `%d`", targetUsername, conf.PointsName, parsed.Args[1].Int()), nil
		},
	},
	{
		CmdCategory:               commands.CategoryFun,
		Name:                      "DelRep",
		Description:               "Deletes someone from the reputation list completely, this cannot be undone.",
		RequiredArgs:              1,
		ApplicationCommandEnabled: true,
		DefaultEnabled:            false,
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.UserID},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			conf, err := GetConfig(parsed.Context(), parsed.GuildData.GS.ID)
			if err != nil {
				return "An error occurred while finding the server config", err
			}

			if !conf.Enabled {
				return repDisabledError, nil
			}

			if !IsAdmin(parsed.GuildData.GS, parsed.GuildData.MS, conf) {
				return "You're not an reputation admin. (no manage servers perms and no rep admin role)", nil
			}

			target := parsed.Args[0].Int64()

			err = DelRep(parsed.Context(), parsed.GuildData.GS.ID, target)
			if err != nil {
				return nil, err
			}

			return fmt.Sprintf("Deleted all of %d's %s.", target, conf.PointsName), nil
		},
	},
	{
		CmdCategory:               commands.CategoryFun,
		Name:                      "RepLog",
		Aliases:                   []string{"replogs"},
		Description:               "Shows the rep log for the specified user. Times are in UTC.",
		RequiredArgs:              0,
		ApplicationCommandEnabled: true,
		DefaultEnabled:            false,
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.UserID},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "raw", Help: "Raw output"},
			{Name: "p", Type: dcmd.Int, Help: "Page number", Default: 0},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			var paginatedView, pageBool bool
			var err error
			limiter := 20
			pageNum := 1
			paginatedView = true

			if parsed.Switches["raw"].Value != nil && parsed.Switches["raw"].Value.(bool) {
				paginatedView = false
			}

			if parsed.Switch("p").Int() > 0 {
				pageBool = true
				pageNum = parsed.Switch("p").Int()
			}

			conf, err := GetConfig(parsed.Context(), parsed.GuildData.GS.ID)
			if err != nil {
				return "An error occurred while finding the server config", err
			}

			if !IsAdmin(parsed.GuildData.GS, parsed.GuildData.MS, conf) {
				return "You're not an reputation admin. (no manage servers perms and no rep admin role)", nil
			}

			targetID := parsed.Author.ID
			if parsed.Args[0].Value != nil {
				targetID = parsed.Args[0].Int64()
			}

			guildID := parsed.GuildData.GS.ID

			var targetUsername string
			target, err := bot.GetMember(guildID, targetID)
			if err != nil {
				targetUsername = strconv.FormatInt(targetID, 10)
			} else {
				targetUsername = target.User.String()
			}

			var logEntries []*RepLogEntry
			if pageBool || !paginatedView {
				logEntries, err = RepLog(guildID, targetID, paginatedView, pageNum, limiter)
				if err != nil {
					return nil, err
				}
			} else {
				logEntries, err = RepLog(guildID, targetID, paginatedView)
				if err != nil {
					return nil, err
				}
			}

			if len(logEntries) < 1 {
				return "No entries...", nil
			}

			var pm *paginatedmessages.PaginatedMessage
			if paginatedView {
				pm, err = paginatedmessages.CreatePaginatedMessage(
					parsed.GuildData.GS.ID, parsed.ChannelID, 1, int(math.Ceil(float64(len(logEntries))/float64(limiter))), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
						i := page - 1

						out := fmt.Sprint("Starting from page ", pageNum)
						out += (repLogCreator(logEntries, guildID, targetID, limiter, i, paginatedView)).String()
						paginatedEmbed := &discordgo.MessageEmbed{
							Title:       "Reputation Log for " + targetUsername,
							Description: out,
						}
						return paginatedEmbed, nil
					})
				if err != nil {
					return "Something went wrong...", err
				}
			} else {
				out := repLogCreator(logEntries, guildID, targetID, limiter, pageNum-1, paginatedView)
				out.WriteString(fmt.Sprint("Page ", pageNum))
				return out.String(), nil
			}

			return pm, nil
		},
	},

	{
		CmdCategory: commands.CategoryFun,
		Name:        "Rep",
		Description: "Shows member's current reputation and rank...",
		Arguments: []*dcmd.ArgDef{
			{Name: "User", Type: dcmd.User},
		},
		ApplicationCommandEnabled: true,
		ApplicationCommandType:    2,
		DefaultEnabled:            false,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			target := parsed.Author
			if parsed.Args[0].Value != nil {
				target = parsed.Args[0].Value.(*discordgo.User)
			}

			conf, err := GetConfig(parsed.Context(), parsed.GuildData.GS.ID)
			if err != nil {
				return "An error occurred finding the server config", err
			}

			score, rank, err := GetUserStats(parsed.GuildData.GS.ID, target.ID)
			if err != nil {
				if err == ErrUserNotFound {
					rank = -1
				} else {
					return nil, err
				}
			}

			receiver, err := bot.GetMember(parsed.GuildData.GS.ID, target.ID)
			if err != nil {
				return nil, err
			}

			rankStr := "#ω"
			if rank != -1 {
				rankStr = strconv.FormatInt(int64(rank), 10)
			}

			return fmt.Sprintf("**%s**: **%d** %s (#**%s**)", receiver.User.Username, score, conf.PointsName, rankStr), nil
		},
	},
	{
		CmdCategory: commands.CategoryFun,
		Name:        "TopRep",
		Description: "Shows rep leader board on the server",
		Arguments: []*dcmd.ArgDef{
			{Name: "Page", Type: dcmd.Int, Default: 0},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "user", Help: "User to search for in the leader board", Type: dcmd.UserID},
		},
		ApplicationCommandEnabled: true,
		DefaultEnabled:            true,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			const (
				countEntriesQuery = "SELECT COUNT(1) FROM reputation_users WHERE guild_id = $1"
				userQuery         = `
					SELECT pos
					FROM (
						SELECT ROW_NUMBER() OVER (ORDER BY points DESC) AS pos, user_id
						FROM reputation_users
						WHERE guild_id = $1
					) as ordered_users
					WHERE user_id = $2`
			)
			page := parsed.Args[0].Int()
			var maxCount int
			err := common.PQ.QueryRow(countEntriesQuery, parsed.GuildData.GS.ID).Scan(&maxCount)
			if err != nil {
				return "Failed finding leader board entries...", err
			}

			if uID := parsed.Switch("user").Int64(); uID != 0 {
				var pos int
				err := common.PQ.QueryRow(userQuery, parsed.GuildData.GS.ID, uID).Scan(&pos)
				if err != nil {
					if err == sql.ErrNoRows {
						return "Could not find that user on the leader board", nil
					}
					return "Failed finding that user on the leader board, try again", err
				}

				page = (pos-1)/15 + 1 // pos and page are both one-based
			}

			if page < 1 {
				page = 1
			}

			if parsed.Context().Value(paginatedmessages.CtxKeyNoPagination) != nil {
				return topRepPager(parsed.GuildData.GS.ID, nil, page)
			}

			pm, err := paginatedmessages.CreatePaginatedMessage(parsed.GuildData.GS.ID, parsed.ChannelID, page, int(math.Ceil(float64(maxCount)/15)), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
				return topRepPager(parsed.GuildData.GS.ID, p, page)
			})

			return pm, err
		},
	},
}

func topRepPager(guildID int64, p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
	offset := (page - 1) * 15
	entries, err := TopUsers(guildID, offset, 15)
	if err != nil {
		return nil, err
	}

	detailed, err := DetailedLeaderboardEntries(guildID, entries)
	if err != nil {
		return nil, err
	}

	if len(entries) < 1 && p != nil && p.LastResponse != nil { //Dont send No Results error on first execution
		return nil, paginatedmessages.ErrNoResults
	}

	embed := &discordgo.MessageEmbed{
		Title: "Reputation leaderboard",
	}

	leaderboardURL := web.BaseURL() + "/public/" + discordgo.StrID(guildID) + "/reputation/leaderboard"
	out := "```\n# -- Points -- User\n"
	for _, v := range detailed {
		user := v.Username
		if user == "" {
			user = "unknown ID:" + strconv.FormatInt(v.UserID, 10)
		}
		out += fmt.Sprintf("#%02d: %6d - %s\n", v.Rank, v.Points, user)
	}
	out += "```\n" + "Full leaderboard: <" + leaderboardURL + ">"

	embed.Description = out

	return embed, nil
}

func repLogCreator(logEntries []*RepLogEntry, guildID, targetID int64, limiter, page int, paginatedView bool) *strings.Builder {
	offset := page * limiter
	scope := offset
	if !paginatedView {
		scope = 0
	}
	// grab the up to date info on as many users as we can
	membersToGrab := make([]int64, 1, len(logEntries))
	membersToGrab[0] = targetID

OUTER:
	for i, entry := range logEntries[scope:] {
		if i < limiter {
			for _, v := range membersToGrab {
				if entry.ReceiverID == targetID {
					if v == entry.SenderID {
						continue OUTER
					}
				} else {
					if v == entry.ReceiverID {
						continue OUTER
					}
				}
			}

			if entry.ReceiverID == targetID {
				membersToGrab = append(membersToGrab, entry.SenderID)
			} else {
				membersToGrab = append(membersToGrab, entry.ReceiverID)
			}
		}
	}

	members, _ := bot.GetMembers(guildID, membersToGrab...)

	// finally display the results
	var out strings.Builder
	out.WriteString("```\n")

	for i, entry := range logEntries[scope:] {
		if i < limiter {
			receiver := entry.ReceiverUsername
			sender := entry.SenderUsername

			for _, v := range members {
				if v.User.ID == entry.ReceiverID {
					receiver = v.User.String()
				}
				if v.User.ID == entry.SenderID {
					sender = v.User.String()
				}
			}

			if receiver == "" {
				receiver = discordgo.StrID(entry.ReceiverID)
			}

			if sender == "" {
				sender = discordgo.StrID(entry.SenderID)
			}

			f := "#%2d: %-15s: %s gave %s: %d points"
			if entry.SetFixedAmount {
				f = "#%2d: %-15s: %s set %s points to: %d"
			}
			out.WriteString(fmt.Sprintf(f, i+offset+1, entry.CreatedAt.UTC().Format("02 Jan 06 15:04"), sender, receiver, entry.Amount))
			out.WriteRune('\n')
		}
	}

	out.WriteString("```")

	return &out
}

func CmdGiveRep(parsed *dcmd.Data) (interface{}, error) {
	target := parsed.Args[0].Value.(*discordgo.User)

	conf, err := GetConfig(parsed.Context(), parsed.GuildData.GS.ID)
	if err != nil {
		return nil, err
	}

	if !conf.Enabled {
		return repDisabledError, nil
	}

	pointsName := conf.PointsName

	if target.ID == parsed.Author.ID {
		return fmt.Sprintf("You can't modify your own %s... **Silly**", pointsName), nil
	}

	sender := parsed.GuildData.MS
	receiver, err := bot.GetMember(parsed.GuildData.GS.ID, target.ID)
	if err != nil {
		return nil, err
	}

	amount := parsed.Args[1].Int()

	err = ModifyRep(parsed.Context(), conf, parsed.GuildData.GS.ID, sender, receiver, int64(amount))
	if err != nil {
		if cast, ok := err.(UserError); ok {
			return cast, nil
		}

		return nil, err
	}

	newScore, newRank, err := GetUserStats(parsed.GuildData.GS.ID, target.ID)
	if err != nil {
		newScore = -1
		newRank = -1
		return nil, err
	}

	actionStr := ""
	targetStr := "to"
	if amount > 0 {
		actionStr = "Gave"
	} else {
		actionStr = "Took away"
		amount = -amount
		targetStr = "from"
	}

	msg := fmt.Sprintf("%s `%d` %s %s **%s** (current: `#%d` - `%d`)", actionStr, amount, pointsName, targetStr, receiver.User.Username, newRank, newScore)
	return msg, nil
}

// Function that checks if the given user has ever received/gave rep in the given server
func userPresentInRepLog(userID int64, guildID int64, parsed *dcmd.Data) (found bool, err error) {
	logEntries, err := models.ReputationLogs(qm.Where("guild_id = ? AND (receiver_id = ? OR sender_id = ?)", guildID, userID, userID), qm.OrderBy("id desc"), qm.Limit(1)).AllG(parsed.Context())
	if err != nil {
		return false, err
	}

	if len(logEntries) < 1 {
		return false, nil
	}
	return true, nil
}

// Checks if the thanks detection is allowed to be run in the given channel
func isThanksDetectionAllowedInChannel(config *models.ReputationConfig, channelID int64) bool {
	if len(config.BlacklistedThanksChannels) > 0 {
		if common.ContainsInt64Slice(config.BlacklistedThanksChannels, channelID) {
			return false
		}
	}
	if len(config.WhitelistedThanksChannels) > 0 {
		return common.ContainsInt64Slice(config.WhitelistedThanksChannels, channelID)
	}
	return true
}
