package discordlogger

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/bot/models"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
)

var (
	// Send bot leaves joins to this discord channel
	confBotLeavesJoins = config.RegisterOption("yagpdb.botleavesjoins", "Channel to log added/left servers to", 0)

	logger = common.GetPluginLogger(&Plugin{})
)

func Register() {
	if confBotLeavesJoins.GetInt() != 0 {
		logger.Info("Listening for bot leaves and join")
		common.RegisterPlugin(&Plugin{})
	}
}

var _ bot.BotInitHandler = (*Plugin)(nil)

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Discord Logger",
		SysName:  "discord_logger",
		Category: common.PluginCategoryCore,
	}
}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLast(p, EventHandler, eventsystem.EventNewGuild, eventsystem.EventGuildDelete)
}

func EventHandler(evt *eventsystem.EventData) (retry bool, err error) {
	count, err := common.GetJoinedServerCount()
	if err != nil {
		return false, errors.WithStackIf(err)
	}

	msg := ""
	switch evt.Type {
	case eventsystem.EventGuildDelete:
		if evt.GuildDelete().Unavailable {
			// Just a guild outage
			return
		}

		guildData, err := models.FindJoinedGuildG(evt.Context(), evt.GuildDelete().ID)
		if err != nil {
			guildData = &models.JoinedGuild{
				ID:   evt.GuildDelete().ID,
				Name: "unknown",
			}

			logger.WithError(err).Error("failed fetching guild data")
		}

		count = count - 1
		msg = fmt.Sprintf(":x: Left guild **%s** :(", guildData.Name)

		leftGuildOwnerName := "Unknown"
		guildOwnerName, err := common.BotSession.User(guildData.OwnerID)
		if err == nil {
			leftGuildOwnerName = guildOwnerName.Username
		}

		msg += fmt.Sprintf(" owned by %s ID: %d", leftGuildOwnerName, guildData.OwnerID)
		//From master code
		//msg = fmt.Sprintf(":x: Left guild **%s** :(", common.ReplaceServerInvites(guildData.Name, 0, "[removed-server-invite]"))
	case eventsystem.EventNewGuild:
		guild := evt.GuildCreate().Guild
		msg = fmt.Sprintf(":white_check_mark: Joined guild **%s** :D", common.ReplaceServerInvites(guild.Name, 0, "[removed-server-invite]"))
		msg += fmt.Sprintf(" owned by %s ID: %d", (bot.State.GetMember(guild.ID, guild.OwnerID)).User.Username, guild.OwnerID)
	}
	msg += fmt.Sprintf(" (now connected to %d servers)", count)
	_, err = common.BotSession.ChannelMessageSend(int64(confBotLeavesJoins.GetInt()), msg)
	return bot.CheckDiscordErrRetry(err), errors.WithStackIf(err)
}
