package reddit

import (
	"context"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/feeds"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/go-reddit"
	"github.com/mrbentarikau/pagst/reddit/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/oauth2"
)

var (
	confClientID     = config.RegisterOption("yagpdb.reddit.clientid", "Client ID for the reddit api application", "")
	confClientSecret = config.RegisterOption("yagpdb.reddit.clientsecret", "Client Secret for the reddit api application", "")
	confRedirectURI  = config.RegisterOption("yagpdb.reddit.redirect", "Redirect URI for the reddit api application", "")
	confRefreshToken = config.RegisterOption("yagpdb.reddit.refreshtoken", "RefreshToken for the reddit api application, you need to ackquire this manually, should be set to permanent", "")

	confMaxPostsHourFast = config.RegisterOption("yagpdb.reddit.fast_max_posts_hour", "Max posts per hour per guild for fast feed", 60)
	confMaxPostsHourSlow = config.RegisterOption("yagpdb.reddit.slow_max_posts_hour", "Max posts per hour per guild for slow feed", 120)

	feedLock sync.Mutex
	fastFeed *PostFetcher
	slowFeed *PostFetcher
)

func (p *Plugin) StartFeed() {
	go p.runBot()
}

func (p *Plugin) StopFeed(wg *sync.WaitGroup) {
	wg.Add(1)

	feedLock.Lock()

	if fastFeed != nil {
		ff := fastFeed
		go func() {
			ff.StopChan <- wg
		}()
		fastFeed = nil
	} else {
		wg.Done()
	}

	if slowFeed != nil {
		sf := slowFeed
		go func() {
			sf.StopChan <- wg
		}()
		slowFeed = nil
	} else {
		wg.Done()
	}

	feedLock.Unlock()
}

func UserAgent() string {
	return fmt.Sprintf("%s:%s:%s (by /u/caubert)", common.ConfBotName.GetString(), confClientID.GetString(), common.VERSION)
}

func setupClient() *reddit.Client {
	authenticator := reddit.NewAuthenticator(UserAgent(), confClientID.GetString(), confClientSecret.GetString(), confRedirectURI.GetString(),
		"a", reddit.ScopeEdit+" "+reddit.ScopeRead)
	redditClient := authenticator.GetAuthClient(&oauth2.Token{RefreshToken: confRefreshToken.GetString()}, UserAgent())
	return redditClient
}

func (p *Plugin) runBot() {
	feedLock.Lock()

	if os.Getenv("YAGPDB_REDDIT_FAST_FEED_DISABLED") == "" {
		fastFeed = NewPostFetcher(p.redditClient, false, NewPostHandler(false))
		go fastFeed.Run()
	}

	slowFeed = NewPostFetcher(p.redditClient, true, NewPostHandler(true))
	go slowFeed.Run()

	feedLock.Unlock()
}

type KeySlowFeeds string
type KeyFastFeeds string

var configCache sync.Map

type PostHandlerImpl struct {
	Slow        bool
	ratelimiter *Ratelimiter
}

func NewPostHandler(slow bool) PostHandler {
	rl := NewRatelimiter()
	go rl.RunGCLoop()

	return &PostHandlerImpl{
		Slow:        slow,
		ratelimiter: rl,
	}
}

func (p *PostHandlerImpl) HandleRedditPosts(links []*reddit.Link) {
	for _, v := range links {
		if strings.EqualFold(v.Selftext, "[removed]") || strings.EqualFold(v.Selftext, "[deleted]") {
			continue
		}

		if !v.IsRobotIndexable {
			continue
		}

		// since := time.Since(time.Unix(int64(v.CreatedUtc), 0))
		// logger.Debugf("[%5.2fs %6s] /r/%-20s: %s", since.Seconds(), v.ID, v.Subreddit, v.Title)
		go p.handlePost(v, 0)
	}
}

func (p *PostHandlerImpl) getConfigs(subreddit string) ([]*models.RedditFeed, error) {
	var key interface{}
	key = KeySlowFeeds(subreddit)
	if !p.Slow {
		key = KeyFastFeeds(subreddit)
	}

	v, ok := configCache.Load(key)
	if ok {
		return v.(models.RedditFeedSlice), nil
	}

	qms := []qm.QueryMod{
		models.RedditFeedWhere.Subreddit.EQ(strings.ToLower(subreddit)),
		models.RedditFeedWhere.Slow.EQ(p.Slow),
		models.RedditFeedWhere.Disabled.EQ(false),
	}

	config, err := models.RedditFeeds(qms...).AllG(context.Background())
	if err != nil {
		logger.WithError(err).Error("failed retrieving Reddit feeds for subreddit")
		return nil, err
	}

	configCache.Store(key, config)

	return config, nil
}

