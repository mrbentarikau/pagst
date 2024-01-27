package twitter

//go:generate sqlboiler --no-hooks psql

import (
	"sync"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/twitter/models"

	twitterscraper "github.com/n0madic/twitter-scraper"
	//twitterscraper "github.com/mrbentarikau/pagst/lib/twitter-scraper"
)

var (
	logger                   = common.GetPluginLogger(&Plugin{})
	confTwitterProxy         = config.RegisterOption("yagpdb.twitter.proxy", "Proxy URL to scrape feeds from twitter", "")
	confTwitterBatchSize     = config.RegisterOption("yagpdb.twitter.batch_size", "Batch Size for scraping feeds", 50)
	confTwitterPollFrequency = config.RegisterOption("yagpdb.twitter.poll_frequency", "Minimum Delay in each feed poll for all feeds in minutes", 15)
	confTwitterBatchDelay    = config.RegisterOption("yagpdb.twitter.batch_delay", "Delay in seconds between each batch", 10)

	confTwitterUsername = config.RegisterOption("yagpdb.twitter.username", "Twitter username for login", "")
	confTwitterPassword = config.RegisterOption("yagpdb.twitter.password", "Twitter password for login", "")
	confTwitterEmail    = config.RegisterOption("yagpdb.twitter.email", "Twitter e-mail for login", "")
	confTwitterActive   = config.RegisterOption("yagpdb.twitter.enabled", "Twitter Plugin enabled", true)
)

func ShouldRegister() bool {
	return confTwitterActive.GetBool()
}

type Plugin struct {
	Stop chan *sync.WaitGroup

	feeds          []*models.TwitterFeed
	feedsLock      sync.Mutex
	twitterScraper *twitterscraper.Scraper
}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Twitter",
		SysName:  "twitter",
		Category: common.PluginCategoryFeeds,
	}
}

func RegisterPlugin() {
	if !ShouldRegister() {
		common.GetPluginLogger(&Plugin{}).Warn("Twitter disabled, skipping plugin init...")
		return
	}

	twitterScraper := twitterscraper.New()
	twitterScraper.WithReplies(true)

	twitterProxy := confTwitterProxy.GetString()
	if len(twitterProxy) > 0 {
		twitterScraper.SetProxy(twitterProxy)
	}

	if confTwitterUsername.GetString() != "" {
		err := twitterScraper.Login(confTwitterUsername.GetString(), confTwitterPassword.GetString(), confTwitterEmail.GetString())
		if err != nil {
			logger.WithError(err).Error("Failed initializing TWITTER plugin, probably login error")
			return
		}
		logger.Info("TWITTER plugin - using username to connect...")
	} else {
		logger.Info("TWITTER plugin - using LoginOpenAccount to connect...")
		twitterScraper.LoginOpenAccount()
	}

	p := &Plugin{
		twitterScraper: twitterScraper,
	}

	common.RegisterPlugin(p)
	mqueue.RegisterSource("twitter", p)
	common.InitSchemas("twitter", DBSchemas...)
}

const TwitterIcon = `data:image/png;base64,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`

const XIcon = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAMAAAD04JH5AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAydpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDkuMS1jMDAxIDc5LjE0NjI4OTk3NzcsIDIwMjMvMDYvMjUtMjM6NTc6MTQgICAgICAgICI+IDxyZGY6UkRGIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyI+IDxyZGY6RGVzY3JpcHRpb24gcmRmOmFib3V0PSIiIHhtbG5zOnhtcD0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bXA6Q3JlYXRvclRvb2w9IkFkb2JlIFBob3Rvc2hvcCAyNS4zIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpFMEE1NzE3MkE3RjcxMUVFQjA4RUIwRDk0RUM5RkIxMCIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDpFMEE1NzE3M0E3RjcxMUVFQjA4RUIwRDk0RUM5RkIxMCI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOkUwQTU3MTcwQTdGNzExRUVCMDhFQjBEOTRFQzlGQjEwIiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOkUwQTU3MTcxQTdGNzExRUVCMDhFQjBEOTRFQzlGQjEwIi8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+g3/nLgAAAl5QTFRFAAAA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`

var _ mqueue.PluginWithWebhookAvatar = (*Plugin)(nil)

func (p *Plugin) WebhookAvatar() string {
	return XIcon
}
