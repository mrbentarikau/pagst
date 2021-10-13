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
	Description:          "Sets the bot's status and streaming url",
	HideFromHelp:         true,
	Arguments: []*dcmd.ArgDef{
		{Name: "status", Type: dcmd.String, Default: ""},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "url", Type: dcmd.String, Default: ""},
		{Name: "idle", Help: "Switches idle status"},
	},
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		var idle, idleState string

		err := common.RedisPool.Do(radix.Cmd(&idle, "GET", "status_idle"))
		if err != nil {
			fmt.Println((fmt.Errorf("failed retrieving bot streaming status")).Error())
		}
		idleState = idle

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
		bot.SetStatus(streamingURL, data.Args[0].Str(), idle)
		if idleState != "" {
			return "Doneso... your idle state is " + idleState, nil
		} else {
			return "Doneso", nil
		}

	}),
}
