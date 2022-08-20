package inspire

import (
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
	Name:                "Inspire",
	Aliases:             []string{"insp"},
	Description:         "Shows 'inspirational' quotes from InspiroBot API...",
	RunInDM:             true,
	DefaultEnabled:      true,
	SlashCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		inspireURL := "https://" + common.ConfHost.GetString() + "/static/img/quacknotfound.png"
		descr := "Here's an AI generated inspirational quote:"

		// remove escape sequences
		inspired, err := inspireFromAPI()
		if err != nil {
			descr = fmt.Sprintf("%s\nInspireAPI wonky... ducks are sad : /", err)
		} else {
			inspireURL = inspired
		}

		embed := &discordgo.MessageEmbed{
			Description: descr,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: inspireURL,
			},
		}

		return embed, nil
	},
}

func inspireFromAPI() (string, error) {
	query := "https://inspirobot.me/api?generate=true"
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

	//inspireReturn := vtclean.Clean(string(body), false)
	inspireReturn := string(body)

	return inspireReturn, nil
}
