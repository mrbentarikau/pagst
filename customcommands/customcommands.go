package customcommands

//go:generate sqlboiler --no-hooks psql

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"emperror.dev/errors"
	"github.com/karlseguin/ccache"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/featureflags"
	"github.com/mrbentarikau/pagst/customcommands/models"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/mrbentarikau/pagst/premium"
	"github.com/mrbentarikau/pagst/web"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	RegexCache = *ccache.New(ccache.Configure())
	logger     = common.GetPluginLogger(&Plugin{})
)

// Setting it to 1 Month approx
const (
	MinIntervalTriggerDurationMinutes = 5
	MinIntervalTriggerDurationHours   = 1
	MaxIntervalTriggerDurationHours   = 744
	MaxIntervalTriggerDurationMinutes = 44640

	dbPageMaxDisplayLength = 64
)

func KeyCommands(guildID int64) string { return "custom_commands:" + discordgo.StrID(guildID) }

type Plugin struct{}

func RegisterPlugin() {
	common.InitSchemas("customcommands", DBSchemas...)

	plugin := &Plugin{}
	common.RegisterPlugin(plugin)
}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Custom Commands",
		SysName:  "custom_commands",
		Category: common.PluginCategoryCore,
	}
}

type CommandTriggerType int

const (
	// The ordering of these might seem weird, but they're used in a database so changes would require migrations of a lot of data
	// yeah... i wish i was smarter when i made this originally

	CommandTriggerNone CommandTriggerType = 10

	CommandTriggerCommand    CommandTriggerType = 0
	CommandTriggerStartsWith CommandTriggerType = 1
	CommandTriggerContains   CommandTriggerType = 2
	CommandTriggerRegex      CommandTriggerType = 3
	CommandTriggerExact      CommandTriggerType = 4
	CommandTriggerInterval   CommandTriggerType = 5
	CommandTriggerReaction   CommandTriggerType = 6
)

var (
	AllTriggerTypes = []CommandTriggerType{
		CommandTriggerNone,
		CommandTriggerCommand,
		CommandTriggerStartsWith,
		CommandTriggerContains,
		CommandTriggerRegex,
		CommandTriggerExact,
		CommandTriggerInterval,
		CommandTriggerReaction,
	}

	triggerStrings = map[CommandTriggerType]string{
		CommandTriggerNone:       "None",
		CommandTriggerCommand:    "Command",
		CommandTriggerStartsWith: "StartsWith",
		CommandTriggerContains:   "Contains",
		CommandTriggerRegex:      "Regex",
		CommandTriggerExact:      "Exact",
		CommandTriggerInterval:   "Interval",
		CommandTriggerReaction:   "Reaction",
	}

	embedTriggerStrings = map[CommandTriggerType]string{
		CommandTriggerNone:       "none",
		CommandTriggerCommand:    "comm",
		CommandTriggerStartsWith: "star",
		CommandTriggerContains:   "cont",
		CommandTriggerRegex:      "regx",
		CommandTriggerExact:      "exac",
		CommandTriggerInterval:   "intv",
		CommandTriggerReaction:   "reac",
	}
)

const (
	ReactionModeBoth       = 0
	ReactionModeAddOnly    = 1
	ReactionModeRemoveOnly = 2
)

func (t CommandTriggerType) String() string {
	return triggerStrings[t]
}

func (t CommandTriggerType) EmbedString() string {
	return embedTriggerStrings[t]
}

