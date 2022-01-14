package toggledbg

import (
	"github.com/mrbentarikau/pagst/common"
	"github.com/sirupsen/logrus"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/mrbentarikau/pagst/lib/dcmd"
)

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "toggledbg",
	Description:          "Toggles Debug Logging",
	HideFromHelp:         true,
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			common.SetLoggingLevel(logrus.InfoLevel)
			return "Disabled debug logging", nil
		}

		common.SetLoggingLevel(logrus.DebugLevel)
		return "Enabled debug logging", nil

	}),
}
