// Random duck image from random-d.uk API
package quack

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "Duck",
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

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("HTTP err: ", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	queryErr := json.Unmarshal([]byte(body), &quack)
	if queryErr != nil {
		return "", err
	}

	return quack.URL, nil
}
