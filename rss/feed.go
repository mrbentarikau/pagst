package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/common/templates"
	"github.com/mrbentarikau/pagst/feeds"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/rss/models"
	"github.com/mediocregopher/radix/v3"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const GoFeedQueryInterval = time.Minute * 5

func (p *Plugin) StartFeed() {
	p.Stop = make(chan *sync.WaitGroup)
	p.runRSSFeedInit()
}

func (p *Plugin) StopFeed(wg *sync.WaitGroup) {
	if p.Stop != nil {
		p.Stop <- wg
	} else {
		wg.Done()
	}
}

// keeps the subscriptions up to date by updating the ones soon to be expiring
func (p *Plugin) runRSSFeedInit() {
	goFeedQueryTicker := time.NewTicker(GoFeedQueryInterval)
	startDelay := time.After(time.Second * 2)

	for {
		select {
		case <-startDelay:
			p.checkGoFeed()
		case wg := <-p.Stop:
			wg.Done()
			return
		case <-goFeedQueryTicker.C:
			p.checkGoFeed()
		}
	}
}

func (p *Plugin) checkGoFeed() {
	activeRSSFeeds, err := models.RSSFeeds(qm.Distinct("feed_url")).AllG(context.Background())
	if err != nil {
		logger.WithError(err).Error("Failed syncing WebSubs, failed retrieving subbed RSS feeds")
		return
	}

	for _, feed := range activeRSSFeeds {
		rssClient := p.rssClient
		parsedFeed, err := rssClient.ParseURL(feed.FeedURL)
		if err != nil {
			logger.WithError(err).WithField("feed_url", feed.FeedURL).Warn("goFeed fetching RSS feed erred")
			return
		}
		go p.CheckRSSFeed(nil, parsedFeed, feed.FeedURL)
	}
}

func (p *Plugin) AddRSSFeed(guildID, discordChannelID, mentionRole int64, feedName, feedTitle, feedURL string) (*models.RSSFeed, error) {
	feedSub := &models.RSSFeed{
		GuildID:     guildID,
		ChannelID:   discordChannelID,
		MentionRole: mentionRole,
		Enabled:     true,
		FeedName:    feedName,
		FeedTitle:   feedTitle,
		FeedURL:     feedURL,
	}

	err := common.BlockingLockRedisKey(RedisRSSFeedLockKey, 0, 10)
	if err != nil {
		return nil, err
	}
	defer common.UnlockRedisKey(RedisRSSFeedLockKey)

	err = feedSub.InsertG(context.Background(), boil.Infer())
	if err != nil {
		return nil, err
	}

	return feedSub, nil
}

func (p *Plugin) CheckRSSFeed(feedByte []byte, directFeed *gofeed.Feed, feedURL string) error {
	parsedFeedURLProtocol := common.LinkRegexBotlabs.FindString(feedURL)
	subs, err := models.RSSFeeds(qm.Where("feed_url = ?", parsedFeedURLProtocol)).AllG(context.Background())
	if err != nil || len(subs) < 1 {
		return err
	}

	feed := directFeed

	if feedByte != nil {
		feedParser := p.rssClient
		toFeedParser := strings.NewReader(string(feedByte))
		feed, err = feedParser.Parse(toFeedParser)
		if err != nil {
			return err
		}
	}

	// no feed items to parse
	if len(feed.Items) == 0 {
		return nil
	}

	parsedFeedURL := removeProtocol(parsedFeedURLProtocol)
	lastRSSLink, lastRSSTime, err := p.getLastRSSFeedTimes(parsedFeedURL)
	if err != nil {
		return err
	}

	if lastRSSLink == feed.Items[0].Link {
		// the feed was already posted and was probably just edited
		return nil
	}

	parsedPublishedAt := feed.Items[0].PublishedParsed
	if parsedPublishedAt == nil {
		// the feed does not have any time information
		return nil
	}

	if time.Since(*parsedPublishedAt) > time.Hour && lastRSSLink != "" {
		// just a safeguard against empty parsedPublishedAt
		return nil
	}

	if lastRSSTime.After(*parsedPublishedAt) {
		// wasn't a new feed
		return nil
	}

	// This is a new feed, post it
	return p.postRSSFeed(subs, parsedPublishedAt, parsedFeedURLProtocol, feed)

}

func (p *Plugin) postRSSFeed(subs models.RSSFeedSlice, publishedAt *time.Time, feedURL string, feed *gofeed.Feed) error {
	var itemLink string
	if len(feed.Items) > 0 {
		itemLink = feed.Items[0].Link
	}
	err := common.MultipleCmds(
		radix.FlatCmd(nil, "SET", KeyLastRSSFeedTime(removeProtocol(feedURL)), publishedAt.Unix()),
		radix.FlatCmd(nil, "SET", KeyLastRSSFeedLink(removeProtocol(feedURL)), itemLink),
	)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.Enabled {
			p.sendNewRSSFeedMessage(sub.GuildID, sub.ChannelID, sub.MentionRole, feed, feedURL, sub.FeedName)
		}
	}

	return nil
}

