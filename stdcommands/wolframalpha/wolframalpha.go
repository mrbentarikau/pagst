package wolframalpha

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/jonas747/dcmd/v4"
	"github.com/jonas747/discordgo/v2"
	"github.com/mediocregopher/radix/v3"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "WolframAlpha",
	Aliases:     []string{"wolfram", "wa"},
	Description: `Queries the API of WolframAlpha for results on ...anything!

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
	var result string
	appID := wolframID

	queryURL := baseURL + "?appid=" + appID + "&input=" + input + "&format=plaintext"
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
		return "Wolfram has no good answer for this query", nil
	}

	/*if waQuery.Queryresult.Pod[0].Title == "Input interpretation" {
		result = waQuery.Queryresult.Pod[0].Subpod[0].Plaintext
	}

	result += waQuery.Queryresult.Pod[1].Subpod[0].Plaintext
	if result == "" {
		result = waQuery.Queryresult.Pod[0].Subpod[0].Plaintext
	}*/
	//if len(waQuery.Queryresult.Pod) > 2 {
	for k, v := range waQuery.Queryresult.Pod {
		if v.Subpod[0].Plaintext != "" && k <= 6 {
			result += "\n\n" + v.Title + ":\n"
			for _, vv := range v.Subpod {
				result += fmt.Sprintf("%s\n", vv.Plaintext)
			}
			//result += fmt.Sprintf("%s\n", v.Subpod[0].Plaintext)
		}

		/*if v.Title == "Decimal approximation" {
			result += "\n\nApproximation: " + v.Subpod[0].Plaintext
		}

		if v.Title == "Unit conversions" {
			result += "\n\nUnit conversions:\n"
			for _, vv := range v.Subpod {
				result += fmt.Sprintf("%s\n", vv.Plaintext)
			}
		}

		if v.Title == "Basic elemental properties" {
			result += "\n\nBasic elemental properties:\n"
			result += fmt.Sprintf("%s\n", v.Subpod[0].Plaintext)
		}*/
	}
	//}
	return result, nil
}
