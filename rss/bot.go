package rss

import (
	"context"
	"fmt"

	"github.com/mrbentarikau/pagst/rss/models"
)

func (p *Plugin) Status() (string, string) {
	//var numFeeds int
	numFeeds, err := models.RSSFeeds(models.RSSFeedWhere.Enabled.EQ(true)).CountG(context.Background())
	if err != nil {
		logger.WithError(err).Error("failed fetching status")
		return "RSS feeds", "error"
	}
	//common.GORM.Model(&RSSFeed{}).Count(&numFeeds)

	return "RSS feeds", fmt.Sprintf("%d", numFeeds)
}
