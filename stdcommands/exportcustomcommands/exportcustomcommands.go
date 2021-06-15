package exportcustomcommands

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jonas747/dcmd/v3"
	"github.com/jonas747/discordgo"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
)

type ExportCC struct {
	GuildID     int64  `json:"GuildID"`
	CCID        int64  `json:"CCID"`
	GroupName   string `json:"GroupName"`
	TriggerType int    `json:"TriggerType"`
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
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		guildIDToMatch := data.GuildData.GS.ID
		if data.Args[0].Value != nil {
			if common.IsOwner(data.Author.ID) {
				guildIDToMatch = data.Args[0].Int64()
			} else {
				return "Only for owner of the bot", nil
			}

		}
		result, err := dbQuery(guildIDToMatch)
		if err != nil {
			return nil, err
		}
		if result != nil {
			buf, _ := json.Marshal(result)
			send := &discordgo.MessageSend{
				Content: "Custom Commands Export",
				File: &discordgo.File{
					ContentType: "application/json",
					Name:        fmt.Sprintf("custom_commands_%d.json", data.GuildData.GS.ID),
					Reader:      bytes.NewReader(buf),
				},
			}
			_, err = common.BotSession.ChannelMessageSendComplex(data.ChannelID, send)
			return nil, err
		}

		return fmt.Sprintf("No CCs found."), err
	},
}

func dbQuery(guildID int64) ([]*ExportCC, error) {
	const q = `SELECT
						   		ccs.guild_id, 
						   		ccs.local_id as ccID,
						   		case(select 1 where ccg.name is null) when 1 then 'ungrouped' else ccg.name end as group_name,
						   		ccs.trigger_type, 
						   		ccs.text_trigger, 
						   		ccs.responses
						   FROM custom_commands AS ccs 
						   LEFT JOIN custom_command_groups AS ccg 
						   ON (ccs.guild_id = ccg.guild_id and ccs.group_id = ccg.id)
						   WHERE ccs.guild_id=$1`

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
		var guildID, ccID int64
		var groupName, textTrigger string
		var triggerType int
		var responses string
		var err = rows.Scan(&guildID, &ccID, &groupName, &triggerType, &textTrigger, &responses)
		if err != nil {
			return []*ExportCC{}, err
		}

		result = append(result, &ExportCC{
			GuildID:     guildID,
			CCID:        ccID,
			GroupName:   groupName,
			TriggerType: triggerType,
			TextTrigger: textTrigger,
			Responses:   responses,
		})
	}
	if len(result) > 0 {
		return result, nil
	}
	return nil, err
}
