package catfact

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
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
		var cf string
		//var cf, cPicLink string

		queryFact := "https://catfact.ninja/fact"
		// queryPic := "https://aws.random.cat/meow"

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

		catsOnCocaine, err := catScarper()
		if err != nil {
			return cf, nil
		}

		//return cf, nil
		/*
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
		*/

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

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Something wonky happened getting results, random.cat is probably down", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("HTTP Response was not 200, but ", resp.StatusCode)
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	toReader := strings.NewReader(string(bytes))
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
