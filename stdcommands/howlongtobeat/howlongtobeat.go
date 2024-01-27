package howlongtobeat

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var (
	hltbScheme   = "https"
	hltbHost     = "howlongtobeat.com"
	hltbURL      = fmt.Sprintf("%s://%s/", hltbScheme, hltbHost)
	hltbHostPath = "api/search"
	hltbRawQuery = ""
)

// Command var needs a comment for lint :)
var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "HowLongToBeat",
	Aliases:                   []string{"hltb"},
	RequiredArgs:              1,
	Description:               "Game information based on query from howlongtobeat.com.\nResults are sorted by popularity, it's their default. Without -p returns the first result.\nSwitch -p gives paginated output using the Jaro-Winkler similarity metric sorting max 20 results.",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "Game-Title", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "c", Help: "Compact output"},
		{Name: "p", Help: "Paginated output"},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var compactView, paginatedView bool
		gameName := data.Args[0].Str()

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			compactView = false
			paginatedView = true
		}

		getData, err := getGameData(gameName)
		if err != nil {
			return nil, err
		}

		var hltbQuery HowlongToBeat
		queryErr := json.Unmarshal(getData, &hltbQuery)
		if queryErr != nil {
			return nil, queryErr
		}

		if len(hltbQuery.Data) == 0 {
			return "No results", nil
		}

		parsedData := parseQueryData(hltbQuery.Data, gameName)

		if compactView {
			compactData := fmt.Sprintf("%s: %s | %s | %s | <%s>",
				normaliseTitle(hltbQuery.Data[0].GameName),
				parsedData[0].CompMainHumanize,
				parsedData[0].CompPlusHumanize,
				parsedData[0].Comp100Humanize,
				parsedData[0].GameURL,
			)
			return compactData, nil
		}

		footerText := "howlongtobeat.com"
		var footerSimilarity string
		if hltbQuery.Data[0].JaroWinklerSimilarity > 0 {
			footerSimilarity = fmt.Sprintf("Term similarity: %.2f%%", hltbQuery.Data[0].JaroWinklerSimilarity*100)
		}
		footerExtra := fmt.Sprintf("%s\n%s", footerText, footerSimilarity)
		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(hltbQuery.Data), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
					i := page - 1
					sort.SliceStable(parsedData, func(i, j int) bool {
						return parsedData[i].JaroWinklerSimilarity > parsedData[j].JaroWinklerSimilarity
					})
					paginatedEmbed := embedCreator(parsedData, i, paginatedView)

					var footerSimilarity string
					if hltbQuery.Data[i].JaroWinklerSimilarity*100 > 0 {
						footerSimilarity = fmt.Sprintf("Term similarity: %.2f%%", hltbQuery.Data[i].JaroWinklerSimilarity*100)
					}

					p.FooterExtra = []string{fmt.Sprintf("%s\n%s", footerText, footerSimilarity)}
					return paginatedEmbed, nil
				}, footerExtra)
			if err != nil {
				return "Something went wrong making the paginated message", nil
			}
		} else {
			hltbEmbed := embedCreator(parsedData, 0, paginatedView)
			return hltbEmbed, nil
		}

		return pm, nil
	},
}

func embedCreator(hltbData []HowlongToBeatData, i int, paginated bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "HowLongToBeat",
			URL:     hltbData[i].GameURL,
			IconURL: "https://howlongtobeat.com/img/icons/favicon-96x96.png",
		},

		Title: normaliseTitle(hltbData[i].GameName),
		Color: int(rand.Int63n(16777215)),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: hltbData[i].ImageURL,
		},
	}

	if hltbData[i].CompMain > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Main Story", Value: hltbData[i].CompMainHumanize + fmt.Sprintf(" (%.2fh)", hltbData[i].CompMainDur.Hours())})
	}

	if hltbData[i].CompPlus > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Main + Extra", Value: hltbData[i].CompPlusHumanize + fmt.Sprintf(" (%.2fh)", hltbData[i].CompPlusDur.Hours())})
	}

	if hltbData[i].Comp100 > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Completionist", Value: hltbData[i].Comp100Humanize + fmt.Sprintf(" (%.2fh)", hltbData[i].Comp100Dur.Hours())})
	}

	if hltbData[i].ReviewScore > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Review Score", Value: fmt.Sprintf("%d%%", hltbData[i].ReviewScore)})
	}

	if hltbData[i].ProfileSteam > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Steam", Value: fmt.Sprintf("[link](https://store.steampowered.com/app/%d)", hltbData[i].ProfileSteam)})
	}

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "howlongtobeat.com"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}

func normaliseTitle(t string) string {
	return strings.Join(strings.Fields(t), " ")
}
