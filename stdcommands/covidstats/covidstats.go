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

type coronaStruct struct {
	Updated                int64
	Country                string
	CountryInfo            countryInfoStruct
	Cases                  int64
	TodayCases             int64
	Deaths                 int64
	TodayDeaths            int64
	Recovered              int64
	TodayRecovered         int64
	Active                 int64
	Critical               int64
	CasesPerOneMillion     int64
	DeathsPerOneMillion    int64
	Tests                  int64
	TestsPerOneMillion     int64
	Population             int64
	Continent              string
	OneCasePerPeople       int64
	OneDeathPerPeople      int64
	OneTestPerPeople       int64
	ActivePerOneMillion    float64
	RecoveredPerOneMillion float64
	CriticalPerOneMillion  float64
}

type countryInfoStruct struct {
	_id  int64 //does not parse this not important
	Iso2 string
	Iso3 string
	Lat  int64
	Long int64
	Flag string
}

var (
	diseaseAPIHost = "https://disease.sh/v3/covid-19/"
	queryType1     = "countries/"
	queryType2     = "states/"
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
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var cStat = coronaStruct{}
		where := data.Args[0].Str()

		var yesterday = "false"
		queryURL := fmt.Sprintf(diseaseAPIHost + "countries/" + where + "?yesterday=" + yesterday + "&strict=true")
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

		queryErr := json.Unmarshal(body, &cStat)
		if queryErr != nil {
			return nil, queryErr
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s (%s)", cStat.Country, cStat.CountryInfo.Iso2),
			Description: "showing corona statistics for current day",
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s", cStat.CountryInfo.Flag)},
			Color: 0x7b0e4e,
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{Name: "Total Cases", Value: fmt.Sprintf("%d", cStat.Cases), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Cases", Value: fmt.Sprintf("%d", cStat.TodayCases), Inline: true},
				&discordgo.MessageEmbedField{Name: "Total Deaths", Value: fmt.Sprintf("%d", cStat.Deaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "New Deaths", Value: fmt.Sprintf("%d", cStat.TodayDeaths), Inline: true},
				&discordgo.MessageEmbedField{Name: "Recovered", Value: fmt.Sprintf("%d", cStat.Recovered), Inline: true},
				&discordgo.MessageEmbedField{Name: "Active", Value: fmt.Sprintf("%d", cStat.Active), Inline: true},
				&discordgo.MessageEmbedField{Name: "Critical", Value: fmt.Sprintf("%d", cStat.Critical), Inline: true},
				&discordgo.MessageEmbedField{Name: "Cases/1M pop", Value: fmt.Sprintf("%d", cStat.CasesPerOneMillion), Inline: true},
			},
			Footer: &discordgo.MessageEmbedFooter{Text: "Stay safe, protect yourself and others!"},
		}
		return embed, nil
	},
}
