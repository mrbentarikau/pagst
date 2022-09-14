package howlongtobeat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/jarowinkler"
)

func getGameData(searchTitle string) ([]byte, error) {
	u := &url.URL{
		Scheme:   hltbScheme,
		Host:     hltbHost,
		Path:     hltbHostPath,
		RawQuery: hltbRawQuery,
	}

	urlStr := u.String()

	client := &http.Client{}

	var jsonQuery HowlongToBeatQuery
	jsonQuery.SearchTerms = []string{searchTitle}
	jsonQuery.SearchType = "games"
	jsonQuery.SearchPage = 1
	jsonQuery.Size = 25
	jsonQuery.SearchOptions.Games.SortCategory = "popular"
	jsonQuery.SearchOptions.Games.RangeCategory = "main"

	jsonData, err := json.Marshal(jsonQuery)
	if err != nil {
		return nil, err
	}

	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(string(jsonData))) // URL-encoded payload
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
	r.Header.Add("Content-Length", strconv.Itoa(len(jsonData)))
	r.Header.Add("User-Agent", common.ConfBotUserAgent.GetString())
	r.Header.Add("Authority", hltbURL)
	r.Header.Add("Origin", hltbURL)
	r.Header.Add("Referer", hltbURL)

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Unable to fetch data from howlongtobeat.com, status code:", resp.StatusCode)
	}
	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func parseQueryData(hltbQuery []HowlongToBeatData, gameName string) []HowlongToBeatData {

	for i, j := range hltbQuery {
		hltbQuery[i].CompMainDur = time.Second * time.Duration(j.CompMain)
		hltbQuery[i].CompPlusDur = time.Second * time.Duration(j.CompPlus)
		hltbQuery[i].Comp100Dur = time.Second * time.Duration(j.Comp100)

		hltbQuery[i].CompMainHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].CompMainDur)
		hltbQuery[i].CompPlusHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].CompPlusDur)
		hltbQuery[i].Comp100Humanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].Comp100Dur)

		hltbQuery[i].JaroWinklerSimilarity = jarowinkler.Similarity([]rune(gameName), []rune(j.GameName))
		hltbQuery[i].GameURL = fmt.Sprintf("%sgame/%d", hltbURL, hltbQuery[i].GameID)
		hltbQuery[i].ImageURL = fmt.Sprintf("%sgames/%s", hltbURL, hltbQuery[i].GameImage)
	}

	return hltbQuery
}
