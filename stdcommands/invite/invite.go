package invite

import (
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/jonas747/dcmd/v4"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryGeneral,
	Name:        "Invite",
	Description: "Responds with bot invite link",
	RunInDM:     true,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		return "Please add the bot through the website\nhttps://" + common.ConfHost.GetString(), nil

		//		return "Please add the bot through the website\nhttps://localhost:5001", nil

	},
}
