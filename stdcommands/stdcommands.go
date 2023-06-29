package stdcommands

import (
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/advice"
	"github.com/mrbentarikau/pagst/stdcommands/allocstat"
	"github.com/mrbentarikau/pagst/stdcommands/antifish"
	"github.com/mrbentarikau/pagst/stdcommands/banserver"
	"github.com/mrbentarikau/pagst/stdcommands/banserverowner"
	"github.com/mrbentarikau/pagst/stdcommands/bashquotes"
	"github.com/mrbentarikau/pagst/stdcommands/calc"
	"github.com/mrbentarikau/pagst/stdcommands/catfact"
	"github.com/mrbentarikau/pagst/stdcommands/ccreqs"

	// "github.com/mrbentarikau/pagst/stdcommands/covidstats"
	"github.com/mrbentarikau/pagst/stdcommands/createinvite"
	"github.com/mrbentarikau/pagst/stdcommands/currentshard"
	"github.com/mrbentarikau/pagst/stdcommands/currenttime"
	"github.com/mrbentarikau/pagst/stdcommands/customembed"
	"github.com/mrbentarikau/pagst/stdcommands/dadjoke"
	"github.com/mrbentarikau/pagst/stdcommands/dcallvoice"
	"github.com/mrbentarikau/pagst/stdcommands/define"
	"github.com/mrbentarikau/pagst/stdcommands/dogfact"
	"github.com/mrbentarikau/pagst/stdcommands/editchannelpermissions"
	"github.com/mrbentarikau/pagst/stdcommands/editrole"
	"github.com/mrbentarikau/pagst/stdcommands/evilinsult"
	"github.com/mrbentarikau/pagst/stdcommands/exchange"
	"github.com/mrbentarikau/pagst/stdcommands/exportcustomcommands"
	"github.com/mrbentarikau/pagst/stdcommands/exportuserdatabase"
	"github.com/mrbentarikau/pagst/stdcommands/findserver"
	"github.com/mrbentarikau/pagst/stdcommands/getiplocation"
	"github.com/mrbentarikau/pagst/stdcommands/globalrl"
	"github.com/mrbentarikau/pagst/stdcommands/guildunavailable"
	"github.com/mrbentarikau/pagst/stdcommands/howlongtobeat"
	"github.com/mrbentarikau/pagst/stdcommands/info"
	"github.com/mrbentarikau/pagst/stdcommands/inspire"
	"github.com/mrbentarikau/pagst/stdcommands/invite"
	"github.com/mrbentarikau/pagst/stdcommands/isthereanydeal"
	"github.com/mrbentarikau/pagst/stdcommands/leaveserver"
	"github.com/mrbentarikau/pagst/stdcommands/listflags"
	"github.com/mrbentarikau/pagst/stdcommands/listroles"
	"github.com/mrbentarikau/pagst/stdcommands/memstats"
	"github.com/mrbentarikau/pagst/stdcommands/myanimelist"
	"github.com/mrbentarikau/pagst/stdcommands/openweathermap"
	"github.com/mrbentarikau/pagst/stdcommands/owldictionary"

	//"github.com/mrbentarikau/pagst/stdcommands/paginationtest"
	"github.com/mrbentarikau/pagst/stdcommands/pagstatus"
	"github.com/mrbentarikau/pagst/stdcommands/ping"
	"github.com/mrbentarikau/pagst/stdcommands/poll"
	"github.com/mrbentarikau/pagst/stdcommands/quack"
	"github.com/mrbentarikau/pagst/stdcommands/roll"
	"github.com/mrbentarikau/pagst/stdcommands/setstatus"
	"github.com/mrbentarikau/pagst/stdcommands/simpleembed"
	"github.com/mrbentarikau/pagst/stdcommands/sleep"
	"github.com/mrbentarikau/pagst/stdcommands/specialservers"
	"github.com/mrbentarikau/pagst/stdcommands/statedbg"
	"github.com/mrbentarikau/pagst/stdcommands/stateinfo"
	"github.com/mrbentarikau/pagst/stdcommands/thanks"
	"github.com/mrbentarikau/pagst/stdcommands/themoviedb"
	"github.com/mrbentarikau/pagst/stdcommands/throw"
	"github.com/mrbentarikau/pagst/stdcommands/toggledbg"
	"github.com/mrbentarikau/pagst/stdcommands/topcommands"
	"github.com/mrbentarikau/pagst/stdcommands/topevents"
	"github.com/mrbentarikau/pagst/stdcommands/topgames"
	"github.com/mrbentarikau/pagst/stdcommands/topic"
	"github.com/mrbentarikau/pagst/stdcommands/topservers"
	"github.com/mrbentarikau/pagst/stdcommands/unbanserver"
	"github.com/mrbentarikau/pagst/stdcommands/unbanserverowner"
	"github.com/mrbentarikau/pagst/stdcommands/undelete"
	"github.com/mrbentarikau/pagst/stdcommands/viewperms"
	"github.com/mrbentarikau/pagst/stdcommands/voidservers"
	"github.com/mrbentarikau/pagst/stdcommands/weather"
	"github.com/mrbentarikau/pagst/stdcommands/wolframalpha"
	"github.com/mrbentarikau/pagst/stdcommands/wouldyourather"
	"github.com/mrbentarikau/pagst/stdcommands/xkcd"
)

