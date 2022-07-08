package yageconomy

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const JoinHeistEmoji = "✅"

type HeistProgressState int

const (
	HeistProgressStateWaiting HeistProgressState = iota
	HeistProgressStateStarting
	HeistProgressStateInvading
	HeistProgressStateCollecting
	HeistProgressStateLeaving
	HeistProgressStateGetaway
	HeistProgressStateEnded
)

type HeistSession struct {
	sync.Mutex

	GuildID   int64
	ChannelID int64
	MessageID int64

	Author *discordgo.User
	Users  []*HeistUser

	CreatedAt time.Time
	StartsAt  time.Time

	ProgressState  HeistProgressState
	StateChangedAt time.Time

	EventCursor int

	ExtraEventChance    int
	MoneyLostPercentage int
	MoneyLostFixed      int

	BotAccount *models.EconomyUser
	Config     *models.EconomyConfig
}

type HeistUser struct {
	User     *discordgo.User
	Dead     bool
	Injured  bool
	Captured bool

	Account  *models.EconomyUser
	Winnings int64
}

var (
	activeHeists   []*HeistSession
	activeHeistsmu sync.Mutex
)

func NewHeist(conf *models.EconomyConfig, guildID, channelID int64, author *discordgo.User, account *models.EconomyUser, waitUntilStart time.Duration) (resp string, err error) {
	activeHeistsmu.Lock()
	defer activeHeistsmu.Unlock()

	for _, v := range activeHeists {
		if v.ChannelID == channelID {
			return "Already a heist going on in this channel", nil
		}

		if v.GuildID == guildID {
			for _, m := range v.Users {
				if m.User.ID == author.ID {
					return "You're already in another heist on this server", nil
				}
			}
		}
	}

	if locked, resp := TryLockMoneyAltering(guildID, author.ID, "You can't use any money-altering commands while in a heist."); !locked {
		return resp, nil
	}
	createdHeist := false
	defer func() {
		if !createdHeist {
			UnlockMoneyAltering(guildID, author.ID)
		}
	}()

	msg, err := common.BotSession.ChannelMessageSendEmbed(channelID, SimpleEmbedResponse(author, "Setting up heist..."))
	if err != nil {
		return "", err
	}

	cdLeft := conf.HeistLastUsage.Add(time.Minute * time.Duration(conf.HeistServerCooldown)).Sub(time.Now())
	if cdLeft > 0 {
		return "Still on cooldown for another " + common.HumanizeDuration(common.DurationPrecisionSeconds, cdLeft), nil
	}

	conf.HeistLastUsage = time.Now()
	_, err = conf.UpdateG(context.Background(), boil.Whitelist("heist_last_usage"))
	if err != nil {
		return "", err
	}

	heist := &HeistSession{
		GuildID:   guildID,
		ChannelID: channelID,
		Author:    author,
		MessageID: msg.ID,
		Users: []*HeistUser{
			&HeistUser{
				Account: account,
				User:    author,
			},
		},

		StartsAt:  time.Now().Add(waitUntilStart),
		CreatedAt: time.Now(),
		Config:    conf,
	}
	createdHeist = true
	activeHeists = append(activeHeists, heist)
	go heist.Run()
	return "", nil
}

func removeHeist(h *HeistSession) {

	h.Lock()
	for _, user := range h.Users {
		go UnlockMoneyAltering(h.GuildID, user.User.ID)
	}
	h.Unlock()

	activeHeistsmu.Lock()
	defer activeHeistsmu.Unlock()

	for i, v := range activeHeists {
		if v == h {
			activeHeists = append(activeHeists[:i], activeHeists[i+1:]...)
			return
		}
	}
}

func handleReactionAddRemove(evt *eventsystem.EventData) {

	mID := int64(0)
	switch t := evt.EvtInterface.(type) {
	case *discordgo.MessageReactionAdd:
		mID = t.MessageID
	case *discordgo.MessageReactionRemove:
		mID = t.MessageID
	}

	activeHeistsmu.Lock()
	defer activeHeistsmu.Unlock()
	for _, v := range activeHeists {
		if v.MessageID == mID {
			go v.handleReaction(evt)
		}
	}
}

