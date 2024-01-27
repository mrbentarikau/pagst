package trivia

import (
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
)

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p, &commands.YAGCommand{
		CmdCategory:               commands.CategoryFun,
		Cooldown:                  10,
		Name:                      "Trivia",
		Description:               "Asks a random question, you have got 30 seconds to answer!",
		RunInDM:                   false,
		ApplicationCommandEnabled: true,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			err := manager.NewTrivia(parsed.GuildData.GS.ID, parsed.ChannelID)
			if err != nil {
				if err == ErrSessionInChannel {
					logger.WithError(err).Error("Failed to create new trivia")
					return "There's already a trivia session in this channel", nil
				}
				return "Failed Running Trivia, unknown error", err
			}
			return nil, nil
		},
	})
}
