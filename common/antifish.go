/*
Using https://anti-fish.bitflow.dev/ API and its /check endpoint.
This service functions without restriction, query just needs to be regexp-filtered to avoid unnecessary requests.
*/

package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type AntiFish struct {
	Match   bool `json:"match"`
	Matches []struct {
		Followed    bool    `json:"followed"`
		Domain      string  `json:"domain"`
		URL         string  `json:"url"`
		Source      string  `json:"source"`
		Type        string  `json:"type"`
		TrustRating float64 `json:"trust_rating"`
	} `json:"matches"`
}

type TransparencyReport struct {
	UnsafeContent       int
	VisitorToHarmfulWeb int
	InstallsUnwanted    int
	TricksVisitor       int
	ContainsUnwanted    int
	UncommonDownloads   int
	ScoreTotal          int
}

var (
	AntiFishLinkRegex = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

	antiFishScheme   = "https"
	antiFishHost     = "anti-fish.bitflow.dev"
	antiFishHostPath = "check"

	sinkingYachtsScheme   = "https"
	sinkingYachtsHost     = "phish.sinking.yachts"
	sinkingYachtsHostPath = "/v2/check"

	transparencyReportHost     = "transparencyreport.google.com"
	transparencyReportHostPath = "transparencyreport/api/v3/safebrowsing/status"
)

func AntiFishQuery(phishingQuery string) (*AntiFish, error) {
	antiFish := AntiFish{}
	phishingQuery = strings.Replace(phishingQuery, "\n", " ", -1)
	queryBytes, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{phishingQuery})
	queryString := string(queryBytes)

	u := &url.URL{
		Scheme: antiFishScheme,
		Host:   antiFishHost,
		Path:   antiFishHostPath,
	}

	urlStr := u.String()
	client := &http.Client{}

	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(queryString)) // URL-encoded payload
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
	r.Header.Add("Content-Length", strconv.Itoa(len(queryString)))
	r.Header.Add("User-Agent", "PAGSTDB/20.42.6702")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		antiFish.Match = false
		return &antiFish, nil
	}

	if resp.StatusCode != 200 {
		respError := fmt.Errorf("unable to fetch data from AntiFish API, status-code %d", resp.StatusCode)
		return nil, respError
	}

	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stringBody := json.Unmarshal(bytes, &antiFish)
	if stringBody != nil {
		return nil, stringBody
	}

	return &antiFish, nil
}

func SinkingYachtsQuery(phishingQuery string) (bool, error) {
	var antiFish bool
	phishingQuery = strings.Replace(phishingQuery, "\n", " ", -1)

	u := &url.URL{
		Scheme: sinkingYachtsScheme,
		Host:   sinkingYachtsHost,
		Path:   sinkingYachtsHostPath,
	}

	urlStr := u.String() + "/" + phishingQuery
	client := &http.Client{}

	r, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "PAGSTDB/20.42.6702")

	resp, err := client.Do(r)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == 404 {
		return antiFish, nil
	}

	if resp.StatusCode != 200 {
		respError := fmt.Errorf("unable to fetch data from SinkingYachts API, status-code %d", resp.StatusCode)
		return false, respError
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	stringBody := json.Unmarshal(bytes, &antiFish)
	if stringBody != nil {
		return false, stringBody
	}

	return antiFish, nil
}

/*This one is really based on maybes, there's no official google documentation how this works : ) */
func TransparencyReportQuery(phishingQuery string) (*TransparencyReport, error) {
	transparencyReport := TransparencyReport{}
	queryString := fmt.Sprintf(`?site=%s`, phishingQuery)

	u := &url.URL{
		Scheme: antiFishScheme,
		Host:   transparencyReportHost,
		Path:   transparencyReportHostPath,
	}

	urlStr := u.String() + queryString
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "pagst/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reSplit := regexp.MustCompile(`,`).Split(string(body), -1)
	if len(reSplit) >= 7 {
		transparencyReport.UnsafeContent, _ = strconv.Atoi(reSplit[1])
		transparencyReport.VisitorToHarmfulWeb, _ = strconv.Atoi(reSplit[2])
		transparencyReport.InstallsUnwanted, _ = strconv.Atoi(reSplit[3])
		transparencyReport.TricksVisitor, _ = strconv.Atoi(reSplit[4])
		transparencyReport.ContainsUnwanted, _ = strconv.Atoi(reSplit[5])
		transparencyReport.UncommonDownloads, _ = strconv.Atoi(reSplit[6])

		for i, v := range reSplit {
			if i > 1 && i <= 6 {
				sum, _ := strconv.Atoi(v)
				transparencyReport.ScoreTotal += sum
			}
		}
	}

	return &transparencyReport, nil
}
