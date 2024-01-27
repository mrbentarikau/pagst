// Random duck image from random-d.uk API
package quack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "Duck",
	Aliases:                   []string{"quack"},
	Description:               "Random duck images from random-d.uk API",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var quack struct {
			URL string `json:"url"`
		}
		var descr, quackURL string

		quackURL = "https://" + common.ConfHost.GetString() + "/static/img/quacknotfound.png"
		query := "https://random-d.uk/api/quack"

		responseBytes, err := util.RequestFromAPI(query)
		if err != nil {
			return "", err
		}

		readerToDecoder := bytes.NewReader(responseBytes)
		err = json.NewDecoder(readerToDecoder).Decode(&quack)
		if err != nil {
			descr = fmt.Sprintf("%s\nQuackAPI wonky... ducks are sad : /", err)
		} else {
			quackURL = quack.URL
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
