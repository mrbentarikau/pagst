package automod

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/mrbentarikau/pagst/antiphishing"
	"github.com/mrbentarikau/pagst/automod/models"
	"github.com/mrbentarikau/pagst/automod_basic"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/mrbentarikau/pagst/safebrowsing"
)

// var tChain = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var forwardSlashReplacer = strings.NewReplacer("\\", "")

/////////////////////////////////////////////////////////////

type BaseRegexTriggerData struct {
	Regex            string `valid:",1,250"`
	NormalizeUnicode bool
}

type BaseRegexTrigger struct {
	Inverse bool
}

func (r BaseRegexTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (r BaseRegexTrigger) DataType() interface{} {
	return &BaseRegexTriggerData{}
}

func (r BaseRegexTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "Regex",
			Key:  "Regex",
			Kind: SettingTypeString,
			Min:  1,
			Max:  250,
		},
		{
			Name:    "Normalize Unicode accents & confusables",
			Key:     "NormalizeUnicode",
			Kind:    SettingTypeBool,
			Default: false,
		},
	}
}

//////////////

type MentionsTriggerData struct {
	Threshold int
}

var _ MessageTrigger = (*MentionsTrigger)(nil)

type MentionsTrigger struct{}

func (mc *MentionsTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (mc *MentionsTrigger) DataType() interface{} {
	return &MentionsTriggerData{}
}

func (mc *MentionsTrigger) Name() string {
	return "Message mentions"
}

func (mc *MentionsTrigger) Description() string {
	return "Triggers when a message includes x or more unique mentions."
}

func (mc *MentionsTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name:    "Threshold",
			Key:     "Threshold",
			Kind:    SettingTypeInt,
			Default: 4,
		},
	}
}

func (mc *MentionsTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*MentionsTriggerData)
	if len(m.Mentions) >= dataCast.Threshold {
		return true, nil
	}

	return false, nil
}

func (mc *MentionsTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*AnyLinkTrigger)(nil)

type AnyLinkTrigger struct{}

func (alc *AnyLinkTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (alc *AnyLinkTrigger) DataType() interface{} {
	return nil
}

func (alc *AnyLinkTrigger) Name() (name string) {
	return "Any Link"
}

func (alc *AnyLinkTrigger) Description() (description string) {
	return "Triggers when a message contains any valid link"
}

func (alc *AnyLinkTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (alc *AnyLinkTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	if common.LinkRegex.MatchString(forwardSlashReplacer.Replace(m.Content)) {
		return true, nil
	}

	return false, nil
}

func (alc *AnyLinkTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*WordListTrigger)(nil)

type WordListTrigger struct {
	Blacklist bool
}
type WorldListTriggerData struct {
	ListID           int64
	NormalizeUnicode bool
}

func (wl *WordListTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (wl *WordListTrigger) DataType() interface{} {
	return &WorldListTriggerData{}
}

func (wl *WordListTrigger) Name() (name string) {
	if wl.Blacklist {
		return "Word blacklist"
	}

	return "Word whitelist"
}

func (wl *WordListTrigger) Description() (description string) {
	if wl.Blacklist {
		return "Triggers on messages containing words in the specified list"
	}

	return "Triggers on messages containing words not in the specified list"
}

func (wl *WordListTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "List",
			Key:  "ListID",
			Kind: SettingTypeList,
		},
		{
			Name:    "Normalize Unicode accents & confusables",
			Key:     "NormalizeUnicode",
			Kind:    SettingTypeBool,
			Default: false,
		},
	}
}

func (wl *WordListTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*WorldListTriggerData)

	list, err := FindFetchGuildList(triggerCtx.GS.ID, dataCast.ListID)
	if err != nil {
		return false, nil
	}

	messageFields := strings.Fields(mdStripped)

	for _, mf := range messageFields {
		if dataCast.NormalizeUnicode {
			mf = common.NormalizeAccents(mf)
			mf = common.NormalizeConfusables(mf)
		}

		contained := false
		for _, w := range list.Content {
			if strings.EqualFold(mf, w) {
				if wl.Blacklist {
					// contains a blacklisted word, trigger
					return true, nil
				} else {
					contained = true
					break
				}
			}
		}

		if !wl.Blacklist && !contained {
			// word not whitelisted, trigger
			return true, nil
		}
	}

	// did not contain a blacklisted word, or contained just whitelisted words
	return false, nil
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*DomainTrigger)(nil)

type DomainTrigger struct {
	Blacklist bool
}
type DomainTriggerData struct {
	ListID int64
}

func (dt *DomainTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (dt *DomainTrigger) DataType() interface{} {
	return &DomainTriggerData{}
}

func (dt *DomainTrigger) Name() (name string) {
	if dt.Blacklist {
		return "Website blacklist"
	}

	return "Website whitelist"
}

func (dt *DomainTrigger) Description() (description string) {
	if dt.Blacklist {
		return "Triggers on messages containing links to websites in the specified list"
	}

	return "Triggers on messages containing links to websites NOT in the specified list"
}

func (dt *DomainTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "List",
			Key:  "ListID",
			Kind: SettingTypeList,
		},
	}
}

