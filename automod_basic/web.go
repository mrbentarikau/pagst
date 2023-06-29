package automod_basic

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/mrbentarikau/pagst/common/cplogs"
	"github.com/mrbentarikau/pagst/common/featureflags"
	"github.com/mrbentarikau/pagst/common/pubsub"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/web"
	"goji.io"
	"goji.io/pat"
)

//go:embed assets/automod_basic.html
var PageHTML string

type GeneralForm struct {
	Enabled bool
}

var panelLogKeyUpdatedSettings = cplogs.RegisterActionFormat(&cplogs.ActionFormat{Key: "automod_basic_settings_updated", FormatString: "Updated settings for basic automod"})

func (p *Plugin) InitWeb() {
	web.AddHTMLTemplate("automod_basic/assets/automod_basic.html", PageHTML)

	web.AddSidebarItem(web.SidebarCategoryTools, &web.SidebarItem{
		Name: "Basic Automoderator",
		URL:  "automod_basic",
		Icon: "fa-brands fa-android",
	})

	autmodMux := goji.SubMux()
	web.CPMux.Handle(pat.New("/automod_basic/*"), autmodMux)
	web.CPMux.Handle(pat.New("/automod_basic"), autmodMux)

	// Alll handlers here require guild channels present
	autmodMux.Use(web.RequireBotMemberMW)
	autmodMux.Use(web.RequirePermMW(discordgo.PermissionManageRoles, discordgo.PermissionKickMembers, discordgo.PermissionBanMembers, discordgo.PermissionManageMessages))

	getHandler := web.RenderHandler(HandleAutomod, "cp_automod_basic")

	autmodMux.Handle(pat.Get("/"), getHandler)
	autmodMux.Handle(pat.Get(""), getHandler)

	// Post handlers
	autmodMux.Handle(pat.Post("/"), ExtraPostMW(web.SimpleConfigSaverHandler(Config{}, getHandler, panelLogKeyUpdatedSettings)))
	autmodMux.Handle(pat.Post(""), ExtraPostMW(web.SimpleConfigSaverHandler(Config{}, getHandler, panelLogKeyUpdatedSettings)))
}

func HandleAutomod(w http.ResponseWriter, r *http.Request) interface{} {
	g, templateData := web.GetBaseCPContextData(r.Context())

	config, err := GetConfig(g.ID)
	web.CheckErr(templateData, err, "Failed retrieving rules", web.CtxLogger(r.Context()).Error)

	templateData["AutomodConfig"] = config
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(g.ID) + "/automod_basic/"

	return templateData
}

// Invalidates the cache
func ExtraPostMW(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		activeGuild, _ := web.GetBaseCPContextData(r.Context())
		pubsub.Publish("update_automod_basic_rules", activeGuild.ID, nil)
		featureflags.MarkGuildDirty(activeGuild.ID)
		inner.ServeHTTP(w, r)
	}
	return http.HandlerFunc(mw)
}

var _ web.PluginWithServerHomeWidget = (*Plugin)(nil)

func (p *Plugin) LoadServerHomeWidget(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	g, templateData := web.GetBaseCPContextData(r.Context())

	templateData["WidgetTitle"] = "Basic Automod"
	templateData["SettingsPath"] = "/automod_basic"

	config, err := GetConfig(g.ID)
	if err != nil {
		return templateData, err
	}

	if config.Enabled {
		templateData["WidgetEnabled"] = true
	} else {
		templateData["WidgetDisabled"] = true
	}

	const format = `<ul>
	<li>Slowmode: %s</li>
	<li>Mass mention: %s</li>
	<li>Server invites: %s</li>
	<li>Any links: %s</li>
	<li>Banned words: %s</li>
	<li>Banned websites: %s</li>
</ul>`

	slowmode := web.EnabledDisabledSpanStatus(config.Spam.Enabled)
	massMention := web.EnabledDisabledSpanStatus(config.Mention.Enabled)
	invites := web.EnabledDisabledSpanStatus(config.Invite.Enabled)
	links := web.EnabledDisabledSpanStatus(config.Links.Enabled)
	words := web.EnabledDisabledSpanStatus(config.Words.Enabled)
	sites := web.EnabledDisabledSpanStatus(config.Sites.Enabled)

	templateData["WidgetBody"] = template.HTML(fmt.Sprintf(format, slowmode,
		massMention, invites, links, words, sites))

	return templateData, nil
}