func (p *Plugin) sendNewRSSFeedMessage(guildID, channelID, mentionRole int64, feed *gofeed.Feed, feedURL, feedName string) {
	var content string
	var rssEmbed *discordgo.MessageEmbed

	guildState := bot.State.GetGuild(guildID) //GuildSet
	if guildState == nil {
		logger.Error("sendNewVidMessage for guild not in state")
		return
	}

	var rssFeedTemplate RSSFeedTemplate
	j, err := json.Marshal(&feed)
	if err != nil {
		logger.Error("rssFeed marshalling error")
		return
	}
	err = json.Unmarshal(j, &rssFeedTemplate)
	if err != nil {
		logger.Error("rssFeed unMarshalling error")
		return
	}

	if len(feed.Items) > 0 {
		rssFeedTemplate.Item = rssFeedTemplate.Items[0]
	}

	ctx := templates.NewContext(guildState, guildState.GetChannel(channelID), nil) //needs GuildSet, ChannelState, MemberState
	ctx.Data["RSSFeed"] = rssFeedTemplate
	ctx.Data["RSSFeedItem"] = rssFeedTemplate.Item
	ctx.Data["RSSName"] = feedName
	ctx.Data["SelectedRoleID"] = mentionRole

	var announceMsg AnnouncementMessage
	err = common.GetRedisJson(RedisRSSAnnounceMessage+":"+discordgo.StrID(guildID), &announceMsg)
	if err != nil {
		return
	}

	rssEmbed = createRSSEmbed(feed, feedName)

	webhookUsername := "RSS Feed • " + common.ConfBotName.GetString()
	// not really good using current mqueue-webhooking
	if (feedName != "" && feedName != "No name") && !announceMsg.Enabled {
		webhookUsername = "RSS Feed • " + feedName
		rssEmbed.Footer.Text += " • " + common.ConfBotName.GetString()
	}

	webhookAvatarURL := ""
	if feed.Image != nil {
		webhookAvatarURL = feed.Image.URL
	}

	if announceMsg.Enabled {
		rssEmbed = nil
		content, err = ctx.Execute(announceMsg.AnnounceMsg)
		if err != nil {
			logger.WithError(err).WithField("guild", guildID).Warn("Failed executing template on sendNewRSSFeedMessage")
			return
		}
		if content == "" { // Nothing to do
			return
		}
	} else if mentionRole > 0 {
		content += fmt.Sprintf("Hey <@&%d>, incoming RSS.\n", mentionRole)
	}

	parseMentions := []discordgo.AllowedMentionType{}
	if mentionRole > 0 {
		parseMentions = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles}
	}

	go analytics.RecordActiveUnit(guildID, p, "posted_rssfeeds_message")

	qm := &mqueue.QueuedElement{
		GuildID:          guildID,
		ChannelID:        channelID,
		Source:           "rss",
		SourceItemID:     "",
		MessageStr:       content,
		MessageEmbed:     rssEmbed,
		UseWebhook:       true,
		WebhookUsername:  webhookUsername,
		WebhookAvatarURL: webhookAvatarURL,
		Priority:         3,
		AllowedMentions: discordgo.AllowedMentions{
			Parse: parseMentions,
		},
	}

	/*if announceMsg.Enabled {
		qm.UseWebhook = false
	}*/

	mqueue.QueueMessage(qm)

	feeds.MetricPostedMessages.With(prometheus.Labels{"source": "rssfeeds"}).Inc()
}

func createRSSEmbed(feed *gofeed.Feed, feedName string) *discordgo.MessageEmbed {
	var feedItem *gofeed.Item
	var feedAuthor string

	if len(feed.Items) > 0 {
		feedItem = feed.Items[0]
	}

	feedTime := feedItem.PublishedParsed
	if len(feedItem.Authors) > 0 {
		feedAuthor = feedItem.Authors[0].Name
	} else {
		feedAuthor = feed.Title

	}
	bm := bluemonday.StripTagsPolicy()
	feedDescription := common.CutStringShort(feedItem.Description, 4000)

	if feedName == "No name" || feedName == "" {
		feedName = feed.Title
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "RSS Feed: " + feedName,
			URL:  feed.FeedLink,
		},
		Title:       common.CutStringShort(html.UnescapeString(feedItem.Title), 240),
		URL:         feedItem.Link,
		Description: html.UnescapeString(bm.Sanitize(feedDescription)),
		Timestamp:   feedTime.Format(time.RFC3339),
		Color:       0xfba114,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "By: " + common.CutStringShort(html.UnescapeString(feedAuthor), 240),
		},
	}

	if feedItem.Image != nil {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: feedItem.Image.URL,
		}
	}

	if feed.Image != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: feed.Image.URL,
		}
	}

	return embed
}

func (p *Plugin) getLastRSSFeedTimes(feedURL string) (lastRSSLink string, lastRSSTime time.Time, err error) {
	// Find the last time feed was posted
	var unixSeconds int64
	err = common.RedisPool.Do(radix.Cmd(&unixSeconds, "GET", KeyLastRSSFeedTime(feedURL)))

	var lastProcessedRSSTime time.Time
	if err != nil || unixSeconds == 0 {
		lastProcessedRSSTime = time.Time{}
	} else {
		lastProcessedRSSTime = time.Unix(unixSeconds, 0)
	}

	var lastProcessedRSSLink string
	err = common.RedisPool.Do(radix.Cmd(&lastProcessedRSSLink, "GET", KeyLastRSSFeedLink(feedURL)))
	return lastProcessedRSSLink, lastProcessedRSSTime, err
}

func removeProtocol(url string) string {
	var replaceHTTPS = regexp.MustCompile(`http(s)?:\/\/`)
	return replaceHTTPS.ReplaceAllString(url, "")
}
