package twitter

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/feeds"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/premium"
	"github.com/mrbentarikau/pagst/twitter/models"
	"github.com/mediocregopher/radix/v3"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var _ feeds.Plugin = (*Plugin)(nil)

func KeyLastTweetTime(id string) string { return "twitter_last_tweet_time:" + id }
func KeyLastTweetID(id string) string   { return "twitter_last_tweet_id:" + id }

func (p *Plugin) StartFeed() {
	logrus.Info("STARTING TWITTER FEED")
	p.Stop = make(chan *sync.WaitGroup)
	go p.updateConfigsLoop()
	go p.runFeedLoop()
}

func (p *Plugin) StopFeed(wg *sync.WaitGroup) {

	if p.Stop != nil {
		wg.Add(1)
		p.twitterScraper.Logout()
		p.Stop <- wg
		p.Stop <- wg
	} else {
		p.twitterScraper.Logout()
		wg.Done()
	}
}

func (p *Plugin) runFeedLoop() {
	logrus.Info("STARTING TWITTER FEED LOOP")
	ticker := time.NewTicker(time.Minute * time.Duration(confTwitterPollFrequency.GetInt()))
	startDelay := time.After(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			p.feedsLock.Lock()
			newFeeds := p.feeds
			p.feedsLock.Unlock()
			p.runFeed(newFeeds)
		case <-startDelay:
		case wg := <-p.Stop:
			wg.Done()
			return
		}
	}
}

func (p *Plugin) getLastTweetInfo(username string) (tweetId string, tweetTime time.Time, err error) {
	// Find the last video time for this channel
	var unixSeconds int64
	err = common.RedisPool.Do(radix.Cmd(&unixSeconds, "GET", KeyLastTweetTime(username)))

	var lastProcessedTweetTime time.Time
	if err != nil || unixSeconds == 0 {
		lastProcessedTweetTime = time.Time{}
	} else {
		lastProcessedTweetTime = time.Unix(unixSeconds, 0)
	}

	var lastTweetID string
	err = common.RedisPool.Do(radix.Cmd(&lastTweetID, "GET", KeyLastTweetID(username)))
	return lastTweetID, lastProcessedTweetTime, err
}

func (p *Plugin) checkTweet(tweet *twitterscraper.Tweet) *twitterscraper.Tweet {
	lastTweetID, lastTweetTime, err := p.getLastTweetInfo(tweet.Username)
	if err != nil {
		logrus.WithError(err).Errorf("Failed getting last tweet info for username %s", tweet.Username)
		return nil
	}

	if lastTweetID == tweet.ID {
		// the tweet has already been processed
		return nil
	}

	if time.Since(tweet.TimeParsed) > time.Hour {
		// just a safeguard against empty last tweet time's
		return nil
	}

	if lastTweetTime.After(tweet.TimeParsed) {
		// wasn't a new tweet
		return nil
	}

	// This is a new tweet, post it
	// p.handleTweet(tweet)
	return tweet
}

func (p *Plugin) getTweetsForUser(username string, attempt int, delay time.Duration) {
	var tweets = make([]*twitterscraper.Tweet, 0)
	logrus.Infof("Getting tweets for user %s", username)
	for tweet := range p.twitterScraper.GetTweets(context.Background(), username, 50) {
		if tweet.Error != nil {
			errString := tweet.Error.Error()
			isNotFound := strings.Contains(errString, "not found")
			isSuspended := strings.Contains(errString, "User has been suspended")
			if isNotFound || isSuspended {
				_, err := models.TwitterFeeds(models.TwitterFeedWhere.TwitterUsername.EQ(username)).UpdateAllG(context.Background(), models.M{"enabled": false})
				if err != nil {
					logrus.WithError(err).Errorf("Failed suspending feed for user %s", username)
				} else {
					logrus.WithError(tweet.Error).Errorf("Disabled feed for %s", username)
				}
			} else {
				logrus.WithError(tweet.Error).Errorf("Failed getting tweets for user %s, ", username)
				if attempt < 3 {
					logrus.Infof("Retrying to get tweets for user %s with attempt %d and delay of %d seconds", username, attempt+1, delay)
					time.Sleep(delay * time.Second)
					//retry if ratelimited after delay
					go p.getTweetsForUser(username, attempt+1, 2*delay)
				}
			}

			break
		}
		tweets = append(tweets, p.checkTweet(&tweet.Tweet))
	}

	sort.SliceStable(tweets, func(i, j int) bool {
		if tweets[i] != nil && tweets[j] != nil {
			return tweets[i].TimeParsed.Before(tweets[j].TimeParsed)
		} else if tweets[i] != nil {
			return false
		} else if tweets[j] != nil {
			return true
		} else {
			return false
		}
	})

	for _, t := range tweets {
		if t != nil {
			p.handleTweet(t)
		}
	}
}