var (
	_ bot.BotInitHandler       = (*Plugin)(nil)
	_ commands.CommandProvider = (*Plugin)(nil)
)

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Standard Commands",
		SysName:  "standard_commands",
		Category: common.PluginCategoryCore,
	}
}

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p,
		// Info
		info.Command,
		invite.Command,

		// Standard
		advice.Command,
		bashquotes.Command,
		calc.Command,
		catfact.Command,
		// covidstats.Command,
		currenttime.Command,
		customembed.Command,
		dadjoke.Command,
		define.Command,
		dogfact.Command,
		editchannelpermissions.Command,
		editrole.Command,
		evilinsult.Command,
		exchange.Command,
		getiplocation.Command,
		howlongtobeat.Command,
		inspire.Command,
		listroles.Command,
		openweathermap.Command,
		ping.Command,
		poll.Command,
		quack.Command,
		roll.Command,
		simpleembed.Command,
		thanks.Command,
		throw.Command,
		topgames.Command,
		topic.Command,
		undelete.Command,
		viewperms.Command,
		weather.Command,
		wolframalpha.Command,
		wouldyourather.Command,
		xkcd.Command,

		// Maintenance
		antifish.Command,
		allocstat.Command,
		banserver.Command,
		banserverowner.Command,
		ccreqs.Command,
		createinvite.Command,
		currentshard.Command,
		dcallvoice.Command,
		exportcustomcommands.Command,
		exportuserdatabase.Command,
		findserver.Command,
		globalrl.Command,
		guildunavailable.Command,
		leaveserver.Command,
		listflags.Command,
		memstats.Command,
		pagstatus.Command,
		// paginationtest.Command,
		setstatus.Command,
		sleep.Command,
		stateinfo.Command,
		toggledbg.Command,
		topcommands.Command,
		topevents.Command,
		topservers.Command,
		unbanserver.Command,
		unbanserverowner.Command,
		voidservers.Command,
	)

	specialservers.Commands()
	statedbg.Commands()

	if !owldictionary.ShouldRegister() {
		common.GetPluginLogger(p).Warn("Owlbot API token not provided, skipping adding owldictionary command...")
		return
	}

	if !isthereanydeal.ShouldRegister() {
		common.GetPluginLogger(p).Warn("IsThereAnyDeal API key not provided, skipping adding isthereanydeal command...")
		return
	}

	if !myanimelist.ShouldRegister() {
		common.GetPluginLogger(p).Warn("MyAnimelist API key not provided, skipping adding myanimelist command...")
		return
	}

	if !themoviedb.ShouldRegister() {
		common.GetPluginLogger(p).Warn("The Movie DB API key not provided, skipping adding tmdb command...")
		return
	}

	commands.AddRootCommands(p, isthereanydeal.Command)
	commands.AddRootCommands(p, myanimelist.Command)
	commands.AddRootCommands(p, owldictionary.Command)
	commands.AddRootCommands(p, themoviedb.Command)

}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLastLegacy(p, ping.HandleMessageCreate, eventsystem.EventMessageCreate)
}

func RegisterPlugin() {
	common.RegisterPlugin(&Plugin{})
}
