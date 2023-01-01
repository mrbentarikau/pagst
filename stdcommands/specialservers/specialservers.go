package specialservers

import (
	"fmt"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"

	"github.com/mediocregopher/radix/v3"
)

func Commands() *dcmd.Container {
	container, _ := commands.CommandSystem.Root.Sub("specialserver", "specialservers")
	container.Description = "utility for adding special permissions to server/guild"
	container.AddMidlewares(util.RequireBotAdmin)
	container.AddCommand(addServer, addServer.GetTrigger())
	container.AddCommand(listServers, listServers.GetTrigger())
	container.AddCommand(removeServer, removeServer.GetTrigger())

	return container
}

var addServer = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "add",
	Description:          ";))",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.Int},
	},
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {
		var whitelisted bool
		err := common.RedisPool.Do(radix.FlatCmd(&whitelisted, "SADD", "special_servers", data.Args[0].Int64()))
		if err != nil {
			return "", err
		}

		if !whitelisted {
			return "Server was already whitelisted", nil
		}
		return "Whitelisted: " + data.Args[0].Str(), nil
	}),
}

var removeServer = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "remove",
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
			return "Server wasn't whitelisted", nil
		}

		return "UnWhitelisted server", nil
	}),
}

type List struct {
	ID int64
}

var listServers = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "list",
	Description:          ";))",
	HideFromHelp:         true,

	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {

		var whitelistSlice []int64
		err := common.RedisPool.Do(radix.Cmd(&whitelistSlice, "SMEMBERS", "special_servers"))
		if err != nil {
			return nil, err
		}

		if len(whitelistSlice) == 0 {
			return "Server wasn't whitelisted", nil
		}

		whitelisted := "```"
		for i, j := range whitelistSlice {
			i++
			gs := bot.State.GetGuild(j)
			if gs == nil {
				return fmt.Sprintf("Guild %d does not exist in whitelisted list", j), nil
			}
			whitelisted += fmt.Sprintf("%d. %d - %s, ownerID:%d\n", i, j, gs.Name, gs.OwnerID)
		}
		whitelisted += "```"

		send := &discordgo.MessageSend{Content: whitelisted}

		return send, nil
	}),
}
