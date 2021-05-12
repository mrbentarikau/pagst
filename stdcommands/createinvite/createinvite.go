package createinvite

import (
	"github.com/jonas747/dcmd/v2"
	"github.com/jonas747/discordgo"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "createinvite",
	Description:          "Maintenance command, creates a invite for the specified server",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.BigInt},
	},
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		channels, err := common.BotSession.GuildChannels(data.Args[0].Int64())
		if err != nil {
			return nil, err
		}

		channelID := int64(0)
		for _, v := range channels {
			if v.Type == discordgo.ChannelTypeGuildText {
				channelID = v.ID
				break
			}
		}

		if channelID == 0 {
			return "No possible channel :(", nil
		}

		invite, err := common.BotSession.ChannelInviteCreate(channelID, discordgo.Invite{
			MaxAge:    120,
			MaxUses:   1,
			Temporary: true,
			Unique:    true,
		})

		if err != nil {
			return nil, err
		}

		bot.SendDM(data.Author.ID, "discord.gg/"+invite.Code)
		return "Sent invite expiring in 120 seconds and with 1 use in DM", nil
	}),
}
