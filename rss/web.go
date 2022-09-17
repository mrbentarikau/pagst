package rss

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/cplogs"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/rss/models"
	"github.com/mrbentarikau/pagst/web"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"goji.io"
	"goji.io/pat"
)

//go:embed assets/rssfeeds.html
var PageHTML string

type ContextKey int

const (
	ContextKeySub          ContextKey = iota
	DefaultAnnounceMessage            = `Hey, {{with .SelectedRoleID}}{{(getRole .).Mention}}{{else}}everyone{{end}}...
Incoming RSS feed, **{{.RSSFeed.Title}}** just posted {{.RSSFeedItem.Link}} !`
)

var (
	panelLogKeyAddedFeed        = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "rss_added_feed", FormatString: "Added RSS feed %s"})
	panelLogKeyFeedAnnouncement = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "rss_feed_announcement", FormatString: "Updated RSS feed announcement"})
	panelLogKeyRemovedFeed      = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "rss_removed_feed", FormatString: "Removed RSS feed %s"})
	panelLogKeyUpdatedFeed      = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "rss_updated_feed", FormatString: "Updated RSS feed %s"})
)

type Form struct {
	FeedName           string `valid:",0,50"`
	FeedURL            string
	RSSFeedAnnounceMsg string `json:"rss_feed_announce_msg" valid:"template,2000"`
	AnnounceEnabled    bool
	DiscordChannel     int64 `valid:"channel,true"`
	ID                 uint
	MentionRole        int64 `valid:"role,true"`
	Enabled            bool
}

type FormEdit struct {
	FeedName       string `valid:",0,50"`
	DiscordChannel int64  `valid:"channel,true"`
	MentionRole    int64  `valid:"role,true"`
	Enabled        bool
}

func (p *Plugin) InitWeb() {
	web.AddHTMLTemplate("rss/assets/rssfeeds.html", PageHTML)
	web.AddSidebarItem(web.SidebarCategoryFeeds, &web.SidebarItem{
		Name: "RSS Feeds",
		URL:  "rssfeeds",
		Icon: "fa-solid fa-square-rss",
	})

	rssMux := goji.SubMux()
	web.CPMux.Handle(pat.New("/rssfeeds/*"), rssMux)
	web.CPMux.Handle(pat.New("/rssfeeds"), rssMux)

	// All handlers here require guild channels present
	rssMux.Use(web.RequireBotMemberMW)
	//rssMux.Use(web.RequirePermMW(discordgo.PermissionMentionEveryone))

	mainGetHandler := web.ControllerHandler(p.HandleRSS, "cp_rssfeeds")

	rssMux.Handle(pat.Get("/"), mainGetHandler)
	rssMux.Handle(pat.Get(""), mainGetHandler)

	addHandler := web.ControllerPostHandler(p.HandleNew, mainGetHandler, Form{})

	rssMux.Handle(pat.Post(""), addHandler)
	rssMux.Handle(pat.Post("/"), addHandler)
	rssMux.Handle(pat.Post("/handle_announce"), web.ControllerPostHandler(p.HandleAnnouncement, mainGetHandler, Form{}))
	rssMux.Handle(pat.Post("/:item/update"), web.ControllerPostHandler(BaseEditHandler(p.HandleEdit), mainGetHandler, FormEdit{}))
	rssMux.Handle(pat.Post("/:item/delete"), web.ControllerPostHandler(BaseEditHandler(p.HandleRemove), mainGetHandler, nil))
	rssMux.Handle(pat.Get("/:item/delete"), web.ControllerPostHandler(BaseEditHandler(p.HandleRemove), mainGetHandler, nil))
}

func (p *Plugin) HandleRSS(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	ag, templateData := web.GetBaseCPContextData(ctx)

	subs, err := models.RSSFeeds(models.RSSFeedWhere.GuildID.EQ(ag.ID), qm.OrderBy("id asc")).AllG(ctx)
	if err != nil {
		return templateData, err
	}

	templateData["RSSFeedAnnounceMsg"] = DefaultAnnounceMessage

	dbAnnounceMsg, err := models.RSSAnnouncements(qm.Where("guild_id = ?", ag.ID)).OneG(ctx)
	if err != nil {
		return templateData, err
	}

	if dbAnnounceMsg.Announcement != "" {
		templateData["RSSFeedAnnounceMsg"] = dbAnnounceMsg.Announcement
	}

	templateData["Subs"] = subs
	templateData["AnnounceEnabled"] = dbAnnounceMsg.Enabled
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(ag.ID) + "/rssfeeds"

	return templateData, nil
}

func (p *Plugin) HandleAnnouncement(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	guild, templateData := web.GetBaseCPContextData(ctx)
	data := ctx.Value(common.ContextKeyParsedForm).(*Form)

	dbAnnounceMsg := &models.RSSAnnouncement{
		GuildID:      guild.ID,
		Announcement: data.RSSFeedAnnounceMsg,
		Enabled:      data.AnnounceEnabled,
	}

	err = dbAnnounceMsg.UpsertG(ctx, true, []string{"guild_id"}, boil.Infer(), boil.Infer())
	if err != nil {
		return nil, err
	}

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyFeedAnnouncement, &cplogs.Param{}))

	return
}

