package main

import (
	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/antiphishing"
	"github.com/mrbentarikau/pagst/common/featureflags"
	"github.com/mrbentarikau/pagst/common/prom"
	"github.com/mrbentarikau/pagst/common/run"
	"github.com/mrbentarikau/pagst/web/discorddata"

	// Core yagpdb packages

	"github.com/mrbentarikau/pagst/admin"
	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/common/internalapi"
	"github.com/mrbentarikau/pagst/common/scheduledevents2"

	// Plugin imports
	"github.com/mrbentarikau/pagst/automod"
	"github.com/mrbentarikau/pagst/automod_basic"
	"github.com/mrbentarikau/pagst/autorole"
	"github.com/mrbentarikau/pagst/cah"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/customcommands"
	"github.com/mrbentarikau/pagst/discordlogger"
	"github.com/mrbentarikau/pagst/logs"
	"github.com/mrbentarikau/pagst/moderation"
	"github.com/mrbentarikau/pagst/notifications"
	"github.com/mrbentarikau/pagst/premium"
	"github.com/mrbentarikau/pagst/premium/patreonpremiumsource"
	"github.com/mrbentarikau/pagst/reddit"
	"github.com/mrbentarikau/pagst/reminders"
	"github.com/mrbentarikau/pagst/reputation"
	"github.com/mrbentarikau/pagst/rolecommands"
	"github.com/mrbentarikau/pagst/rss"
	"github.com/mrbentarikau/pagst/rsvp"
	"github.com/mrbentarikau/pagst/safebrowsing"
	"github.com/mrbentarikau/pagst/serverstats"
	"github.com/mrbentarikau/pagst/soundboard"
	"github.com/mrbentarikau/pagst/stdcommands"
	"github.com/mrbentarikau/pagst/streaming"
	"github.com/mrbentarikau/pagst/tickets"
	"github.com/mrbentarikau/pagst/timezonecompanion"
	"github.com/mrbentarikau/pagst/twitter"
	"github.com/mrbentarikau/pagst/verification"
	"github.com/mrbentarikau/pagst/yageconomy"
	"github.com/mrbentarikau/pagst/youtube"
	// External plugins
)

func main() {

	run.Init()

	//BotSession.LogLevel = discordgo.LogInformational
	paginatedmessages.RegisterPlugin()
	discorddata.RegisterPlugin()

	// Setup plugins
	admin.RegisterPlugin()
	analytics.RegisterPlugin()
	antiphishing.RegisterPlugin()
	automod.RegisterPlugin()
	automod_basic.RegisterPlugin()
	autorole.RegisterPlugin()
	cah.RegisterPlugin()
	commands.RegisterPlugin()
	customcommands.RegisterPlugin()
	discordlogger.Register()
	featureflags.RegisterPlugin()
	internalapi.RegisterPlugin()
	logs.RegisterPlugin()
	moderation.RegisterPlugin()
	notifications.RegisterPlugin()
	patreonpremiumsource.RegisterPlugin()
	premium.RegisterPlugin()
	prom.RegisterPlugin()
	reddit.RegisterPlugin()
	reminders.RegisterPlugin()
	reputation.RegisterPlugin()
	rolecommands.RegisterPlugin()
	rss.RegisterPlugin()
	rsvp.RegisterPlugin()
	safebrowsing.RegisterPlugin()
	scheduledevents2.RegisterPlugin()
	serverstats.RegisterPlugin()
	soundboard.RegisterPlugin()
	stdcommands.RegisterPlugin()
	streaming.RegisterPlugin()
	tickets.RegisterPlugin()
	timezonecompanion.RegisterPlugin()
	twitter.RegisterPlugin()
	verification.RegisterPlugin()
	yageconomy.RegisterPlugin()
	youtube.RegisterPlugin()

	run.Run()
}
