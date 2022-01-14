package wolframalpha

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mediocregopher/radix/v3"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "WolframAlpha",
	Aliases:     []string{"wolfram", "wa"},
	Description: `Queries the API of WolframAlpha for results on ...anything!
					Results are given in metric system, link below would use user's local unit-system.

					Needs user created AppID for WolframAlpha.
					To setup a WolframAlpha appID, you must register a Wolfram ID and sign in to the Wolfram|Alpha Developer Portal > https://developer.wolframalpha.com/portal/
					Upon logging in, go to the *My Apps* tab to start creating your first app. 
					
					This free access gives for up to **2 000** non-commercial API calls per month.`,
	RequiredArgs: 1,
	Arguments: []*dcmd.ArgDef{
		{Name: "Expression", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "appid", Help: "Add your Wolfram|Alpha appID case sensitive"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var directURL = "https://www.wolframalpha.com/input/?i="

		if data.Switches["appid"].Value != nil && data.Switches["appid"].Value.(bool) {

			targetID := data.Author.ID
			target, _ := bot.GetMember(data.GuildData.GS.ID, targetID)

			if isAdmin, _ := data.GuildData.GS.GetMemberPermissions(data.GuildData.CS.ID, data.Author.ID, target.Member.Roles); isAdmin&discordgo.PermissionAdministrator != 0 {
				appID := data.Args[0].Str()
				if len(appID) < 8 || len(appID) > 25 {
					return "appID is too short or too long", nil
				}
				err := common.RedisPool.Do(radix.Cmd(nil, "SET", "wolfram_appID:"+strconv.FormatInt(data.GuildData.GS.ID, 10), appID))
				if err != nil {
					return "", err
				}
				return fmt.Sprintln("Wolfram|Alpha appID added"), nil
			} else {
				return "Only a Guild Admin can add appID", nil
			}
		}

		var appID string
		err := common.RedisPool.Do(radix.Cmd(&appID, "GET", "wolfram_appID:"+strconv.FormatInt(data.GuildData.GS.ID, 10)))
		if err != nil {
			return "No Wolfram|Alpha appID", nil
		}

		input := url.QueryEscape(data.Args[0].Str())
		response := "```\n"
		responseTooLong := "\n\n(response too long)"
		responseEnd := "\n```<" + directURL + input + ">"

		query, err := requestWolframAPI(input, appID)
		if err != nil {
			return "", err
		}

		if len(query) > 2000 {
			query = common.CutStringShort(query, 1980-len(responseTooLong+responseEnd)) + responseTooLong
		}

		response += query + responseEnd
		return response, nil
	},
}

func requestWolframAPI(input, wolframID string) (string, error) {
	var baseURL = "http://api.wolframalpha.com/v2/query"
	var waQuery WolframAlpha
	var result, subpodResult string
	appID := wolframID

	queryURL := baseURL + "?appid=" + appID + "&input=" + input + "&format=plaintext&unit=metric"
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return fmt.Sprintf("Wolfram is wonky: status code is not 200! But: %d", resp.StatusCode), nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	xml.Unmarshal(body, &waQuery.Queryresult)

	if waQuery.Queryresult.AttrError == "true" {
		result = fmt.Sprintln("Wolfram is wonky: ", waQuery.Queryresult.Error.Msg)
		return result, nil
	}

	if len(waQuery.Queryresult.Pod) == 0 {
		return "Wolfram has no good answer for this query...", nil
	}
	//Convert response to somewhat general Discord format (maybe separete func.)
	var re = regexp.MustCompile(`\x0a\x20\x7C`)

	for k, v := range waQuery.Queryresult.Pod {
		if v.Subpod[0].Plaintext != "" && k <= 6 {
			result += "\n" + v.Title + ":\n"
			subpodResult = ""
			for _, vv := range v.Subpod {
				subpodResult += fmt.Sprintf("%s\n", re.ReplaceAllString(vv.Plaintext, " |"))
			}
			if len(subpodResult) >= 512 {
				subpodResult = common.CutStringShort(subpodResult, 510) + "\n"
			}
			result += subpodResult
		}
	}

	return result, nil
}