type CustomCommand struct {
	TriggerType        CommandTriggerType `json:"trigger_type"`
	TriggerTypeForm    string             `json:"-" schema:"type"`
	Trigger            string             `json:"trigger" schema:"trigger" valid:",0,256"`
	RegexTrigger       string             `json:"regex_trigger" schema:"regex_trigger" valid:",0,256"`
	Responses          []string           `json:"responses" schema:"responses" valid:"template,20000"`
	CaseSensitive      bool               `json:"case_sensitive" schema:"case_sensitive"`
	RegexCaseSensitive bool               `json:"regex_case_sensitive" schema:"regex_case_sensitive"`
	ID                 int64              `json:"id"`
	Note               string             `json:"note" schema:"note" valid:",0,256"`
	Disabled           bool               `json:"disabled" schema:"disabled"`

	Public   bool   `json:"public" schema:"public"`
	PublicID string `json:"public_id" schema:"public_id"`

	ContextChannel int64 `schema:"context_channel" valid:"channel,true"`

	TimeTriggerInterval       int     `schema:"time_trigger_interval"`
	TimeTriggerExcludingDays  []int64 `schema:"time_trigger_excluding_days"`
	TimeTriggerExcludingHours []int64 `schema:"time_trigger_excluding_hours"`
	TimeTriggerStartsAt       string  `schema:"time_trigger_starts_at"`

	ReactionTriggerMode int `schema:"reaction_trigger_mode"`

	// If set, then the following categories are required, otherwise they are ignored
	RequireCategories bool    `json:"require_categories" schema:"require_categories"`
	Categories        []int64 `json:"categories" schema:"categories"`

	// If set, then the following channels are required, otherwise they are ignored
	RequireChannels bool    `json:"require_channels" schema:"require_channels"`
	Channels        []int64 `json:"channels" schema:"channels"`

	// If set, then one of the following channels are required, otherwise they are ignored
	RequireRoles  bool    `json:"require_roles" schema:"require_roles"`
	Roles         []int64 `json:"roles" schema:"roles"`
	TriggerOnEdit bool    `json:"trigger_on_edit" schema:"trigger_on_edit"`

	GroupID int64

	ShowErrors       bool `schema:"show_errors"`
	ThreadsEnabled   bool `schema:"threads_enabled"`
	NormalizeUnicode bool `json:"normalize_unicode" schema:"normalize_unicode"`
}

var _ web.CustomValidator = (*CustomCommand)(nil)

func validateCCResponseLength(responses []string, guild_id int64) bool {
	combinedSize := 0
	for _, v := range responses {
		combinedSize += utf8.RuneCountInString(v)
	}

	ccMaxLength := MaxCCResponsesLength
	isGuildPremium, _ := premium.IsGuildPremium(guild_id)
	if isGuildPremium {
		ccMaxLength = MaxCCResponsesLengthPremium
	}

	return combinedSize <= ccMaxLength
}

func (cc *CustomCommand) Validate(tmpl web.TemplateData, guild_id int64) (ok bool) {
	/*if len(cc.Responses) > MaxUserMessages {
		tmpl.AddAlerts(web.ErrorAlert(fmt.Sprintf("Too many responses, max %d", MaxUserMessages)))
		return false
	}*/

	foundOkayResponse := false
	for _, v := range cc.Responses {
		if strings.TrimSpace(v) != "" {
			foundOkayResponse = true
			break
		}
	}

	if !foundOkayResponse {
		tmpl.AddAlerts(web.ErrorAlert("No response set"))
		return false
	}

	isValidCCLength := validateCCResponseLength(cc.Responses, guild_id)

	if !cc.Disabled && !isValidCCLength {
		tmpl.AddAlerts(web.ErrorAlert("Max combined command size can be 10k for free servers, and 20k for premium servers"))
		return false
	}

	if cc.TriggerTypeForm == "interval_minutes" && cc.TimeTriggerInterval < 1 {
		tmpl.AddAlerts(web.ErrorAlert("Minimum interval is 1 minute..."))
		return false
	}

	// check max interval limits
	var intvMax bool
	durLimitHours := 2560000 // for 292 years
	intvMult := 1
	if cc.TriggerTypeForm == "interval_hours" {
		intvMult = 60
		if cc.TimeTriggerInterval/60 > durLimitHours {
			intvMax = true
		}
	} else if cc.TimeTriggerInterval > durLimitHours*60 {
		intvMax = true
	}

	if (cc.TriggerTypeForm == "interval_minutes" || cc.TriggerTypeForm == "interval_hours") && (time.Minute*time.Duration(cc.TimeTriggerInterval*intvMult) < 0 || intvMax) {
		tmpl.AddAlerts(web.ErrorAlert(fmt.Sprintf("Interval %d goes beyond limits of negative or 292 years...", cc.TimeTriggerInterval)))
		return false
	}

	return true
}

