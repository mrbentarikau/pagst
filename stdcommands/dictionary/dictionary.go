package dictionary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sirupsen/logrus"
)

var Command = &commands.YAGCommand{
	CmdCategory:  commands.CategoryFun,
	Name:         "dictionary",
	Aliases:      []string{"owldict", "owl", "dict"},
	Description:  "Get the definition of an English word using dictionaryapi.dev",
	RequiredArgs: 1,
	Cooldown:     5,
	Arguments: []*dcmd.ArgDef{
		{Name: "Query", Help: "Word to search for", Type: dcmd.String},
	},
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		query := strings.ToLower(data.Args[0].Str())
		url := "https://api.dictionaryapi.dev/api/v2/entries/en/" + url.QueryEscape(query)

		responseBytes, err := util.RequestFromAPI(url)
		if err != nil {
			if err.Error() == "HTTP err: 404" {
				return fmt.Sprintf("Could not find a definition for word: `%s`.", query), nil
			}
			return nil, err
		}

		var res []DictionaryResponse
		readerToDecoder := bytes.NewReader(responseBytes)
		err = json.NewDecoder(readerToDecoder).Decode(&res)
		if err != nil || len(res[0].Meanings) == 0 {
			logrus.WithError(err).Error("Failed getting response from dictionaryAPI.dev")
			return "Could not find a definition for that word.", err
		}

		var dictionary = &res[0]
		embedColor := int(rand.Int63n(0xffffff))

		if len(dictionary.Meanings) == 1 || data.Context().Value(paginatedmessages.CtxKeyNoPagination) != nil {
			return createDictionaryDefinitionEmbed(dictionary, &dictionary.Meanings[0], embedColor), nil
		}

		pm, err := paginatedmessages.CreatePaginatedMessage(data.GuildData.GS.ID, data.ChannelID, 1, len(dictionary.Meanings), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
			if page > len(dictionary.Meanings) {
				return nil, paginatedmessages.ErrNoResults
			}

			return createDictionaryDefinitionEmbed(dictionary, &dictionary.Meanings[page-1], embedColor), nil
		})

		return pm, err
	},
}

func createDictionaryDefinitionEmbed(res *DictionaryResponse, def *Meaning, color int) *discordgo.MessageEmbed {
	title := strings.Title(normalizeOutput(res.Word))

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: common.CutStringShort(capitalizeSentences(normalizeOutput(def.Definitions[0].Definition)), 2048),
		Color:       color, //0x07AB99
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if len(res.SourceUrls) > 0 {
		embed.URL = res.SourceUrls[0]
	}

	var description string
	for _, d := range def.Definitions {
		if len(description)+len(d.Definition)+len(d.Example) > 2000 {
			// if all definitions along with examples cannot be fit into the description, skip remaining definitions.
			break
		}
		description = fmt.Sprintf("%s\n- %s", description, capitalizeSentences(normalizeOutput(d.Definition)))
		if d.Example != "" {
			var example = capitalizeSentences(normalizeOutput(d.Example))
			if !hasEndOfSentenceSymbol(example) {
				example = example + "." // add period if no other symbol that ends the sentence is present
			}
			description = fmt.Sprintf("%s\n**Example:** *%s*\n", description, example)
		}
	}

	embed.Description = common.CutStringShort(description, 2048)

	if res.Origin != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Origin",
			Value:  normalizeOutput(res.Origin),
			Inline: true,
		})
	}

	if len(res.Phonetics) != 0 {
		var pronunciation = &discordgo.MessageEmbedField{
			Name:   "Pronunciation",
			Value:  "",
			Inline: true,
		}
		for _, v := range res.Phonetics {
			if v.Audio != "" {
				if v.Text == "" {
					v.Text = res.Word
				}
				pronunciation.Value = fmt.Sprintf("%s\n🔊[%s](%s)", pronunciation.Value, normalizeOutput(v.Text), v.Audio)
			} else {
				pronunciation.Value = fmt.Sprintf("%s\n%s", pronunciation.Value, normalizeOutput(v.Text))
			}
		}
		embed.Fields = append(embed.Fields, pronunciation)
	}

	if def.PartOfSpeech != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Type",
			Value:  strings.Title(normalizeOutput(def.PartOfSpeech)),
			Inline: true,
		})
	}

	return embed
}

var policy = bluemonday.StrictPolicy()

func normalizeOutput(s string) string {
	// The API occasionally returns HTML tags and escapes as part of output, remove them.
	decoded := html.UnescapeString(policy.Sanitize(s))
	// It also sometimes returns non-printable characters, strip them out too.
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, decoded)
}
func capitalizeSentences(s string) string {
	var builder strings.Builder
	capitalizeCur := true // whether the current phrase should be capitalized.
	for i, word := range strings.Fields(s) {
		if i > 0 {
			builder.WriteByte(' ')
		}
		if capitalizeCur {
			// strings.Title() does not work properly with punctuation: for example, "what's" becomes 'What'S" when passed to it, which is undesirable.
			// Instead, title-case the first rune manually and write the rest as is, as we know `word` represents a single word.
			r, size := utf8.DecodeRuneInString(word)
			if r == utf8.RuneError {
				// fall back to original text
				builder.WriteString(word)
			} else {
				builder.WriteRune(unicode.ToTitle(r))
				builder.WriteString(word[size:])
			}
		} else {
			builder.WriteString(word)
		}
		capitalizeCur = hasEndOfSentenceSymbol(word)
	}
	return builder.String()
}
func hasEndOfSentenceSymbol(s string) bool {
	if len(s) == 0 {
		return false
	}
	switch s[len(s)-1] {
	case '.', '?', '!':
		return true
	default:
		return false
	}
}

type Phonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type Definition struct {
	Definition string   `json:"definition"`
	Synonyms   []string `json:"synonyms"`
	Antonyms   []string `json:"antonyms"`
	Example    string   `json:"example,omitempty"`
}

type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
	Synonyms     []string     `json:"synonyms"`
	Antonyms     []string     `json:"antonyms"`
}

type DictionaryResponse struct {
	Origin     string     `json:"origin,omitempty"`
	Word       string     `json:"word"`
	Phonetics  []Phonetic `json:"phonetics"`
	Meanings   []Meaning  `json:"meanings"`
	SourceUrls []string   `json:"sourceUrls"`
}
