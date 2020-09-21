package howlongtobeat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/commands"
)

type xkcd struct {
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

var xkcdHost = "https://xkcd.com/"
var xkcdJSON = "info.0.json"

//Command is exported to
var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:			"HowLongToBeat",
	Aliases:		[]string{"hltb"},
	RequiredArgs: 1,
	Description: "An xkcd comic, by default returns random comic strip",
	Arguments: []*dcmd.ArgDef{
		&dcmd.ArgDef{Name: "Comic number", Type: dcmd.Int},
	},
	ArgSwitches: []*dcmd.ArgDef{
		&dcmd.ArgDef{Switch: "noembed", Name: "Latest comic"},
	},
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
			Description: fmt.Sprintf("[%s](%s%d/)", xkcd.Alt, xkcdHost, xkcd.Num),
			Color:       int(rand.Int63n(16777215)),
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
		}
		return embed, nil
	},
}

func getComic(number ...int64) (*xkcd, error) {
	xkcd := xkcd{}
	queryURL := xkcdHost + xkcdJSON

	if len(number) >= 1 {
		queryURL = fmt.Sprintf(xkcdHost+"%d/"+xkcdJSON, number[0])
	}

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "curl/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryErr := json.Unmarshal(body, &xkcd)
	if queryErr != nil {
		return nil, queryErr
	}

	return &xkcd, nil
}