func (cc *CustomCommand) ToDBModel() *models.CustomCommand {
	pqCommand := &models.CustomCommand{
		TriggerType:               int(cc.TriggerType),
		RegexTrigger:              cc.RegexTrigger,
		RegexTriggerCaseSensitive: cc.RegexCaseSensitive,
		TextTrigger:               cc.Trigger,
		TextTriggerCaseSensitive:  cc.CaseSensitive,

		Public:   cc.Public,
		PublicID: cc.PublicID,

		Categories:              cc.Categories,
		CategoriesWhitelistMode: cc.RequireCategories,
		Channels:                cc.Channels,
		ChannelsWhitelistMode:   cc.RequireChannels,
		Roles:                   cc.Roles,
		RolesWhitelistMode:      cc.RequireRoles,

		TimeTriggerInterval:       cc.TimeTriggerInterval,
		TimeTriggerExcludingDays:  cc.TimeTriggerExcludingDays,
		TimeTriggerExcludingHours: cc.TimeTriggerExcludingHours,
		ContextChannel:            cc.ContextChannel,

		ReactionTriggerMode: int16(cc.ReactionTriggerMode),

		Responses: cc.Responses,

		ShowErrors:    cc.ShowErrors,
		Disabled:      !cc.Disabled,
		TriggerOnEdit: cc.TriggerOnEdit,

		ThreadsEnabled:   cc.ThreadsEnabled,
		NormalizeUnicode: cc.NormalizeUnicode,

		DateUpdated: null.TimeFrom(time.Now()),
	}

	if cc.TimeTriggerExcludingDays == nil {
		pqCommand.TimeTriggerExcludingDays = []int64{}
	}

	if cc.TimeTriggerExcludingHours == nil {
		pqCommand.TimeTriggerExcludingHours = []int64{}
	}

	if cc.GroupID != 0 {
		pqCommand.GroupID = null.Int64From(cc.GroupID)
	}

	if cc.Note != "" {
		pqCommand.Note = null.StringFrom(cc.Note)
	} else {
		pqCommand.Note = null.NewString("", false)
	}

	if cc.TriggerTypeForm == "interval_hours" {
		pqCommand.TimeTriggerInterval *= 60
	}

	return pqCommand
}

func CmdRunsInCategory(cc *models.CustomCommand, parentChannel int64) bool {
	gs := bot.State.GetGuild(cc.GuildID)
	cs := gs.GetChannelOrThread(parentChannel)
	threadChannelParent := int64(0)
	if cs != nil {
		threadChannelParent = cs.ParentID
	}

	if cc.GroupID.Valid {
		// check group restrictions
		if common.ContainsInt64Slice(cc.R.Group.IgnoreCategories, parentChannel) {
			return false
		}

		if len(cc.R.Group.WhitelistCategories) > 0 {
			if !common.ContainsInt64Slice(cc.R.Group.WhitelistCategories, parentChannel) {
				return false
			}
		}
	}

	// check command specific restrictions
	for _, v := range cc.Categories {
		if v == parentChannel {
			return cc.CategoriesWhitelistMode
		}

		if threadChannelParent != 0 && v == threadChannelParent {
			return cc.CategoriesWhitelistMode
		}
	}

	// Not found
	return !cc.CategoriesWhitelistMode
}

func CmdRunsInChannel(cc *models.CustomCommand, channel int64) bool {
	gs := bot.State.GetGuild(cc.GuildID)
	cs := gs.GetChannelOrThread(channel)
	if cc.GroupID.Valid {
		// check group restrictions
		if common.ContainsInt64Slice(cc.R.Group.IgnoreChannels, channel) {
			return false
		}

		if len(cc.R.Group.WhitelistChannels) > 0 {
			if !common.ContainsInt64Slice(cc.R.Group.WhitelistChannels, channel) {
				return false
			}
		}
	}

	// check command specific restrictions
	for _, v := range cc.Channels {
		if v == channel {
			return cc.ChannelsWhitelistMode
		}
	}

	if cs != nil && cs.Type.IsThread() {
		isContained := common.ContainsInt64Slice(cc.Channels, common.ChannelOrThreadParentID(cs))
		if cc.ChannelsWhitelistMode && !isContained {
			return false
		}
		if !cc.ChannelsWhitelistMode && isContained {
			return false
		}
		return cc.ThreadsEnabled
	}

	// Not found
	return !cc.ChannelsWhitelistMode
}