func (p *Plugin) HandleNew(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	activeGuild, templateData := web.GetBaseCPContextData(ctx)

	// limit it to MaxFeedsForContext feeds
	count, err := models.RSSFeeds(models.RSSFeedWhere.GuildID.EQ(activeGuild.ID)).CountG(ctx)
	if err != nil {
		return templateData, err
	}

	if count >= MaxRSSFeedsForContext(ctx) {
		return templateData.AddAlerts(web.ErrorAlert(fmt.Sprintf("Max %d RSS feeds allowed (%d for premium servers)", GuildMaxFeeds, GuildMaxFeedsPremium))), nil
	}

	formData := ctx.Value(common.ContextKeyParsedForm).(*Form)
	url := formData.FeedURL

	// validate URL
	if !common.LinkRegexProtocolStrict.MatchString(url) {
		return templateData.AddAlerts(web.ErrorAlert("This is not a valid HTTP(S) URL...")), nil
	}

	reRedditYT := regexp.MustCompile(`(reddit)(.com/)`)
	if reRedditYT.MatchString(url) {
		reMatched := reRedditYT.FindStringSubmatch(url)[1]
		return templateData.AddAlerts(web.ErrorAlert("Use " + reMatched + " feeds plugin...")), nil
	}

	// search for the RSS feed
	parsedURL, err := p.rssClient.ParseURL(url)
	if err != nil {
		return templateData.AddAlerts(web.ErrorAlert("RSS feed not found")), nil
	}

	if len(parsedURL.Items) == 0 {
		return templateData.AddAlerts(web.ErrorAlert("RSS feed has no items")), nil
	}

	if parsedURL.Items[0].PublishedParsed == nil {
		return templateData.AddAlerts(web.ErrorAlert("RSS feed has no valid formatting")), nil
	}

	feedTitle := parsedURL.Title

	sub, err := p.AddRSSFeed(activeGuild.ID, formData.DiscordChannel, formData.MentionRole, strings.TrimSpace(formData.FeedName), feedTitle, url)
	if err != nil {
		if err == ErrNoChannel {
			return templateData.AddAlerts(web.ErrorAlert("No such RSS feed found")), errors.New("channel not found")
		}
		return templateData, err
	}

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyAddedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.FeedURL}))

	return templateData, nil
}

func BaseEditHandler(inner web.ControllerHandlerFunc) web.ControllerHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
		ctx := r.Context()
		activeGuild, templateData := web.GetBaseCPContextData(ctx)

		id, _ := strconv.Atoi(pat.Param(r, "item"))

		sub, err := models.FindRSSFeedG(ctx, int64(id))
		if err != nil {
			return templateData.AddAlerts(web.ErrorAlert("Failed retrieving that feed item")), err
		}

		if sub.GuildID != activeGuild.ID {
			return templateData.AddAlerts(web.ErrorAlert("This appears to belong somewhere else...")), nil
		}

		ctx = context.WithValue(ctx, ContextKeySub, sub)

		return inner(w, r.WithContext(ctx))
	}
}

func (p *Plugin) HandleEdit(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	_, templateData = web.GetBaseCPContextData(ctx)

	sub := ctx.Value(ContextKeySub).(*models.RSSFeed)
	data := ctx.Value(common.ContextKeyParsedForm).(*FormEdit)

	sub.ChannelID = data.DiscordChannel
	sub.MentionRole = data.MentionRole
	sub.FeedName = strings.TrimSpace(data.FeedName)

	if data.DiscordChannel == 0 {
		sub.Enabled = false
	} else {
		sub.Enabled = data.Enabled
	}

	_, err = sub.UpdateG(ctx, boil.Whitelist("channel_id", "feed_name", "mention_role", "enabled"))
	if err == nil {
		go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyUpdatedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.FeedURL}))
	}
	return
}

func (p *Plugin) HandleRemove(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	ctx := r.Context()
	_, templateData = web.GetBaseCPContextData(ctx)

	sub := ctx.Value(ContextKeySub).(*models.RSSFeed)

	_, err = sub.DeleteG(ctx)
	if err != nil {
		return
	}

	go cplogs.RetryAddEntry(web.NewLogEntryFromContext(r.Context(), panelLogKeyRemovedFeed, &cplogs.Param{Type: cplogs.ParamTypeString, Value: sub.FeedURL}))

	return
}

func (p *Plugin) LoadServerHomeWidget(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ag, templateData := web.GetBaseCPContextData(r.Context())

	templateData["WidgetTitle"] = "RSS feeds"
	templateData["SettingsPath"] = "/rssfeeds"

	numFeeds, err := models.RSSFeeds(models.RSSFeedWhere.GuildID.EQ(ag.ID)).CountG(r.Context())
	if err != nil {
		return templateData, err
	}

	if numFeeds > 0 {
		templateData["WidgetEnabled"] = true
	} else {
		templateData["WidgetDisabled"] = true
	}

	const format = `<p>Active RSS feeds: <code>%d</code></p>`
	templateData["WidgetBody"] = template.HTML(fmt.Sprintf(format, numFeeds))

	return templateData, nil
}
