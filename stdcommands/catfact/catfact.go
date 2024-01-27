package catfact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"

	"github.com/PuerkitoBio/goquery"
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

		queryFact := "https://catfact.ninja/fact"
		cf := Catfacts[rand.Intn(len(Catfacts))]

		request := rand.Intn(2)
		if request > 0 {
			responseBytes, err := util.RequestFromAPI(queryFact)
			if err != nil {
				return cf, nil
			}

			readerToDecoder := bytes.NewReader(responseBytes)
			queryErr := json.NewDecoder(readerToDecoder).Decode(&catStuff)
			if queryErr == nil {
				cf = catStuff.Fact
			}
		}

		catsOnCocaine, err := catScarper()
		if err != nil {
			return cf, nil
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			return cf, nil
		}

		embed := &discordgo.MessageEmbed{
			Description: cf,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: catsOnCocaine,
			},
		}

		return embed, nil

	},
}

func catScarper() (string, error) {
	query := fmt.Sprintf("http://random.cat/view/%d", rand.Intn(1677))

	body, err := util.RequestFromAPI(query)
	if err != nil {
		return "", err
	}

	toReader := bytes.NewReader(body)
	parseData, err := goquery.NewDocumentFromReader(toReader)
	if err != nil {
		return "", err
	}

	if parseData.Find("img#cat") == nil {
		return "", fmt.Errorf("no cat on cocaine found")
	}

	cawt := parseData.Find("img#cat").AttrOr("src", "")
	return cawt, nil
}
