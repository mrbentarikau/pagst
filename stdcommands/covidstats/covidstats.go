//inspired by https://disease.sh > https://github.com/disease-sh/api

package covidstats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"math"
	"net/http"
	//"strings"
	//"time"

	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/bot/paginatedmessages"
	"github.com/jonas747/yagpdb/commands"
)

var (
	diseaseAPIHost = "https://disease.sh/v3/covid-19/"
	typeWorld      = "all"
	typeCountries  = "countries"
	typeContinents = "continents"
	typeStates     = "states"

	//These image links could disappear at random times.
	globeImage  = "http://pngimg.com/uploads/globe/globe_PNG63.png"
	africaImage = "http://endlessicons.com/wp-content/uploads/2012/12/africa-icon1-614x460.png"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryTool,
	Name:        "CoronaStatistics",
	Aliases:     []string{"coronastats", "cstats", "cst"},
	Description: "WIP: Shows COVID-19 statistics sourcing Worldometer statistics. Input is country name or their ISO2/3 shorthand.",
	RunInDM:     true,
	Arguments: []*dcmd.ArgDef{
		&dcmd.ArgDef{Name: "Location", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		&dcmd.ArgDef{Switch: "countries", Name: "Countries of the World"},
		&dcmd.ArgDef{Switch: "continents", Name: "The Continents of the World"},
		&dcmd.ArgDef{Switch: "states", Name: "A State name in USA"},
		&dcmd.ArgDef{Switch: "1d", Name: "Stats from yesterday"},
		&dcmd.ArgDef{Switch: "2d", Name: "Stats from the day before yesterday (does not apply to states)"},
	},
	RunFunc: paginatedmessages.PaginatedCommand(0, func(data *dcmd.Data, p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {

		var cStats coronaWorldWideStruct
		var cConts []coronaWorldWideStruct
		var queryType = typeCountries
		var whatDay = "current day"
		var yesterday = "false"
		var twoDaysAgo = "false"
		var where, queryURL string
		var flag string

		//to determine what will happen and what data gets shown
		if data.Switches["countries"].Value != nil && data.Switches["countries"].Value.(bool) {
			flag = typeCountries
		} else if data.Switches["continents"].Value != nil && data.Switches["continents"].Value.(bool) {
			queryType = typeContinents
		} else if data.Switches["states"].Value != nil && data.Switches["states"].Value.(bool) {
			queryType = typeStates
		}

		//day-back switches
		if data.Switches["1d"].Value != nil && data.Switches["1d"].Value.(bool) {
			whatDay = "yesterday"
			yesterday = "true"
		} else if data.Switches["2d"].Value != nil && data.Switches["2d"].Value.(bool) {
			whatDay = "day before yesterday"
			twoDaysAgo = "true"
		}

		if data.Args[0].Str() != "" {
			where = data.Args[0].Str()
			queryURL = fmt.Sprintf("%s%s/%s", diseaseAPIHost, queryType, where+"?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		} else if (data.Args[0].Str() == "") && (flag == typeCountries) {
			queryURL = fmt.Sprintf("%s%s/%s", diseaseAPIHost, queryType, "?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		} else if (data.Args[0].Str() == "") && (queryType == typeCountries) {
			queryType = typeWorld
			queryURL = fmt.Sprintf("%s%s/%s", diseaseAPIHost, queryType, "?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		} else {
			queryURL = fmt.Sprintf("%s%s/%s", diseaseAPIHost, queryType, "?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		}

		//let's get that API data
		body, err := getData(queryURL, where, queryType)
		if err != nil {
			return nil, err
		}

		//voodoo-hoodoo on detecting if JSON's array/object
		jsonDetector := bytes.TrimLeft(body, " \t\r\n")
		mapYes := len(jsonDetector) > 0 && jsonDetector[0] == '['
		mapNo := len(jsonDetector) > 0 && jsonDetector[0] == '{'
		if mapYes {
			queryErr := json.Unmarshal([]byte(body), &cConts)
			if queryErr != nil {
				return nil, queryErr
			}
		} else if mapNo {
			queryErr := json.Unmarshal([]byte(body), &cStats)
			if queryErr != nil {
				return nil, queryErr
			}
		}

		//let's render everything to slice
		cConts = append(cConts, cStats)
		i := page - 1
		if page > len(cConts) && p != nil && p.LastResponse != nil {
			return nil, paginatedmessages.ErrNoResults
		}
		var embed = &discordgo.MessageEmbed{}
		embed = &discordgo.MessageEmbed{
			Description: fmt.Sprintf("showing corona statistics for " + whatDay + ":"),
			Color:       0x7b0e4e,
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{Name: "Population", Value: fmt.Sprintf("%d", cConts[i].Population), Inline: true},
				&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cConts[i].Cases), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cConts[i].TodayCases), Inline: true},
				&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cConts[i].Deaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cConts[i].TodayDeaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cConts[i].Recovered), Inline: true},
				&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cConts[i].Active), Inline: true},
			},
		}
		//this here is to because USA states API does not give critical conditions and to continue proper layout
		if queryType != typeStates {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Critical", Value: fmt.Sprintf("%d", cConts[i].Critical), Inline: true})
		}
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%.0f", cConts[i].CasesPerOneMillion), Inline: true},
			&discordgo.MessageEmbedField{Name: "Total Tests", Value: fmt.Sprintf("%.0f", cConts[i].Tests), Inline: true})

		switch queryType {
		case "all":
			embed.Title = fmt.Sprintf("Whole world")
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: globeImage}
		case "countries":
			embed.Title = fmt.Sprintf("%s (%s)", cConts[i].Country, cConts[i].CountryInfo.Iso2)
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s", cConts[i].CountryInfo.Flag)}
		case "continents":
			embed.Title = fmt.Sprintf("%s", cConts[i].Continent)
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: africaImage}
		case "states":
			embed.Title = fmt.Sprintf("USA, %s", cConts[i].State)
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: "https://disease.sh/assets/img/flags/us.png"}
		}

		return embed, nil
	}),
}

func getData(query, loc, qtype string) ([]byte, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Cannot fetch corona statistics data for the given location:** " + qtype + ": " + loc + "**")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
