package dogfact

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
	Name:                "DogFact",
	Aliases:             []string{"dog", "dogfacts"},
	Description:         "Dog Facts from local database or dog-api.kinduff.com API",
	SlashCommandEnabled: true,
	DefaultEnabled:      true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		request := rand.Intn(2)
		if request > 0 {
			df, err := dogFactFromAPI()
			if err == nil && len(df) > 0 {
				return df[0], nil
			}
		}
		df := dogfacts[rand.Intn(len(dogfacts))]
		return df, nil
	},
}

func dogFactFromAPI() ([]string, error) {
	var df struct {
		Facts   []string `json:"facts"`
		Success bool     `json:"success"`
	}

	query := "https://dog-api.kinduff.com/api/facts"
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Not 200!")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryErr := json.Unmarshal([]byte(body), &df)
	if queryErr != nil {
		return nil, err
	}

	return df.Facts, nil
}
