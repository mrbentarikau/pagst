package removespecialserver

import (
	"github.com/mediocregopher/radix/v3"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "removespecialserver",
	Description:          ";))",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.String},
	},
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {

		var whitelisted bool
		err := common.RedisPool.Do(radix.FlatCmd(&whitelisted, "SREM", "special_servers", data.Args[0].Str()))
		if err != nil {
			return nil, err
		}

		if !whitelisted {
			return "Server wasnt whitelisted", nil
		}

		return "UnWhitelisted server", nil
	}),
}
