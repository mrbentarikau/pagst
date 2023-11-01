package moderation

import (
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

func (p *Plugin) createDeleteBotDMInteraction() *commands.YAGCommand {
	return &commands.YAGCommand{
		CmdCategory:          commands.CategoryGeneral,
		Name:                 "Delete Bot DMs",
		Cooldown:             15,
		CmdNoRun:             true,
		RunInDM:              true,
		HideFromHelp:         true,
		IsResponseEphemeral:  true,
		HideFromCommandsPage: true,
		DefaultEnabled:       false,

		ApplicationCommandEnabled: true,
		ApplicationCommandType:    3,

		Arguments: []*dcmd.ArgDef{
			{Name: "Message", Type: dcmd.String},
		},

		RunFunc: func(data *dcmd.Data) (interface{}, error) {
			i := data.SlashCommandTriggerData.Interaction

			if i.Type != discordgo.InteractionApplicationCommand {
				return "Something went wrong!", nil
			}
			if i.GuildID != 0 {
				return "This can only be used in DMs!", nil
			}

			channel, err := common.BotSession.UserChannelCreate(i.User.ID)
			if err != nil {
				return nil, err
			}

			targetMessageId := i.ApplicationCommandData().TargetID
			targetMessage, err := common.BotSession.ChannelMessage(channel.ID, targetMessageId)
			if err != nil {
				return nil, err
			}
			if targetMessage.Author.ID != common.BotApplication.ID {
				return "You can only use this on " + common.ConfBotName.GetString() + " messages!", nil
			}

			if err := common.BotSession.ChannelMessageDelete(channel.ID, targetMessageId); err != nil {
				return nil, err
			}

			return "Successfully deleted message!", nil
		},
	}
}
