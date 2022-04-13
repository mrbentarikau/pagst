package antifish

import (
	"fmt"
	"strings"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var backSlashReplacer = strings.NewReplacer("\\", "")

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "AntiPhish",
	Aliases:             []string{"antifish", "af", "ap"},
	Description:         "Anti-Fish information from anti-fish.bitflow.dev and phish.sinking.yachts (abbr. SY)...",
	DefaultEnabled:      true,
	RequireDiscordPerms: []int64{discordgo.PermissionManageGuild},
	RequiredArgs:        1,
	SlashCommandEnabled: false,
	Arguments: []*dcmd.ArgDef{
		{Name: "URL", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		query := data.Args[0].Str()
		var err error
		response := "not a valid URL"
		var af *common.AntiFish

		if common.LinkRegexJonas.MatchString(backSlashReplacer.Replace(query)) {
			af, err = common.AntiFishQuery(query)
			if err != nil {
				return nil, commands.NewPublicError(fmt.Sprintf("%s", err))
			}
			if !af.Match {
				response = "Match: false"
			} else {
				response = fmt.Sprintf("Match: %t\nFollowed: %t\nDomain: %s\nSource: %s\nType: %s\nTrust rating: %.2f\n", af.Match, af.Matches[0].Followed, af.Matches[0].Domain, af.Matches[0].Source, af.Matches[0].Type, af.Matches[0].TrustRating)
			}

			sinkingYachts, _ := common.SinkingYachtsQuery(common.DomainFinderRegex.FindString(query))
			response += fmt.Sprintf("\nSY Match: %t", sinkingYachts)
		}

		return response, nil
	},
}
