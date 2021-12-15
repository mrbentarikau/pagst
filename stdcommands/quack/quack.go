//Random duck image from random-d.uk API
package quack

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/jonas747/dcmd/v4"
	"github.com/jonas747/discordgo/v2"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "duck",
	Aliases:             []string{"quack"},
	Description:         "Random duck images from random-d.uk API",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		quacknotfound := "https://" + common.ConfHost.GetString() + "/static/img/quacknotfound.png"

		quack, err := duckFromAPI()
		if err != nil {
			embed := &discordgo.MessageEmbed{
				Description: "QuackAPI wonky... ducks are sad : /",
				Color:       int(rand.Int63n(16777215)),
				Image: &discordgo.MessageEmbedImage{
					URL: quacknotfound,
				},
			}
			return embed, nil
		}

		embed := &discordgo.MessageEmbed{
			Color: int(rand.Int63n(16777215)),
			Image: &discordgo.MessageEmbedImage{
				URL: quack,
			},
		}
		return embed, nil
	},
}

func duckFromAPI() (string, error) {
	var quack struct {
		URL string `json:"url"`
	}

	query := "https://random-d.uk/api/quack"
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("Not 200!")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	queryErr := json.Unmarshal([]byte(body), &quack)
	if queryErr != nil {
		return "", err
	}

	return quack.URL, nil
}
