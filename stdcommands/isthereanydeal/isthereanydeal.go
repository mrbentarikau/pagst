package isthereanydeal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var (
	confItadAPIKey = config.RegisterOption("yagpdb.itad_api_key", "IsThereAnyDeal API key", "")
	itadAPIHost    = "https://api.isthereanydeal.com"
)

func ShouldRegister() bool {
	return confItadAPIKey.GetString() != ""
}

var Currency = map[string]string{
	"EUR": "\u20ac",
	"GBP": "\u00a3",
	"USD": "$",
	"CAD": "C$",
	"BRL": "R$",
	"AUD": "A$",
	"TRY": "\u20BA",
	"CNY": "\u00a5",
}

var Command = &commands.YAGCommand{
	CmdCategory:  commands.CategoryFun,
	Name:         "IsThereAnyDeal",
	Aliases:      []string{"itad", "anydeal", "gamedeal"},
	Description:  "Queries [IsThereAnyDeal](https://isthereanydeal.com) API for current prices of targeted game.",
	RequiredArgs: 1,
	Cooldown:     5,
	Arguments: []*dcmd.ArgDef{
		{Name: "Title", Help: "Search prices based on game's title", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "c", Help: "Compact output"},
		{Name: "p", Help: "Paginated output"},
		{Name: "ccode", Help: "Country code determines price currency if available, default is US for dollar", Type: dcmd.String},
	},
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var compactView, paginatedView bool
		var queryLimit = 25
		var country = "US"
		var itadAPIKey = confItadAPIKey.GetString()

		queryTitle := strings.ToLower(data.Args[0].Str())

		switchCountry := data.Switch("ccode")
		if switchCountry.Value != nil {
			country = strings.ToUpper(switchCountry.Str())
		}

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			paginatedView = true
		}

		// Searching for the game - new API
		var itadSearchResults ItadSearchResults
		querySearch := fmt.Sprintf("%s/games/search/v1?key=%s&title=%s&results=%d", itadAPIHost, itadAPIKey, url.QueryEscape(queryTitle), queryLimit)
		body, err := util.RequestFromAPI(querySearch)
		if err != nil {
			return nil, err
		}

		readerToDecoder := bytes.NewReader(body)
		err = json.NewDecoder(readerToDecoder).Decode(&itadSearchResults)
		if err != nil {
			return nil, err
		}

		// NEW ITAD PRICES
		var uuIDsSlice = make([]string, 0)
		for _, v := range itadSearchResults {
			uuIDsSlice = append(uuIDsSlice, v.ID)
		}

		if len(itadSearchResults) == 0 {
			return "Game title not found", nil
		}

		var itadPrices ItadPrices
		var uuIDsForPrice []string
		if compactView && !paginatedView {
			uuIDsForPrice = []string{itadSearchResults[0].ID}
		} else {
			uuIDsForPrice = uuIDsSlice
		}

		queryURL := itadAPIHost + "/games/prices/v2?key=" + itadAPIKey + "&nondeals=true&country=" + country
		getPrices, err := getItadData(queryURL, uuIDsForPrice)
		if err != nil {
			return nil, err
		}

		queryErr := json.Unmarshal(getPrices, &itadPrices)
		if queryErr != nil {
			return nil, err
		}

		if compactView && !paginatedView {
			var compactPriceNew string
			compactPrice := itadPrices[0].Deals[0]
			if compactPrice.Price.Currency == "EUR" {
				compactPriceNew = fmt.Sprintf("%.2f%s", compactPrice.Price.Amount, Currency[compactPrice.Price.Currency])
			} else {
				compactPriceNew = fmt.Sprintf("%s%.2f", Currency[compactPrice.Price.Currency], compactPrice.Price.Amount)

			}
			compactData := fmt.Sprintf("%s: %s | CUT: %d%% | %s: <%s> | ITAD: <%s>",
				html.UnescapeString(itadSearchResults[0].Title),
				compactPriceNew,
				compactPrice.Cut,
				compactPrice.Shop.Name,
				compactPrice.URL,
				fmt.Sprintf("https://isthereanydeal.com/game/%s/info/", itadSearchResults[0].Slug),
			)
			return compactData, nil
		}

		// Query for thumbnail image (appID) and other game info
		var itadGameInfo ItadGameInfo
		if !compactView && !paginatedView {
			queryInfo := fmt.Sprintf("%s/games/info/v2?key=%s&id=%s", itadAPIHost, itadAPIKey, itadSearchResults[0].ID)
			body, err = util.RequestFromAPI(queryInfo)
			if err != nil {
				return nil, err
			}

			readerToDecoder = bytes.NewReader(body)
			err = json.NewDecoder(readerToDecoder).Decode(&itadGameInfo)
			if err != nil {
				return nil, err
			}
		}

		itadComplete := &ItadComplete{
			GameInfo:   &itadGameInfo,
			Price:      &itadPrices,
			SearchData: &itadSearchResults,
			// PriceLow:   &itadPriceLowOld,
		}
		itadEmbed := embedCreator(itadComplete, 0, paginatedView, compactView)
		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(*itadComplete.SearchData), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
					i := page - 1
					paginatedEmbed := embedCreator(itadComplete, i, paginatedView, compactView)
					return paginatedEmbed, nil
				})
			if err != nil {
				return "Something went wrong making the paginated messages", nil
			}
		} else {
			return itadEmbed, nil
		}

		return pm, nil
	},
}