func (dt *DomainTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*DomainTriggerData)

	list, err := FindFetchGuildList(triggerCtx.GS.ID, dataCast.ListID)
	if err != nil {
		return false, nil
	}

	matches := common.LinkRegex.FindAllString(forwardSlashReplacer.Replace(m.Content), -1)

	for _, v := range matches {
		if contains, _ := dt.containsDomain(v, list.Content); contains {
			if dt.Blacklist {
				return true, nil
			}
		} else if !dt.Blacklist {
			// whitelist mode, unknown link
			return true, nil
		}

	}

	// did not contain any link, or no blacklisted links
	return false, nil
}

func (dt *DomainTrigger) containsDomain(link string, list []string) (bool, string) {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") && !strings.HasPrefix(link, "steam://") {
		link = "http://" + link
	}

	parsed, err := url.ParseRequestURI(link)
	if err != nil {
		logger.WithError(err).WithField("url", link).Error("Failed parsing request url matched with regex")
		return false, ""
	}

	host := parsed.Host
	if index := strings.Index(host, ":"); index > -1 {
		host = host[:index]
	}

	host = strings.ToLower(host)

	for _, v := range list {
		if strings.HasSuffix(host, "."+v) {
			return true, v
		}

		if v == host {
			return true, v
		}
	}

	return false, ""
}

/////////////////////////////////////////////////////////////

type ViolationsTriggerData struct {
	Name           string `valid:",1,100,trimspace"`
	Threshold      int
	Interval       int
	IgnoreIfLesser bool
}

var _ ViolationListener = (*ViolationsTrigger)(nil)

type ViolationsTrigger struct{}

func (vt *ViolationsTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (vt *ViolationsTrigger) DataType() interface{} {
	return &ViolationsTriggerData{}
}

func (vt *ViolationsTrigger) Name() string {
	return "x Violations in y minutes"
}

func (vt *ViolationsTrigger) Description() string {
	return "Triggers when a user has x or more violations within y minutes."
}

func (vt *ViolationsTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name:    "Violation name",
			Key:     "Name",
			Kind:    SettingTypeString,
			Default: "name",
			Min:     1,
			Max:     50,
		},
		{
			Name:    "Number of violations",
			Key:     "Threshold",
			Kind:    SettingTypeInt,
			Default: 4,
		},
		{
			Name:    "Within (minutes)",
			Key:     "Interval",
			Kind:    SettingTypeInt,
			Default: 60,
		},
		{
			Name:    "Ignore if a higher violation trigger of this name was activated",
			Key:     "IgnoreIfLesser",
			Kind:    SettingTypeBool,
			Default: true,
		},
	}
}

func (vt *ViolationsTrigger) CheckUser(ctxData *TriggeredRuleData, violations []*models.AutomodViolation, settings interface{}, triggeredOnHigher bool) (isAffected bool, err error) {
	settingsCast := settings.(*ViolationsTriggerData)
	if triggeredOnHigher && settingsCast.IgnoreIfLesser {
		return false, nil
	}

	numRecent := 0
	for _, v := range violations {
		if v.Name != settingsCast.Name {
			continue
		}

		if time.Since(v.CreatedAt).Minutes() > float64(settingsCast.Interval) {
			continue
		}

		numRecent++
	}

	if numRecent >= settingsCast.Threshold {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

type AllCapsTriggerData struct {
	MinLength  int
	Percentage int
}

var _ MessageTrigger = (*AllCapsTrigger)(nil)

type AllCapsTrigger struct{}

func (caps *AllCapsTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (caps *AllCapsTrigger) DataType() interface{} {
	return &AllCapsTriggerData{}
}

func (caps *AllCapsTrigger) Name() string {
	return "All Caps"
}

func (caps *AllCapsTrigger) Description() string {
	return "Triggers when a message contains x% or more of just capitalized letters"
}

func (caps *AllCapsTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name:    "Min number of all caps",
			Key:     "MinLength",
			Kind:    SettingTypeInt,
			Default: 3,
		},
		{
			Name:    "Percentage of all caps",
			Key:     "Percentage",
			Kind:    SettingTypeInt,
			Default: 100,
			Min:     1,
			Max:     100,
		},
	}
}

func (caps *AllCapsTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*AllCapsTriggerData)

	if len(m.Content) < dataCast.MinLength {
		return false, nil
	}

	totalCapitalisableChars := 0
	numCaps := 0

	// count the number of upper case characters, note that this dosen't include other characters such as punctuation
	for _, r := range m.Content {
		if unicode.IsUpper(r) {
			numCaps++
			totalCapitalisableChars++
		} else {
			if unicode.ToUpper(r) != unicode.ToLower(r) {
				totalCapitalisableChars++
			}
		}
	}

	if totalCapitalisableChars < 1 {
		return false, nil
	}

	percentage := (numCaps * 100) / (totalCapitalisableChars)
	if numCaps >= dataCast.MinLength && percentage >= dataCast.Percentage {
		return true, nil
	}

	return false, nil
}

func (caps *AllCapsTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*ServerInviteTrigger)(nil)

type ServerInviteTrigger struct{}

