package setstatus

import (
	"fmt"

	"github.com/mrbentarikau/pagst/bot"
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
	Name:                 "setstatus",
	Description:          "Sets the bot's presence type, status text, online status, and optional streaming URL. Bot Admin Only",
	HideFromHelp:         true,
	Arguments: []*dcmd.ArgDef{
		{Name: "status", Type: dcmd.String, Default: ""},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "url", Type: dcmd.String, Help: "The URL to the stream. Must be on twitch.tv or youtube.com. Activity type will always be streaming if this is set.", Default: ""},
		{Name: "type", Type: dcmd.String, Help: "Set activity type. Allowed values are 'playing', 'streaming', 'listening', 'watching', 'custom', 'competing'.", Default: "custom"},
		{Name: "state", Type: dcmd.String, Help: "Additional activity state. User's current party status, or text used for a custom status.", Default: ""},
		{Name: "status", Type: dcmd.String, Help: "Set online status. Allowed values are 'online', 'idle', 'dnd', 'offline'.", Default: "online"},
		{Name: "default", Help: "Defaults to online with version number as custom status"},
	},
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		var statusType, activityType string

		statusText := data.Args[0].Str()
		streamingUrl := data.Switch("url").Str()
		stateText := data.Switch("state").Str()
		if stateText == "" {
			stateText = statusText
		}

		err := common.RedisPool.Do(radix.Cmd(&statusType, "GET", "status_type"))
		if err != nil {
			fmt.Println((fmt.Errorf("failed retrieving bot status_type")).Error())
		}

		if statusType == "" || data.Switches["status"].Raw != nil {
			statusType = data.Switch("status").Str()
		}

		err2 := common.RedisPool.Do(radix.Cmd(&activityType, "GET", "status_activity_type"))
		if err2 != nil {
			fmt.Println((fmt.Errorf("failed retrieving bot status_activity_type")).Error())
		}

		if activityType == "" || data.Switches["type"].Raw != nil {
			activityType = data.Switch("type").Str()
		}

		// default all
		if data.Switches["default"].Value != nil && data.Switches["default"].Value.(bool) {
			activityType = data.Switch("type").Str()
			statusType = data.Switch("status").Str()
		}

		switch activityType {
		case "playing", "streaming", "listening", "watching", "custom", "competing":
			// Valid activity type, do nothing
		default:
			return nil, commands.NewUserError(fmt.Sprintf("Invalid activity type %q. Allowed values are 'playing', 'streaming', 'listening', 'watching', 'custom', 'competing'", activityType))
		}

		switch statusType {
		case "online", "idle", "dnd", "offline":
			// Valid status type, do nothing
		default:
			return nil, commands.NewUserError(fmt.Sprintf("Invalid status type %q. Allowed values are 'online', 'idle', 'dnd', 'offline'", statusType))
		}

		bot.SetStatus(activityType, statusType, statusText, stateText, streamingUrl)
		return "Doneso", nil
	}),
}
