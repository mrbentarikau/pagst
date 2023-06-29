package web

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Masterminds/go-fileserver"
	"github.com/NYTimes/gziphandler"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/common/patreon"
	yagtmpl "github.com/mrbentarikau/pagst/common/templates"

	"github.com/mrbentarikau/pagst/frontend"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/web/discordblog"
	"github.com/natefinch/lumberjack"
	"goji.io"
	"goji.io/middleware"
	"goji.io/pat"
	"golang.org/x/crypto/acme/autocert"
)

var (
	// Core template files
	Templates *template.Template

	Debug              = true // Turns on debug mode
	ListenAddressHTTP  = ":5000"
	ListenAddressHTTPS = ":5001"

	// Muxers
	RootMux           *goji.Mux
	CPMux             *goji.Mux
	ServerPublicMux   *goji.Mux
	ServerPubliAPIMux *goji.Mux
	Error404Mux       *goji.Mux

	properAddresses bool

	https    bool
	exthttps bool

	acceptingRequests *int32

	globalTemplateData = TemplateData(make(map[string]interface{}))

	StartedAt = time.Now()

	logger = common.GetFixedPrefixLogger("web")

	confAnnouncementsChannel       = config.RegisterOption("yagpdb.announcements_channel", "Channel to pull announcements from and display on the control panel homepage", 0)
	confReverseProxyClientIPHeader = config.RegisterOption("yagpdb.web.reverse_proxy_client_ip_header", "If were behind a reverse proxy, this is the header field with the real ip that the proxy passes along", "")
	confDemoServerID               = config.RegisterOption("yagpdb.web.demo_server_id", "Server ID for live demo links", 0)

	confDisableRequestLogging = config.RegisterOption("yagpdb.disable_request_logging", "Disable logging of http requests to web server", false)

	// can be overriden by plugins
	// main prurpose is to plug in a onboarding process through a properietary plugin
	SelectServerHomePageHandler http.Handler = RenderHandler(HandleSelectServer, "cp_selectserver")
)

func init() {
	b := int32(1)
	acceptingRequests = &b

	Templates = template.New("")
	Templates = Templates.Funcs(template.FuncMap{
		"mTemplate":         mTemplate,
		"hasPerm":           hasPerm,
		"formatTime":        prettyTime,
		"checkbox":          tmplCheckbox,
		"checkboxWithInput": tmplCheckboxWithInput,
		"roleOptions":       tmplRoleDropdown,
		"roleOptionsMulti":  tmplRoleDropdownMutli,

		"textChannelOptions":      tmplChannelOpts([]discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews, discordgo.ChannelTypeGuildVoice, discordgo.ChannelTypeGuildForum}),
		"textChannelOptionsMulti": tmplChannelOptsMulti([]discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews, discordgo.ChannelTypeGuildVoice, discordgo.ChannelTypeGuildForum}),

		"voiceChannelOptions":      tmplChannelOpts([]discordgo.ChannelType{discordgo.ChannelTypeGuildVoice}),
		"voiceChannelOptionsMulti": tmplChannelOptsMulti([]discordgo.ChannelType{discordgo.ChannelTypeGuildVoice}),

		"catChannelOptions":      tmplChannelOpts([]discordgo.ChannelType{discordgo.ChannelTypeGuildCategory}),
		"catChannelOptionsMulti": tmplChannelOptsMulti([]discordgo.ChannelType{discordgo.ChannelTypeGuildCategory}),
	})

	Templates = Templates.Funcs(yagtmpl.StandardFuncMap)

	flag.BoolVar(&properAddresses, "pa", false, "Sets the listen addresses to 80 and 443")
	flag.BoolVar(&https, "https", true, "Serve web on HTTPS. Only disable when using an HTTPS reverse proxy.")
	flag.BoolVar(&exthttps, "exthttps", false, "Set if the website uses external https (through reverse proxy) but should only listen on http.")
}

func loadTemplates() {

	coreTemplates := []string{
		"templates/index.html", "templates/cp_main.html",
		"templates/cp_nav.html", "templates/cp_selectserver.html", "templates/cp_logs.html",
		"templates/status.html", "templates/cp_server_home.html", "templates/cp_core_settings.html",
		"templates/error404.html", "templates/privacy_policy.html", "templates/chat.html", "templates/tos.html",
	}

	for _, v := range coreTemplates {
		loadCoreHTMLTemplate(v)
	}
}

