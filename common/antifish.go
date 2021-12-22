/*
Using https://anti-fish.bitflow.dev/ API and its /check endpoint.
This service functions without restriction, query just needs to be regexp-filtered to avoid unnecessary requests.
*/

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	antiFishScheme   = "https"
	antiFishHost     = "anti-fish.bitflow.dev"
	antiFishURL      = fmt.Sprintf("%s://%s/", antiFishScheme, antiFishHost)
	antiFishHostPath = "check"

	transparencyReportHost     = "transparencyreport.google.com"
	transparencyReportURL      = fmt.Sprintf("%s://%s/", antiFishScheme, transparencyReportHost)
	transparencyReportHostPath = "transparencyreport/api/v3/safebrowsing/status"
)

func AntiFishQuery(phisingQuery string) (*AntiFish, error) {
	antiFish := AntiFish{}
	phisingQuery = strings.Replace(phisingQuery, "\n", " ", -1)
	queryString := fmt.Sprintf(`{"message":"%s"}`, phisingQuery)

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
	r.Header.Add("User-Agent", "Mozilla-PAGST1.12")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		antiFish.Match = false
		return &antiFish, nil

	} else if resp.StatusCode != 200 {
		respError := fmt.Errorf("Unable to fetch data from that site, status-code %d", resp.StatusCode)
		return nil, respError
	}
	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stringBody := json.Unmarshal(bytes, &antiFish)
	if stringBody != nil {
		return nil, stringBody
	}

	return &antiFish, nil
}

/*This one is really based on maybes, there's no official google documentation how this works : ) */
func TransparencyReportQuery(phisingQuery string) (*TransparencyReport, error) {
	transparencyReport := TransparencyReport{}
	queryString := fmt.Sprintf(`?site=%s`, phisingQuery)

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

	body, err := ioutil.ReadAll(resp.Body)
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
