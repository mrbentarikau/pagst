package exportuserdatabase

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

type ExportTUD struct {
	EntryID string `json:"EntryID"`
	UserID  string `json:"UserID"`
	Key     string `json:"Key"`
}

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "exportuserdatabase",
	Aliases:              []string{"exportud", "extud"},
	RequireDiscordPerms:  []int64{discordgo.PermissionAdministrator},
	Description:          "Exports all your custom database id, userID, key entries as JSON,\nuser has to be serverAdmin.\nServerID argument is for the owner of the bot...",
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
			send := &discordgo.MessageSend{Content: "User Database Export"}
			if exportCSV {
				var in strings.Builder
				in.WriteString("EntryID\tUserID\tKey")
				for _, r := range result {
					in.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.EntryID, r.UserID, r.Key))
				}
				send.File = &discordgo.File{
					ContentType: "text/csv",
					Name:        fmt.Sprintf("database_entries_%d.csv", data.GuildData.GS.ID),
					Reader:      strings.NewReader(in.String()),
				}
			} else {
				buf, _ := json.Marshal(result)
				send.File = &discordgo.File{
					ContentType: "application/json",
					Name:        fmt.Sprintf("database_entries_%d.json", data.GuildData.GS.ID),
					Reader:      bytes.NewReader(buf),
				}
			}

			_, err = common.BotSession.ChannelMessageSendComplex(data.ChannelID, send)
			return nil, err
		}

		return "No database entries found...", err
	},
}

func dbQuery(guildID int64) ([]*ExportTUD, error) {
	const q = `
			SELECT
				tud.id,
				tud.user_id as userID,
				tud.key
			FROM templates_user_database AS tud
			WHERE tud.guild_id=$1 ORDER BY id`

	rows, err := common.PQ.Query(q, guildID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*ExportTUD{}, nil
		}
		return []*ExportTUD{}, err
	}
	defer rows.Close()

	result := make([]*ExportTUD, 0)
	for rows.Next() {
		var entryID, userID, key string
		var err = rows.Scan(&entryID, &userID, &key)
		if err != nil {
			return []*ExportTUD{}, err
		}

		result = append(result, &ExportTUD{
			EntryID: entryID,
			UserID:  userID,
			Key:     key,
		})
	}
	if len(result) > 0 {
		return result, nil
	}
	return nil, err
}