func (hs *HeistSession) init() {
	err := common.BotSession.MessageReactionAdd(hs.ChannelID, hs.MessageID, JoinHeistEmoji)
	if err != nil {
		logger.WithError(err).Error("failed adding reaction")
	}
}

func (hs *HeistSession) Run() {
	hs.init()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	defer removeHeist(hs)

	for {
		select {
		case <-ticker.C:
			if hs.update() {
				return
			}
		}
	}

}

func (hs *HeistSession) update() (stop bool) {
	hs.Lock()
	defer hs.Unlock()

	switch hs.ProgressState {
	case HeistProgressStateWaiting:
		hs.tickWaiting()
		return false
	case HeistProgressStateEnded:
		return true
	}

	hs.tickEvents()

	return false
}

func (hs *HeistSession) tickWaiting() {
	timeUntilStart := hs.StartsAt.Sub(time.Now())
	if timeUntilStart < 0 {
		hs.Start()
		return
	}

	timeBetweenUpdates := time.Minute
	if timeUntilStart < 10*time.Second {
		timeBetweenUpdates = time.Second
	} else if timeUntilStart < 20*time.Second {
		timeBetweenUpdates = time.Second * 5
	} else if timeUntilStart < 40*time.Second {
		timeBetweenUpdates = time.Second * 10
	} else if timeUntilStart < time.Minute {
		timeBetweenUpdates = time.Second * 20
	} else if timeUntilStart < time.Minute*2 {
		timeUntilStart = time.Second * 30
	}

	if time.Since(hs.StateChangedAt) > timeBetweenUpdates {
		hs.StateChangedAt = time.Now()
		hs.updateWaitingMessage()
	}
}

func (hs *HeistSession) Start() {
	account, _, err := GetCreateAccount(context.Background(), common.BotUser.ID, hs.GuildID, hs.Config.StartBalance)
	if err != nil {
		logger.WithError(err).Error("failed retrieving account")
		hs.ProgressState = HeistProgressStateEnded
		return
	}
	hs.BotAccount = account
	hs.ProgressState = HeistProgressStateStarting
	hs.StateChangedAt = time.Now()
	hs.EventCursor = 0
}

func (hs *HeistSession) updateWaitingMessage() {
	timeUntilStart := hs.StartsAt.Sub(time.Now())
	precision := common.DurationPrecisionSeconds
	if timeUntilStart > time.Minute*2 {
		precision = common.DurationPrecisionMinutes
	}

	embed := SimpleEmbedResponse(hs.Author, "A heist is being set up by **%s#%s**\nIt's scheduled to start in `%s` at `%s` UTC",
		hs.Author.Username, hs.Author.Discriminator, common.HumanizeDuration(precision, timeUntilStart), hs.StartsAt.UTC().Format(time.Kitchen))

	embed.Description += "\n\nYou can join the heist by reacting below, doing so will put all your money in your wallet at stake.\nThe more money and the more people the higher the chance of succeeding."

	embed.Title = "Heist being set up"

	membersBuf := &strings.Builder{}
	for _, v := range hs.Users {
		membersBuf.WriteString(fmt.Sprintf("**%s#%s**: %d\n", v.User.Username, v.User.Discriminator, v.Account.MoneyWallet))
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:  "Crew",
		Value: membersBuf.String(),
	})

	_, err := common.BotSession.ChannelMessageEditEmbed(hs.ChannelID, hs.MessageID, embed)
	if err != nil {
		logger.WithError(err).WithField("guil", hs.GuildID).Error("Failed updating heist message")
	}
}

