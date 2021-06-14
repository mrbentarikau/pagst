package stdcommands

import (
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/stdcommands/addspecialserver"
	"github.com/mrbentarikau/pagst/stdcommands/advice"
	"github.com/mrbentarikau/pagst/stdcommands/allocstat"
	"github.com/mrbentarikau/pagst/stdcommands/banserver"
	"github.com/mrbentarikau/pagst/stdcommands/bashquotes"
	"github.com/mrbentarikau/pagst/stdcommands/calc"
	"github.com/mrbentarikau/pagst/stdcommands/catfact"
	"github.com/mrbentarikau/pagst/stdcommands/ccreqs"
	"github.com/mrbentarikau/pagst/stdcommands/covidstats"
	"github.com/mrbentarikau/pagst/stdcommands/createinvite"
	"github.com/mrbentarikau/pagst/stdcommands/currentshard"
	"github.com/mrbentarikau/pagst/stdcommands/currenttime"
	"github.com/mrbentarikau/pagst/stdcommands/customembed"
	"github.com/mrbentarikau/pagst/stdcommands/dcallvoice"
	"github.com/mrbentarikau/pagst/stdcommands/define"
	"github.com/mrbentarikau/pagst/stdcommands/dogfact"
	"github.com/mrbentarikau/pagst/stdcommands/editrole"
	"github.com/mrbentarikau/pagst/stdcommands/exportcustomcommands"
	"github.com/mrbentarikau/pagst/stdcommands/findserver"
	"github.com/mrbentarikau/pagst/stdcommands/getiplocation"
	"github.com/mrbentarikau/pagst/stdcommands/globalrl"
	"github.com/mrbentarikau/pagst/stdcommands/guildunavailable"
	"github.com/mrbentarikau/pagst/stdcommands/howlongtobeat"
	"github.com/mrbentarikau/pagst/stdcommands/info"
	"github.com/mrbentarikau/pagst/stdcommands/invite"
	"github.com/mrbentarikau/pagst/stdcommands/leaveserver"
	"github.com/mrbentarikau/pagst/stdcommands/listroles"
	"github.com/mrbentarikau/pagst/stdcommands/memstats"
	"github.com/mrbentarikau/pagst/stdcommands/openweathermap"
	"github.com/mrbentarikau/pagst/stdcommands/pagstatus"
	"github.com/mrbentarikau/pagst/stdcommands/ping"
	"github.com/mrbentarikau/pagst/stdcommands/poll"
	"github.com/mrbentarikau/pagst/stdcommands/removespecialserver"
	"github.com/mrbentarikau/pagst/stdcommands/roll"
	"github.com/mrbentarikau/pagst/stdcommands/setstatus"
	"github.com/mrbentarikau/pagst/stdcommands/simpleembed"
	"github.com/mrbentarikau/pagst/stdcommands/sleep"
	"github.com/mrbentarikau/pagst/stdcommands/stateinfo"
	"github.com/mrbentarikau/pagst/stdcommands/throw"
	"github.com/mrbentarikau/pagst/stdcommands/toggledbg"
	"github.com/mrbentarikau/pagst/stdcommands/topcommands"
	"github.com/mrbentarikau/pagst/stdcommands/topevents"
	"github.com/mrbentarikau/pagst/stdcommands/topgames"
	"github.com/mrbentarikau/pagst/stdcommands/topic"
	"github.com/mrbentarikau/pagst/stdcommands/topservers"
	"github.com/mrbentarikau/pagst/stdcommands/unbanserver"
	"github.com/mrbentarikau/pagst/stdcommands/undelete"
	"github.com/mrbentarikau/pagst/stdcommands/viewperms"
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
		define.Command,
		weather.Command,
		openweathermap.Command,
		calc.Command,
		topic.Command,
		catfact.Command,
		dogfact.Command,
		advice.Command,
		ping.Command,
		throw.Command,
		roll.Command,
		customembed.Command,
		simpleembed.Command,
		currenttime.Command,
		editrole.Command,
		listroles.Command,
		memstats.Command,
		wouldyourather.Command,
		poll.Command,
		undelete.Command,
		viewperms.Command,
		topgames.Command,
		xkcd.Command,
		howlongtobeat.Command,
		covidstats.Command,
		getiplocation.Command,
		wolframalpha.Command,
		bashquotes.Command,

		// Maintenance
		stateinfo.Command,
		leaveserver.Command,
		banserver.Command,
		allocstat.Command,
		unbanserver.Command,
		topservers.Command,
		topcommands.Command,
		topevents.Command,
		currentshard.Command,
		pagstatus.Command,
		guildunavailable.Command,
		setstatus.Command,
		createinvite.Command,
		findserver.Command,
		dcallvoice.Command,
		ccreqs.Command,
		sleep.Command,
		toggledbg.Command,
		globalrl.Command,
		exportcustomcommands.Command,
		addspecialserver.Command,
		removespecialserver.Command,
	)

}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLastLegacy(p, ping.HandleMessageCreate, eventsystem.EventMessageCreate)
}

func RegisterPlugin() {
	common.RegisterPlugin(&Plugin{})
}