func BaseURL() string {
	if https || exthttps {
		return "https://" + common.ConfHost.GetString()
	}

	return "http://" + common.ConfHost.GetString()
}

func Run() {
	common.ServiceTracker.RegisterService(common.ServiceTypeFrontend, "Webserver", "", nil)

	common.RegisterPlugin(&ControlPanelPlugin{})

	loadTemplates()

	AddGlobalTemplateData("BotName", common.ConfBotName.GetString())
	AddGlobalTemplateData("ClientID", common.ConfClientID.GetString())
	AddGlobalTemplateData("Host", common.ConfHost.GetString())
	AddGlobalTemplateData("SupportServerName", common.ConfSupportServerName.GetString())
	AddGlobalTemplateData("SupportServerURL", common.ConfSupportServerURL.GetString())
	AddGlobalTemplateData("Testing", common.Testing)
	AddGlobalTemplateData("Version", common.VERSION)

	if properAddresses {
		ListenAddressHTTP = ":80"
		ListenAddressHTTPS = ":443"
	}

	patreon.Run()

	InitOauth()
	mux := setupRoutes()

	// Start monitoring the bot
	go pollCommandsRan()
	go pollCCsRan()
	go pollAMV2sRan()
	go pollDiscordStatus()
	go pollRedditQuotes()

	blogChannel := confAnnouncementsChannel.GetInt()
	if blogChannel != 0 {
		go discordblog.RunPoller(common.BotSession, int64(blogChannel), time.Minute)
	}

	logger.Info("Running webservers")
	runServers(mux)
}

func Stop() {
	atomic.StoreInt32(acceptingRequests, 0)
}

func IsAcceptingRequests() bool {
	return atomic.LoadInt32(acceptingRequests) != 0
}

func runServers(mainMuxer *goji.Mux) {
	if !https {
		logger.Info("Starting ", common.ConfBotName.GetString(), " web server http:", ListenAddressHTTP)

		server := &http.Server{
			Addr:        ListenAddressHTTP,
			Handler:     mainMuxer,
			IdleTimeout: time.Minute,
		}

		err := server.ListenAndServe()
		if err != nil {
			logger.Error("Failed http ListenAndServe:", err)
		}
	} else {
		logger.Info("Starting ", common.ConfBotName.GetString(), " web server http:", ListenAddressHTTP, ", and https:", ListenAddressHTTPS)

		cache := autocert.DirCache("cert")

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(common.ConfHost.GetString(), "www."+common.ConfHost.GetString()),
			Email:      common.ConfEmail.GetString(),
			Cache:      cache,
		}

		// launch the redis server
		go func() {
			unsafeHandler := &http.Server{
				Addr:        ListenAddressHTTP,
				Handler:     certManager.HTTPHandler(http.HandlerFunc(httpsRedirHandler)),
				IdleTimeout: time.Minute,
			}

			err := unsafeHandler.ListenAndServe()
			if err != nil {
				logger.Error("Failed http ListenAndServe:", err)
			}
		}()

		tlsServer := &http.Server{
			Addr:        ListenAddressHTTPS,
			Handler:     mainMuxer,
			IdleTimeout: time.Minute,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		err := tlsServer.ListenAndServeTLS("", "")
		if err != nil {
			logger.Error("Failed https ListenAndServeTLS:", err)
		}
	}
}

