package howlongtobeat

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var (
	hltbScheme = "https"
	hltbHost   = "howlongtobeat.com"
	hltbURL    = fmt.Sprintf("%s://%s/", hltbScheme, hltbHost)
	//hltbHostPath = "search_results"
	hltbHostPath = "api/search"
	//hltbRawQuery = "page=1"
	hltbRawQuery = ""
)

// Command var needs a comment for lint :)
var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "HowLongToBeat",
	Aliases:             []string{"hltb"},
	RequiredArgs:        1,
	Description:         "Game information based on query from howlongtobeat.com.\nResults are sorted by popularity, it's their default. Without -p returns the first result.\nSwitch -p gives paginated output using the Jaro-Winkler similarity metric sorting max 20 results.",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,
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

		hltbEmbed := embedCreator(parsedData, 0, paginatedView)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(hltbQuery.Data), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					sort.SliceStable(parsedData, func(i, j int) bool {
						return parsedData[i].JaroWinklerSimilarity > parsedData[j].JaroWinklerSimilarity
					})
					paginatedEmbed := embedCreator(parsedData, i, paginatedView)
					return paginatedEmbed, nil
				})
			if err != nil {
				return "Something went wrong", nil
			}
		} else {
			return hltbEmbed, nil
		}

		return pm, nil
	},
}

func embedCreator(hltbData []HowlongToBeatData, i int, paginated bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: normaliseTitle(hltbData[i].GameName),
			URL:  hltbData[i].GameURL,
		},

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
	if paginated {
		embed.Description = fmt.Sprintf("Term similarity: %.1f%%", hltbData[i].JaroWinklerSimilarity*100)
	}
	return embed
}

func normaliseTitle(t string) string {
	return strings.Join(strings.Fields(t), " ")
}
