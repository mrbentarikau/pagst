package xkcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

type Xkcd struct {
	Month      string
	Num        int64
	Link       string
	Year       string
	News       string
	SafeTitle  string
	Transcript string
	Alt        string
	Img        string
	Title      string
	Day        string
}

var XkcdHost = "https://xkcd.com/"
var XkcdJson = "info.0.json"
var XkcdExplainedHost = "https://www.explainxkcd.com/wiki/index.php/"

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "Xkcd",
	Description: "An xkcd comic, by default returns random comic strip",
	Arguments: []*dcmd.ArgDef{
		{Name: "Comic-number", Type: dcmd.Int},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "l", Help: "Latest comic"},
	},
	ApplicationCommandEnabled: true,
	DefaultEnabled:            true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {

		//first query to get latest number
		latest := false
		xkcd, err := getComic()
		if err != nil {
			return "Something happened whilst getting the comic!", err
		}

		xkcdNum := rand.Int63n(xkcd.Num) + 1

		//latest comic strip flag, already got that data
		if data.Switches["l"].Value != nil && data.Switches["l"].Value.(bool) {
			latest = true
		}

		//specific comic strip number
		if data.Args[0].Value != nil {
			n := data.Args[0].Int64()
			if n >= 1 && n <= xkcd.Num {
				xkcdNum = n
				if xkcdNum == 404 {
					return "A treasured and carefully-guarded point in the space of four-character strings not found.", nil
				}
			} else {
				return fmt.Sprintf("There's no comic numbered %d, current range is 1-%d", n, xkcd.Num), nil
			}
		}

		//if no latest flag is set, fetches a comic by number
		if !latest {
			xkcd, err = getComic(xkcdNum)
			if err != nil {
				return "Something happened whilst getting the comic!", err
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("#%d: %s", xkcd.Num, xkcd.Title),
			URL:         fmt.Sprintf("%s%d/", XkcdHost, xkcd.Num),
			Description: fmt.Sprintf("%s\n\n[#%[2]d explained](%s%[2]d)", xkcd.Alt, xkcd.Num, XkcdExplainedHost),
			Color:       int(rand.Int63n(16777215)),
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
		}
		return embed, nil
	},
}

func getComic(number ...int64) (*Xkcd, error) {
	xkcd := Xkcd{}
	queryURL := XkcdHost + XkcdJson

	if len(number) >= 1 {
		queryURL = fmt.Sprintf(XkcdHost+"%d/"+XkcdJson, number[0])
	}

	responseBytes, err := util.RequestFromAPI(queryURL)
	if err != nil {
		return nil, err
	}

	readerToDecoder := bytes.NewReader(responseBytes)
	queryErr := json.NewDecoder(readerToDecoder).Decode(&xkcd)
	if queryErr != nil {
		return nil, queryErr
	}

	return &xkcd, nil
}