func (p *PostHandlerImpl) handlePost(post *reddit.Link, filterGuild int64) error {

	// createdSince := time.Since(time.Unix(int64(post.CreatedUtc), 0))
	// logger.Printf("[%5.1fs] /r/%-15s: %s, %s", createdSince.Seconds(), post.Subreddit, post.Title, post.ID)

	config, err := p.getConfigs(strings.ToLower(post.Subreddit))
	if err != nil {
		logger.WithError(err).Error("failed retrieving Reddit feeds for subreddit")
		return err
	}

	if filterGuild > 0 {
		filtered := make([]*models.RedditFeed, 0)
		for _, v := range config {
			if v.GuildID == filterGuild {
				filtered = append(filtered, v)
			}
		}

		config = filtered
	}

	// Get the configs that listens to this subreddit, if any
	filteredItems := p.FilterFeeds(config, post)

	// No channels nothing to do...
	if len(filteredItems) < 1 {
		return nil
	}

	logger.WithFields(logrus.Fields{
		"num_channels": len(filteredItems),
		"subreddit":    post.Subreddit,
	}).Debug("Found matched Reddit post")

	message, embed := CreatePostMessage(post)

	for _, item := range filteredItems {
		idStr := strconv.FormatInt(item.ID, 10)

		webhookUsername := "r/" + post.Subreddit + " • PAGSTDB"

		var content string
		parseMentions := []discordgo.AllowedMentionType{}
		if len(item.MentionRole) > 0 {
			parseMentions = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles}
			content += "Hey <@&" + strconv.FormatInt(item.MentionRole[0], 10) + ">, a new Reddit post!\n"
		}

		qm := &mqueue.QueuedElement{

			GuildID:         item.GuildID,
			ChannelID:       item.ChannelID,
			MessageStr:      content,
			Source:          "reddit",
			SourceItemID:    idStr,
			UseWebhook:      true,
			WebhookUsername: webhookUsername,
			AllowedMentions: discordgo.AllowedMentions{

				Parse: parseMentions,
			},
		}

		var matureContentWarning string
		if post.Over18 && item.FilterNSFW == FilterNSFWNone {
			matureContentWarning = "**Mature Content Warning**\n\n"
			if item.UseEmbeds {
				embed.Color = 0xc51717
			}
		}
		if item.UseEmbeds {
			qm.MessageEmbed = embed
			qm.MessageEmbed.Description = matureContentWarning + embed.Description
		} else {
			qm.MessageStr += matureContentWarning + message
		}

		mqueue.QueueMessage(qm)

		feeds.MetricPostedMessages.With(prometheus.Labels{"source": "reddit"}).Inc()
		go analytics.RecordActiveUnit(item.GuildID, &Plugin{}, "posted_reddit_message")
	}

	return nil
}

func (p *PostHandlerImpl) FilterFeeds(feeds []*models.RedditFeed, post *reddit.Link) []*models.RedditFeed {
	filteredItems := make([]*models.RedditFeed, 0, len(feeds))

OUTER:
	for _, c := range feeds {
		// remove duplicates
		for _, v := range filteredItems {
			if v.ChannelID == c.ChannelID {
				continue OUTER
			}
		}

		limit := confMaxPostsHourFast.GetInt()
		if p.Slow {
			limit = confMaxPostsHourSlow.GetInt()
		}

		// apply ratelimiting
		if !p.ratelimiter.CheckIncrement(time.Now(), c.GuildID, limit) {
			continue
		}

		if post.Over18 && c.FilterNSFW == FilterNSFWIgnore {
			// NSFW and we ignore nsfw posts
			continue
		} else if !post.Over18 && c.FilterNSFW == FilterNSFWRequire {
			// Not NSFW and we only care about nsfw posts
			continue
		}

		if p.Slow {
			if post.Score < c.MinUpvotes {
				// less than required upvotes
				continue
			}
		}

		filteredItems = append(filteredItems, c)
	}

	return filteredItems
}

