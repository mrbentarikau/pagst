package thanks

import (
	"fmt"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/mrbentarikau/pagst/reputation"
)

var thanksLocalizations = &map[discordgo.Locale]string{
	"da":    "tak",
	"de":    "danke",
	"en-GB": "thanks",
	"en-US": "thanks",
	"es-ES": "gracias",
	"fr":    "merci",
	"hr":    "hvala",
	"it":    "grazie",
	"lt":    "dėkoju",
	"hu":    "kösz",
	"nl":    "dank",
	"no":    "takk",
	"pl":    "dzięki",
	"pt-BR": "obrigado",
	"ro":    "mulțumesc",
	"fi":    "kiitos",
	"sv-SE": "tack",
	"vi":    "cảm ơn",
	"tr":    "şükür",
	"cs":    "dik",
	"el":    "ευχαριστώ",
	"bg":    "благодаря",
	"ru":    "спасибо",
	"uk":    "спасибі",
	"hi":    "धन्यवाद",
	"th":    "ขอบใจ",
	"zh-CN": "谢谢",
	"ja":    "どうも",
	"zh-TW": "謝謝",
	"ko":    "감사해요",
}

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryFun,
	Name:                 "Thanks",
	CmdNoRun:             true,
	RunInDM:              false,
	HideFromHelp:         true,
	HideFromCommandsPage: false,
	DefaultEnabled:       false,

	ApplicationCommandEnabled: true,
	ApplicationCommandType:    2,
	NameLocalizations:         thanksLocalizations,

	Arguments: []*dcmd.ArgDef{
		{Name: "User", Type: &commands.MemberArg{}},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		sender := data.GuildData.MS
		target := data.Args[0].Value.(*dstate.MemberState)
		var response string
		var repDisabledError = "**Rep command is disabled on this server. Enable it from the control panel.**"

		repConf, err := reputation.GetConfig(data.Context(), data.GuildData.GS.ID)
		if err != nil {
			return nil, err
		}

		if !repConf.Enabled {
			return repDisabledError, nil
		}

		if sender.User.ID == target.User.ID {
			response = fmt.Sprintf("You can't modify your own %s... **Silly**", repConf.PointsName)
			return &commands.EphemeralOrNone{Content: response}, nil
		}

		if err = reputation.CanModifyRep(repConf, sender, target); err != nil {
			response = "You don't have any of the required roles to give points"
			return &commands.EphemeralOrNone{Content: response}, nil
		}

		err = reputation.ModifyRep(data.Context(), repConf, data.GuildData.GS.ID, sender, target, 1)
		if err != nil {
			if err == reputation.ErrCooldown {
				// Ignore this error silently
				return nil, nil
			}
		}

		content := fmt.Sprintf("Gave +1 %s to **%s**", repConf.PointsName, target.User.Mention())
		return content, nil
	},
}