func (hs *HeistSession) tickEvents() {
	if time.Since(hs.StateChangedAt) < time.Second*4 {
		return
	}

	events := OrderedHeistEvents[hs.ProgressState]
	for {
		if len(events) <= hs.EventCursor {
			hs.EventCursor = 0
			if hs.ProgressState == HeistProgressStateGetaway {
				hs.End()
				return
			}
			hs.ProgressState++
			return
		}

		event := events[hs.EventCursor]
		hs.EventCursor++
		hs.StateChangedAt = time.Now()

		result := event.Run(hs)
		if result == nil {
			continue
		}
		hs.applyEffect(result)
		if len(hs.aliveUsers()) < 1 {
			hs.End()
		}

		return
	}
}

func (hs *HeistSession) applyEffect(effect *HeistEventEffect) {
	resp := &strings.Builder{}
	resp.WriteString(effect.TextResponse)

	if len(effect.Dead) > 0 {
		for _, dead := range effect.Dead {
			dead.Dead = true
		}
		resp.WriteString("\n" + joinUsers(effect.Dead) + " has died.")
	}
	if len(effect.Injured) > 0 {
		for _, v := range effect.Injured {
			v.Injured = true
		}

		resp.WriteString("\n" + joinUsers(effect.Injured) + " were injured.")
	}
	if len(effect.Captured) > 0 {
		for _, v := range effect.Captured {
			v.Captured = true
		}

		resp.WriteString("\n" + joinUsers(effect.Injured) + " were captured by the police.")
	}
	if effect.MoneyLostPercentage > 0 {
		resp.WriteString("\n" + fmt.Sprintf("%d%% of the total money was lost.", effect.MoneyLostPercentage))
		hs.MoneyLostPercentage += effect.MoneyLostPercentage
	}
	if effect.MoneyLostFixed > 0 {
		resp.WriteString("\n" + fmt.Sprintf("%d money was lost.", effect.MoneyLostFixed))
		hs.MoneyLostFixed += effect.MoneyLostFixed
	}
	if effect.IncreasedChanceOfEvents > 0 {
		resp.WriteString("\n" + fmt.Sprintf("Chance of something going wrong increased by %d%%.", effect.IncreasedChanceOfEvents))
		hs.ExtraEventChance += effect.IncreasedChanceOfEvents
	}

	embed := SimpleEmbedResponse(hs.Author, "%s", resp.String())
	embed.Title = "Undergoing heist"
	common.BotSession.ChannelMessageSendEmbed(hs.ChannelID, embed)
}

func (hs *HeistSession) handleReaction(evt *eventsystem.EventData) {
	userID := int64(0)
	emoji := ""
	add := false
	guildID := int64(0)
	switch t := evt.EvtInterface.(type) {
	case *discordgo.MessageReactionAdd:
		userID = t.UserID
		emoji = t.Emoji.Name
		add = true
		guildID = t.GuildID
	case *discordgo.MessageReactionRemove:
		guildID = t.GuildID
		userID = t.UserID
		emoji = t.Emoji.Name
		add = false
	}

	if emoji != JoinHeistEmoji {
		return
	}

	users := bot.GetUsers(guildID, userID)
	if len(users) < 1 {
		return
	}
	user := users[0]
	if user.Bot {
		return
	}

	addedToHeist := false
	if add {
		if locked, resp := TryLockMoneyAltering(guildID, userID, "You can't use any money altering commands while in a heist"); !locked {
			bot.SendDM(userID, "Unable to join heist: "+resp)
			return
		}

		defer func() {
			if !addedToHeist {
				UnlockMoneyAltering(guildID, userID)
			}
		}()
	}

	hs.Lock()
	defer hs.Unlock()

	if hs.ProgressState != HeistProgressStateWaiting {
		return
	}

	if add && len(hs.Users) >= 10 {
		return
	}

	config, err := GuildConfigOrDefault(context.Background(), guildID)
	if err != nil {
		logger.WithError(err).Error("failed fetching config")
		return
	}

	if !add && userID == hs.Author.ID {
		hs.ProgressState = HeistProgressStateEnded
		common.BotSession.ChannelMessageEditEmbed(hs.ChannelID, hs.MessageID, SimpleEmbedResponse(hs.Author, "Author cancelled this heist"))
		config.HeistLastUsage = time.Now().Add(-time.Minute * time.Duration(config.HeistServerCooldown))
		_, err = config.UpdateG(context.Background(), boil.Whitelist("heist_last_usage"))
		if err != nil {
			logger.WithError(err).Error("failed updating config")
		}
		return
	}

	account, _, err := GetCreateAccount(context.Background(), user.ID, guildID, config.StartBalance)
	if err != nil {
		logger.WithError(err).Error("failed fetching account")
		return
	}

	if account.MoneyWallet < 1 {
		return
	}

	heistUser := &HeistUser{
		User:    user,
		Account: account,
	}

	if add {
		addedToHeist = true
		hs.Users = addUserIfNotExists(hs.Users, heistUser)
	} else {

		newUsers, removed := removeUserFromSlice(hs.Users, heistUser)
		hs.Users = newUsers
		if removed {
			UnlockMoneyAltering(guildID, userID)
		}
	}

	hs.updateWaitingMessage()
}

