/*
Using datamuse.com API and its /words endpoint.
This service functions without restriction and without an API key for up to 100,000 requests per day.
*/

package common

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
)

func wordFromAPI(wrd string, synSwitch bool) string {
	type wordFromAPI struct {
		Word string `json:"word"`
	}

	var words []wordFromAPI
	var response = wrd
	var code = "ant="

	if synSwitch {
		code = "syn="
	}

	queryURL := "https://api.datamuse.com/words?rel_" + code + wrd
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return response
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response
	}

	if resp.StatusCode != 200 {
		return response
	}
	resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response
	}

	queryErr := json.Unmarshal([]byte(body), &words)
	if queryErr != nil {
		return response
	}

	if len(words) > 0 {
		response = words[rand.Intn(len(words))].Word
	}

	return response
}