func embedCreator(itadComplete *ItadComplete, i int, paginated, compact bool) *discordgo.MessageEmbed {
	var currencyFromShop string
	var embedDescription, embedPriceLow, embedPriceNew, embedPriceOld string
	var embedHistLow, embedPriceNewAvg, embedPriceLowAvg, embedPriceOldAvg, embedPricingHeader string
	var priceNewAvg, priceLowAvg, priceOldAvg float64

	embedCurrency := Currency["USD"]
	if len(*itadComplete.Price) > 0 && len((*itadComplete.Price)[i].Deals) > 0 {
		currencyFromShop = (*itadComplete.Price)[i].Deals[0].Price.Currency
		currencySymbol := Currency[currencyFromShop]
		if currencySymbol != "" {
			embedCurrency = currencySymbol
		}
	}

	itadCSD := (*itadComplete.SearchData)[i]
	var itadPriceDealsSlice []ItadPricesDeal

	for _, value := range *itadComplete.Price {
		if value.ID == itadCSD.ID {
			itadPriceDealsSlice = value.Deals
		}
	}

	// let's put the historical low sky high
	priceHistoryLow := 9999.99
	for _, v := range itadPriceDealsSlice {
		if v.HistoryLow.Amount < priceHistoryLow {
			priceHistoryLow = v.HistoryLow.Amount
		}
	}

	currencyFormatting := "%[1]s%.2[2]f"
	if currencyFromShop == "EUR" {
		currencyFormatting = "%.2[2]f%[1]s"
	}

	for pos, v := range itadPriceDealsSlice {
		embedPriceNew = fmt.Sprintf(currencyFormatting, embedCurrency, v.Price.Amount)

		priceLowAvg += v.StoreLow.Amount
		priceNewAvg += v.Price.Amount
		priceOldAvg += v.Regular.Amount

		if !compact {
			if pos == 0 {
				embedPricingHeader += fmt.Sprintf("`Price Cut |%8s |%8s |%8s |` Store\n", "Current", "Lowest", "Regular")
				embedDescription = embedPricingHeader
			}

			embedPriceLow = fmt.Sprintf(currencyFormatting, embedCurrency, v.StoreLow.Amount)
			embedPriceOld = fmt.Sprintf(currencyFormatting, embedCurrency, v.Regular.Amount)

			embedDescription += fmt.Sprintf("`%9s |%8s |%8s |%8s |` [%s](%s)\n",
				fmt.Sprintf("%d%%", v.Cut),
				embedPriceNew,
				embedPriceLow,
				embedPriceOld,
				v.Shop.Name,
				v.URL,
			)
		} else {
			if pos == 0 {
				embedPricingHeader += "`Sale| Current|` Store\n"
				embedDescription = embedPricingHeader
			}

			embedDescription += fmt.Sprintf("`%4s|%8s|` [%s](%s)\n",
				fmt.Sprintf("%d%%", v.Cut),
				embedPriceNew,
				v.Shop.Name,
				v.URL,
			)
		}
	}

	lenItadPDSlice := len(itadPriceDealsSlice)
	if lenItadPDSlice != 0 {
		embedHistLow = fmt.Sprintf(currencyFormatting, embedCurrency, priceHistoryLow)
		embedPriceNewAvg = fmt.Sprintf(currencyFormatting, embedCurrency, priceNewAvg/float64(lenItadPDSlice))
		if !compact {
			embedPriceLowAvg = fmt.Sprintf(currencyFormatting, embedCurrency, priceLowAvg/float64(lenItadPDSlice))
			embedPriceOldAvg = fmt.Sprintf(currencyFormatting, embedCurrency, priceOldAvg/float64(lenItadPDSlice))

			embedDescription += fmt.Sprintf("`Average     %7s |%8s |%8s |`\n", embedPriceNewAvg, embedPriceLowAvg, embedPriceOldAvg)
			embedDescription += fmt.Sprintf("`History low %7s |`\n", embedHistLow)
		} else {
			embedDescription += fmt.Sprintf("`Avg: %8s|`\n", embedPriceNewAvg)
			embedDescription += fmt.Sprintf("`Min: %8s|`\n", embedHistLow)
		}
	} else {
		embedDescription = "```No tracked stores are currently selling this game in this region/country...```"
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "IsThereAnyDeal",
			URL:     fmt.Sprintf("https://isthereanydeal.com/game/%s/info/", itadCSD.Slug),
			IconURL: "https://d2uym1p5obf9p8.cloudfront.net/images/favicon.png",
			//IconURL: "https://raw.githubusercontent.com/IsThereAnyDeal/AugmentedSteam/master/src/img/itad.png",
		},

		Title:       html.UnescapeString(itadCSD.Title),
		Description: embedDescription,

		Color: int(rand.Int63n(16777215)),
	}

	if !compact && itadComplete.GameInfo.Appid != 0 {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/steam/apps/%d/header.jpg", itadComplete.GameInfo.Appid),
		}
	}

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "isthereanydeal.com"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}

func getItadData(urlStr string, uuIDs []string) ([]byte, error) {
	client := &http.Client{}

	jsonData, err := json.Marshal(uuIDs)
	if err != nil {
		return nil, err
	}
	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(string(jsonData)))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
	r.Header.Add("Content-Length", strconv.Itoa(len(jsonData)))
	r.Header.Add("User-Agent", common.ConfBotUserAgent.GetString())
	r.Header.Add("Authority", itadAPIHost)
	r.Header.Add("Origin", itadAPIHost)
	r.Header.Add("Referer", itadAPIHost)

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("unable to fetch data from isthereanydeal.com, status code:", resp.StatusCode)
	}
	r.Body.Close()
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, nil

}
