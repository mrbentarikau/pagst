package youtube

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/analytics"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/common/templates"
	"github.com/mrbentarikau/pagst/feeds"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mediocregopher/radix/v3"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	MaxChannelsPerPoll  = 30
	PollInterval        = time.Second * 10
	WebSubCheckInterval = time.Second * 10
	// PollInterval = time.Second * 5 // <- used for debug purposes
)

func (p *Plugin) StartFeed() {
	p.Stop = make(chan *sync.WaitGroup)
	p.runWebsubChecker()
}

func (p *Plugin) StopFeed(wg *sync.WaitGroup) {

	if p.Stop != nil {
		p.Stop <- wg
	} else {
		wg.Done()
	}
}

func (p *Plugin) SetupClient() error {
	/*httpClient, err := google.DefaultClient(context.Background(), youtube.YoutubeScope)
	if err != nil {
		return common.ErrWithCaller(err)
	}*/

	//yt, err := youtube.New(httpClient)
	yt, err := youtube.NewService(context.Background(), option.WithScopes(youtube.YoutubeScope))
	if err != nil {
		return common.ErrWithCaller(err)
	}

	p.YTService = yt

	return nil
}

// keeps the subscriptions up to date by updating the ones soon to be expiring
func (p *Plugin) runWebsubChecker() {
	p.syncWebSubs()

	websubTicker := time.NewTicker(WebSubCheckInterval)
	for {
		select {
		case wg := <-p.Stop:
			wg.Done()
			return
		case <-websubTicker.C:
			p.checkExpiringWebsubs()
		}
	}
}

func (p *Plugin) checkExpiringWebsubs() {
	err := common.BlockingLockRedisKey(RedisChannelsLockKey, 0, 10)
	if err != nil {
		logger.WithError(err).Error("Failed locking channels lock")
		return
	}
	defer common.UnlockRedisKey(RedisChannelsLockKey)

	maxScore := time.Now().Unix()

	var expiring []string
	err = common.RedisPool.Do(radix.FlatCmd(&expiring, "ZRANGEBYSCORE", RedisKeyWebSubChannels, "-inf", maxScore))
	if err != nil {
		logger.WithError(err).Error("Failed checking websubs")
		return
	}

	for _, v := range expiring {
		err := p.WebSubSubscribe(v)
		if err != nil {
			logger.WithError(err).WithField("yt_channel", v).Error("Failed subscribing to channel")
		}
		time.Sleep(time.Second)
	}
}

func (p *Plugin) syncWebSubs() {
	var activeChannels []string
	err := common.SQLX.Select(&activeChannels, "SELECT DISTINCT(youtube_channel_id) FROM youtube_channel_subscriptions")
	if err != nil {
		logger.WithError(err).Error("Failed syncing websubs, failed retrieving subbed channels")
		return
	}

	common.RedisPool.Do(radix.WithConn(RedisKeyWebSubChannels, func(client radix.Conn) error {

		locked := false

		for _, channel := range activeChannels {
			if !locked {
				err := common.BlockingLockRedisKey(RedisChannelsLockKey, 0, 5000)
				if err != nil {
					logger.WithError(err).Error("Failed locking channels lock")
					return err
				}
				locked = true
			}

			mn := radix.MaybeNil{}
			client.Do(radix.Cmd(&mn, "ZSCORE", RedisKeyWebSubChannels, channel))
			if mn.Nil {
				// Not added
				err := p.WebSubSubscribe(channel)
				if err != nil {
					logger.WithError(err).WithField("yt_channel", channel).Error("Failed subscribing to channel")
				}

				common.UnlockRedisKey(RedisChannelsLockKey)
				locked = false

				time.Sleep(time.Second)
			}
		}

		if locked {
			common.UnlockRedisKey(RedisChannelsLockKey)
		}

		return nil
	}))
}

