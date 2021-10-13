package bashquotes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/jonas747/dcmd/v4"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "BashQuote",
	Aliases:     []string{"bash", "bquote", "bq"},
	Description: `Returns a quote (NSFW probability high) from Bash Quotes Database > 
				http://bash.org`,
	Arguments: []*dcmd.ArgDef{
		{Name: "Quote-number", Type: dcmd.Int},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var response, quote, number, votes string
		var bashHost = "http://bash.org/"
		var bashRandom = "?random"

		query := bashHost + bashRandom
		if data.Args[0].Value != nil {
			query = fmt.Sprintf("%s?%d", bashHost, data.Args[0].Int64())
		}

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
			return "", commands.NewPublicError("HTTP Response was not 200, but ", resp.StatusCode)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		toReader := strings.NewReader(string(bytes))
		parseData, err := goquery.NewDocumentFromReader(toReader)
		if err != nil {
			return nil, err
		}

		if !parseData.Find("p").HasClass("quote") {
			return "No quote with such number found", nil
		}

		sel := parseData.Find(".quote").First()
		number = sel.Find("a").AttrOr("href", "")
		votes = sel.Text()
		quote = sel.Next().Text()
		if len(quote) > 1925 {
			quote = quote[0:1925] + "...\n\n(quote too long)"
		}

		r, err := regexp.Compile(`\((\d+)\)`)
		if err != nil {
			return "", err
		}
		votes = (r.FindStringSubmatch(votes))[1]

		response = fmt.Sprintf("```\n%s```<%s%s> | %s votes", quote, bashHost, number, votes)

		return response, nil
	},
}
