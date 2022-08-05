package youtube

import (
	"context"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/cplogs"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/web"
	"github.com/jinzhu/gorm"
	"github.com/mediocregopher/radix/v3"
	"goji.io"
	"goji.io/pat"
	"google.golang.org/api/youtube/v3"
)

type CtxKey int

const (
	CurrentConfig CtxKey = iota
)

//go:embed assets/youtube.html
var PageHTML string

var (
	panelLogKeyAddedFeed        = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "youtube_added_feed", FormatString: "Added YouTube feed %s"})
	panelLogKeyFeedAnnouncement = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "youtube_feed_announcement", FormatString: "Updated YouTube feed announcement"})
	panelLogKeyRemovedFeed      = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "youtube_removed_feed", FormatString: "Removed YouTube feed from %s"})
	panelLogKeyUpdatedFeed      = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "youtube_updated_feed", FormatString: "Updated YouTube feed %s"})
)

type Form struct {
	YoutubeChannelID   string
	YoutubeChannelUser string
	YoutubeCustomURL   string
	YoutubeVideoURL    string
	YoutubeURL         string
	YoutubeAnnounceMsg string `json:"yt_announce_msg" valid:"template,2000"`
	AnnounceEnabled    bool
	DiscordChannel     int64 `valid:"channel,true"`
	ID                 uint
	MentionEveryone    bool
	MentionRole        int64 `valid:"role,true"`
	PublishLivestream  bool
	Enabled            bool
}

type ytUrlType int

const (
	ytUrlTypeVideo ytUrlType = iota
	ytUrlTypeCustom
	ytUrlTypeChannel
	ytUrlTypeUser
	ytUrlTypeInvalid
)

var (
	ytUrlRegex        = regexp.MustCompile(`^(https?:\/\/)?((www|m)\.)?youtube\.com`)
	ytVideoUrlRegex   = regexp.MustCompile(`^(https?:\/\/)?((www|m)\.)?youtube\.com\/watch\?.*v=([a-zA-Z0-9_-]+).*`)
	ytChannelUrlRegex = regexp.MustCompile(`^(https?:\/\/)?((www|m)\.)?youtube\.com\/(channel)\/(UC[\w-]{21}[AQgw])$`)
	ytCustomUrlRegex  = regexp.MustCompile(`^(https?:\/\/)?((www|m)\.)?youtube\.com\/(c\/)?([\w-]+)$`)
	ytUserUrlRegex    = regexp.MustCompile(`^(https?:\/\/)?((www|m)\.)?youtube\.com\/(user\/)([\w-]+)$`)
)

func (p *Plugin) InitWeb() {
	web.AddHTMLTemplate("youtube/assets/youtube.html", PageHTML)
	web.AddSidebarItem(web.SidebarCategoryFeeds, &web.SidebarItem{
		Name: "Youtube",
		URL:  "youtube",
		Icon: "fab fa-youtube",
	})

	ytMux := goji.SubMux()
	web.CPMux.Handle(pat.New("/youtube/*"), ytMux)
	web.CPMux.Handle(pat.New("/youtube"), ytMux)

	// All handlers here require guild channels present
	ytMux.Use(web.RequireBotMemberMW)
	ytMux.Use(web.RequirePermMW(discordgo.PermissionMentionEveryone))

	mainGetHandler := web.ControllerHandler(p.HandleYoutube, "cp_youtube")

	ytMux.Handle(pat.Get("/"), mainGetHandler)
	ytMux.Handle(pat.Get(""), mainGetHandler)

	addHandler := web.ControllerPostHandler(p.HandleNew, mainGetHandler, Form{})

	ytMux.Handle(pat.Post(""), addHandler)
	ytMux.Handle(pat.Post("/"), addHandler)
	ytMux.Handle(pat.Post("/handle_announce"), web.ControllerPostHandler(p.HandleAnnouncement, mainGetHandler, Form{}))
	ytMux.Handle(pat.Post("/:item/update"), web.ControllerPostHandler(BaseEditHandler(p.HandleEdit), mainGetHandler, Form{}))
	ytMux.Handle(pat.Post("/:item/delete"), web.ControllerPostHandler(BaseEditHandler(p.HandleRemove), mainGetHandler, nil))
	ytMux.Handle(pat.Get("/:item/delete"), web.ControllerPostHandler(BaseEditHandler(p.HandleRemove), mainGetHandler, nil))

	// The handler from pubsubhub
	web.RootMux.Handle(pat.New("/yt_new_upload/"+confWebsubVerifytoken.GetString()), http.HandlerFunc(p.HandleFeedUpdate))
}

const DefaultAnnounceMessage = `**{{.ChannelName}}** just uploaded a new video!
{{.URL}}`

