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

func adjectiveAntonymsFromAPI(adj string) string {
	type wordFromAPI struct {
		Word string `json:"word"`
	}

	var antonyms []wordFromAPI
	var response = adj

	queryURL := "https://api.datamuse.com/words?rel_ant=" + adj
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response
	}

	queryErr := json.Unmarshal([]byte(body), &antonyms)
	if queryErr != nil {
		return response
	}

	if len(antonyms) > 0 {
		response = antonyms[rand.Intn(len(antonyms))].Word
	}

	return response
}