func CmdRunsForUser(cc *models.CustomCommand, ms *dstate.MemberState) bool {
	if cc.GroupID.Valid {
		// check group restrictions
		if common.ContainsInt64SliceOneOf(cc.R.Group.IgnoreRoles, ms.Member.Roles) {
			return false
		}

		if len(cc.R.Group.WhitelistRoles) > 0 && !common.ContainsInt64SliceOneOf(cc.R.Group.WhitelistRoles, ms.Member.Roles) {
			return false
		}
	}

	// check command specific restrictions
	if len(cc.Roles) == 0 {
		// Fast path
		return !cc.RolesWhitelistMode
	}

	for _, v := range cc.Roles {
		if common.ContainsInt64Slice(ms.Member.Roles, v) {
			return cc.RolesWhitelistMode
		}
	}

	// Not found
	return !cc.RolesWhitelistMode
}

type CustomCommandSlice []*CustomCommand

// Len is the number of elements in the collection.
func (c CustomCommandSlice) Len() int {
	return len(c)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (c CustomCommandSlice) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}

// Swap swaps the elements with indexes i and j.
func (c CustomCommandSlice) Swap(i, j int) {
	temp := c[i]
	c[i] = c[j]
	c[j] = temp
}

func filterEmptyResponses(s string, ss ...string) []string {
	result := make([]string, 0, len(ss)+1)
	if s != "" {
		result = append(result, s)
	}

	for _, s := range ss {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

const (
	MaxCCResponsesLength        = 10000
	MaxCCResponsesLengthPremium = 20000
	MaxCommands                 = 100
	MaxCommandsPremium          = 250
	MaxGroups                   = 50
	MaxUserMessages             = 20
)

func MaxCommandsForContext(ctx context.Context) int {
	if premium.ContextPremium(ctx) {
		return MaxCommandsPremium
	}

	return MaxCommands
}

var _ featureflags.PluginWithFeatureFlags = (*Plugin)(nil)

const (
	featureFlagHasCommands = "custom_commands_has_commands"
)

func (p *Plugin) UpdateFeatureFlags(guildID int64) ([]string, error) {

	var flags []string
	count, err := models.CustomCommands(qm.Where("guild_id = ?", guildID)).CountG(context.Background())
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	if count > 0 {
		flags = append(flags, featureFlagHasCommands)
	}

	return flags, nil
}

func (p *Plugin) AllFeatureFlags() []string {
	return []string{
		featureFlagHasCommands, // set if this server has any custom commands at all
	}
}

func getDatabaseEntries(ctx context.Context, guildID int64, page int, queryType, query string, limit int) (models.TemplatesUserDatabaseSlice, int64, error) {
	qms := []qm.QueryMod{
		models.TemplatesUserDatabaseWhere.GuildID.EQ(guildID),
	}

	if len(query) > 0 {
		switch queryType {
		case "id":
			qms = append(qms, qm.Where("id = ?", query))
		case "user_id":
			qms = append(qms, qm.Where("user_id = ?", query))
		case "key":
			qms = append(qms, qm.Where("key ILIKE ?", query))
		}
	}

	count, err := models.TemplatesUserDatabases(qms...).CountG(ctx)
	if int64(page) > (count / 100) {
		page = int(math.Ceil(float64(count) / 100))
	}

	if page > 1 {
		qms = append(qms, qm.Offset((limit * (page - 1))))
	}
	if err != nil {
		return nil, 0, err
	}

	qms = append(qms, qm.OrderBy("id desc"), qm.Limit(limit))
	entries, err := models.TemplatesUserDatabases(qms...).AllG(ctx)
	return entries, count, err
}

func convertEntries(result models.TemplatesUserDatabaseSlice) []*LightDBEntry {
	entries := make([]*LightDBEntry, 0, len(result))
	for _, v := range result {
		converted, err := ToLightDBEntry(v)
		if err != nil {
			logger.WithError(err).Warn("[cc/web] failed converting to light db entry")
			continue
		}

		b, err := json.Marshal(converted.Value)
		if err != nil {
			logger.WithError(err).Warn("[cc/web] failed converting to light db entry")
			continue
		}

		converted.Value = common.CutStringShort(string(b), dbPageMaxDisplayLength)

		entries = append(entries, converted)
	}

	return entries
}
