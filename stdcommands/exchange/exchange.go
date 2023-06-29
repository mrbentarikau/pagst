package exchange

//go:generate go run gen/symbols_gen.go -o exchange_symbols.go

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Cooldown:                  5,
	Name:                      "Forex",
	Aliases:                   []string{"fx", "exchange", "money"},
	Description:               "💱Shows Currency Exchange Rate and converts to target currency...",
	RunInDM:                   true,
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	RequiredArgs:              3,
	Arguments: []*dcmd.ArgDef{
		{Name: "Amount", Type: &dcmd.FloatArg{Min: 1, Max: 1000000000000000}},
		{Name: "From", Type: dcmd.String}, {Name: "To", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		amount := data.Args[0].Float64()
		from := Symbols[strings.ToUpper(data.Args[1].Str())]
		to := Symbols[strings.ToUpper(data.Args[2].Str())]

		if len(to) == 0 || len(from) == 0 {
			pm, err := paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, int(math.Ceil(float64(len(Symbols))/float64(15))), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
					return errEmbed(Symbols, page)
				})
			if err != nil {
				return nil, err
			}
			return pm, nil
		}

		query := "https://api.exchangerate.host/convert?from=" + from["Code"] + "&to=" + to["Code"] + "&amount=" + fmt.Sprintf("%.3f", amount)
		body, err := util.RequestFromAPI(query)
		if err != nil {
			return nil, err
		}

		output := &Result{}
		err = json.Unmarshal([]byte(body), &output)
		if err != nil {
			return nil, err
		}

		printer := message.NewPrinter(language.English)
		exchangeFrom := printer.Sprintf("%.2f", amount)
		exchangeTo := printer.Sprintf("%.3f", output.Result)

		embed := &discordgo.MessageEmbed{
			Title:       "💱Currency Exchange Rate",
			Description: fmt.Sprintf("\n%s **%s** (%s) is %s **%s** (%s).", exchangeFrom, from["Description"], output.Query.From, exchangeTo, to["Description"], output.Query.To),
			Color:       int(rand.Int63n(0xffffff)),
			Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Based on currency rate 1 : %f", output.Info.Rate)},
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}

		return embed, nil
	},
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

func errEmbed(check map[string]map[string]string, page int) (*discordgo.MessageEmbed, error) {
	desc := "CODE | Description\n------------------\n"
	codes := make([]string, 0, len(Symbols))
	for k := range Symbols {
		codes = append(codes, k)
	}
	sort.Strings(codes)
	start := (page * 15) - 15
	end := page * 15
	for i, c := range codes {
		if i < end && i >= start {
			desc += fmt.Sprintf("%4s | %s\n", c, Symbols[c]["Description"])
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       "Invalid currency code",
		URL:         "https://api.exchangerate.host/symbols",
		Color:       0xAE27FF,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Description: fmt.Sprintf("Check out available codes on: https://api.exchangerate.host/symbols ```\n%s```", desc),
	}
	return embed, nil
}
