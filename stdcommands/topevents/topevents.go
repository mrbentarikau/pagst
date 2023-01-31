package topevents

import (
	"fmt"
	"math"
	"sort"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	Cooldown:     2,
	CmdCategory:  commands.CategoryDebug,
	Name:         "topevents",
	Description:  "Shows gateway event processing stats for all or one shard",
	HideFromHelp: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "shard", Type: dcmd.Int},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw, legacy output"},
	},

	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {

		//func cmdFuncTopEvents(data *dcmd.Data) (interface{}, error) {

		shardsTotal, lastPeriod := bot.EventLogger.GetStats()

		sortable := make([]*DiscordEvtEntry, len(eventsystem.AllDiscordEvents))
		for i := range sortable {
			sortable[i] = &DiscordEvtEntry{
				Name: eventsystem.AllDiscordEvents[i].String(),
			}
		}

		for i := range shardsTotal {
			if data.Args[0].Value != nil && data.Args[0].Int() != i {
				continue
			}

			for de, j := range eventsystem.AllDiscordEvents {
				sortable[de].Total += shardsTotal[i][j]
				sortable[de].PerSecond += float64(lastPeriod[i][j]) / bot.EventLoggerPeriodDuration.Seconds()
			}
		}

		sort.Sort(DiscordEvtEntrySortable(sortable))

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {

			out := "Total event stats across all shards:\n"
			if data.Args[0].Value != nil {
				out = fmt.Sprintf("Stats for shard %d:\n", data.Args[0].Int())
			}

			out += "```\n#     Total  -   /s  - Event\n"
			sum := int64(0)
			sumPerSecond := float64(0)
			for k, entry := range sortable {
				out += fmt.Sprintf("#%-2d: %7d - %5.1f - %s\n", k+1, entry.Total, entry.PerSecond, entry.Name)
				sum += entry.Total
				sumPerSecond += entry.PerSecond
			}

			out += fmt.Sprintf("\nTotal: %d, Events per second: %.1f", sum, sumPerSecond)
			out += "\n```"

			return out, nil
		}

		maxLength := 25
		pm, err := paginatedmessages.CreatePaginatedMessage(
			data.GuildData.GS.ID, data.ChannelID, 1, int(math.Ceil(float64(len(sortable))/float64(maxLength))), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
				i := page - 1
				description := rangeResults(sortable, i, maxLength, 0, false)
				paginatedEmbed := embedCreator(description)
				return paginatedEmbed, nil
			})
		if err != nil {
			return "Something went wrong", nil
		}

		return pm, nil
	}),
}

type DiscordEvtEntry struct {
	Name      string
	Total     int64
	PerSecond float64
}

type DiscordEvtEntrySortable []*DiscordEvtEntry

func (d DiscordEvtEntrySortable) Len() int {
	return len(d)
}

func (d DiscordEvtEntrySortable) Less(i, j int) bool {
	return d[i].Total > d[j].Total
}

func (d DiscordEvtEntrySortable) Swap(i, j int) {
	temp := d[i]
	d[i] = d[j]
	d[j] = temp
}

func rangeResults(cmdResults []*DiscordEvtEntry, i, ml, hours int, raw bool) string {
	description := fmt.Sprintf("```\n\n%-3s %8s - %5s - %s\n", "#", "Total", "Sec", "Event")
	total := int64(0)
	sumPerSecond := float64(0)
	for k, entry := range cmdResults[i*ml:] {
		if k <= ml-1 || raw {
			description += fmt.Sprintf("#%-2d: %7d - %5.1f - %s\n", k+1, entry.Total, entry.PerSecond, entry.Name)
			total += entry.Total
			sumPerSecond += entry.PerSecond
		}
	}

	description += fmt.Sprintf("\nTotal: %d, Events per second: %.1f", total, sumPerSecond)
	description += "\n```"
	return description
}

func embedCreator(description string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Total event stats across all shards:\n",
		Description: description,
	}
	return embed
}
