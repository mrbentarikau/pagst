package dcallvoice

import (
	"fmt"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "dcallvoice",
	Description:          "Disconnects from all the voice channels the bot is in. Bot Admin Only.",
	HideFromHelp:         true,
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {

		vcs := make([]*discordgo.VoiceState, 0)

		processShards := bot.ReadyTracker.GetProcessShards()
		for _, shard := range processShards {
			guilds := bot.State.GetShardGuilds(int64(shard))
			for _, g := range guilds {
				vc := g.GetVoiceState(common.BotUser.ID)
				if vc != nil {
					vcs = append(vcs, vc)
					go bot.ShardManager.SessionForGuild(g.ID).GatewayManager.ChannelVoiceLeave(g.ID)
				}
			}
		}

		return fmt.Sprintf("Leaving %d voice channels...", len(vcs)), nil
	}),
}
