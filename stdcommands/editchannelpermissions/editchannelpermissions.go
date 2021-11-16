package editchannelpermissions

import (
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/jonas747/dcmd/v4"
	"github.com/jonas747/discordgo/v2"
	"github.com/jonas747/dstate/v4"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryTool,
	Name:                "EditChannelPermissions",
	Aliases:             []string{"EditChannel", "EChannel", "Ecpo"},
	Description:         "Edits channel permission overwrites",
	LongDescription:     "\nRequires the manage channels permission.\nOverwrite permissions follow Discord's standard [encoding](https://discordapp.com/developers/docs/topics/permissions) and can be calculated on websites like [this](https://discordapi.com/permissions.html).",
	RequiredArgs:        1,
	SlashCommandEnabled: false,
	Arguments: []*dcmd.ArgDef{
		{Name: "Channel", Type: dcmd.Channel},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "member", Help: "Member - mention/ID", Type: &commands.MemberArg{}},
		{Name: "role", Help: "Role - mention/ID", Type: &commands.RoleArg{}},
		{Name: "allow", Help: "Allow permissions - 0 to 535529258065", Default: 0, Type: &dcmd.IntArg{Min: 0, Max: 535529258065}},
		{Name: "deny", Help: "Deny permissions - 0 to 535529258065", Default: 0, Type: &dcmd.IntArg{Min: 0, Max: 535529258065}},
	},
	RunFunc:            cmdFuncEditRole,
	GuildScopeCooldown: 30,
}

func cmdFuncEditRole(data *dcmd.Data) (interface{}, error) {
	cID := data.ChannelID

	if ok, err := bot.AdminOrPermMS(data.GuildData.GS.ID, cID, data.GuildData.MS, discordgo.PermissionManageChannels); err != nil {
		return "Failed checking perms", err
	} else if !ok {
		return "You need manage roles perms to use this command", nil
	}

	var targetID, allow, deny int64
	var overwriteType discordgo.PermissionOverwriteType

	channel := data.Args[0].Value.(*dstate.ChannelState)
	cID = channel.ID

	if data.Switch("member").Value != nil {
		member := data.Switch("member").Value.(*dstate.MemberState)
		targetID = member.User.ID
		overwriteType = discordgo.PermissionOverwriteTypeMember
	}

	if data.Switch("role").Value != nil && targetID == 0 {
		role := data.Switch("role").Value.(*discordgo.Role)
		targetID = role.ID
		overwriteType = discordgo.PermissionOverwriteTypeRole
	}

	if data.Switch("allow").Raw != nil {
		allow = data.Switch("allow").Int64()
	}

	if data.Switch("deny").Raw != nil {
		deny = data.Switch("deny").Int64()
	}

	if targetID == 0 {
		return "Member or Role switches mishandled.", nil
	}

	err := common.BotSession.ChannelPermissionSet(cID, targetID, overwriteType, allow, deny)
	if err != nil {
		return nil, err
	}
	return "Doneso", nil
}
