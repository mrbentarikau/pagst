package catfact

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
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

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("Not 200!")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	queryErr := json.Unmarshal([]byte(body), &cf)
	if queryErr != nil {
		return "", err
	}

	return cf.Fact, nil
}
