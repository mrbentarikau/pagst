package catfact

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/jonas747/dcmd/v4"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "CatFact",
	Aliases:             []string{"cf", "cat", "catfacts"},
	Description:         "Cat Facts from local database or catfact.ninja API",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var cf string
		var err error
		request := rand.Intn(2)
		if request > 0 {
			cf, err = catFactFromAPI()
			if err == nil {
				return cf, nil
			}
		}
		cf = Catfacts[rand.Intn(len(Catfacts))]
		return cf, nil
	},
}

func catFactFromAPI() (string, error) {
	var cf struct {
		Fact string `json:"fact"`
	}

	query := "https://catfact.ninja/fact"
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("Not 200!")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	queryErr := json.Unmarshal([]byte(body), &cf)
	if queryErr != nil {
		return "", err
	}

	return cf.Fact, nil
}
