package setstatus

import (
	"fmt"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/jonas747/dcmd/v4"
	"github.com/mediocregopher/radix/v3"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "setstatus",
	Description:          "Sets the bot's status, streaming url,\nactivity type (gaming/listening/watching)\nand toggles idle state.",
	HideFromHelp:         true,
	Arguments: []*dcmd.ArgDef{
		{Name: "status", Type: dcmd.String, Default: ""},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "url", Type: dcmd.String, Default: "", Help: "Status text"},
		{Name: "activity", Type: dcmd.String, Default: "", Help: "Activity type"},
		{Name: "idle", Help: "Idle switch"},
	},
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		var idle, idleState, activityType string

		err := common.RedisPool.Do(radix.Cmd(&idle, "GET", "status_idle"))
		if err != nil {
			fmt.Println((fmt.Errorf("failed retrieving bot streaming status")).Error())
		}
		idleState = idle

		err2 := common.RedisPool.Do(radix.Cmd(&activityType, "GET", "status_activity"))
		if err2 != nil {
			fmt.Println((fmt.Errorf("failed retrieving bot streaming status")).Error())
		}

		if data.Switch("activity").Str() != "" {
			activityType = data.Switch("activity").Str()
		}

		streamingURL := data.Switch("url").Str()
		if data.Switches["idle"].Value != nil && data.Switches["idle"].Value.(bool) {
			if idle == "" {
				idle = "enabled"
				idleState = "enabled"
			} else {
				idle = ""
				idleState = "disabled"
			}
		}

		bot.SetStatus(streamingURL, data.Args[0].Str(), idle, activityType)
		if idleState != "" {
			return "Doneso... Your idle state is " + idleState, nil
		} else {
			return "Doneso", nil
		}

	}),
}
