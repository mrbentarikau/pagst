package bashquotes

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

func ShouldRegister() bool {
	dialHost := "bash.org:http"
	timeout := 1 * time.Second
	conn, err := net.DialTimeout("tcp", dialHost, timeout)
	if conn != nil {
		defer conn.Close()
	}
	return err == nil
}

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

		body, err := util.RequestFromAPI(query)
		if err != nil {
			return "", err
		}

		toReader := bytes.NewReader(body)
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

		r, err := regexp.Compile(`\((-?\d+)\)`)
		if err != nil {
			return "", err
		}
		votesArray := (r.FindStringSubmatch(votes))
		if len(votesArray) > 0 {
			votes = votesArray[1]
		}

		response = fmt.Sprintf("```\n%s```<%s%s> | %s votes", quote, bashHost, number, votes)

		return response, nil
	},
}
