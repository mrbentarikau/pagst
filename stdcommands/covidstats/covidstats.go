//inspired by disea.sh > https://github.com/disease-sh/api

package covidstats

import (
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
	diseaseAPIHost = "https://disease.sh/v3/covid-19"
	queryType1     = "countries"
	queryType2     = "states"
	footerImage    = "https://upload-icon.s3.us-east-2.amazonaws.com/uploads/icons/png/2129370911599778130-512.png"
)

var Command = &commands.YAGCommand{
	CmdCategory:  commands.CategoryFun,
	Name:         "CoronaStatistics",
	Aliases:      []string{"coronastats", "cstats", "cst"},
	Description:  "WIP: Shows COVID-19 statistics sourcing Worldometer statistics. Input is country name or their ISO2/3 shorthand.",
	RunInDM:      true,
	RequiredArgs: 1,
	Arguments: []*dcmd.ArgDef{
		&dcmd.ArgDef{Name: "Location", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		&dcmd.ArgDef{Switch: "states", Name: "State name in USA"},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var cWorld = coronaWorldWideStruct{}
		var cStates = coronaStatesStruct{}
		var queryType = queryType1
		var yesterday = "false"
		where := data.Args[0].Str()

		if data.Switches["states"].Value != nil && data.Switches["states"].Value.(bool) {
			queryType = queryType2
		}

		queryURL := fmt.Sprintf(diseaseAPIHost + "/" + queryType + "/" + where + "?yesterday=" + yesterday + "&strict=true")
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
		var embed = &discordgo.MessageEmbed{}
		switch queryType {
		case "countries":

			queryErr := json.Unmarshal(body, &cWorld)
			if queryErr != nil {
				return nil, queryErr
			}

			embed = &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%s (%s)", cWorld.Country, cWorld.CountryInfo.Iso2),
				Description: "showing corona statistics for current day:",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s", cWorld.CountryInfo.Flag)},
				Color: 0x7b0e4e,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{Name: "Population", Value: fmt.Sprintf("%d", cWorld.Population), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cWorld.Cases), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cWorld.TodayCases), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cWorld.Deaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cWorld.TodayDeaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cWorld.Recovered), Inline: true},
					&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cWorld.Active), Inline: true},
					&discordgo.MessageEmbedField{Name: "Critical", Value: fmt.Sprintf("%d", cWorld.Critical), Inline: true},
					&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%d", cWorld.CasesPerOneMillion), Inline: true},
				},
				Footer: &discordgo.MessageEmbedFooter{Text: "Stay safe, protect yourself and others!", IconURL: footerImage},
			}
		case "states":
			queryErr := json.Unmarshal(body, &cStates)
			if queryErr != nil {
				return nil, queryErr
			}

			embed = &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("USA, %s", cStates.State),
				Description: "showing corona statistics for current day:",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://disease.sh/assets/img/flags/us.png"},
				Color: 0x7b0e4e,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{Name: "Population", Value: fmt.Sprintf("%d", cStates.Population), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cStates.Cases), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cStates.TodayCases), Inline: true},
					&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cStates.Deaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cStates.TodayDeaths), Inline: true},
					&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cStates.Recovered), Inline: true},
					&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cStates.Active), Inline: true},
					&discordgo.MessageEmbedField{Name: "Tests", Value: fmt.Sprintf("%d", cStates.Tests), Inline: true},
					&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%d", cStates.CasesPerOneMillion), Inline: true},
				},
				Footer: &discordgo.MessageEmbedFooter{Text: "Stay safe, protect yourself and others!", IconURL: footerImage},
			}
		}
		return embed, nil
	},
}
