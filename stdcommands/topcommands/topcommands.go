package topcommands

import (
	"fmt"
	"math"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var Command = &commands.YAGCommand{
	Cooldown:     2,
	CmdCategory:  commands.CategoryDebug,
	Name:         "topcommands",
	Description:  "Shows command usage stats",
	HideFromHelp: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "hours", Type: dcmd.Int, Default: 1},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw, legacy output"},
	},

	RunFunc: cmdFuncTopCommands,
}

func cmdFuncTopCommands(data *dcmd.Data) (interface{}, error) {
	hours := data.Args[0].Int()
	within := time.Now().Add(time.Duration(-hours) * time.Hour)

	var results []*TopCommandsResult
	err := common.GORM.Table(common.LoggedExecutedCommand{}.TableName()).Select("command, COUNT(id)").Where("created_at > ?", within).Group("command").Order("count(id) desc").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var raw bool
	if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
		raw = true
		out := rangeResults(results, 0, 0, hours, raw)

		return out, nil
	}

	if len(results) < 1 {
		return fmt.Sprintf("No commands executed in %d hours...", hours), nil
	}

	maxLength := 25
	pm, err := paginatedmessages.CreatePaginatedMessage(
		data.GuildData.GS.ID, data.ChannelID, 1, int(math.Ceil(float64(len(results))/float64(maxLength))), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
			i := page - 1
			description := rangeResults(results, i, maxLength, hours, raw)
			paginatedEmbed := embedCreator(description)
			return paginatedEmbed, nil
		})
	if err != nil {
		return "Something went wrong", nil
	}

	return pm, nil
}

type TopCommandsResult struct {
	Command string
	Count   int
}

func rangeResults(cmdResults []*TopCommandsResult, i, ml, hours int, raw bool) string {
	description := fmt.Sprintf("```\nCommand stats from now to %d hour(s) ago\n#    Total - Command\n", hours)
	total := 0
	for k, result := range cmdResults[i*ml:] {
		if k <= ml-1 || raw {
			description += fmt.Sprintf("#%02d: %5d - %s\n", i*ml+k+1, result.Count, result.Command)
			total += result.Count
		}
	}

	cpm := float64(total) / float64(hours) / 60

	description += fmt.Sprintf("\nTotal: %d, Commands per minute: %.1f", total, cpm)
	description += "\n```"
	return description
}

func embedCreator(description string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Description: description,
	}
	return embed
}
