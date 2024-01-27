package poll

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var (
	pollReactions = [...]string{"1⃣", "2⃣", "3⃣", "4⃣", "5⃣", "6⃣", "7⃣", "8⃣", "9⃣", "🔟"}
	Command       = &commands.YAGCommand{
		CmdCategory:               commands.CategoryTool,
		Name:                      "Poll",
		Description:               "Create very simple reaction poll. Example: `poll \"favorite color?\" blue red pink`",
		RequiredArgs:              3,
		ApplicationCommandEnabled: true,
		Arguments: []*dcmd.ArgDef{
			{
				Name: "Topic",
				Type: dcmd.String,
				Help: "Description of the poll",
			},
			{Name: "Option1", Type: dcmd.String},
			{Name: "Option2", Type: dcmd.String},
			{Name: "Option3", Type: dcmd.String},
			{Name: "Option4", Type: dcmd.String},
			{Name: "Option5", Type: dcmd.String},
			{Name: "Option6", Type: dcmd.String},
			{Name: "Option7", Type: dcmd.String},
			{Name: "Option8", Type: dcmd.String},
			{Name: "Option9", Type: dcmd.String},
			{Name: "Option10", Type: dcmd.String},
		},
		ArgSwitches: []*dcmd.ArgDef{
			{Name: "desc", Type: dcmd.String, Help: "Text in the 'description' field"},
		},
		RunFunc: createPoll,
	}
)

func createPoll(data *dcmd.Data) (interface{}, error) {
	// Helper function to return value of a switch + whether it was set.
	getSwitch := func(key string) (value interface{}, set bool) {
		value = data.Switch(key).Value
		set = value != nil
		return
	}

	topic := data.Args[0].Str()
	options := data.Args[1:]
	for i, option := range options {
		if option.Str() == "" || i >= len(pollReactions) {
			options = options[:i]
			break
		}
	}

	var description string
	if desc, set := getSwitch("desc"); set {
		description = fmt.Sprintf("%s\n\n", desc.(string))
	}

	for i, option := range options {
		if i != 0 {
			description += "\n"
		}
		description += pollReactions[i] + " " + option.Str()
	}

	authorName := data.GuildData.MS.Member.Nick
	if authorName == "" {
		authorName = data.GuildData.MS.User.Username
	}

	response := discordgo.MessageEmbed{
		Title:       topic,
		Description: description,
		Color:       0x65f442,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorName,
			IconURL: discordgo.EndpointUserAvatar(data.GuildData.MS.User.ID, data.Author.Avatar),
		},
	}

	if data.TraditionalTriggerData != nil {
		common.BotSession.ChannelMessageDelete(data.ChannelID, data.TraditionalTriggerData.Message.ID)
	}

	pollMsg, err := common.BotSession.ChannelMessageSendEmbed(data.ChannelID, &response)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to add poll description")
	}
	for i := range options {
		common.BotSession.MessageReactionAdd(pollMsg.ChannelID, pollMsg.ID, pollReactions[i])
	}

	pm := &paginatedmessages.PaginatedMessage{
		MessageID: pollMsg.ID,
	}
	return pm, nil
}