func (p *Plugin) runFeed(feeds []*models.TwitterFeed) {
	uniqueFeeds := make(map[string]int)
	for _, v := range feeds {
		if uniqueFeeds[v.TwitterUsername] == 0 {
			uniqueFeeds[v.TwitterUsername] = 1
		}
		uniqueFeeds[v.TwitterUsername]++
	}

	logger.Info("NUMBER OF Unique Twitter Feeds: ", len(uniqueFeeds))
	batchSize := confTwitterBatchSize.GetInt()
	batches := make([][]string, 0)
	currentChunk := make([]string, 0, batchSize)
	for user := range uniqueFeeds {
		currentChunk = append(currentChunk, user)
		if len(currentChunk) == batchSize {
			batches = append(batches, currentChunk)
			currentChunk = make([]string, 0, batchSize)
		}
	}
	if len(currentChunk) > 0 {
		batches = append(batches, currentChunk)
	}

	for idx, batch := range batches {
		logrus.Infof("Running batch %d of %d for twitter feeds", idx+1, len(batches))
		for _, user := range batch {
			go p.getTweetsForUser(user, 0, 10)
		}
		time.Sleep(time.Duration(confTwitterBatchDelay.GetInt()) * time.Second)
	}

}

func (p *Plugin) handleTweet(t *twitterscraper.Tweet) {
	if t.UserID == "" {
		logger.Errorf("Twitter user is nil?: %#v", t)
		return
	}

	p.feedsLock.Lock()
	tFeeds := p.feeds
	p.feedsLock.Unlock()

	relevantFeeds := make([]*models.TwitterFeed, 0, 4)

OUTER:
	for _, f := range tFeeds {
		tweetUser, _ := strconv.ParseInt(t.UserID, 10, 64)
		if tweetUser != f.TwitterUserID {
			continue
		}

		for _, r := range relevantFeeds {
			// skip multiple feeds to the same channel
			if f.ChannelID == r.ChannelID {
				continue OUTER
			}
		}

		isRetweet := t.RetweetedStatus != nil
		if isRetweet && !f.IncludeRT {
			continue
		}

		if t.IsReply && !f.IncludeReplies {
			continue
		}

		relevantFeeds = append(relevantFeeds, f)
	}

	if len(relevantFeeds) < 1 {
		return
	}
	err := common.RedisPool.Do(radix.FlatCmd(nil, "SET", KeyLastTweetTime(t.Username), time.Now().Unix()))
	if err != nil {
		logrus.WithError(err).Errorf("Failed Saving tweet time for %s ", t.UserID)
		return
	}

	err = common.RedisPool.Do(radix.FlatCmd(nil, "SET", KeyLastTweetID(t.Username), t.ID))
	if err != nil {
		logrus.WithError(err).Errorf("Failed Saving tweet id for %s ", t.Username)
		return
	}

	user, err := p.twitterScraper.GetProfile(t.Username)
	if err != nil {
		logrus.WithError(err).Errorf("Failed getting user info for userID %s", t.Username)
	}
	webhookUsername := "Twitter • " + common.ConfBotName.GetString()
	embed := p.createTweetEmbed(t, &user)
	for _, v := range relevantFeeds {
		go analytics.RecordActiveUnit(v.GuildID, p, "posted_twitter_message")

		parseMentions := []discordgo.AllowedMentionType{}
		var content string
		if len(v.MentionRole) > 0 {
			parseMentions = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles}
			content = fmt.Sprintf("Hey <@&%d>, a new tweet!", v.MentionRole[0])
		}

		mqueue.QueueMessage(&mqueue.QueuedElement{
			Source:       "twitter",
			SourceItemID: strconv.FormatInt(v.ID, 10),

			GuildID:   v.GuildID,
			ChannelID: v.ChannelID,

			MessageStr:      content,
			MessageEmbed:    embed,
			UseWebhook:      true,
			WebhookUsername: webhookUsername,

			AllowedMentions: discordgo.AllowedMentions{
				Parse: parseMentions,
			},

			Priority: 5, // above youtube and reddit
		})
	}

	feeds.MetricPostedMessages.With(prometheus.Labels{"source": "twitter"}).Add(float64(len(relevantFeeds)))

	logger.Infof("Handled tweet %q from %s on %d channels", t.Text, t.Username, len(relevantFeeds))
}