func setupRoutes() *goji.Mux {

	// setup the root routes and middle-wares
	setupRootMux()
	RootMux.Use(NotFound())
	// Guild specific public routes, does not require admin or being logged in at all
	serverPublicMux := goji.SubMux()
	serverPublicMux.Use(ActiveServerMW)
	serverPublicMux.Use(RequireActiveServer)
	serverPublicMux.Use(LoadCoreConfigMiddleware)
	serverPublicMux.Use(SetGuildMemberMiddleware)
	serverPublicMux.Use(NotFound())

	RootMux.Handle(pat.New("/public/:server"), serverPublicMux)
	RootMux.Handle(pat.New("/public/:server/*"), serverPublicMux)
	ServerPublicMux = serverPublicMux

	// same as above but for API stuff
	ServerPubliAPIMux = goji.SubMux()
	ServerPubliAPIMux.Use(ActiveServerMW)
	ServerPubliAPIMux.Use(RequireActiveServer)
	ServerPubliAPIMux.Use(LoadCoreConfigMiddleware)
	ServerPubliAPIMux.Use(SetGuildMemberMiddleware)
	ServerPubliAPIMux.Use(NotFound())

	RootMux.Handle(pat.Get("/api/:server"), ServerPubliAPIMux)
	RootMux.Handle(pat.Get("/api/:server/*"), ServerPubliAPIMux)

	ServerPubliAPIMux.Handle(pat.Get("/channelperms/:channel"), RequireActiveServer(APIHandler(HandleChanenlPermissions)))

	// Server selection has its own handler
	RootMux.Handle(pat.Get("/manage"), SelectServerHomePageHandler)
	RootMux.Handle(pat.Get("/manage/"), SelectServerHomePageHandler)
	RootMux.Handle(pat.Get("/status"), ControllerHandler(HandleStatusHTML, "cp_status"))
	RootMux.Handle(pat.Get("/status/"), ControllerHandler(HandleStatusHTML, "cp_status"))
	RootMux.Handle(pat.Get("/privacy_policy"), ControllerHandler(HandleStatusHTML, "cp_privacy_policy"))
	RootMux.Handle(pat.Get("/privacy_policy/"), ControllerHandler(HandleStatusHTML, "cp_privacy_policy"))
	RootMux.Handle(pat.Get("/tos"), ControllerHandler(HandleStatusHTML, "cp_terms_and_conditions"))
	RootMux.Handle(pat.Get("/tos/"), ControllerHandler(HandleStatusHTML, "cp_terms_and_conditions"))
	RootMux.Handle(pat.Get("/status.json"), APIHandler(HandleStatusJSON))
	RootMux.Handle(pat.Get("/error404"), RenderHandler(HandleError404, "error404"))
	RootMux.Handle(pat.Get("/error404/"), RenderHandler(HandleError404, "error404"))
	RootMux.Handle(pat.Post("/shard/:shard/reconnect"), ControllerHandler(HandleReconnectShard, "cp_status"))
	RootMux.Handle(pat.Post("/shard/:shard/reconnect/"), ControllerHandler(HandleReconnectShard, "cp_status"))

	RootMux.HandleFunc(pat.Get("/cp"), legacyCPRedirHandler)
	RootMux.HandleFunc(pat.Get("/cp/*"), legacyCPRedirHandler)

	// Server control panel, requires you to be an admin for the server (owner or have server management role)
	CPMux = goji.SubMux()
	CPMux.Use(ActiveServerMW)
	CPMux.Use(RequireActiveServer)
	CPMux.Use(LoadCoreConfigMiddleware)
	CPMux.Use(SetGuildMemberMiddleware)
	CPMux.Use(RequireServerAdminMiddleware)
	CPMux.Use(NotFound())

	RootMux.Handle(pat.New("/manage/:server"), CPMux)
	RootMux.Handle(pat.New("/manage/:server/*"), CPMux)

	CPMux.Handle(pat.Get("/cplogs"), RenderHandler(HandleCPLogs, "cp_action_logs"))
	CPMux.Handle(pat.Get("/cplogs/"), RenderHandler(HandleCPLogs, "cp_action_logs"))
	CPMux.Handle(pat.Get("/home"), ControllerHandler(HandleServerHome, "cp_server_home"))
	CPMux.Handle(pat.Get("/home/"), ControllerHandler(HandleServerHome, "cp_server_home"))

	coreSettingsHandler := RenderHandler(nil, "cp_core_settings")

	CPMux.Handle(pat.Get("/core/"), coreSettingsHandler)
	CPMux.Handle(pat.Get("/core"), coreSettingsHandler)
	CPMux.Handle(pat.Post("/core"), ControllerPostHandler(HandlePostCoreSettings, coreSettingsHandler, CoreConfigPostForm{}))

	RootMux.Handle(pat.Get("/guild_selection"), RequireSessionMiddleware(ControllerHandler(HandleGetManagedGuilds, "cp_guild_selection")))
	CPMux.Handle(pat.Get("/guild_selection"), RequireSessionMiddleware(ControllerHandler(HandleGetManagedGuilds, "cp_guild_selection")))

	// Set up the routes for the per server home widgets
	for _, p := range common.Plugins {
		if cast, ok := p.(PluginWithServerHomeWidget); ok {
			handler := GuildScopeCacheMW(p, ControllerHandler(cast.LoadServerHomeWidget, "cp_server_home_widget"))

			if mwares, ok2 := p.(PluginWithServerHomeWidgetMiddlewares); ok2 {
				handler = mwares.ServerHomeWidgetApplyMiddlewares(handler)
			}

			CPMux.Handle(pat.Get("/homewidgets/"+p.PluginInfo().SysName), handler)
			CPMux.Use(NotFound())
		}
	}

	AddSidebarItem(SidebarCategoryCore, &SidebarItem{
		Name: "Core",
		URL:  "core",
		Icon: "fas fa-cog",
	})

	AddSidebarItem(SidebarCategoryCore, &SidebarItem{
		Name: "Control panel logs",
		URL:  "cplogs",
		Icon: "fas fa-copy",
	})

	for _, plugin := range common.Plugins {
		if webPlugin, ok := plugin.(Plugin); ok {
			webPlugin.InitWeb()
			logger.Info("Initialized web plugin:", plugin.PluginInfo().Name)
		}
	}

	return RootMux
}

