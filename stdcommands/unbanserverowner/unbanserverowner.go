package unbanserverowner

import (
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
	"github.com/mediocregopher/radix/v3"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "unbanserverowner",
	Description:          ";))\n\nRemoves the bot ban from the specified owner. Bot Owner Only",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.String},
	},
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {

		var unbanned bool
		err := common.RedisPool.Do(radix.Cmd(&unbanned, "SREM", "banned_server_owners", data.Args[0].Str()))
		if err != nil {
			return nil, err
		}

		if !unbanned {
			return "User wasn't banned", nil
		}

		return "Unbanned owner", nil
	}),
}
