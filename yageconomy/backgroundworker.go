package yageconomy

import (
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/backgroundworkers"
)

var _ backgroundworkers.BackgroundWorkerPlugin = (*Plugin)(nil)

func (p *Plugin) RunBackgroundWorker() {
	go p.updateInterestsLoop()
}

func (p *Plugin) StopBackgroundWorker(wg *sync.WaitGroup) {
	wg.Done()
}

func (p *Plugin) updateInterestsLoop() {
	for {
		time.Sleep(time.Minute)
		err := p.updateInterests()
		if err != nil {
			logger.WithError(err).Error("failed updating interest")
		}
	}
}

func (p *Plugin) updateInterests() error {
	result, err := common.PQ.Exec("UPDATE economy_users SET money_bank = money_bank * 0.95, last_interest_update = now() WHERE (now() - last_interest_update) > interval '1 day'")
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Infof("Updated %d rows when adding interest", rows)
	return nil
}
