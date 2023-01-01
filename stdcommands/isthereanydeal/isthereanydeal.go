package isthereanydeal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var confItadAPIKey = config.RegisterOption("yagpdb.itad_api_key", "IsThereAnyDeal API key", "")

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
		var itadAPIKey = confItadAPIKey.GetString()
		var itadAPIHost = "https://api.isthereanydeal.com"
		var queryLimit = 25
		var country = "US"

		queryTitle := strings.ToLower(data.Args[0].Str())

		switchCountry := data.Switch("ccode")
		if switchCountry.Value != nil {
			country = strings.ToUpper(switchCountry.Str())
		}

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
			paginatedView = false
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			compactView = false
			paginatedView = true
		}

		// Searching for the game
		var itadSearch ItadSearch
		querySearch := fmt.Sprintf("%s/v02/search/search/?key=%s&q=%s&limit=%d", itadAPIHost, itadAPIKey, url.QueryEscape(queryTitle), queryLimit)
		body, err := util.RequestFromAPI(querySearch)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(body), &itadSearch)
		if err != nil {
			return nil, err
		}

		if len(itadSearch.Data.Results) == 0 {
			return "Game title not found", nil
		}

		itadSearchResults := itadSearch.Data.Results

		var plainsBuilder strings.Builder
		var plainsSlice = make([]string, 0)
		for _, v := range itadSearchResults {
			if plainsBuilder.Len() != 0 {
				plainsBuilder.WriteString(",")
			}

			plainsBuilder.WriteString(v.Plain)
			plainsSlice = append(plainsSlice, v.Plain)
		}

		// Query for lowest price in stores (plains)
		var itadPriceLow ItadShopPriceLowest
		if !compactView {
			for i := 0; i <= len(plainsSlice)/5; i++ {
				var joinedPlains string
				if (len(plainsSlice)-i*5)/5 != 0 {
					joinedPlains = strings.Join(plainsSlice[i*5:i*5+5], ",")
				} else {
					joinedPlains = strings.Join(plainsSlice[i*5:], ",")
				}

				if joinedPlains != "" {
					queryPriceLow := fmt.Sprintf("%s/v01/game/storelow/?key=%s&plains=%s", itadAPIHost, itadAPIKey, joinedPlains)
					body, err = util.RequestFromAPI(queryPriceLow)
					if err != nil {
						return nil, err
					}

					err = json.Unmarshal([]byte(body), &itadPriceLow)
					if err != nil {
						continue
					}

					itadPriceLow.Compact = make(map[string]map[string]float64)
					for k, v := range itadPriceLow.Data {
						itadPriceLow.Compact[k] = make(map[string]float64)
						for _, vv := range v {
							itadPriceLow.Compact[k][vv.Shop] = vv.Price
						}
					}
				}
				if !paginatedView {
					break
				}
			}
		}

		// Query for store prices (discount, regular)
		var itadPrice ItadPrice
		queryPrice := fmt.Sprintf("%s/v01/game/prices/?key=%s&plains=%s&country=%s", itadAPIHost, itadAPIKey, plainsBuilder.String(), country)
		body, err = util.RequestFromAPI(queryPrice)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(body), &itadPrice)
		if err != nil {
			return nil, err
		}

		if compactView && !paginatedView {
			var compactPriceNew string
			compactPrice := itadPrice.Data[itadSearchResults[0].Plain].List[0]
			if itadPrice.Meta.Currency == "EUR" {
				compactPriceNew = fmt.Sprintf("%.2f%s", compactPrice.PriceNew, Currency[itadPrice.Meta.Currency])
			} else {
				compactPriceNew = fmt.Sprintf("%s%.2f", Currency[itadPrice.Meta.Currency], compactPrice.PriceNew)

			}
			compactData := fmt.Sprintf("%s: %s | CUT: %d%% | %s: <%s> | ITAD: <%s>",
				itadSearchResults[0].Title,
				compactPriceNew,
				compactPrice.PriceCut,
				compactPrice.Shop.Name,
				compactPrice.URL,
				itadPrice.Data[itadSearchResults[0].Plain].Urls.Game,
			)
			return compactData, nil
		}

		// Query for thumbnail image and other game info
		var itadGameInfo ItadGameInfo
		if !compactView {
			queryInfo := fmt.Sprintf("%s/v01/game/info/?key=%s&plains=%s", itadAPIHost, itadAPIKey, plainsBuilder.String())
			body, err = util.RequestFromAPI(queryInfo)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal([]byte(body), &itadGameInfo)
			if err != nil {
				return nil, err
			}
		}

		itadComplete := &ItadComplete{
			Search:   itadSearch,
			Price:    itadPrice,
			PriceLow: itadPriceLow,
			GameInfo: itadGameInfo,
		}

		itadEmbed := embedCreator(itadComplete, 0, paginatedView, compactView)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(itadComplete.Search.Data.Results), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
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
	var embedDescription, embedPriceLow, embedPriceNew, embedPriceOld string
	var embedPriceNewAvg, embedPriceLowAvg, embedPriceOldAvg, embedPricingHeader string
	var priceNewAvg, priceLowAvg, priceOldAvg float64

	embedCurrency := Currency[itadComplete.Price.Meta.Currency]
	if !compact {
		embedPricingHeader = "`Price Cut | Current | Lowest | Regular |` Store\n"
	} else {
		embedPricingHeader = "`Sale|Current|` Store\n"
	}

	embedDescription = embedPricingHeader

	itadPriceDataListSlice := itadComplete.Price.Data[itadComplete.Search.Data.Results[i].Plain].List
	lenItadPDL := len(itadPriceDataListSlice)

	for _, v := range itadPriceDataListSlice {
		priceLow := itadComplete.PriceLow.Compact[itadComplete.Search.Data.Results[i].Plain][v.Shop.ID]
		if itadComplete.Price.Meta.Currency == "EUR" {
			embedPriceNew = fmt.Sprintf("%.2f%s", v.PriceNew, embedCurrency)
			if !compact {
				embedPriceLow = fmt.Sprintf("%.2f%s", priceLow, embedCurrency)
				embedPriceOld = fmt.Sprintf("%.2f%s", v.PriceOld, embedCurrency)
			}
		} else {
			embedPriceNew = fmt.Sprintf("%s%.2f", embedCurrency, v.PriceNew)
			if !compact {
				embedPriceLow = fmt.Sprintf("%s%.2f", embedCurrency, priceLow)
				embedPriceOld = fmt.Sprintf("%s%.2f", embedCurrency, v.PriceOld)
			}
		}
		priceLowAvg += priceLow
		priceNewAvg += v.PriceNew
		priceOldAvg += v.PriceOld

		if !compact {
			embedDescription += fmt.Sprintf("`%9s | %7s | %6s | %7s |` [%s](%s)\n",
				fmt.Sprintf("%d%%", v.PriceCut),
				embedPriceNew,
				embedPriceLow,
				embedPriceOld,
				v.Shop.Name,
				v.URL,
			)
		} else {
			embedDescription += fmt.Sprintf("`%4s|%7s|` [%s](%s)\n",
				fmt.Sprintf("%d%%", v.PriceCut),
				embedPriceNew,
				v.Shop.Name,
				v.URL,
			)

		}
	}

	if lenItadPDL != 0 {
		if itadComplete.Price.Meta.Currency == "EUR" {
			embedPriceNewAvg = fmt.Sprintf("%.2f%s", priceNewAvg/float64(lenItadPDL), embedCurrency)
			if !compact {
				embedPriceLowAvg = fmt.Sprintf("%.2f%s", priceLowAvg/float64(lenItadPDL), embedCurrency)
				embedPriceOldAvg = fmt.Sprintf("%.2f%s", priceOldAvg/float64(lenItadPDL), embedCurrency)
			}
		} else {
			embedPriceNewAvg = fmt.Sprintf("%s%.2f", embedCurrency, priceNewAvg/float64(lenItadPDL))
			if !compact {
				embedPriceLowAvg = fmt.Sprintf("%s%.2f", embedCurrency, priceLowAvg/float64(lenItadPDL))
				embedPriceOldAvg = fmt.Sprintf("%s%.2f", embedCurrency, priceOldAvg/float64(lenItadPDL))
			}
		}
		if !compact {
			embedDescription += fmt.Sprintf("`Average     %7s | %6s | %7s |`\n", embedPriceNewAvg, embedPriceLowAvg, embedPriceOldAvg)
		} else {
			embedDescription += fmt.Sprintf("`Avg: %7s|`\n", embedPriceNewAvg)
		}

	} else {
		embedDescription = "```No tracked stores are currently selling this game in this region/country...```"
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "IsThereAnyDeal",
			URL:  itadComplete.Price.Data[itadComplete.Search.Data.Results[i].Plain].Urls.Game,
			//IconURL: "https://raw.githubusercontent.com/IsThereAnyDeal/AugmentedSteam/master/src/img/itad.png",
			IconURL: "https://d2uym1p5obf9p8.cloudfront.net/images/favicon.png",
		},

		Title:       itadComplete.Search.Data.Results[i].Title,
		Description: embedDescription,

		Color: int(rand.Int63n(16777215)),
	}

	if !compact {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: itadComplete.GameInfo.Data[itadComplete.Search.Data.Results[i].Plain].Image,
		}
	}

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "isthereanydeal.com"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}
