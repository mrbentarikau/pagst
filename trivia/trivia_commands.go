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
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "local", Help: "In-built Trivia"},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			var local bool
			if parsed.Switches["local"].Value != nil && parsed.Switches["local"].Value.(bool) {
				local = true
			}

			err := manager.NewTrivia(parsed.GuildData.GS.ID, parsed.ChannelID, local)
			if err != nil {
				if err == ErrSessionInChannel {
					logger.WithError(err).Error("Failed to create new trivia")
					return "There's already a trivia session in this channel", nil
				}
				return "Failed Running Trivia, error: " + err.Error(), err
			}
			return nil, nil
		},
	})
}
