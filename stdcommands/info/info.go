package info

import (
	"fmt"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/jonas747/dcmd/v4"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryGeneral,
	Name:        "Info",
	Description: "Responds with bot information",
	RunInDM:     true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {

		info := fmt.Sprintf(`**PAGSTDB - People Against Generally Shitty Things Discord Bot**

Control panel: <https://%s/manage>
				`, common.ConfHost.GetString())

		return info, nil
	},
}