func (p *Plugin) sendNewVidMessage(guild, discordChannel, channelTitle, channelID, videoID, mentionRole, liveBroadcastContent string, mentionEveryone bool) {
	var content string

	parsedChannel, _ := strconv.ParseInt(discordChannel, 10, 64)
	parsedGuild, _ := strconv.ParseInt(guild, 10, 64)
	parseMentionRole, _ := strconv.ParseInt(mentionRole, 10, 64)

	guildState := bot.State.GetGuild(parsedGuild) //GuildSet
	if guildState == nil {
		logger.Error("sendNewVidMessage for guild not in state")
		return
	}

	videoYT := "https://www.youtube.com/watch?v=" + videoID
	ctx := templates.NewContext(guildState, guildState.GetChannel(parsedChannel), nil) //needs GuildSet, ChannelState, MemberState
	ctx.Data["URL"] = "https://www.youtube.com/watch?v=" + videoID
	ctx.Data["ChannelName"] = channelTitle
	ctx.Data["VideoID"] = videoID
	ctx.Data["ChannelID"] = channelID
	ctx.Data["LiveStream"] = liveBroadcastContent
	ctx.Data["SelectedRoleID"] = parseMentionRole

	var dbAnnounceMsg YoutubeAnnouncements
	var announceMsg AnnouncementMessage

	err := common.GORM.Model(&YoutubeAnnouncements{}).Where("guild_id = ?", parsedGuild).First(&dbAnnounceMsg).Error
	if err == nil {
		announceMsg.Enabled = dbAnnounceMsg.Enabled
		announceMsg.AnnounceMsg = dbAnnounceMsg.Announcement
	}

	if announceMsg.Enabled {
		content, err = ctx.Execute(announceMsg.AnnounceMsg)
		if err != nil {
			logger.WithError(err).WithField("guild", parsedGuild).Warn("Failed executing template on sendNewVidMessage")
			return
		}
		if content == "" { // Nothing to do
			return
		}
	} else {
		if mentionEveryone {
			if liveBroadcastContent != "none" {
				content += fmt.Sprintf("Hey @everyone, incoming %s YouTube video by **%s**!\n%s\n", liveBroadcastContent, channelTitle, videoYT)
			} else {
				content += fmt.Sprintf("Hey @everyone, incoming YouTube video by **%s**!\n%s\n", channelTitle, videoYT)
			}
		} else if parseMentionRole > 0 {
			if liveBroadcastContent != "none" {
				content += fmt.Sprintf("Hey <@&%s>, incoming %s YouTube video by **%s**!\n%s\n", mentionRole, liveBroadcastContent, channelTitle, videoYT)
			} else {
				content += fmt.Sprintf("Hey <@&%s>, incoming YouTube video by **%s!**\n%s\n", mentionRole, channelTitle, videoYT)
			}
		} else {
			content += fmt.Sprintf("**%s** uploaded a new YouTube video!\n%s", channelTitle, videoYT)
		}
	}

	parseMentions := []discordgo.AllowedMentionType{}
	if mentionEveryone {
		parseMentions = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone}
	} else if parseMentionRole > 0 {
		parseMentions = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles}
	}

	go analytics.RecordActiveUnit(parsedGuild, p, "posted_youtube_message")
	feeds.MetricPostedMessages.With(prometheus.Labels{"source": "youtube"}).Inc()

	mqueue.QueueMessage(&mqueue.QueuedElement{
		GuildID:      parsedGuild,
		ChannelID:    parsedChannel,
		Source:       "youtube",
		SourceItemID: "",
		MessageStr:   content,
		Priority:     2,
		AllowedMentions: discordgo.AllowedMentions{
			Parse: parseMentions,
		},
	})
}

var (
	ErrIDNotFound = errors.New("ID not found")
)

func SubsForChannel(channel string) (result []*ChannelSubscription, err error) {
	err = common.GORM.Where("youtube_channel_id = ?", channel).Find(&result).Error
	return
}

var (
	ErrNoChannel = errors.New("no channel with that id found")
)

