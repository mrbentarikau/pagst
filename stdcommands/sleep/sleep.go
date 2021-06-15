package sleep

import (
	"time"

	"github.com/jonas747/dcmd/v3"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "sleep",
	Description:          "Maintenance command, used to test command queueing",
	HideFromHelp:         true,
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		time.Sleep(time.Second * 5)
		return "Slept, Done", nil
	}),
}
