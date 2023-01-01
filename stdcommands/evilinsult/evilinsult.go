package evilinsult

import (
	"fmt"
	"html"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "Roast",
	Aliases:     []string{"insult"},
	Description: "sends delicious roasts from EvilInsult.com",
	Arguments: []*dcmd.ArgDef{
		{Name: "Target", Type: dcmd.User},
	},
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	NSFW:                      true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		target := "a random person nearby"
		if data.Args[0].Value != nil {
			target = data.Args[0].Value.(*discordgo.User).Username
		}

		queryInsult := "https://evilinsult.com/generate_insult.php?lang=en"

		body, err := util.RequestFromAPI(queryInsult)
		if err != nil {
			return "Not enough heat for a roast", err
		}

		embed := &discordgo.MessageEmbed{}
		embed.Title = fmt.Sprintf(`%s roasted %s`, data.Author.Username, target)
		embed.Description = html.UnescapeString(string(body))

		return embed, nil
	},
}