func (p *Plugin) HandleYoutube(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	ag, templateData := web.GetBaseCPContextData(ctx)

	var subs []*ChannelSubscription
	err := common.GORM.Where("guild_id = ?", ag.ID).Order("id desc").Find(&subs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return templateData, err
	}

	templateData["YoutubeAnnounceMsg"] = DefaultAnnounceMessage
	var announceMsg AnnouncementMessage
	err = common.GetRedisJson("youtube_announce_message:"+discordgo.StrID(ag.ID), &announceMsg)
	if err == nil && announceMsg.AnnounceMsg != "" {
		templateData["YoutubeAnnounceMsg"] = announceMsg.AnnounceMsg
	}

	templateData["Subs"] = subs
	templateData["AnnounceEnabled"] = announceMsg.Enabled
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(ag.ID) + "/youtube"

	return templateData, nil
}

type LegacyYTStruct struct {
	YTChannelID string
	YTUsername  string
	YTCustomURL string
	YTVideoURL  string
}

func (p *Plugin) HandleNew(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	activeGuild, templateData := web.GetBaseCPContextData(ctx)

	// limit it to max 25 feeds
	var count int
	common.GORM.Model(&ChannelSubscription{}).Where("guild_id = ?", activeGuild.ID).Count(&count)

	if count >= MaxFeedsForContext(ctx) {
		return templateData.AddAlerts(web.ErrorAlert(fmt.Sprintf("Max %d YouTube feeds allowed (%d for premium servers)", GuildMaxFeeds, GuildMaxFeedsPremium))), nil
	}

	data := ctx.Value(common.ContextKeyParsedForm).(*Form)

	cID := trimYouTubeURLParts(data.YoutubeChannelID)
	username := trimYouTubeURLParts(data.YoutubeChannelUser)
	cURL := trimYouTubeURLParts(data.YoutubeCustomURL)
	vURL := trimYouTubeURLParts(data.YoutubeVideoURL)

	urlYT := data.YoutubeURL
	if !ytUrlRegex.MatchString(urlYT) && cID == "" && username == "" && cURL == "" && vURL == "" {
		return templateData.AddAlerts(web.ErrorAlert("This is not a YouTube link...")), nil
	}

	var ytChannel, legacyYTChannel *youtube.Channel
	var err error
	tChain := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	if urlYT != "" {
		urlYT, _ = url.QueryUnescape(urlYT)
		urlYT, _, _ = transform.String(tChain, urlYT)
		ytChannel, err = p.getYTChannel(urlYT)
	} else {
		customURL, _ := url.QueryUnescape(data.YoutubeCustomURL)
		customURL, _, _ = transform.String(tChain, customURL)
		legacyYT := LegacyYTStruct{
			YTChannelID: data.YoutubeChannelID,
			YTUsername:  data.YoutubeChannelUser,
			YTCustomURL: customURL,
			YTVideoURL:  data.YoutubeVideoURL,
		}
		legacyYTChannel, err = p.legacyGetYTChannel(legacyYT)
	}

	if err != nil {
		logger.WithError(err).Errorf("error occurred fetching channel for URL %s", urlYT)
		return templateData.AddAlerts(web.ErrorAlert("No channel found for that link")), err
	}

	sub, err := p.AddFeed(activeGuild.ID, data.DiscordChannel, legacyYTChannel, ytChannel, data.MentionEveryone, data.MentionRole, data.PublishLivestream)
	if err != nil {
		if err == ErrNoChannel {
			return templateData.AddAlerts(web.ErrorAlert("No channel by that id/username found")), errors.New("channel not found")
		}
		return templateData, err
	}

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyAddedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.YoutubeChannelName}))

	return templateData, nil
}

// See https://regex101.com/r/18Ttrq/1/ for some examples of what this matches.
var youtubeURLPartRegexp = regexp.MustCompile(`(?i)\A(?:https?://)?(?:www\.)?youtube\.com/(?:(?:c|channel|user)/)?`)

// trimYouTubeURLParts removes the leading YouTube URL parts from v if present.
// For example, 'youtube.com/user/123' will be transformed to '123'.
func trimYouTubeURLParts(v string) string {
	loc := youtubeURLPartRegexp.FindStringIndex(v)
	if loc == nil {
		return v
	}

	return v[loc[1]:]
}

type ContextKey int

const (
	ContextKeySub ContextKey = iota
)

func BaseEditHandler(inner web.ControllerHandlerFunc) web.ControllerHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
		ctx := r.Context()
		activeGuild, templateData := web.GetBaseCPContextData(ctx)

		id := pat.Param(r, "item")

		// Get the actual watch item from the config
		var sub ChannelSubscription
		err := common.GORM.Model(&ChannelSubscription{}).Where("id = ?", id).First(&sub).Error
		if err != nil {
			return templateData.AddAlerts(web.ErrorAlert("Failed retrieving that feed item")), err
		}

		if sub.GuildID != discordgo.StrID(activeGuild.ID) {
			return templateData.AddAlerts(web.ErrorAlert("This appears to belong somewhere else...")), nil
		}

		ctx = context.WithValue(ctx, ContextKeySub, &sub)

		return inner(w, r.WithContext(ctx))
	}
}

