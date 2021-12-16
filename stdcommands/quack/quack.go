//Random duck image from random-d.uk API
package quack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/jonas747/dcmd/v4"
	"github.com/jonas747/discordgo/v2"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "duck",
	Aliases:             []string{"quack"},
	Description:         "Random duck images from random-d.uk API",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var descr, quackURL string
		quackURL = "https://" + common.ConfHost.GetString() + "/static/img/quacknotfound.png"

		quack, err := duckFromAPI()
		if err != nil {
			descr = fmt.Sprintf("%s\nQuackAPI wonky... ducks are sad : /", err)
		} else {
			quackURL = quack
		}

		embed := &discordgo.MessageEmbed{
			Description: descr,
			Color:       int(rand.Int63n(16777215)),
			Image: &discordgo.MessageEmbedImage{
				URL: quackURL,
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
		return "", commands.NewPublicError("HTTP err: ", resp.StatusCode)
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
