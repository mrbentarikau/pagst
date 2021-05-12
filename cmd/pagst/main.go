package main

import (
	"github.com/mrbentarikau/pagst/analytics"
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
	"github.com/mrbentarikau/pagst/automod_legacy"
	"github.com/mrbentarikau/pagst/autorole"
	"github.com/mrbentarikau/pagst/aylien"
	"github.com/mrbentarikau/pagst/cah"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/customcommands"
	"github.com/mrbentarikau/pagst/discordlogger"
	"github.com/mrbentarikau/pagst/logs"
	"github.com/mrbentarikau/pagst/moderation"
	"github.com/mrbentarikau/pagst/notifications"
	"github.com/mrbentarikau/pagst/stdcommands/owlbot"
	"github.com/mrbentarikau/pagst/premium"
	"github.com/mrbentarikau/pagst/premium/patreonpremiumsource"
	"github.com/mrbentarikau/pagst/reddit"
	"github.com/mrbentarikau/pagst/reminders"
	"github.com/mrbentarikau/pagst/reputation"
	"github.com/mrbentarikau/pagst/rolecommands"
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
	"github.com/mrbentarikau/pagst/youtube"
	// External plugins
)

func main() {

	run.Init()

	//BotSession.LogLevel = discordgo.LogInformational
	paginatedmessages.RegisterPlugin()
	discorddata.RegisterPlugin()

	// Setup plugins
	analytics.RegisterPlugin()
	owlbot.RegisterPlugin()
	safebrowsing.RegisterPlugin()
	discordlogger.Register()
	commands.RegisterPlugin()
	stdcommands.RegisterPlugin()
	serverstats.RegisterPlugin()
	notifications.RegisterPlugin()
	customcommands.RegisterPlugin()
	reddit.RegisterPlugin()
	moderation.RegisterPlugin()
	reputation.RegisterPlugin()
	aylien.RegisterPlugin()
	streaming.RegisterPlugin()
	automod_legacy.RegisterPlugin()
	automod.RegisterPlugin()
	logs.RegisterPlugin()
	autorole.RegisterPlugin()
	reminders.RegisterPlugin()
	soundboard.RegisterPlugin()
	youtube.RegisterPlugin()
	rolecommands.RegisterPlugin()
	cah.RegisterPlugin()
	tickets.RegisterPlugin()
	verification.RegisterPlugin()
	premium.RegisterPlugin()
	patreonpremiumsource.RegisterPlugin()
	scheduledevents2.RegisterPlugin()
	twitter.RegisterPlugin()
	rsvp.RegisterPlugin()
	timezonecompanion.RegisterPlugin()
	admin.RegisterPlugin()
	internalapi.RegisterPlugin()
	prom.RegisterPlugin()
	featureflags.RegisterPlugin()

	run.Run()
}