func addUserIfNotExists(users []*HeistUser, target *HeistUser) []*HeistUser {
	for _, v := range users {
		if v.User.ID == target.User.ID {
			return users
		}
	}

	return append(users, target)
}

func removeUserFromSlice(users []*HeistUser, target *HeistUser) ([]*HeistUser, bool) {
	for i, v := range users {
		if v.User.ID == target.User.ID {
			return append(users[:i], users[i+1:]...), true
		}
	}

	return users, false
}

func joinUsers(users []*HeistUser) string {
	var builder strings.Builder
	for i, v := range users {
		if i != 0 && i == len(users)-1 {
			builder.WriteString(" and ")
		} else if i != 0 {
			builder.WriteString(", ")
		}

		builder.WriteString("`" + v.User.Username + "`")
	}

	return builder.String()
}

func inUsersSlice(users []*discordgo.User, id int64) bool {
	for _, v := range users {
		if v.ID == id {
			return true
		}
	}

	return false
}

func (hs *HeistSession) End() {
	config := hs.Config

	hs.ProgressState = HeistProgressStateEnded

	builder := &strings.Builder{}

	alive := hs.aliveUsers()
	if len(alive) < 1 {
		builder.WriteString("You all died or got captured by the police, seems like you should re-evaluate your decisions.")
		for _, v := range hs.Users {
			_, err := common.PQ.Exec("UPDATE economy_users SET last_failed_heist = now() WHERE user_id = $2 AND guild_id = $1", config.GuildID, v.User.ID)
			if err == nil {
				err = TransferMoneyWallet(context.Background(), nil, config, false, v.User.ID, common.BotUser.ID, v.Account.MoneyWallet, v.Account.MoneyWallet)
			}
			if err != nil {
				logger.WithError(err).Error("failed updating users")
			}
		}
	} else {
		builder.WriteString(joinUsers(alive) + " made it out alive")

		botAccount := hs.BotAccount

		profit := (botAccount.MoneyWallet + botAccount.MoneyBank) / 20
		if config.HeistFixedPayout > 0 {
			profit = int64(config.HeistFixedPayout)
		}

		if hs.MoneyLostPercentage >= 100 {
			builder.WriteString(", but you lost all the money...")
		} else {
			multiplier := 1 - float64(hs.MoneyLostPercentage)/100
			profit = int64(multiplier * float64(profit))
			profitsFiltered := hs.calcWinnings(profit)

			// perPerson := profit / int64(len(hs.Users))
			// remainder := profit % int64(len(hs.Users))

			builder.WriteString(fmt.Sprintf("\n\n**Total earnings: %s%d (-%d%%)**\n\n", config.CurrencySymbol, profitsFiltered, hs.MoneyLostPercentage))
			for _, v := range hs.Users {
				hs.finishHeistUser(config, v, builder)
			}
		}
	}

	embed := SimpleEmbedResponse(hs.Author, "%s", builder.String())
	embed.Title = "Heist finished"
	_, err := common.BotSession.ChannelMessageSendEmbed(hs.ChannelID, embed)
	if err != nil {
		logger.WithError(err).Error("error sending message")
	}
}

