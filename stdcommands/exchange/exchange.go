package exchange

//go:generate go run gen/symbols_gen.go -o exchange_symbols.go

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Cooldown:            5,
	Name:                "exchange",
	Aliases:             []string{"exch", "money"},
	Description:         "💱Shows Currency Exchange Rates",
	RunInDM:             true,
	DefaultEnabled:      true,
	SlashCommandEnabled: true,
	RequiredArgs:        3,
	Arguments: []*dcmd.ArgDef{
		{Name: "Amount", Type: dcmd.Float},
		{Name: "From", Type: dcmd.String}, {Name: "To", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		amount := fmt.Sprintf("%.2f", data.Args[0].Float64())

		from := Symbols[strings.ToUpper(data.Args[1].Str())]
		to := Symbols[strings.ToUpper(data.Args[2].Str())]

		if len(to) == 0 || len(from) == 0 {
			return "Invalid currency code.\nCheck out available codes on: <https://api.exchangerate.host/symbols>", nil
		}
		output, err := requestAPI("https://api.exchangerate.host/convert?from=" + from["Code"] + "&to=" + to["Code"] + "&amount=1")
		if err != nil {
			return nil, err
		}

		embed := &discordgo.MessageEmbed{
			Title:       "💱Currency Exhange Rate",
			Description: fmt.Sprintf("\n%s **%s** (%s) is %0.3f **%s** (%s).", amount, from["Description"], output.Query.From, data.Args[0].Float64()*output.Result, to["Description"], output.Query.To),
			Color:       int(rand.Int63n(0xffffff)),
			Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Based on currency rate 1:%0.3f", output.Result)},
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}

		return embed, nil
	},
}

func requestAPI(query string) (*Result, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("HTTP err: ", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &Result{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SymbolInfo struct {
	Description string `json:"description,omitempty"`
	Code        string `json:"code,omitempty"`
}

type Result struct {
	Motd *struct {
		Msg string `json:"msg"`
		URL string `json:"url"`
	} `json:"motd"`
	Success bool `json:"success"`
	Query   *struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	} `json:"query,omitempty"`
	Info *struct {
		Rate float64 `json:"rate"`
	} `json:"info,omitempty"`
	Historical bool                   `json:"historical,omitempty"`
	Date       string                 `json:"date,omitempty"`
	Result     float64                `json:"result,omitempty"`
	Symbols    map[string]*SymbolInfo `json:"symbols,omitempty"`
}