func (p *Plugin) createTweetEmbed(tweet *twitterscraper.Tweet, user *twitterscraper.Profile) *discordgo.MessageEmbed {
	timeStr := time.Unix(tweet.Timestamp, 0).Format(time.RFC3339)

	var text string

	if tweet.RetweetedStatus != nil {
		text += fmt.Sprintf("[Retweet:](https://twitter.com/%s/status/%s) ", tweet.RetweetedStatus.Username, tweet.RetweetedStatus.ID)
	} else if tweet.InReplyToStatus != nil {
		text += fmt.Sprintf("[Reply:](https://twitter.com/%s/status/%s) ", tweet.InReplyToStatus.Username, tweet.InReplyToStatus.ID)
	} else if tweet.QuotedStatus != nil {
		text += fmt.Sprintf("[Quote:](https://twitter.com/%s/status/%s) ", tweet.QuotedStatus.Username, tweet.QuotedStatus.ID)
	}

	text += tweet.Text

	for _, ht := range tweet.Hashtags {
		text = strings.Replace(text, (`#` + ht), fmt.Sprintf("[#%[1]s](https://twitter.com/hashtag/%[1]s?src=hashtag_click)", ht), -1)
		// just in case
		//re := regexp.MustCompile(`#` + ht)
		//text = re.ReplaceAllString(text, fmt.Sprintf("[#%[1]s](https://twitter.com/hashtag/%[1]s?src=hashtag_click)", ht))
	}

	author := &discordgo.MessageEmbedAuthor{
		Name: "@" + tweet.Username,
		URL:  tweet.PermanentURL,
	}
	if user != nil {
		author.IconURL = user.Avatar
	}
	embed := &discordgo.MessageEmbed{
		Author:      author,
		Description: html.UnescapeString(text),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://abs.twimg.com/icons/apple-touch-icon-192x192.png",
			Text:    "Twitter"},
		Timestamp: timeStr,
		Color:     0x38A1F3,
	}

	var img string
	if tweet.Photos != nil && len(tweet.Photos) > 0 {
		img = tweet.Photos[0].URL
	} else if tweet.QuotedStatus != nil && (tweet.QuotedStatus.Photos != nil && len(tweet.QuotedStatus.Photos) > 0) {
		img = tweet.QuotedStatus.Photos[0].URL
	}

	embed.Image = &discordgo.MessageEmbedImage{
		URL: img,
	}

	return embed
}

func (p *Plugin) updateConfigsLoop() {
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()
	for {
		p.updateConfigs()

		select {
		case <-ticker.C:
		case wg := <-p.Stop:
			wg.Done()
			logger.Info("Twitter updateConfigsLoop shut down")
			return
		}
	}
}

func (p *Plugin) updateConfigs() {
	configs, err := models.TwitterFeeds(models.TwitterFeedWhere.Enabled.EQ(true), qm.OrderBy("id asc")).AllG(context.Background())
	if err != nil {
		logger.WithError(err).Error("failed updating configs")
		return
	}

	filtered := make([]*models.TwitterFeed, 0, len(configs))
	for _, v := range configs {
		isPremium, err := premium.IsGuildPremium(v.GuildID)
		if err != nil {
			logger.WithError(err).Error("failed checking if guild is premium")
			return
		}

		if !isPremium {
			v.Enabled = false
			_, err = v.UpdateG(context.Background(), boil.Whitelist("enabled"))
			if err != nil {
				logger.WithError(err).Error("failed disabling non-premium feed")
			}
			continue
		}

		filtered = append(filtered, v)
	}

	p.feedsLock.Lock()
	p.feeds = filtered
	p.feedsLock.Unlock()
}
