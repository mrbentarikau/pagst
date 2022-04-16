package exportcustomcommands

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

type ExportCC struct {
	GuildID     string `json:"GuildID"`
	CCID        string `json:"CCID"`
	GroupName   string `json:"GroupName"`
	TriggerType string `json:"TriggerType"`
	TextTrigger string `json:"TextTrigger"`
	Responses   string `json:"Responses"`
}

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "exportcustomscommands",
	Aliases:              []string{"exportccs", "eccs"},
	RequireDiscordPerms:  []int64{discordgo.PermissionAdministrator},
	Description:          "Exports all your custom commands data's reasonable fields as JSON,\nuser has to be serverAdmin.\nServerID argument is for the owner of the bot...",
	HideFromHelp:         true,
	Arguments: []*dcmd.ArgDef{
		{Name: "ServerID", Type: dcmd.Int},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "csv", Help: "Export to CSV with TAB as delimiter"},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		guildIDToMatch := data.GuildData.GS.ID
		if data.Args[0].Value != nil {
			if common.IsOwner(data.Author.ID) {
				guildIDToMatch = data.Args[0].Int64()
			} else {
				return "Only for owner of the bot", nil
			}

		}
		var exportCSV bool
		if data.Switches["csv"].Value != nil && data.Switches["csv"].Value.(bool) {
			exportCSV = true
		}

		result, err := dbQuery(guildIDToMatch)
		if err != nil {
			return nil, err
		}
		if result != nil {
			send := &discordgo.MessageSend{Content: "Custom Commands Export"}
			if exportCSV {
				in := fmt.Sprintln("GuildID\tCCID\tGroupName\tTriggerType\tTextTrigger\tResponses")
				for _, r := range result {
					in += fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", r.GuildID, r.CCID, r.GroupName, r.TriggerType, r.TextTrigger, strings.ReplaceAll(strings.ReplaceAll(r.Responses, "\r\n", " "), "\t", " "))
				}
				send.File = &discordgo.File{
					ContentType: "text/csv",
					Name:        fmt.Sprintf("custom_commands_%d.csv", data.GuildData.GS.ID),
					Reader:      strings.NewReader(in),
				}
			} else {
				buf, _ := json.Marshal(result)
				send.File = &discordgo.File{
					ContentType: "application/json",
					Name:        fmt.Sprintf("custom_commands_%d.json", data.GuildData.GS.ID),
					Reader:      bytes.NewReader(buf),
				}
			}

			_, err = common.BotSession.ChannelMessageSendComplex(data.ChannelID, send)
			return nil, err
		}

		return "No CCs found...", err
	},
}

func dbQuery(guildID int64) ([]*ExportCC, error) {
	const q = `
			SELECT
				ccs.guild_id, 
				ccs.local_id as ccID,
				case(select 1 where ccg.name is null) when 1 then 'ungrouped' else ccg.name end as group_name,
				ccs.trigger_type, 
				ccs.text_trigger, 
				ccs.responses
			FROM custom_commands AS ccs 
			LEFT JOIN custom_command_groups AS ccg 
			ON (ccs.guild_id = ccg.guild_id and ccs.group_id = ccg.id)
			WHERE ccs.guild_id=$1 ORDER BY ccID`

	rows, err := common.PQ.Query(q, guildID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*ExportCC{}, nil
		}
		return []*ExportCC{}, err
	}
	defer rows.Close()

	result := make([]*ExportCC, 0)
	for rows.Next() {
		//var guildID, ccID int64
		var guildID, ccID, groupName, textTrigger string
		var triggerType int
		var responses string
		var err = rows.Scan(&guildID, &ccID, &groupName, &triggerType, &textTrigger, &responses)
		if err != nil {
			return []*ExportCC{}, err
		}

		types := map[int]string{
			10: "None",
			0:  "Command",
			1:  "Starts With",
			2:  "Contains",
			3:  "Regex",
			4:  "Exact Match",
			5:  "Interval",
			6:  "Reaction",
		}

		triggerTypeString := types[triggerType]

		result = append(result, &ExportCC{
			GuildID:     guildID,
			CCID:        ccID,
			GroupName:   groupName,
			TriggerType: triggerTypeString,
			TextTrigger: textTrigger,
			Responses:   responses,
		})
	}
	if len(result) > 0 {
		return result, nil
	}
	return nil, err
}
