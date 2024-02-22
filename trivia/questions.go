package trivia

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

type TriviaQuestion struct {
	Answer     string   `json:"correct_answer"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
	Question   string   `json:"question"`
	Type       string   `json:"type"`
	Options    []string `json:"incorrect_answers"`
}

type TriviaResponse struct {
	Code      int               `json:"response_code"`
	Questions []*TriviaQuestion `json:"results"`
}

func FetchQuestions(amount int) ([]*TriviaQuestion, error) {
	url := fmt.Sprintf("https://opentdb.com/api.php?amount=%d&encode=base64", amount)
	responseBytes, err := util.RequestFromAPI(url)
	if err != nil {
		return nil, err
	}

	var triviaResponse TriviaResponse
	readerToDecoder := strings.NewReader(string(responseBytes))
	err = json.NewDecoder(readerToDecoder).Decode(&triviaResponse)
	if err != nil {
		return nil, err
	}

	if triviaResponse.Code != 0 {
		return nil, commands.NewPublicError("Error from Trivia API")
	}

	for _, question := range triviaResponse.Questions {
		question.Decode()
		question.Category = strings.ReplaceAll(question.Category, ": ", " - ")
		if question.Type == "boolean" {
			question.Options = []string{"True", "False"}
		} else {
			question.RandomizeOptionOrder()
		}
	}

	return triviaResponse.Questions, nil
}

func (q *TriviaQuestion) Decode() {
	q.Question, _ = common.Base64DecodeToString(q.Question)
	q.Answer, _ = common.Base64DecodeToString(q.Answer)
	q.Category, _ = common.Base64DecodeToString(q.Category)
	q.Difficulty, _ = common.Base64DecodeToString(q.Difficulty)
	q.Type, _ = common.Base64DecodeToString(q.Type)
	for index, option := range q.Options {
		q.Options[index], _ = common.Base64DecodeToString(option)
	}
}

// RandomizeOptionOrder randomizes the option order and returns the result
// this also adds the answer to the list of options
func (q *TriviaQuestion) RandomizeOptionOrder() {
	cop := make([]string, len(defaultOptionEmojis))
	copy(cop, q.Options[:len(defaultOptionEmojis)-1])
	cop[len(cop)-1] = q.Answer

	rand.Shuffle(len(cop), func(i, j int) {
		cop[i], cop[j] = cop[j], cop[i]
	})

	q.Options = cop
}