func (p *Plugin) parseYtUrl(url string) (t ytUrlType, id string, err error) {
	if ytVideoUrlRegex.MatchString(url) {
		capturingGroups := ytVideoUrlRegex.FindAllStringSubmatch(url, -1)
		if len(capturingGroups) > 0 && len(capturingGroups[0]) > 0 && len(capturingGroups[0][4]) > 0 {
			return ytUrlTypeVideo, capturingGroups[0][4], nil
		}
	} else if ytChannelUrlRegex.MatchString(url) {
		capturingGroups := ytChannelUrlRegex.FindAllStringSubmatch(url, -1)
		if len(capturingGroups) > 0 && len(capturingGroups[0]) > 0 && len(capturingGroups[0][5]) > 0 {
			return ytUrlTypeChannel, capturingGroups[0][5], nil
		}
	} else if ytCustomUrlRegex.MatchString(url) {
		capturingGroups := ytCustomUrlRegex.FindAllStringSubmatch(url, -1)
		if len(capturingGroups) > 0 && len(capturingGroups[0]) > 0 && len(capturingGroups[0][5]) > 0 {
			return ytUrlTypeCustom, capturingGroups[0][5], nil
		}
	} else if ytUserUrlRegex.MatchString(url) {
		capturingGroups := ytUserUrlRegex.FindAllStringSubmatch(url, -1)
		if len(capturingGroups) > 0 && len(capturingGroups[0]) > 0 && len(capturingGroups[0][5]) > 0 {
			return ytUrlTypeUser, capturingGroups[0][5], nil
		}
	}
	return ytUrlTypeInvalid, "", errors.New("invalid or incomplete url")
}

func (p *Plugin) getYTChannel(url string) (channel *youtube.Channel, err error) {
	urlType, id, err := p.parseYtUrl(url)
	if err != nil {
		return nil, err
	}

	var cResp *youtube.ChannelListResponse
	channelListCall := p.YTService.Channels.List([]string{"snippet"})

	switch urlType {
	case ytUrlTypeChannel:
		channelListCall = channelListCall.Id(id)
	case ytUrlTypeUser:
		channelListCall = channelListCall.ForUsername(id)
	case ytUrlTypeCustom:
		searchListCall := p.YTService.Search.List([]string{"snippet"})
		searchListCall = searchListCall.Q(id).Type("channel")
		sResp, err := searchListCall.Do()
		if err != nil {
			return nil, common.ErrWithCaller(err)
		}
		if len(sResp.Items) < 1 {
			return nil, ErrNoChannel
		}
		channelListCall = channelListCall.Id(sResp.Items[0].Id.ChannelId)
	case ytUrlTypeVideo:
		searchListCall := p.YTService.Search.List([]string{"snippet"})
		searchListCall = searchListCall.Q(id).Type("video")
		sResp, err := searchListCall.Do()

		if err != nil {
			return nil, common.ErrWithCaller(err)
		}
		if len(sResp.Items) < 1 {
			return nil, ErrNoChannel
		}
		channelListCall = channelListCall.Id(sResp.Items[0].Snippet.ChannelId)
	default:
		return nil, common.ErrWithCaller(errors.New("invalid youtube Url"))
	}
	cResp, err = channelListCall.Do()

	if err != nil {
		return nil, common.ErrWithCaller(err)
	}
	if len(cResp.Items) < 1 {
		return nil, ErrNoChannel
	}
	return cResp.Items[0], nil
}

