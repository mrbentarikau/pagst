package logs

import (
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/backgroundworkers"
	"github.com/mrbentarikau/pagst/logs/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/net/context"
)

var _ backgroundworkers.BackgroundWorkerPlugin = (*Plugin)(nil)

func (p *Plugin) RunBackgroundWorker() {
	ticker := time.NewTicker(time.Minute * 5)
	if !ConfRestrictLogsThirtyDays.GetBool() {
		logger.Infof("[logs] Disabling background worker for message log purge, set yagpdb.enable_restrict_logs to true for this ")
		(<-p.stopWorkers).Done()
		return
	}
	for {
		select {
		case <-ticker.C:
			go p.DeleteOldMessages()
			go p.DeleteOldMessageLogs()
			//go p.DeleteOldWarningsLinks()
		case wg := <-p.stopWorkers:
			wg.Done()
			return
		}

		p.RunBackgroundWorker()
	}
}

func (p *Plugin) DeleteOldMessages() {
	started := time.Now()
	deleted, err := models.Messages2s(qm.SQL("DELETE FROM messages2 WHERE created_at < now() - interval '30 days';")).DeleteAll(context.Background(), common.PQ)
	if err != nil {
		logger.WithError(err).Error("failed deleting older messages from messages2")
		return
	}
	logger.Infof("[logs] Took %s to delete %v old messages from message2", time.Since(started), deleted)
}

func (p *Plugin) DeleteOldMessageLogs() {
	started := time.Now()
	deleted, err := models.MessageLogs2s(qm.SQL("DELETE FROM message_logs2 WHERE created_at < now() - interval '30 days';")).DeleteAll(context.Background(), common.PQ)
	if err != nil {
		logger.WithError(err).Error("failed deleting older message logs from message_logs2")
		return
	}
	logger.Infof("[logs] Took %s to delete %v old message_logs2", time.Since(started), deleted)
}

func (p *Plugin) DeleteOldWarningsLinks() {
	started := time.Now()
	rows := common.GORM.Table("moderation_warnings").Where("created_at < now() - interval '30 days' AND logs_link != ''").Update("logs_link", "").RowsAffected

	logger.Infof("[logs] Took %s to delete %v old moderation_warnings links", time.Since(started), rows)

}

func (p *Plugin) StopBackgroundWorker(wg *sync.WaitGroup) {
	p.stopWorkers <- wg
}
