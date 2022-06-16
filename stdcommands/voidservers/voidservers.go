package voidservers

import (
	"fmt"

	"github.com/mrbentarikau/pagst/bot/models"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var Command = &commands.YAGCommand{
	Cooldown:    5,
	CmdCategory: commands.CategoryDebug,
	Name:        "VoidServers",
	Description: "Shows void guilds...",
	Arguments: []*dcmd.ArgDef{
		{Name: "Skip", Help: "Entries to skip", Type: dcmd.Int, Default: 0},
	},

	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		skip := data.Args[0].Int()

		results, err := models.JoinedGuilds(qm.Where("left_at is null"), qm.OrderBy("member_count desc"), qm.Limit(50), qm.Offset(skip)).AllG(data.Context())
		if err != nil {
			return nil, err
		}

		out := "```"
		for k, v := range results {
			notVoid, _ := common.BotSession.Guild(v.ID)
			if notVoid == nil {
				out += fmt.Sprintf("\n#%-2d: %-18d %-25s (%d members)", k+skip+1, v.ID, v.Name, v.MemberCount)
			}
		}
		if len(out) == 3 {
			out += "0, void, peace on Earth :)"
		}
		return "Void guilds still 'with' this bot:\n" + out + "\n```", nil
	}),
}