func (p *Plugin) legacyGetYTChannel(url LegacyYTStruct) (channel *youtube.Channel, err error) {
	call := p.YTService.Channels.List([]string{"snippet"})
	if url.YTChannelID != "" {
		call = call.Id(url.YTChannelID)
	} else if url.YTUsername != "" {
		call = call.ForUsername(url.YTUsername)
	} else if url.YTCustomURL != "" {
		searchCall := p.YTService.Search.List([]string{"snippet"})
		searchCall = searchCall.Q(url.YTCustomURL).Type("channel")
		sResp, err := searchCall.Do()
		if err != nil {
			return nil, common.ErrWithCaller(err)
		}
		if len(sResp.Items) < 1 {
			return nil, ErrNoChannel
		}

		call = call.Id(sResp.Items[0].Id.ChannelId)
	} else {
		searchCall := p.YTService.Search.List([]string{"snippet"})
		searchCall = searchCall.Q(url.YTVideoURL).Type("video")

		sResp, err := searchCall.Do()
		if err != nil {
			return nil, common.ErrWithCaller(err)
		}
		if len(sResp.Items) < 1 {
			return nil, ErrNoChannel
		}

		call = call.Id(sResp.Items[0].Snippet.ChannelId)
	}

	cResp, err := call.Do()
	if err != nil {
		return nil, common.ErrWithCaller(err)
	}

	if len(cResp.Items) < 1 {
		return nil, ErrNoChannel
	}
	return cResp.Items[0], nil
}

func (p *Plugin) AddFeed(guildID, discordChannelID int64, legacyYTChannel, ytChannel *youtube.Channel, mentionEveryone bool, mentionRole int64, publishLivestream bool) (*ChannelSubscription, error) {
	sub := &ChannelSubscription{
		GuildID:           discordgo.StrID(guildID),
		ChannelID:         discordgo.StrID(discordChannelID),
		MentionEveryone:   mentionEveryone,
		MentionRole:       discordgo.StrID(mentionRole),
		PublishLivestream: publishLivestream,
		Enabled:           true,
	}

	if ytChannel != nil {
		sub.YoutubeChannelName = ytChannel.Snippet.Title
		sub.YoutubeChannelID = ytChannel.Id
	} else {
		sub.YoutubeChannelName = legacyYTChannel.Snippet.Title
		sub.YoutubeChannelID = legacyYTChannel.Id
	}

	err := common.BlockingLockRedisKey(RedisChannelsLockKey, 0, 10)
	if err != nil {
		return nil, err
	}
	defer common.UnlockRedisKey(RedisChannelsLockKey)

	err = common.GORM.Create(sub).Error
	if err != nil {
		return nil, err
	}

	err = p.MaybeAddChannelWatch(false, sub.YoutubeChannelID)
	return sub, err
}

// maybeRemoveChannelWatch checks the channel for subs, if it has none then it removes it from the watchlist in redis.
func (p *Plugin) MaybeRemoveChannelWatch(channel string) {
	err := common.BlockingLockRedisKey(RedisChannelsLockKey, 0, 10)
	if err != nil {
		return
	}
	defer common.UnlockRedisKey(RedisChannelsLockKey)

	var count int
	err = common.GORM.Model(&ChannelSubscription{}).Where("youtube_channel_id = ?", channel).Count(&count).Error
	if err != nil || count > 0 {
		if err != nil {
			logger.WithError(err).WithField("yt_channel", channel).Error("Failed getting sub count")
		}
		return
	}

	err = common.MultipleCmds(
		radix.Cmd(nil, "DEL", KeyLastVidTime(channel)),
		radix.Cmd(nil, "DEL", KeyLastVidID(channel)),
		radix.Cmd(nil, "ZREM", RedisKeyWebSubChannels, channel),
	)

	if err != nil {
		return
	}

	err = p.WebSubUnsubscribe(channel)
	if err != nil {
		logger.WithError(err).Error("Failed unsubscribing to channel ", channel)
	}

	logger.WithField("yt_channel", channel).Info("Removed orphaned youtube channel from subbed channel sorted set")
}