func (inv *ServerInviteTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (inv *ServerInviteTrigger) DataType() interface{} {
	return nil
}

func (inv *ServerInviteTrigger) Name() string {
	return "Server invites"
}

func (inv *ServerInviteTrigger) Description() string {
	return "Triggers on messages containing invites to other servers, also includes some 3rd party server lists."
}

func (inv *ServerInviteTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (inv *ServerInviteTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	containsBadInvited := automod_basic.CheckMessageForBadInvites(m.Content, m.GuildID)
	return containsBadInvited, nil
}

func (inv *ServerInviteTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*AntiPhishingLinkTrigger)(nil)

type AntiPhishingLinkTrigger struct{}

func (a *AntiPhishingLinkTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (a *AntiPhishingLinkTrigger) Name() string {
	return "Anti-Fish API flagged bad links"
}

func (a *AntiPhishingLinkTrigger) DataType() interface{} {
	return nil
}

func (a *AntiPhishingLinkTrigger) Description() string {
	return "Triggers on messages that have scam links flagged by SinkingYachts and BitFlow AntiPhishing APIs and uses Google's Transparency Report as last resort."
}

func (a *AntiPhishingLinkTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (a *AntiPhishingLinkTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	badDomain, err := antiphishing.CheckMessageForPhishingDomains(forwardSlashReplacer.Replace(m.Content))
	if err != nil {
		logger.WithError(err).Error("Failed to check url ")
		return false, nil
	}

	if badDomain != "" {
		return true, nil
	}

	/*
		matches := common.LinkRegexJonas.FindAllString(forwardSlashReplacer.Replace(m.Content), -1)

		for _, v := range matches {
			trasparencyReport, err := common.TransparencyReportQuery(v)
			if err != nil {
				logger.WithError(err).Error("Failed checking URLs from Google's Transparency Report API.")
				return false, nil
			}

			if trasparencyReport.UnsafeContent == 2 { // || trasparencyReport.ScoreTotal >= 2
				return true, nil
			}

		}
	*/

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*GoogleSafeBrowsingTrigger)(nil)

type GoogleSafeBrowsingTrigger struct{}

func (g *GoogleSafeBrowsingTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (g *GoogleSafeBrowsingTrigger) DataType() interface{} {
	return nil
}

func (g *GoogleSafeBrowsingTrigger) Name() string {
	return "Google flagged bad links"
}

func (g *GoogleSafeBrowsingTrigger) Description() string {
	return "Triggers on messages containing links that are flagged by Google Safebrowsing as unsafe."
}

func (g *GoogleSafeBrowsingTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (g *GoogleSafeBrowsingTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	threat, err := safebrowsing.CheckString(forwardSlashReplacer.Replace(m.Content))
	if err != nil {
		logger.WithError(err).Error("Failed checking urls against google safebrowser")
		return false, nil
	}

	if threat != nil {
		return true, nil
	}

	return false, nil
}

func (g *GoogleSafeBrowsingTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

type SlowmodeTriggerData struct {
	Threshold                int
	Interval                 int
	SingleMessageAttachments bool
	SingleMessageLinks       bool
}

var _ MessageTrigger = (*SlowmodeTrigger)(nil)

type SlowmodeTrigger struct {
	ChannelBased bool
	Attachments  bool // whether this trigger checks any messages or just attachments
	Links        bool // whether this trigger checks any messages or just links
	Stickers     bool
}

func (s *SlowmodeTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (s *SlowmodeTrigger) DataType() interface{} {
	return &SlowmodeTriggerData{}
}

func (s *SlowmodeTrigger) Name() string {
	if s.ChannelBased {
		if s.Attachments {
			return "x channel attachments in y seconds"
		}

		if s.Links {
			return "x channel links in y seconds"
		}

		if s.Stickers {
			return "x channel stickers in y seconds"
		}

		return "x channel messages in y seconds"
	}

	if s.Attachments {
		return "x user attachments in y seconds"
	}

	if s.Links {
		return "x user links in y seconds"
	}

	if s.Stickers {
		return "x user stickers in y seconds"
	}

	return "x user messages in y seconds"
}

func (s *SlowmodeTrigger) Description() string {
	if s.ChannelBased {
		if s.Attachments {
			return "Triggers when a channel has x or more attachments within y seconds"
		}

		if s.Links {
			return "Triggers when a channel has x or more links within y seconds"
		}

		if s.Stickers {
			return "Triggers when a channel has x or more stickers within y seconds"
		}

		return "Triggers when a channel has x or more messages in y seconds."
	}

	if s.Attachments {
		return "Triggers when a user has x or more attachments within y seconds in a single channel"
	}

	if s.Links {
		return "Triggers when a user has x or more links within y seconds in a single channel"
	}

	if s.Stickers {
		return "Triggers when a user has x or more stickers within y seconds"
	}

	return "Triggers when a user has x or more messages in y seconds in a single channel."
}

func (s *SlowmodeTrigger) UserSettings() []*SettingDef {
	defaultMessages := 5
	defaultInterval := 5
	thresholdName := "Messages"

	if s.Attachments {
		defaultMessages = 10
		defaultInterval = 60
		thresholdName = "Attachments"
	} else if s.Links {
		defaultInterval = 60
		thresholdName = "Links"
	}

	settings := []*SettingDef{
		{
			Name:    thresholdName,
			Key:     "Threshold",
			Kind:    SettingTypeInt,
			Default: defaultMessages,
		},
		{
			Name:    "Within (seconds)",
			Key:     "Interval",
			Kind:    SettingTypeInt,
			Default: defaultInterval,
		},
	}

	if s.Attachments {
		settings = append(settings, &SettingDef{
			Name:    "Also count multiple attachments in single message",
			Key:     "SingleMessageAttachments",
			Kind:    SettingTypeBool,
			Default: false,
		})
	} else if s.Links {
		settings = append(settings, &SettingDef{
			Name:    "Also count multiple links in single message",
			Key:     "SingleMessageLinks",
			Kind:    SettingTypeBool,
			Default: false,
		})
	}

	return settings
}

func (s *SlowmodeTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	if s.Attachments && len(m.Attachments) < 1 {
		return false, nil
	}

	if s.Links && !common.LinkRegex.MatchString(forwardSlashReplacer.Replace(m.Content)) {
		return false, nil
	}

	if s.Stickers && len(m.StickerItems) < 1 {
		return false, nil
	}

	settings := triggerCtx.Data.(*SlowmodeTriggerData)

	within := time.Duration(settings.Interval) * time.Second
	now := time.Now()

	amount := 0

	messages := bot.State.GetMessages(cs.GuildID, cs.ID, &dstate.MessagesQuery{
		Limit: 1000,
	})

	// New messages are at the end
	for _, v := range messages {
		age := now.Sub(v.ParsedCreatedAt)
		if age > within {
			break
		}

		if !s.ChannelBased && v.Author.ID != triggerCtx.MS.User.ID {
			continue
		}

		if s.Attachments {
			if len(v.Attachments) < 1 {
				continue // we're only checking messages with attachments
			}
			if settings.SingleMessageAttachments {
				// Add the count of all attachments of this message to the amount
				amount += len(v.Attachments)
				continue
			}
		}

		if s.Links {
			linksLen := len(common.LinkRegexJonas.FindAllString(forwardSlashReplacer.Replace(v.Content), -1))
			if linksLen < 1 {
				continue // we're only checking messages with links
			}
			if settings.SingleMessageLinks {
				// Add the count of all links of this message to the amount
				amount += linksLen
				continue
			}
		}

		if s.Stickers && len(v.Stickers) < 1 {
			continue // were only checking messages with stickers
		}

		amount++
	}

	if amount >= settings.Threshold {
		return true, nil
	}

	return false, nil
}

func (s *SlowmodeTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

type MultiMsgMentionTriggerData struct {
	Threshold       int
	Interval        int
	CountDuplicates bool
	IgnoreReplies   bool
}

var _ MessageTrigger = (*MultiMsgMentionTrigger)(nil)

type MultiMsgMentionTrigger struct {
	ChannelBased bool
}

func (mt *MultiMsgMentionTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (mt *MultiMsgMentionTrigger) DataType() interface{} {
	return &MultiMsgMentionTriggerData{}
}

func (mt *MultiMsgMentionTrigger) Name() string {
	if mt.ChannelBased {
		return "channel: x mentions within y seconds"
	}

	return "user: x mentions within y seconds"
}

func (mt *MultiMsgMentionTrigger) Description() string {
	if mt.ChannelBased {
		return "Triggers when a channel has x or more unique mentions in y seconds"
	}

	return "Triggers when a user has sent x or more unique mentions in y seconds in a single channel"
}

func (mt *MultiMsgMentionTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name:    "Mentions",
			Key:     "Threshold",
			Kind:    SettingTypeInt,
			Default: 20,
		},
		{
			Name:    "Within (seconds)",
			Key:     "Interval",
			Kind:    SettingTypeInt,
			Default: 10,
		},
		{
			Name: "Count multiple mentions to the same user",
			Key:  "CountDuplicates",
			Kind: SettingTypeBool,
		},
		{
			Name: "Ignore reply mentions",
			Key:  "IgnoreReplies",
			Kind: SettingTypeBool,
		},
	}
}

func (mt *MultiMsgMentionTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	if len(m.Mentions) < 1 {
		return false, nil
	}

	settings := triggerCtx.Data.(*MultiMsgMentionTriggerData)

	within := time.Duration(settings.Interval) * time.Second
	now := time.Now()

	mentions := make([]int64, 0)

	messages := bot.State.GetMessages(cs.GuildID, cs.ID, &dstate.MessagesQuery{
		Limit: 1000,
	})

	// New messages are at the end
	for _, v := range messages {
		age := now.Sub(v.ParsedCreatedAt)
		if age > within {
			break
		}

		if settings.IgnoreReplies && v.ReferencedMessage != nil {
			continue
		}

		if mt.ChannelBased || v.Author.ID == triggerCtx.MS.User.ID {
			// we only care about unique mentions, e.g mentioning the same user a ton wont do anything
			for _, msgMention := range v.Mentions {
				if settings.CountDuplicates || !common.ContainsInt64Slice(mentions, msgMention.ID) {
					mentions = append(mentions, msgMention.ID)
				}
			}
		}
		if len(mentions) >= settings.Threshold {
			return true, nil
		}
	}

	if len(mentions) >= settings.Threshold {
		return true, nil
	}

	return false, nil
}

func (mt *MultiMsgMentionTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*MessageRegexTrigger)(nil)

type MessageRegexTrigger struct {
	BaseRegexTrigger
}

func (r *MessageRegexTrigger) Name() string {
	if r.BaseRegexTrigger.Inverse {
		return "Message not matching regex"
	}

	return "Message matches regex"
}

func (r *MessageRegexTrigger) Description() string {
	if r.BaseRegexTrigger.Inverse {
		return "Triggers when a message does not match the provided regex"
	}

	return "Triggers when a message matches the provided regex"
}

func (r *MessageRegexTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*BaseRegexTriggerData)

	item, err := RegexCache.Fetch(dataCast.Regex, time.Minute*10, func() (interface{}, error) {
		re, err := regexp.Compile(dataCast.Regex)
		if err != nil {
			return nil, err
		}

		return re, nil
	})

	if err != nil {
		return false, nil
	}

	re := item.Value().(*regexp.Regexp)

	mContent := m.Content
	if dataCast.NormalizeUnicode {
		mContent = common.NormalizeAccents(mContent)
		mContent = common.NormalizeConfusables(mContent)
	}

	if re.MatchString(mContent) {
		if r.BaseRegexTrigger.Inverse {
			return false, nil
		}
		return true, nil
	}

	if r.BaseRegexTrigger.Inverse {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

type SpamTriggerData struct {
	Threshold int
	TimeLimit int
}

var _ MessageTrigger = (*SpamTrigger)(nil)

type SpamTrigger struct{}

func (spam *SpamTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (spam *SpamTrigger) DataType() interface{} {
	return &SpamTriggerData{}
}

func (spam *SpamTrigger) Name() string {
	return "x consecutive identical messages"
}

func (spam *SpamTrigger) Description() string {
	return "Triggers when a user sends x identical messages after each other"
}

func (spam *SpamTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name:    "Threshold",
			Key:     "Threshold",
			Kind:    SettingTypeInt,
			Min:     1,
			Max:     250,
			Default: 4,
		},
		{
			Name:    "Within seconds (0 = infinity)",
			Key:     "TimeLimit",
			Kind:    SettingTypeInt,
			Min:     0,
			Max:     10000,
			Default: 30,
		},
	}
}

func (spam *SpamTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {

	settingsCast := triggerCtx.Data.(*SpamTriggerData)

	mToCheckAgainst := strings.TrimSpace(strings.ToLower(m.Content))

	count := 1

	timeLimit := time.Now().Add(-time.Second * time.Duration(settingsCast.TimeLimit))

	messages := bot.State.GetMessages(cs.GuildID, cs.ID, &dstate.MessagesQuery{
		Limit: 1000,
	})

	for _, v := range messages {
		if v.ID == m.ID {
			continue
		}

		if v.Author.ID != m.Author.ID {
			continue
		}

		if settingsCast.TimeLimit > 0 && timeLimit.After(v.ParsedCreatedAt) {
			// if this message was created before the time limit, then break out
			break
		}

		if len(v.Attachments) > 0 {
			break // treat any attachment as a different message, in the future i may download them and check hash or something? maybe too much
		}

		if strings.ToLower(strings.TrimSpace(v.Content)) == mToCheckAgainst {
			count++
		} else {
			break
		}
	}

	if count >= settingsCast.Threshold {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ NameListener = (*NameRegexTrigger)(nil)

type NameRegexTrigger struct {
	Inverse bool
}

type NameRegexTriggerData struct {
	Regex            string `valid:",1,250"`
	ExcludeUsername  bool
	NormalizeUnicode bool
}

func (nr NameRegexTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (nr NameRegexTrigger) DataType() interface{} {
	return &NameRegexTriggerData{}
}

func (nr *NameRegexTrigger) Name() string {
	if nr.Inverse {
		return "Name/Nick not matching regex"
	}

	return "Name/Nick matches regex"
}

func (nr *NameRegexTrigger) Description() string {
	if nr.Inverse {
		return "Triggers when a member's name does not match the provided regex"
	}

	return "Triggers when a member's name matches the provided regex"
}

func (nr *NameRegexTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "Regex",
			Key:  "Regex",
			Kind: SettingTypeString,
			Min:  1,
			Max:  250,
		},
		{
			Name:    "Exclude username",
			Key:     "ExcludeUsername",
			Kind:    SettingTypeBool,
			Default: false,
		},
		{
			Name:    "Normalize Unicode accents & confusables",
			Key:     "NormalizeUnicode",
			Kind:    SettingTypeBool,
			Default: false,
		},
	}
}

func (nr *NameRegexTrigger) CheckName(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*NameRegexTriggerData)

	item, err := RegexCache.Fetch(dataCast.Regex, time.Minute*10, func() (interface{}, error) {
		re, err := regexp.Compile(dataCast.Regex)
		if err != nil {
			return nil, err
		}

		return re, nil
	})

	if err != nil {
		return false, nil
	}

	username := t.MS.User.Username
	nickname := t.MS.Member.Nick

	if dataCast.NormalizeUnicode {
		username = common.NormalizeAccents(username)
		nickname = common.NormalizeAccents(nickname)
		username = common.NormalizeConfusables(username)
		nickname = common.NormalizeConfusables(nickname)
	}

	if dataCast.ExcludeUsername {
		username = ""
	}

	re := item.Value().(*regexp.Regexp)
	if re.MatchString(username) || re.MatchString(nickname) {
		if nr.Inverse {
			return false, nil
		}
		return true, nil
	}

	if nr.Inverse {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ NameListener = (*NameWordlistTrigger)(nil)

type NameWordlistTrigger struct {
	Blacklist bool
}
type NameWordlistTriggerData struct {
	ListID           int64
	ExcludeUsername  bool
	NormalizeUnicode bool
}

func (nwl *NameWordlistTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (nwl *NameWordlistTrigger) DataType() interface{} {
	return &NameWordlistTriggerData{}
}

func (nwl *NameWordlistTrigger) Name() (name string) {
	if nwl.Blacklist {
		return "Name/Nick word blacklist"
	}

	return "Name/Nick word whitelist"
}

func (nwl *NameWordlistTrigger) Description() (description string) {
	if nwl.Blacklist {
		return "Triggers when a member has a name containing words in the specified list, this is currently very easy to circumvent atm, and will likely be improved in the future."
	}

	return "Triggers when a member has a name containing words not in the specified list, this is currently very easy to circumvent atm, and will likely be improved in the future."
}

func (nwl *NameWordlistTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "List",
			Key:  "ListID",
			Kind: SettingTypeList,
		},
		{
			Name:    "Exclude username",
			Key:     "ExcludeUsername",
			Kind:    SettingTypeBool,
			Default: false,
		},
		{
			Name:    "Normalize Unicode accents & confusables",
			Key:     "NormalizeUnicode",
			Kind:    SettingTypeBool,
			Default: false,
		},
	}
}

func (nwl *NameWordlistTrigger) CheckName(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*NameWordlistTriggerData)

	list, err := FindFetchGuildList(t.GS.ID, dataCast.ListID)
	if err != nil {
		return false, nil
	}

	uNameFields := strings.Fields(PrepareMessageForWordCheck(t.MS.User.Username))
	nNameFields := strings.Fields(PrepareMessageForWordCheck(t.MS.Member.Nick))

	if dataCast.ExcludeUsername {
		uNameFields = []string{}
	}

	fields := append(nNameFields, uNameFields...)

	var contained bool
	for _, mf := range fields {
		if dataCast.NormalizeUnicode {
			mf = common.NormalizeAccents(mf)
			mf = common.NormalizeConfusables(mf)
		}

		for _, w := range list.Content {
			if strings.EqualFold(mf, w) {
				if nwl.Blacklist {
					// contains a blacklisted word, trigger
					return true, nil
				} else {
					contained = true
					break
				}
			}
		}
	}

	if !nwl.Blacklist && !contained {
		// word not whitelisted, trigger
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ UsernameListener = (*UsernameRegexTrigger)(nil)

type UsernameRegexTrigger struct {
	BaseRegexTrigger
}

func (r *UsernameRegexTrigger) Name() string {
	if r.BaseRegexTrigger.Inverse {
		return "Join username not matching regex"
	}

	return "Join username matches regex"
}

func (r *UsernameRegexTrigger) Description() string {
	if r.BaseRegexTrigger.Inverse {
		return "Triggers when a member joins with a username that does not match the provided regex"
	}

	return "Triggers when a member joins with a username that matches the provided regex"
}

func (r *UsernameRegexTrigger) CheckUsername(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*BaseRegexTriggerData)

	item, err := RegexCache.Fetch(dataCast.Regex, time.Minute*10, func() (interface{}, error) {
		re, err := regexp.Compile(dataCast.Regex)
		if err != nil {
			return nil, err
		}

		return re, nil
	})

	if err != nil {
		return false, nil
	}

	username := t.MS.User.Username
	if dataCast.NormalizeUnicode {
		username = common.NormalizeAccents(username)
		username = common.NormalizeConfusables(username)
	}

	re := item.Value().(*regexp.Regexp)
	if re.MatchString(username) {
		if r.BaseRegexTrigger.Inverse {
			return false, nil
		}
		return true, nil
	}

	if r.BaseRegexTrigger.Inverse {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ UsernameListener = (*UsernameWordlistTrigger)(nil)

type UsernameWordlistTrigger struct {
	Blacklist bool
}
type UsernameWorldlistData struct {
	ListID           int64
	NormalizeUnicode bool
}

func (uwl *UsernameWordlistTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (uwl *UsernameWordlistTrigger) DataType() interface{} {
	return &UsernameWorldlistData{}
}

func (uwl *UsernameWordlistTrigger) Name() (name string) {
	if uwl.Blacklist {
		return "Join username word blacklist"
	}

	return "Join username word whitelist"
}

func (uwl *UsernameWordlistTrigger) Description() (description string) {
	if uwl.Blacklist {
		return "Triggers when a member joins with a username that contains a word in the specified list"
	}

	return "Triggers when a member joins with a username that contains a words not in the specified list"
}

func (uwl *UsernameWordlistTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "List",
			Key:  "ListID",
			Kind: SettingTypeList,
		},
		{
			Name:    "Normalize Unicode accents & confusables",
			Key:     "NormalizeUnicode",
			Kind:    SettingTypeBool,
			Default: false,
		},
	}
}

func (uwl *UsernameWordlistTrigger) CheckUsername(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*UsernameWorldlistData)

	list, err := FindFetchGuildList(t.GS.ID, dataCast.ListID)
	if err != nil {
		return false, nil
	}

	fields := strings.Fields(PrepareMessageForWordCheck(t.MS.User.Username))

	for _, mf := range fields {
		if dataCast.NormalizeUnicode {
			mf = common.NormalizeAccents(mf)
			mf = common.NormalizeConfusables(mf)
		}

		contained := false
		for _, w := range list.Content {
			if strings.EqualFold(mf, w) {
				if uwl.Blacklist {
					// contains a blacklisted word, trigger
					return true, nil
				} else {
					contained = true
					break
				}
			}
		}

		if !uwl.Blacklist && !contained {
			// word not whitelisted, trigger
			return true, nil
		}
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ UsernameListener = (*UsernameInviteTrigger)(nil)

type UsernameInviteTrigger struct {
}

func (uv *UsernameInviteTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (uv *UsernameInviteTrigger) DataType() interface{} {
	return nil
}

func (uv *UsernameInviteTrigger) Name() (name string) {
	return "Join username invite"
}

func (uv *UsernameInviteTrigger) Description() (description string) {
	return "Triggers when a member joins with a username that contains a server invite"
}

func (uv *UsernameInviteTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (uv *UsernameInviteTrigger) CheckUsername(t *TriggerContext) (bool, error) {
	if common.ContainsInvite(t.MS.User.Username, true, true) != nil {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ JoinListener = (*MemberJoinTrigger)(nil)

type MemberJoinTrigger struct {
}

func (mj *MemberJoinTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (mj *MemberJoinTrigger) DataType() interface{} {
	return nil
}

func (mj *MemberJoinTrigger) Name() (name string) {
	return "New Member"
}

func (mj *MemberJoinTrigger) Description() (description string) {
	return "Triggers when a new member join"
}

func (mj *MemberJoinTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (mj *MemberJoinTrigger) CheckJoin(t *TriggerContext) (isAffected bool, err error) {
	return true, nil
}

/////////////////////////////////////////////////////////////

var _ VoiceStateListener = (*VoiceStateUpdateTrigger)(nil)

type VoiceStateUpdateTrigger struct {
	UserJoin bool
}

func (vs *VoiceStateUpdateTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (vs *VoiceStateUpdateTrigger) DataType() interface{} {
	return nil
}

func (vs *VoiceStateUpdateTrigger) Name() (name string) {
	if vs.UserJoin {
		return "Member joins VC"
	}
	return "Member leaves VC"
}

func (vs *VoiceStateUpdateTrigger) Description() (description string) {
	if vs.UserJoin {
		return "Triggers when a member joins voice-channel"
	}
	return "Triggers when a member leaves voice-channel"
}

func (vs *VoiceStateUpdateTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{}
}

func (vs *VoiceStateUpdateTrigger) CheckVoiceState(t *TriggerContext, cs *dstate.ChannelState) (isAffected bool, err error) {
	//func (vs *VoiceStateUpdateTrigger) CheckVoiceState(t *TriggerContext, gs *dstate.GuildState, ms *dstate.MemberState) (isAffected bool, err error) {
	gs := bot.State.GetGuild(t.MS.GuildID)
	userVoiceState := gs.GetVoiceState(t.MS.User.ID)
	if vs.UserJoin && userVoiceState != nil {
		return true, nil
	}

	if !vs.UserJoin && userVoiceState == nil {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*MessageAttachmentTrigger)(nil)

type MessageAttachmentTriggerData struct {
	FilenameRegex string `valid:",0,256"`
	InverseMatch  bool
}

type MessageAttachmentTrigger struct {
	RequiresAttachment bool
}

func (mat *MessageAttachmentTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (mat *MessageAttachmentTrigger) DataType() interface{} {
	return &MessageAttachmentTriggerData{}
}

func (mat *MessageAttachmentTrigger) Name() string {
	if mat.RequiresAttachment {
		return "Message with attachments"
	}

	return "Message without attachments"
}

func (mat *MessageAttachmentTrigger) Description() string {
	if mat.RequiresAttachment {
		return "Triggers when a message contains an attachment"
	}

	return "Triggers when a message does not contain an attachment"
}

func (mat *MessageAttachmentTrigger) UserSettings() []*SettingDef {
	if mat.RequiresAttachment {
		return []*SettingDef{
			{
				Name: "Filename regex",
				Key:  "FilenameRegex",
				Kind: SettingTypeString,
				Min:  0,
				Max:  256,
			},
			{
				Name:    "Inverse match",
				Key:     "InverseMatch",
				Kind:    SettingTypeBool,
				Default: false,
			},
		}
	}
	return []*SettingDef{}
}

func (mat *MessageAttachmentTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	contains := len(m.Attachments) > 0
	if contains && mat.RequiresAttachment {

		dataCast := triggerCtx.Data.(*MessageAttachmentTriggerData)
		if dataCast.FilenameRegex != "" {
			item, err := RegexCache.Fetch(dataCast.FilenameRegex, time.Minute*10, func() (interface{}, error) {
				re, err := regexp.Compile(dataCast.FilenameRegex)
				if err != nil {
					return nil, err
				}

				return re, nil
			})

			if err != nil {
				return false, nil
			}

			re := item.Value().(*regexp.Regexp)
			for _, ma := range m.Attachments {
				if re.MatchString(ma.Filename) {
					if dataCast.InverseMatch {
						return false, nil
					}

					return true, nil
				}
			}

			if dataCast.InverseMatch {
				return true, nil
			}

			return false, nil
		}

		return true, nil

	} else if !contains && !mat.RequiresAttachment {
		return true, nil
	}

	return false, nil
}

func (mat *MessageAttachmentTrigger) MergeDuplicates(data []interface{}) interface{} {
	return data[0] // no point in having duplicates of this
}

/////////////////////////////////////////////////////////////

var _ MessageTrigger = (*MessageLengthTrigger)(nil)

type MessageLengthTrigger struct {
	Inverted bool
}
type MessageLengthTriggerData struct {
	Length int
}

func (ml *MessageLengthTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (ml *MessageLengthTrigger) DataType() interface{} {
	return &MessageLengthTriggerData{}
}

func (ml *MessageLengthTrigger) Name() (name string) {
	if ml.Inverted {
		return "Message length less than x characters"
	}

	return "Message length more than x characters"
}

func (ml *MessageLengthTrigger) Description() (description string) {
	if ml.Inverted {
		return "Triggers on messages where the content length is lesser than the specified value"
	}

	return "Triggers on messages where the content length is greater than the specified value"
}

func (ml *MessageLengthTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "Length",
			Key:  "Length",
			Kind: SettingTypeInt,
		},
	}
}

func (ml *MessageLengthTrigger) CheckMessage(triggerCtx *TriggerContext, cs *dstate.ChannelState, m *discordgo.Message, mdStripped string) (bool, error) {
	dataCast := triggerCtx.Data.(*MessageLengthTriggerData)

	if ml.Inverted {
		return utf8.RuneCountInString(m.Content) < dataCast.Length, nil
	}

	return utf8.RuneCountInString(m.Content) > dataCast.Length, nil
}

/////////////////////////////////////////////////////////////
/*
var _ UserStatusListener = (*UserStatusRegexTrigger)(nil)

type UserStatusRegexTrigger struct {
	BaseRegexTrigger
}

func (r *UserStatusRegexTrigger) Name() string {
	if r.BaseRegexTrigger.Inverse {
		return "UserStatus not matching regex"
	}

	return "UserStatus matches regex"
}

func (r *UserStatusRegexTrigger) Description() string {
	if r.BaseRegexTrigger.Inverse {
		return "Triggers when a members UserStatus does not match the provided regex"
	}

	return "Triggers when a members UserStatus matches the provided regex"
}

func (r *UserStatusRegexTrigger) CheckUserStatus(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*BaseRegexTriggerData)

	item, err := RegexCache.Fetch(dataCast.Regex, time.Minute*10, func() (interface{}, error) {
		re, err := regexp.Compile(dataCast.Regex)
		if err != nil {
			return nil, err
		}

		return re, nil
	})

	if err != nil {
		return false, nil
	}

	re := item.Value().(*regexp.Regexp)

	if t.MS.Presence.Game != nil {
		if re.MatchString(t.MS.Presence.Game.State) {
			if r.BaseRegexTrigger.Inverse {
				return false, nil
			}
			return true, nil
		}
	}

	if r.BaseRegexTrigger.Inverse {
		return true, nil
	}

	return false, nil
}

/////////////////////////////////////////////////////////////

var _ UserStatusListener = (*UserStatusWordlistTrigger)(nil)

type UserStatusWordlistTrigger struct {
	Blacklist bool
}
type UserStatusWordlistTriggerData struct {
	ListID int64
}

func (nwl *UserStatusWordlistTrigger) Kind() RulePartType {
	return RulePartTrigger
}

func (nwl *UserStatusWordlistTrigger) DataType() interface{} {
	return &UserStatusWordlistTriggerData{}
}

func (nwl *UserStatusWordlistTrigger) Name() (name string) {
	if nwl.Blacklist {
		return "UserStatus word blacklist"
	}

	return "UserStatus word whitelist"
}

func (nwl *UserStatusWordlistTrigger) Description() (description string) {
	if nwl.Blacklist {
		return "Triggers when a member has a UserStatus containing words in the specified list, this is currently very easy to circumvent atm, and will likely be improved in the future."
	}

	return "Triggers when a member has a UserStatus containing words not in the specified list, this is currently very easy to circumvent atm, and will likely be improved in the future."
}

func (nwl *UserStatusWordlistTrigger) UserSettings() []*SettingDef {
	return []*SettingDef{
		&SettingDef{
			Name: "List",
			Key:  "ListID",
			Kind: SettingTypeList,
		},
	}
}

func (nwl *UserStatusWordlistTrigger) CheckUserStatus(t *TriggerContext) (bool, error) {
	dataCast := t.Data.(*UserStatusWordlistTriggerData)

	list, err := FindFetchGuildList(t.MS.GuildID, dataCast.ListID)
	if err != nil {
		return false, nil
	}

	fields := strings.Fields(PrepareMessageForWordCheck(t.MS.Presence.Game.State))

	for _, mf := range fields {
		contained := false
		for _, w := range list.Content {
			if strings.EqualFold(mf, w) {
				if nwl.Blacklist {
					// contains a blacklisted word, trigger
					return true, nil
				} else {
					contained = true
					break
				}
			}
		}

		if !nwl.Blacklist && !contained {
			// word not whitelisted, trigger
			return true, nil
		}
	}

	return false, nil
}
*/

/////////////////////////////////////////////////////////////

var _ AutomodListener = (*AutomodExecution)(nil)

type AutomodExecution struct {
}
type AutomodExecutionData struct {
	RuleID string
}

func (am *AutomodExecution) Kind() RulePartType {
	return RulePartTrigger
}

func (am *AutomodExecution) DataType() interface{} {
	return &AutomodExecutionData{}
}
func (am *AutomodExecution) Name() (name string) {
	return "Message triggers Discord Automod"
}

func (am *AutomodExecution) Description() (description string) {
	return "Triggers when a message is detected by Discord Automod"
}
func (am *AutomodExecution) UserSettings() []*SettingDef {
	return []*SettingDef{
		{
			Name: "Rule ID (leave blank for all)",
			Key:  "RuleID",
			Kind: SettingTypeString,
		},
	}
}

func (am *AutomodExecution) CheckRuleID(triggerCtx *TriggerContext, ruleID int64) (bool, error) {
	dataCast := triggerCtx.Data.(*AutomodExecutionData)

	if dataCast.RuleID == fmt.Sprint(ruleID) {
		return true, nil
	}

	if dataCast.RuleID == "" {
		return true, nil
	}

	return false, nil
}
