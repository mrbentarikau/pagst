package allocstat

import (
	"fmt"
	"runtime"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "allocstat",
	Description:          "Memory statistics. Bot Admin Only.",
	HideFromHelp:         true,
	RunFunc: util.RequireBotAdmin(func(data *dcmd.Data) (interface{}, error) {
		common.BotSession.ChannelTyping(data.ChannelID)
		var memstatsStarted runtime.MemStats
		runtime.ReadMemStats(&memstatsStarted)

		time.Sleep(time.Second * 10)

		var memstatsStopped runtime.MemStats
		runtime.ReadMemStats(&memstatsStopped)

		bytesAlloc := (memstatsStopped.TotalAlloc - memstatsStarted.TotalAlloc) / 1000
		numAlloc := memstatsStopped.Mallocs - memstatsStarted.Mallocs

		lastGC := time.Unix(0, int64(memstatsStopped.LastGC))
		numGC := memstatsStopped.NumGC

		return fmt.Sprintf("Bytes allocated(10s): %dKB\nNum allocs (10s): %d\nLast gc: %s\nNum gc (from start): %d", bytesAlloc, numAlloc, time.Since(lastGC), numGC), nil
	}),
}