// maybeAddChannelWatch adds a channel watch to redis, if there wasn't one before
func (p *Plugin) MaybeAddChannelWatch(lock bool, channel string) error {
	if lock {
		err := common.BlockingLockRedisKey(RedisChannelsLockKey, 0, 10)
		if err != nil {
			return common.ErrWithCaller(err)
		}
		defer common.UnlockRedisKey(RedisChannelsLockKey)
	}

	now := time.Now().Unix()

	mn := radix.MaybeNil{}
	err := common.RedisPool.Do(radix.Cmd(&mn, "ZSCORE", RedisKeyWebSubChannels, channel))
	if err != nil {
		return err
	}

	if !mn.Nil {
		// Websub subscription already active, don't do anything more
		return nil
	}

	err = common.RedisPool.Do(radix.FlatCmd(nil, "SET", KeyLastVidTime(channel), now))
	if err != nil {
		return err
	}

	// Also add websub subscription
	err = p.WebSubSubscribe(channel)
	if err != nil {
		logger.WithError(err).Error("Failed subscribing to channel ", channel)
	}

	logger.WithField("yt_channel", channel).Info("Added new youtube channel watch")
	return nil
}

func (p *Plugin) CheckVideo(videoID string, channelID string) error {
	subs, err := p.getRemoveSubs(channelID)
	if err != nil || len(subs) < 1 {
		return err
	}

	lastVid, lastVidTime, err := p.getLastVidTimes(channelID)
	if err != nil {
		return err
	}

	if lastVid == videoID {
		// the video was already posted and was probably just edited
		return nil
	}

	resp, err := p.YTService.Videos.List([]string{"snippet"}).Id(videoID).Do()
	if err != nil || len(resp.Items) < 1 {
		return err
	}

	item := resp.Items[0]

	parsedPublishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		return errors.New("Failed parsing youtube timestamp: " + err.Error() + ": " + item.Snippet.PublishedAt)
	}

	if time.Since(parsedPublishedAt) > time.Hour {
		// just a safeguard against empty lastVidTime's
		return nil
	}

	if lastVidTime.After(parsedPublishedAt) {
		// wasn't a new vid
		return nil
	}

	// This is a new video, post it
	return p.postVideo(subs, parsedPublishedAt, item, channelID)
}

func (p *Plugin) postVideo(subs []*ChannelSubscription, publishedAt time.Time, video *youtube.Video, channelID string) error {
	err := common.MultipleCmds(
		radix.FlatCmd(nil, "SET", KeyLastVidTime(channelID), publishedAt.Unix()),
		radix.FlatCmd(nil, "SET", KeyLastVidID(channelID), video.Id),
	)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.Enabled {
			//if livestream notifications are disabled, and the video is a livestream, don't post it
			if !sub.PublishLivestream && video.Snippet.LiveBroadcastContent != "none" {
				continue
			}

			p.sendNewVidMessage(sub.GuildID, sub.ChannelID, video.Snippet.ChannelTitle, sub.YoutubeChannelID, video.Id, sub.MentionRole, video.Snippet.LiveBroadcastContent, sub.MentionEveryone)
		}
	}

	return nil
}

func (p *Plugin) getRemoveSubs(channelID string) ([]*ChannelSubscription, error) {
	var subs []*ChannelSubscription
	err := common.GORM.Where("youtube_channel_id = ?", channelID).Find(&subs).Error
	if err != nil {
		return subs, err
	}

	if len(subs) < 1 {
		time.AfterFunc(time.Second*10, func() {
			p.MaybeRemoveChannelWatch(channelID)
		})
		return subs, nil
	}

	return subs, nil
}

func (p *Plugin) getLastVidTimes(channelID string) (lastVid string, lastVidTime time.Time, err error) {
	// Find the last video time for this channel
	var unixSeconds int64
	err = common.RedisPool.Do(radix.Cmd(&unixSeconds, "GET", KeyLastVidTime(channelID)))

	var lastProcessedVidTime time.Time
	if err != nil || unixSeconds == 0 {
		lastProcessedVidTime = time.Time{}
	} else {
		lastProcessedVidTime = time.Unix(unixSeconds, 0)
	}

	var lastVidID string
	err = common.RedisPool.Do(radix.Cmd(&lastVidID, "GET", KeyLastVidID(channelID)))
	return lastVidID, lastProcessedVidTime, err
}
