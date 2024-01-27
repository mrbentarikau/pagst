package dogfact

import (
	"bytes"
	"encoding/json"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

type DogStuff struct {
	Facts   []string `json:"facts"`
	Message string   `json:"message"`
}

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "DogFact",
	Aliases:                   []string{"dog", "dogfacts"},
	Description:               "Dog Facts from local database or dog-api.kinduff.com API",
	ApplicationCommandEnabled: true,
	DefaultEnabled:            true,
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw output"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var dogStuff *DogStuff
		var df, dPicLink string
		var err error

		queryFact := "https://dog-api.kinduff.com/api/facts"
		queryPic := "https://dog.ceo/api/breeds/image/random"

		df = dogfacts[rand.Intn(len(dogfacts))]

		request := rand.Intn(2)
		if request > 0 {
			responseBytes, err := util.RequestFromAPI(queryFact)
			if err != nil {
				return df, nil
			}

			readerToDecoder := bytes.NewReader(responseBytes)
			bodyErr := json.NewDecoder(readerToDecoder).Decode(&dogStuff)
			if bodyErr == nil && len(dogStuff.Facts) > 0 {
				df = dogStuff.Facts[0]
			}
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			return df, nil
		}

		responseBytes, err := util.RequestFromAPI(queryPic)
		if err != nil {
			return df, nil
		}

		readerToDecoder := bytes.NewReader(responseBytes)
		queryErr := json.NewDecoder(readerToDecoder).Decode(&dogStuff)
		if queryErr == nil {
			dPicLink = dogStuff.Message
		} else {
			return df, nil
		}

		embed := &discordgo.MessageEmbed{
			Description: df,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: dPicLink,
			},
		}

		return embed, nil
	},
}