func CreatePostMessage(post *reddit.Link) (string, *discordgo.MessageEmbed) {
	plainMessage := fmt.Sprintf("**%s**\n*by %s (<%s>)*\n",
		html.UnescapeString(post.Title), post.Author, "https://redd.it/"+post.ID)

	plainBody := ""
	parentSpoiler := false

	if post.IsSelf {
		postSelftext := strings.ReplaceAll(post.Selftext, "&amp;#x200B;", " ")
		plainBody = common.CutStringShort(html.UnescapeString(postSelftext), 250)
	} else if post.CrosspostParent != "" && len(post.CrosspostParentList) > 0 {
		// Handle cross posts
		parent := post.CrosspostParentList[0]
		plainBody += "**" + html.UnescapeString(parent.Title) + "**\n"

		if parent.IsSelf {
			parentSelftext := strings.ReplaceAll(post.CrosspostParentList[0].Selftext, "&amp;#x200B;", " ")
			plainBody += common.CutStringShort(html.UnescapeString(parentSelftext), 250)
		} else {
			plainBody += parent.URL
		}

		if parent.Spoiler {
			parentSpoiler = true
		}
	} else {
		plainBody = post.URL
	}

	if post.Spoiler || parentSpoiler {
		plainMessage += "|| " + plainBody + " ||"
	} else {
		plainMessage += plainBody
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:  "https://reddit.com/u/" + post.Author,
			Name: fmt.Sprint("u/", post.Author),
		},
		Provider: &discordgo.MessageEmbedProvider{
			Name: "Reddit",
			URL:  "https://reddit.com",
		},
		//Description: "**" + html.UnescapeString(post.Title) + "**\n",
		Description: "",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Type: ",
		},
		Timestamp: time.Unix(int64(post.CreatedUtc), 0).UTC().Format(time.RFC3339),
	}

	if post.Spoiler {
		embed.Title = " [spoiler]\n"
	}

	embed.Title += common.CutStringShort(html.UnescapeString(post.Title), 240)
	embed.URL = "https://redd.it/" + post.ID

	if post.IsSelf {
		//  Handle Self posts
		embed.Footer.Text += "new self post"
		postSelftext := strings.ReplaceAll(post.Selftext, "&amp;#x200B;", " ")
		if post.Spoiler {
			embed.Description += "|| " + common.CutStringShort(html.UnescapeString(postSelftext), 250) + " ||"
		} else {
			embed.Description += common.CutStringShort(html.UnescapeString(postSelftext), 250)
		}

		embed.Color = 0xc3fc7e
	} else if post.CrosspostParent != "" && len(post.CrosspostParentList) > 0 {
		//  Handle crossposts
		embed.Footer.Text += "new crosspost"

		parent := post.CrosspostParentList[0]
		embed.Description += "**" + html.UnescapeString(parent.Title) + "**\n"
		if parent.IsSelf {
			// Cropsspost was a self post
			embed.Color = 0xc3fc7e
			parentSelftext := strings.ReplaceAll(post.CrosspostParentList[0].Selftext, "&amp;#x200B;", " ")
			if parent.Spoiler {
				embed.Description += "|| " + common.CutStringShort(html.UnescapeString(parentSelftext), 250) + " ||"
			} else {
				embed.Description += common.CutStringShort(html.UnescapeString(parentSelftext), 250)
			}
		} else {
			// cross post was a link most likely
			embed.Color = 0x88c0d0
			embed.Description += parent.URL
			if parent.Media.Type == "" && !parent.Spoiler && parent.PostHint == "image" {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: parent.URL,
				}
			}
		}
	} else {
		//  Handle Link posts
		embed.Color = 0x88c0d0
		embed.Footer.Text += "new link post"
		embed.Description += post.URL

		if post.Media.Type == "" && !post.Spoiler && post.PostHint == "image" {
			embed.Image = &discordgo.MessageEmbedImage{
				URL: post.URL,
			}
		}
	}

	return plainMessage, embed
}

type RedditIdSlice []string

// Len is the number of elements in the collection.
func (r RedditIdSlice) Len() int {
	return len(r)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (r RedditIdSlice) Less(i, j int) bool {
	a, err1 := strconv.ParseInt(r[i], 36, 64)
	b, err2 := strconv.ParseInt(r[j], 36, 64)
	if err1 != nil {
		logger.WithError(err1).Error("Failed parsing id")
	}
	if err2 != nil {
		logger.WithError(err2).Error("Failed parsing id")
	}

	return a > b
}

// Swap swaps the elements with indexes i and j.
func (r RedditIdSlice) Swap(i, j int) {
	temp := r[i]
	r[i] = r[j]
	r[j] = temp
}
