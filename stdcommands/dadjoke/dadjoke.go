package dadjoke

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
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
	ApplicationCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		//Define the request and website we will navigate to.
		req, err := http.NewRequest("GET", "https://icanhazdadjoke.com", nil)
		if err != nil {
			return nil, err
		}

		//Set the headers that will be sent to the API to determine the response.
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "PAGST/20.42.6702")

		client := &http.Client{}
		apiResp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		//Once the rest of the function is done close our connection the API.
		defer apiResp.Body.Close()

		//Read the API response.
		bytes, err := io.ReadAll(apiResp.Body)
		if err != nil {
			return nil, err
		}

		//Create our struct and unmarshal the content into it.
		joke := Joke{}
		err = json.Unmarshal(bytes, &joke)
		if err != nil {
			return nil, err
		}

		//Return the joke - the other pieces are unneeded and ignored.
		return joke.Joke, nil
	},
}
