package catfact

import (
	"encoding/json"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

type CatStuff struct {
	Fact string `json:"fact"`
	File string `json:"file"`
}

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "CatFact",
	Aliases:                   []string{"cf", "cat", "catfacts"},
	Description:               "Cat Facts from local database or catfact.ninja API",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw output"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var catStuff *CatStuff
		var cf, cPicLink string

		queryFact := "https://catfact.ninja/fact"
		queryPic := "https://aws.random.cat/meow"

		cf = Catfacts[rand.Intn(len(Catfacts))]

		request := rand.Intn(2)
		if request > 0 {
			catStuffBody, err := util.RequestFromAPI(queryFact)
			if err != nil {
				return cf, nil
			}

			queryErr := json.Unmarshal([]byte(catStuffBody), &catStuff)
			if queryErr == nil {
				cf = catStuff.Fact
			}
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			return cf, nil
		}

		catStuffBody, err := util.RequestFromAPI(queryPic)
		if err != nil {
			return cf, nil
		}

		queryErr := json.Unmarshal([]byte(catStuffBody), &catStuff)
		if queryErr == nil {
			cPicLink = catStuff.File
		} else {
			return cf, nil
		}

		embed := &discordgo.MessageEmbed{
			Description: cf,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: cPicLink,
			},
		}

		return embed, nil
	},
}