func (p *Plugin) HandleAnnouncement(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	guild, templateData := web.GetBaseCPContextData(ctx)
	data := ctx.Value(common.ContextKeyParsedForm).(*Form)

	var announceMsg AnnouncementMessage
	announceMsg.AnnounceMsg = data.YoutubeAnnounceMsg
	announceMsg.Enabled = data.AnnounceEnabled
	err = common.SetRedisJson("youtube_announce_message:"+discordgo.StrID(guild.ID), announceMsg)

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyFeedAnnouncement, &cplogs.Param{}))

	return
}

func (p *Plugin) HandleEdit(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	_, templateData = web.GetBaseCPContextData(ctx)

	sub := ctx.Value(ContextKeySub).(*ChannelSubscription)
	data := ctx.Value(common.ContextKeyParsedForm).(*Form)

	sub.MentionEveryone = data.MentionEveryone
	sub.PublishLivestream = data.PublishLivestream
	sub.ChannelID = discordgo.StrID(data.DiscordChannel)
	sub.MentionRole = discordgo.StrID(data.MentionRole)
	if data.DiscordChannel == 0 {
		sub.Enabled = false
	} else {
		sub.Enabled = data.Enabled
	}

	err = common.GORM.Save(sub).Error
	if err == nil {
		go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyUpdatedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.YoutubeChannelName}))
	}
	return
}

func (p *Plugin) HandleRemove(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	_, templateData = web.GetBaseCPContextData(ctx)

	sub := ctx.Value(ContextKeySub).(*ChannelSubscription)
	err = common.GORM.Delete(sub).Error
	if err != nil {
		return
	}

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyRemovedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.YoutubeChannelName}))

	p.MaybeRemoveChannelWatch(sub.YoutubeChannelID)
	return
}

func (p *Plugin) HandleFeedUpdate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ctx := r.Context()
	switch query.Get("hub.mode") {
	case "subscribe":
		if query.Get("hub.verify_token") != confWebsubVerifytoken.GetString() {
			return // We don't want no intruders here
		}

		web.CtxLogger(ctx).Info("Responding to challenge: ", query.Get("hub.challenge"))
		p.ValidateSubscription(w, r, query)
		return
	case "unsubscribe":
		if query.Get("hub.verify_token") != confWebsubVerifytoken.GetString() {
			return // We don't want no intruders here
		}

		w.Write([]byte(query.Get("hub.challenge")))

		topicURI, err := url.ParseRequestURI(query.Get("hub.topic"))
		if err != nil {
			web.CtxLogger(ctx).WithError(err).Error("Failed parsing websub topic URI")
			return
		}

		common.RedisPool.Do(radix.Cmd(nil, "ZREM", RedisKeyWebSubChannels, topicURI.Query().Get("channel_id")))
		return
	}

	// Handle new/updated video
	defer r.Body.Close()
	bodyReader := io.LimitReader(r.Body, 0xffff1)

	result, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		web.CtxLogger(ctx).WithError(err).Error("Failed reading body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var parsed XMLFeed

	err = xml.Unmarshal(result, &parsed)
	if err != nil {
		web.CtxLogger(ctx).WithError(err).Error("Failed parsing feed body: ", string(result))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if parsed.VideoId == "" || parsed.ChannelID == "" {
		return
	}

	err = p.CheckVideo(parsed.VideoId, parsed.ChannelID)
	if err != nil {
		web.CtxLogger(ctx).WithError(err).Error("Failed parsing checking new YouTube video")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) ValidateSubscription(w http.ResponseWriter, r *http.Request, query url.Values) {
	w.Write([]byte(query.Get("hub.challenge")))

	lease := query.Get("hub.lease_seconds")
	if lease != "" {
		parsed, err := strconv.ParseInt(lease, 10, 64)
		if err != nil {
			web.CtxLogger(r.Context()).WithError(err).Error("Failed parsing websub lease time")
			return
		}

		expires := time.Now().Add(time.Second * time.Duration(parsed-10)).Unix()

		topicURI, err := url.ParseRequestURI(query.Get("hub.topic"))
		if err != nil {
			web.CtxLogger(r.Context()).WithError(err).Error("Failed parsing websub topic URI")
			return
		}

		common.RedisPool.Do(radix.FlatCmd(nil, "ZADD", RedisKeyWebSubChannels, expires, topicURI.Query().Get("channel_id")))
	}
}

var _ web.PluginWithServerHomeWidget = (*Plugin)(nil)

func (p *Plugin) LoadServerHomeWidget(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ag, templateData := web.GetBaseCPContextData(r.Context())

	templateData["WidgetTitle"] = "Youtube feeds"
	templateData["SettingsPath"] = "/youtube"

	var numFeeds int64
	result := common.GORM.Model(&ChannelSubscription{}).Where("guild_id = ?", ag.ID).Count(&numFeeds)
	if numFeeds > 0 {
		templateData["WidgetEnabled"] = true
	} else {
		templateData["WidgetDisabled"] = true
	}

	const format = `<p>Active Youtube feeds: <code>%d</code></p>`
	templateData["WidgetBody"] = template.HTML(fmt.Sprintf(format, numFeeds))

	return templateData, result.Error
}