func (hs *HeistSession) calcWinnings(totalWinnings int64) int64 {
	totalContributed := int64(0)
	for _, v := range hs.Users {
		if v.Dead || v.Captured {
			continue
		}
		totalContributed += v.Account.MoneyWallet
	}

	filteredWinnings := int64(0)
	for _, v := range hs.Users {
		if v.Dead || v.Captured {
			continue
		}

		portion := (float64(v.Account.MoneyWallet) / float64(totalContributed)) * float64(totalWinnings)
		filteredWinnings += int64(portion)
		v.Winnings = int64(portion)
	}

	return filteredWinnings
}

func (hs *HeistSession) finishHeistUser(config *models.EconomyConfig, v *HeistUser, builder *strings.Builder) {
	builder.WriteString("`" + v.User.Username + "`: " + config.CurrencySymbol)

	win := int64(0)
	if v.Captured {
		builder.WriteString("0 (captured)")
	} else if v.Dead {
		builder.WriteString("0 (dead)")
	} else {
		win = v.Winnings

		extraStr := ""
		if v.Injured {
			// take away 10% if they're injured
			win = int64(float64(win) * 0.9)
			extraStr = "(injured, -33%)"
		}

		builder.WriteString(fmt.Sprintf("%d%s", win, extraStr))
	}

	var err error
	if win > 0 {
		err = TransferMoneyWallet(context.Background(), nil, config, false, common.BotUser.ID, v.User.ID, win, win)
	} else if v.Captured || v.Dead {
		_, err = common.PQ.Exec("UPDATE economy_users SET last_failed_heist = now() WHERE user_id = $2 AND guild_id = $1", config.GuildID, v.User.ID)
		if err == nil {
			err = TransferMoneyWallet(context.Background(), nil, config, false, v.User.ID, common.BotUser.ID, v.Account.MoneyWallet, v.Account.MoneyWallet)
		}
	}

	if err != nil {
		logger.WithError(err).WithField("guild", config.GuildID).WithField("win_amount", win).Error("failed updating user")
	}

	builder.WriteString("\n")
}

func (hs *HeistSession) calcEventChance(in float64) bool {
	ceilingExt := (float64(len(hs.Users)) - 1) * 0.25

	sumMoney := int64(0)
	for _, v := range hs.Users {
		sumMoney += v.Account.MoneyWallet
	}

	totalPayout := float64(hs.Config.HeistFixedPayout)
	if totalPayout < 1 {
		totalPayout = float64(hs.BotAccount.MoneyWallet + sumMoney)
	}

	logger.Println("Ceiling: ", ceilingExt)
	ceilingExt += (float64(sumMoney) / totalPayout) - 0.25
	logger.Println("Ceiling: ", ceilingExt)

	if rand.Float64()*(ceilingExt+1) < in+float64(hs.ExtraEventChance) {
		return true
	}

	return false
}

func (hs *HeistSession) deadUsers() []*HeistUser {
	dst := make([]*HeistUser, 0, len(hs.Users))
	for _, v := range hs.Users {
		if v.Dead {
			dst = append(dst, v)
		}
	}

	return dst
}

func (hs *HeistSession) capturedUsers() []*HeistUser {
	dst := make([]*HeistUser, 0, len(hs.Users))
	for _, v := range hs.Users {
		if v.Captured {
			dst = append(dst, v)
		}
	}

	return dst
}

func (hs *HeistSession) injuredUsers() []*HeistUser {
	dst := make([]*HeistUser, 0, len(hs.Users))
	for _, v := range hs.Users {
		if v.Injured {
			dst = append(dst, v)
		}
	}

	return dst
}

func (hs *HeistSession) aliveUsers() []*HeistUser {
	dst := make([]*HeistUser, 0, len(hs.Users))
	for _, v := range hs.Users {
		if !v.Dead && !v.Captured {
			dst = append(dst, v)
		}
	}

	return dst
}
