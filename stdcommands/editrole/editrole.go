package editrole

import (
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/moderation"
	"golang.org/x/image/colornames"
)

var Command = &commands.YAGCommand{
	CmdCategory:     commands.CategoryTool,
	Name:            "EditRole",
	Aliases:         []string{"ERole"},
	Description:     "Edits a role",
	LongDescription: "\nRequires the manage roles permission and the bot and your highest role being above the edited role.\nRole permissions follow Discord's standard [encoding](https://discordapp.com/developers/docs/topics/permissions) and can be calculated on websites like [this](https://discordapi.com/permissions.html).",
	RequiredArgs:    1,
	Arguments: []*dcmd.ArgDef{
		{Name: "Role", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "name", Help: "Role name - String", Type: dcmd.String, Default: ""},
		{Name: "color", Help: "Role color - Either hex code or name", Type: dcmd.String, Default: ""},
		{Name: "mention", Help: "Role Mentionable - 1 for true 0 for false", Type: &dcmd.IntArg{Min: 0, Max: 1}},
		{Name: "hoist", Help: "Role Hoisted - 1 for true 0 for false", Type: &dcmd.IntArg{Min: 0, Max: 1}},
		{Name: "perms", Help: "Role Permissions - 0 to 1099511627775 ", Type: &dcmd.IntArg{Min: 0, Max: 1099511627775}},
	},
	RunFunc:            cmdFuncEditRole,
	GuildScopeCooldown: 30,
}

func cmdFuncEditRole(data *dcmd.Data) (interface{}, error) {
	cID := data.ChannelID
	if ok, err := bot.AdminOrPermMS(data.GuildData.GS.ID, cID, data.GuildData.MS, discordgo.PermissionManageRoles); err != nil {
		return "Failed checking perms", err
	} else if !ok {
		return "You need manage roles perms to use this command", nil
	}

	roleS := data.Args[0].Str()
	role := moderation.FindRole(data.GuildData.GS, roleS)

	if role == nil {
		return "No role with the Name or ID `" + roleS + "` found", nil
	}

	if !bot.IsMemberAboveRole(data.GuildData.GS, data.GuildData.MS, role) {
		return "Can't edit roles above you", nil
	}

	change := false

	name := role.Name
	if n := data.Switch("name").Str(); n != "" {
		name = limitString(n, 100)
		change = true
	}
	color := role.Color
	if c := data.Switch("color").Str(); c != "" {
		if data.Source == dcmd.TriggerSourceDM {
			return nil, errors.New("Cannot use role color edit in custom commands to prevent api abuse")
		}
		parsedColor, ok := ParseColor(c)
		if !ok {
			return "Unknown color: " + c + ", can be either hex color code or name for a known color", nil
		}
		color = parsedColor
		change = true
	}
	mentionable := role.Mentionable
	if m := data.Switch("mention").Bool(); m != false {
		mentionable = m
		change = true
	}
	hoisted := role.Hoist
	if h := data.Switch("hoist").Bool(); h != false {
		hoisted = h
		change = true
	}
	perms := role.Permissions
	if p := data.Switch("perms").Int64(); p != 0 {
		perms = p
		change = true
	}

	if change {
		roleParams := &discordgo.RoleParams{
			Name:        name,
			Color:       &color,
			Hoist:       &hoisted,
			Permissions: &perms,
			Mentionable: &mentionable,
		}
		_, err := common.BotSession.GuildRoleEdit(data.GuildData.GS.ID, role.ID, roleParams)
		if err != nil {
			return nil, err
		}
	}

	_, err := common.BotSession.ChannelMessageSendComplex(cID, &discordgo.MessageSend{
		Content:         fmt.Sprintf("__**Role(%d) properties:**__\n\n**Name **: `%s`\n**Color **: `%d`\n**Mentionable **: `%t`\n**Hoisted **: `%t`\n**Permissions **: `%d`", role.ID, name, color, mentionable, hoisted, perms),
		AllowedMentions: discordgo.AllowedMentions{},
	})

	if err != nil {
		return nil, err
	}

	return nil, err
}

func ParseColor(raw string) (int, bool) {
	if strings.HasPrefix(raw, "#") {
		raw = raw[1:]
	}

	// try to parse as hex color code first
	parsed, err := strconv.ParseInt(raw, 16, 32)
	if err == nil {
		temp := int(parsed)
		if temp > 16777215 {
			temp = 16777215
		}
		return temp, true
	}

	// look up the color code table
	for _, v := range colornames.Names {
		if strings.EqualFold(v, raw) {
			cStruct := colornames.Map[v]

			color := (int(cStruct.R) << 16) | (int(cStruct.G) << 8) | int(cStruct.B)
			return color, true
		}
	}

	return 0, false
}

// limitstring cuts off a string at max l length, supports multi byte characters
func limitString(s string, l int) string {
	if len(s) <= l {
		return s
	}

	lastValidLoc := 0
	for i, _ := range s {
		if i > l {
			break
		}
		lastValidLoc = i
	}

	return s[:lastValidLoc]
}