func AddServer(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("https://discordapp.com/oauth2/authorize?client_id=%s&scope=bot%%20identify%%20guilds%%20applications.commands&permissions=1516122532343&response_type=code&redirect_uri=https://%s/manage", common.ConfClientID.GetString(), common.ConfHost.GetString())
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func NotFound() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := middleware.Handler(r.Context())
			if handler == nil {
				http.Redirect(w, r, "/error404", http.StatusMovedPermanently)

				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

var StaticFilesFS fs.FS = frontend.StaticFiles

func setupRootMux() {
	mux := goji.NewMux()
	RootMux = mux

	if !confDisableRequestLogging.GetBool() {
		requestLogger := &lumberjack.Logger{
			Filename: "access.log",
			MaxSize:  10,
		}

		mux.Use(RequestLogger(requestLogger))
	}
	fileserver.NotFoundHandler = func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/error404", http.StatusMovedPermanently)
	}

	// Setup fileserver
	mux.Handle(pat.Get("/static"), fileserver.FileServer(http.FS(StaticFilesFS)))
	mux.Handle(pat.Get("/static/*"), fileserver.FileServer(http.FS(StaticFilesFS)))
	mux.Handle(pat.Get("/robots.txt"), http.HandlerFunc(handleRobotsTXT))

	// General middleware
	mux.Use(SkipStaticMW(gziphandler.GzipHandler, ".css", ".js", ".map"))
	mux.Use(SkipStaticMW(MiscMiddleware))
	mux.Use(SkipStaticMW(BaseTemplateDataMiddleware))
	mux.Use(SkipStaticMW(SessionMiddleware))
	mux.Use(SkipStaticMW(UserInfoMiddleware))
	mux.Use(SkipStaticMW(CSRFProtectionMW))
	mux.Use(addPromCountMW)

	// General handlers
	mux.Handle(pat.Get("/"), ControllerHandler(HandleLandingPage, "index"))
	mux.Handle(pat.Get("/chat"), ControllerHandler(HandleLandingPage, "chat"))
	mux.HandleFunc(pat.Get("/login"), HandleLogin)
	mux.HandleFunc(pat.Get("/confirm_login"), HandleConfirmLogin)
	mux.HandleFunc(pat.Get("/logout"), HandleLogout)
	mux.HandleFunc(pat.Get("/invite"), AddServer)
}

func httpsRedirHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}

func AddGlobalTemplateData(key string, data interface{}) {
	globalTemplateData[key] = data
}

func legacyCPRedirHandler(w http.ResponseWriter, r *http.Request) {
	logger.Println("Hit cp path: ", r.RequestURI)
	trimmed := strings.TrimPrefix(r.RequestURI, "/cp")
	http.Redirect(w, r, "/manage"+trimmed, http.StatusMovedPermanently)
}

func AddHTMLTemplate(name, contents string) {
	Templates = Templates.New(name)
	Templates = template.Must(Templates.Parse(contents))
}

func loadCoreHTMLTemplate(path string) {
	contents, err := frontend.CoreTemplates.ReadFile(path)
	if err != nil {
		panic(err)
	}
	Templates = Templates.New(path)
	Templates = template.Must(Templates.Parse(string(contents)))
}

const (
	SidebarCategoryTopLevel = "Top"
	SidebarCategoryFeeds    = "Feeds"
	SidebarCategoryTools    = "Tools"
	SidebarCategoryFun      = "Fun"
	SidebarCategoryCore     = "Core"
)

type SidebarItem struct {
	Name            string
	URL             string
	Icon            string
	CustomIconImage string
	New             bool
	External        bool
}

var sideBarItems = make(map[string][]*SidebarItem)

func AddSidebarItem(category string, sItem *SidebarItem) {
	sideBarItems[category] = append(sideBarItems[category], sItem)
}
