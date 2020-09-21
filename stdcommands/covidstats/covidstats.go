//inspired by disea.sh > https://github.com/disease-sh/api

package covidstats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"strings"
	//"time"

	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
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
	footerImage = "https://upload-icon.s3.us-east-2.amazonaws.com/uploads/icons/png/2129370911599778130-512.png"
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
	RunFunc: func(data *dcmd.Data) (interface{}, error) {

		var cStats = coronaWorldWideStruct{}
		var cConts = []coronaWorldWideStruct{}
		var queryType = typeCountries
		var whatDay = "current day"
		var yesterday = "false"
		var twoDaysAgo = "false"
		var where, queryURL string

		if data.Switches["continents"].Value != nil && data.Switches["continents"].Value.(bool) {
			queryType = typeContinents
		} else if data.Switches["states"].Value != nil && data.Switches["states"].Value.(bool) {
			queryType = typeStates
		}

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
		} else if (data.Args[0].Str() == "") && (queryType == typeCountries) {
			queryType = typeWorld
			queryURL = fmt.Sprintf("%s%s/%s", diseaseAPIHost, queryType, "?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		} else {
			queryURL = fmt.Sprintf("%s%s", diseaseAPIHost, queryType+"/?yesterday="+yesterday+"&twoDaysAgo="+twoDaysAgo+"&strict=true")
		}

		req, err := http.NewRequest("GET", queryURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "curlPAGST/7.65.1")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			return "Cannot fetch corona statistics data for the given location: " + where, nil
		}

		body, err := ioutil.ReadAll(resp.Body)
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
			queryErr := json.Unmarshal(body, &cStats)
			if queryErr != nil {
				return nil, queryErr
			}
		}

		var embed = &discordgo.MessageEmbed{}
		embed = &discordgo.MessageEmbed{
			Description: fmt.Sprintf("showing corona statistics for " + whatDay + ":"),
			Color:       0x7b0e4e,
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{Name: "Population", Value: fmt.Sprintf("%d", cStats.Population), Inline: true},
				&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cStats.Cases), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cStats.TodayCases), Inline: true},
				&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cStats.Deaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cStats.TodayDeaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cStats.Recovered), Inline: true},
				&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cStats.Active), Inline: true},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Stay safe, protect yourself and others!", IconURL: footerImage},
		}
		//this here is to because USA states API does not give critical conditions and to continue proper layout
		if queryType != typeStates {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Critical", Value: fmt.Sprintf("%d", cStats.Critical), Inline: true})
		}
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%.0f", cStats.CasesPerOneMillion), Inline: true},
			&discordgo.MessageEmbedField{Name: "Total Tests", Value: fmt.Sprintf("%.0f", cStats.Tests), Inline: true})

		switch queryType {
		case "all":
			embed.Title = fmt.Sprintf("Whole world")
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: globeImage}
		case "countries":
			embed.Title = fmt.Sprintf("%s (%s)", cStats.Country, cStats.CountryInfo.Iso2)
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s", cStats.CountryInfo.Flag)}
		case "continents":
			embed = &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%s", cConts[0].Continent),
				Description: fmt.Sprintf("showing corona statistics for " + whatDay + ":"),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s", cStats.CountryInfo.Flag)},
				Color: 0x7b0e4e,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{Name: "Population", Value: fmt.Sprintf("%d", cConts[0].Population), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cConts[0].Cases), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cConts[0].TodayCases), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cConts[0].Deaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cConts[0].TodayDeaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cConts[0].Recovered), Inline: true},
					&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cConts[0].Active), Inline: true},
					&discordgo.MessageEmbedField{Name: "Critical", Value: fmt.Sprintf("%d", cConts[0].Critical), Inline: true},
					&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%.0f", cConts[0].CasesPerOneMillion), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Tests", Value: fmt.Sprintf("%.0f", cConts[0].Tests), Inline: true},
				},
				Footer: &discordgo.MessageEmbedFooter{Text: "Stay safe, protect yourself and others!", IconURL: footerImage},
			}
		case "states":
			embed.Title = fmt.Sprintf("USA, %s", cStats.State)
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: "https://disease.sh/assets/img/flags/us.png"}
		}
		return embed, nil
	},
}
