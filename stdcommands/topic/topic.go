package topic

import (
	"math/rand"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
)

var Command = &commands.YAGCommand{
	Cooldown:            5,
	CmdCategory:         commands.CategoryFun,
	Name:                "Topic",
	Description:         "Generates a conversation topic to help chat get moving.",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		topic := ChatTopics[rand.Intn(len(ChatTopics))]
		request := rand.Intn(2)
		if request > 0 {
			apiTopic, err := topicAPI()
			if err == nil {
				return "> " + apiTopic, nil
			}
		}

		return "> " + topic, nil
	},
}

func topicAPI() (response string, err error) {
	req, err := http.NewRequest("GET", "https://www.conversationstarters.com/generator.php", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", "curlPAGST/7.65.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	response = doc.Find("div#random").Text()

	return
}
