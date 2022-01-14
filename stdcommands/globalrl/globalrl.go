package globalrl

import (
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	Name:                 "globalrl",
	Description:          "Tests the global ratelimit functionality",
	RequiredArgs:         1,
	HideFromHelp:         true,
	HideFromCommandsPage: true,
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {

		rlEvt := &discordgo.RateLimit{
			URL: "Wew",
			TooManyRequests: &discordgo.TooManyRequests{
				Bucket:     "wewsss",
				Message:    "Too many!",
				RetryAfter: 5,
			},
		}

		go common.BotSession.HandleEvent("__RATE_LIMIT__", rlEvt)

		return "Done", nil
	}),
}
