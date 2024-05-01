package dadjoke

import (
	"bytes"
	"encoding/json"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

// Create the struct that we will serialize the API response into.
type Joke struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

var Command = &commands.YAGCommand{
	Cooldown:                  5,
	CmdCategory:               commands.CategoryFun,
	Name:                      "DadJoke",
	Description:               "Generates a dad joke using the API from icanhazdadjoke.",
	DefaultEnabled:            true,
	RunInDM:                   true,
	ApplicationCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var extraHeaders = map[string]string{"Accept": "application/json"}
		queryURL := "https://icanhazdadjoke.com"

		//Read the API response.
		responseBytes, err := util.RequestFromAPI(queryURL, extraHeaders)
		if err != nil {
			return nil, err
		}

		//Create our struct and unmarshal the content into it.
		joke := Joke{}
		readerToDecoder := bytes.NewReader(responseBytes)
		err = json.NewDecoder(readerToDecoder).Decode(&joke)
		if err != nil {
			return nil, err
		}

		//Return the joke - the other pieces are unneeded and ignored.
		return joke.Joke, nil
	},
}
