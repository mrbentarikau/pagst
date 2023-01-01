package throw

import (
	"fmt"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "Throw",
	Description: "Throwing things is cool.",
	Arguments: []*dcmd.ArgDef{
		{Name: "Target", Type: dcmd.User},
	},
	DefaultEnabled: true,

	ApplicationCommandEnabled: true,
	ApplicationCommandType:    2,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		target := "a random person nearby"
		thrower := data.GuildData.MS.User
		if data.Args[0].Value != nil {
			if thrower.ID == data.Args[0].Value.(*discordgo.User).ID {
				target = "themself"
			} else {
				target = data.Args[0].Value.(*discordgo.User).Username
			}
		}

		var resp string

		rng := rand.Intn(100)
		if rng < 5 {
			resp = fmt.Sprintf("TRIPLE THROW!!! **%s** threw **%s**, **%s** and **%s** at **%s**", thrower.Username, randomThing(), randomThing(), randomThing(), target)
		} else if rng < 15 {
			resp = fmt.Sprintf("DOUBLE THROW!! **%s** threw **%s** and **%s** at **%s**", thrower.Username, randomThing(), randomThing(), target)
		} else {
			resp = fmt.Sprintf("**%s** threw **%s** at **%s**", thrower.Username, randomThing(), target)
		}

		return resp, nil
	},
}

func randomThing() string {
	var query = "https://roger.redevised.com/api/v1/"
	randNum := rand.Intn(3)

	if randNum > 1 {
		return common.RandomNoun()
	} else if randNum > 0 {
		if qrt, err := util.RequestFromAPI(query); err == nil {
			return string(qrt)
		}
	}

	return throwThings[rand.Intn(len(throwThings))]
}
