package exchange

//go:generate go run gen/currency_codes_gen.go -o currency_codes.go

import (
	"bytes"
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

const currenciesAPIURL = "https://api.frankfurter.app/currencies"
const currencyPerPage = 15

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
		{Name: "Amount", Type: &dcmd.FloatArg{Min: 1, Max: 1000000000000000, PreventInfNaN: true}},
		{Name: "From", Type: dcmd.String}, {Name: "To", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		amount := data.Args[0].Float64()
		var exchangeRateResult ExchangeRate
		var err error

		from := strings.ToUpper(data.Args[1].Str())
		to := strings.ToUpper(data.Args[2].Str())

		// Check if the currencies exist in the map
		fromExtended, fromExist := Currencies[from]
		toExtended, toExist := Currencies[to]

		// Checks the max amount of pages by the number of symbols on each page
		maxPages := int(math.Ceil(float64(len(Currencies)) / float64(currencyPerPage)))

		// If the currency isn't supported by API.
		if !toExist || !fromExist {
			_, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, maxPages, func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
					embed, err := errEmbed(Currencies, page)
					if err != nil {
						return nil, err
					}
					return embed, nil
				})
			if err != nil {
				return nil, err
			}
			return nil, nil
		}

		query := fmt.Sprintf("https://api.frankfurter.app/latest?amount=%.3f&from=%s&to=%s", amount, from, to)
		responseBytes, err := util.RequestFromAPI(query)
		if err != nil {
			return nil, err
		}

		readerToDecoder := bytes.NewReader(responseBytes)
		err = json.NewDecoder(readerToDecoder).Decode(&exchangeRateResult)
		if err != nil {
			return nil, err
		}

		p := message.NewPrinter(language.English)
		embed := &discordgo.MessageEmbed{
			Title:       "💱Currency Exchange Rate",
			Description: p.Sprintf("\n%.2f **%s** (%s) is %.3f **%s** (%s).", amount, fromExtended, from, exchangeRateResult.Rates[to], toExtended, to),
			Color:       int(rand.Int63n(0xffffff)),
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		return embed, nil
	},
}

func errEmbed(currenciesResult map[string]string, page int) (*discordgo.MessageEmbed, error) {
	desc := "CODE | Description\n------------------\n"
	codes := make([]string, 0, len(currenciesResult))
	for k := range currenciesResult {
		codes = append(codes, k)
	}
	sort.Strings(codes)
	start := (page * currencyPerPage) - currencyPerPage
	end := page * currencyPerPage
	for i, c := range codes {
		if i < end && i >= start {
			desc += fmt.Sprintf("%4s | %s\n", c, currenciesResult[c])
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       "Invalid currency code",
		URL:         currenciesAPIURL,
		Color:       0xAE27FF,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Description: fmt.Sprintf("Check out available codes on: %s ```\n%s```", currenciesAPIURL, desc),
	}
	return embed, nil
}

type ExchangeRate struct {
	Amount float64
	Base   string
	Date   string
	Rates  map[string]float64
}
