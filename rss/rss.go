package rss

//go:generate sqlboiler --no-hooks psql

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/premium"
	"github.com/mrbentarikau/pagst/rss/models"
	"github.com/mmcdole/gofeed"
)

const (
	GuildMaxFeeds        = 50
	GuildMaxFeedsPremium = 100

	RedisRSSAnnounceMessage = "rss_announce_message"
	RedisRSSFeedLockKey     = "rss_feed_channel_lock"
)

var (
	logger       = common.GetPluginLogger(&Plugin{})
	ErrNoChannel = errors.New("no such feed found")
)

var _ mqueue.PluginWithWebhookAvatar = (*Plugin)(nil)

func (p *Plugin) WebhookAvatar() string {
	return RSSIconPNGB64
}

func KeyLastRSSFeedTime(feedURL string) string {
	return "rss_last_feed_time:" + url.QueryEscape(feedURL)
}
func KeyLastRSSFeedLink(feedURL string) string {
	return "rss_last_feed_link:" + url.QueryEscape(feedURL)
}

type Plugin struct {
	rssClient *gofeed.Parser
	Stop      chan *sync.WaitGroup
}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "RSS",
		SysName:  "rss",
		Category: common.PluginCategoryFeeds,
	}
}

func RegisterPlugin() {
	p := &Plugin{}

	mqueue.RegisterSource("rss", p)
	common.InitSchemas("rssfeeds", DBSchemas...)

	/*if !common.FeedEnabled(p.PluginInfo().Name) {
		return
	}*/

	p.rssClient = gofeed.NewParser()
	p.rssClient.UserAgent = common.ConfBotUserAgent.GetString()

	common.RegisterPlugin(p)
}

type AnnouncementMessage struct {
	Enabled      bool
	Announcement string `json:"rss_feed_announce_msg" valid:"template,2000"`
}

var _ mqueue.PluginWithSourceDisabler = (*Plugin)(nil)

// Remove feeds if they don't point to a proper channel
func (p *Plugin) DisableFeed(elem *mqueue.QueuedElement, err error) {
	// Remove it
	_, err = models.RSSFeeds(models.RSSFeedWhere.ChannelID.EQ(elem.ChannelID)).UpdateAllG(context.Background(), models.M{"enabled": false})
	if err != nil {
		logger.WithError(err).WithField("channel_id", elem.ChannelID).Error("failed disabling feed")
	} else {
		logger.WithField("channel", elem.ChannelID).Info("Disabled RSS feed for non-existent channel")
	}
}

func MaxRSSFeedsForContext(ctx context.Context) int64 {
	if premium.ContextPremium(ctx) {
		return GuildMaxFeedsPremium
	}

	return GuildMaxFeeds

}
