package yageconomy

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/eventsystem"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	prfx "github.com/mrbentarikau/pagst/common/prefix"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/image/font/gofont/goregular"
)

var _ bot.BotInitHandler = (*Plugin)(nil)
var _ commands.CommandProvider = (*Plugin)(nil)

var gofont, _ = truetype.Parse(goregular.TTF)

var CategoryEconomy = &dcmd.Category{
	Name:        "Economy",
	Description: "Ecnonomy commands",
	HelpEmoji:   "$",
	EmbedColor:  0x35afed,
}
var CategoryWaifu = &dcmd.Category{
	Name:        "Waifu",
	Description: "Waifu commands",
	HelpEmoji:   "$",
	EmbedColor:  0xed369e,
}

const (
	ColorBlue = 0x4595e0
	ColorRed  = 0xe04545
)

func (p *Plugin) AddCommands() {

	// commands.AddRootCommands(cmds...)
	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware}, CoreCommands...)
	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, economyAdminMiddleware}, CoreAdminCommands...)
	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, gamblingCmdMiddleware, moneyAlteringMW}, GameCommands...)

	waifuContainer, _ := commands.CommandSystem.Root.Sub("waifu", "wf")
	waifuContainer.NotFound = commands.CommonContainerNotFoundHandler(waifuContainer, "")

	waifuContainer.AddMidlewares(economyCmdMiddleware)

	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware},
		WaifuCmdTop, WaifuCmdInfo, WaifuCmdAffinity)
	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, moneyAlteringMW},
		WaifuCmdClaim, WaifuCmdReset, WaifuCmdTransfer, WaifuCmdDivorce, WaifuCmdGift)

	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, economyAdminMiddleware},
		WaifuShopAdd, WaifuShopEdit, WaifuCmdDel)

	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, moneyAlteringMW}, ShopCommands...)
	commands.AddRootCommandsWithMiddlewares(p, []dcmd.MiddleWareFunc{economyCmdMiddleware, economyAdminMiddleware}, ShopAdminCommands...)
}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLastLegacy(p, handleMessageCreate, eventsystem.EventMessageCreate)
	eventsystem.AddHandlerAsyncLastLegacy(p, handleReactionAddRemove, eventsystem.EventMessageReactionAdd, eventsystem.EventMessageReactionRemove)
}

func economyCmdMiddleware(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		config, err := models.FindEconomyConfigG(data.Context(), data.GuildData.GS.ID)
		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				config = DefaultConfig(data.GuildData.GS.ID)
			} else {
				return "Failed retrieving economy config", err
			}
		}

		if !config.Enabled {
			return "Economy is disabled on this server, you can enable it in the control panel.", nil
		}

		if len(config.EnabledChannels) > 0 {
			if !common.ContainsInt64Slice(config.EnabledChannels, data.GuildData.CS.ID) {
				return "Economy disabled in this channel", nil
			}
		}

		ctx := context.WithValue(data.Context(), CtxKeyConfig, config)
		account, _, err := GetCreateAccount(ctx, data.Author.ID, data.GuildData.GS.ID, config.StartBalance)
		if err != nil {
			return "Failed creating or retrieving your economy account", err
		}

		ctx = context.WithValue(ctx, CtxKeyUser, account)
		return inner(data.WithContext(ctx))
	}
}

func economyAdminMiddleware(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		conf := CtxConfig(data.Context())
		ms := data.GuildData.MS
		if sm := bot.State.GetMember(data.GuildData.GS.ID, ms.User.ID); sm != nil {
			// Prefer state member over the one provided in the message, since it may have presence data
			ms = sm
		}
		if !common.ContainsInt64SliceOneOf(ms.Member.Roles, conf.Admins) {
			return ErrorEmbed(data.Author, "This command requires you to be an economy admin"), nil
		}

		return inner(data)
	}
}

func GetCreateAccount(ctx context.Context, userID int64, guildID int64, startBalance int64) (account *models.EconomyUser, created bool, err error) {
	return GetCreateAccountExec(common.PQ, ctx, userID, guildID, startBalance)
}

func GetCreateAccountExec(exec boil.ContextExecutor, ctx context.Context, userID int64, guildID int64, startBalance int64) (account *models.EconomyUser, created bool, err error) {
	account, err = models.FindEconomyUser(ctx, exec, guildID, userID)
	if err == nil {
		return account, false, nil
	}

	if errors.Cause(err) != sql.ErrNoRows {
		return nil, false, err
	}

	account = &models.EconomyUser{
		GuildID:   guildID,
		UserID:    userID,
		MoneyBank: startBalance,
	}

	err = account.Insert(ctx, exec, boil.Infer())
	if err != nil {
		return nil, false, err
	}

	return account, true, nil
}

type CtxKey int

const (
	CtxKeyConfig CtxKey = iota
	CtxKeyUser
)

func CtxConfig(c context.Context) *models.EconomyConfig {
	return c.Value(CtxKeyConfig).(*models.EconomyConfig)
}

func CtxUser(c context.Context) *models.EconomyUser {
	return c.Value(CtxKeyUser).(*models.EconomyUser)
}

func UserEmebdAuthor(user *discordgo.User) *discordgo.MessageEmbedAuthor {
	return &discordgo.MessageEmbedAuthor{
		Name:    user.Username + "#" + user.Discriminator,
		IconURL: user.AvatarURL("128"),
	}
}

func SimpleEmbedResponse(user *discordgo.User, msgF string, args ...interface{}) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author:      UserEmebdAuthor(user),
		Color:       ColorBlue,
		Description: fmt.Sprintf(msgF, args...),
	}
}

func ErrorEmbed(user *discordgo.User, msgF string, args ...interface{}) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author:      UserEmebdAuthor(user),
		Color:       ColorRed,
		Description: fmt.Sprintf(msgF, args...),
	}
}

func handleMessageCreate(evt *eventsystem.EventData) {
	msg := evt.MessageCreate()
	if msg.Author == nil || msg.Author.Bot {
		return
	}

	conf, err := models.FindEconomyConfigG(evt.Context(), msg.GuildID)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			conf = DefaultConfig(msg.GuildID)
		} else {
			logger.WithError(err).WithField("guild", msg.GuildID).Error("failed retrieving economy config")
			return
		}
	}

	if !conf.Enabled {
		return
	}

	if len(conf.EnabledChannels) > 0 && !common.ContainsInt64Slice(conf.EnabledChannels, msg.ChannelID) {
		return
	}

	if conf.ChatmoneyAmountMax > 0 {
		// gen chat money maybe?
		amount := rand.Int63n(conf.ChatmoneyAmountMax-conf.ChatmoneyAmountMin) + conf.ChatmoneyAmountMin

		result, err := common.PQ.Exec(`UPDATE economy_users SET last_chatmoney_claim = now(), money_bank = money_bank + $4
			WHERE guild_id = $1 AND user_id = $2 AND EXTRACT(EPOCH FROM (now() - last_chatmoney_claim))  > $3`,
			msg.GuildID, msg.Author.ID, conf.ChatmoneyFrequency, amount)
		if err != nil {
			logger.WithField("guild", msg.GuildID).WithError(err).Error("failed claiming chatmoney")
			return
		}

		rows, err := result.RowsAffected()
		if err != nil {
			logger.WithField("guild", msg.GuildID).WithError(err).Error("failed claiming chatmoney, rows")
			return
		}

		if rows > 0 {
			logger.Infof("Gave %s (%d) chat money of %d", msg.Author.Username, msg.Author.ID, amount)
		}
	}

	// maybe plant?
	if common.ContainsInt64Slice(conf.AutoPlantChannels, msg.ChannelID) {

		chance, _ := conf.AutoPlantChance.Float64()
		if rand.Float64() > chance {
			return
		}

		amount := rand.Int63n(conf.AutoPlantMax-conf.AutoPlantMin) + conf.AutoPlantMin

		prefix, _ := prfx.GetCommandPrefixRedis(conf.GuildID)
		msgContent := fmt.Sprintf("**%d** random **%s** appeared! Pick them up with `%spick <password>`", amount, conf.CurrencySymbol, prefix)

		// Plant!
		err = PlantMoney(context.Background(), conf, msg.ChannelID, 0, int(amount), "", msgContent)
		if err != nil {
			logger.WithError(err).WithField("guild", msg.GuildID).WithField("channel", msg.ChannelID).Error("failed planting money")
		}
	}
}

var errAlreadyPlantInChannel = errors.New("Already money planted in this channel")

func PlantMoney(ctx context.Context, conf *models.EconomyConfig, channelID, author int64, amount int, password, prompt string) error {
	if password == "" {
		password = genPlantPassword()
	}

	r, err := getRandomPlantImage(ctx, conf.GuildID)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		return err
	}

	m := &models.EconomyPlant{
		ChannelID: channelID,
		GuildID:   conf.GuildID,
		AuthorID:  author,
		Amount:    int64(amount),
		Password:  strings.ToLower(password),
	}

	// create the drawing context
	var c *gg.Context
	if r != nil {
		img, _, err := image.Decode(bytes.NewReader(r.Image))
		if err != nil {
			return err
		}

		c = gg.NewContextForImage(img)
	} else {
		// just use a white placeholder image
		c = gg.NewContext(350, 150)
		c.SetColor(color.RGBA{
			255, 255, 255, 255,
		})
		c.DrawRectangle(0, 0, 350, 150)
		c.Fill()
	}

	c.SetColor(color.RGBA{
		1, 1, 1, 156,
	})

	tSize := float64(36)
	textWidth := float64(0)
	textHeight := float64(0)

	for {
		// find the best size
		c.SetFontFace(truetype.NewFace(gofont, &truetype.Options{
			Size: tSize,
		}))

		textWidth, textHeight = c.MeasureString(password)
		if int(textWidth) > c.Width() || int(textHeight) > c.Height() {
			tSize /= 2
		} else {
			break
		}

		if tSize <= 2 {
			break
		}
	}

	c.DrawRectangle(0, 0, textWidth+20, textHeight+20)
	c.Fill()

	c.SetColor(color.RGBA{
		255, 255, 255, 255,
	})

	c.DrawString(password, 5, textHeight+5)

	buf := bytes.NewBuffer(nil)
	err = c.EncodePNG(buf)
	if err != nil {
		return err
	}

	msg, err := common.BotSession.ChannelFileSendWithMessage(channelID, prompt, "plant.png", buf)
	if err != nil {
		return err
	}

	m.MessageID = msg.ID
	err = m.InsertG(ctx, boil.Infer())

	return err
}

func getRandomPlantImage(ctx context.Context, guildID int64) (*models.EconomyPickImages2, error) {
	// models.FindEconomyPickImageG(ctx, conf.GuildID)

	imgs, err := models.EconomyPickImages2s(qm.Select("id"), qm.Where("guild_id=?", guildID)).AllG(ctx)
	if err != nil {
		return nil, err
	}

	if len(imgs) < 1 {
		return nil, nil
	}

	imgID := imgs[rand.Intn(len(imgs))].ID

	// get the full image
	img, err := models.FindEconomyPickImages2G(ctx, imgID)
	return img, err
}

var (
	ErrInsufficientFunds = errors.New("Insufficient funds")
)

// TransferMoneyWallet transfers money from one users wallet to another
// both from and to is optional
// out and in amount can be different in certain cases (such as gambling)
func TransferMoneyWallet(ctx context.Context, tx *sql.Tx, conf *models.EconomyConfig, checkFunds bool, from, to int64, outAmount, inAmount int64) (err error) {
	createdTX := false
	if tx == nil {
		createdTX = true
		tx, err = common.PQ.Begin()
		if err != nil {
			return err
		}
	}

	// make sure the origin account is created
	if from != 0 {
		account, _, err := GetCreateAccountExec(tx, ctx, from, conf.GuildID, conf.StartBalance)
		if err != nil {
			if createdTX {
				tx.Rollback()
			}
			return err
		}
		if checkFunds && account.MoneyWallet < outAmount {
			if createdTX {
				tx.Rollback()
			}
			return ErrInsufficientFunds
		}
	}

	// make sure the destination account is created
	if to != 0 {
		_, _, err = GetCreateAccountExec(tx, ctx, to, conf.GuildID, conf.StartBalance)
		if err != nil {
			if createdTX {
				tx.Rollback()
			}
			return err
		}
	}

	// update origin account
	if from != 0 {
		_, err := tx.Exec("UPDATE economy_users SET money_wallet = money_wallet - $3 WHERE user_id = $2 AND guild_id = $1", conf.GuildID, from, outAmount)
		if err != nil {
			if createdTX {
				tx.Rollback()
			}
			return err
		}
	}

	// update the destination account
	if to != 0 {
		_, err = tx.Exec("UPDATE economy_users SET money_wallet = money_wallet + $3 WHERE user_id = $2 AND guild_id = $1", conf.GuildID, to, inAmount)
		if err != nil {
			if createdTX {
				tx.Rollback()
			}
			return err
		}
	}

	if createdTX {
		return tx.Commit()
	}

	return err
}

func genPlantPassword() string {

	availableChars := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	pw := ""
	for i := 0; i < 4; i++ {
		char := availableChars[rand.Intn(len(availableChars))]
		pw += char
	}

	return pw
}

// UserIDArg matches a mention or a plain id, the user does not have to be a part of the server
// The type of the ID is parsed into a int64
type AmountArg struct {
	Min, Max          int64
	InteractionString bool
}

func (ca *AmountArg) CheckCompatibility(def *dcmd.ArgDef, part string) dcmd.CompatibilityResult {
	return dcmd.CompatibilityGood
}

func (ca *AmountArg) Parse(def *dcmd.ArgDef, part string, data *dcmd.Data) (interface{}, error) {
	// read fixed number
	fixed, err := strconv.ParseInt(part, 10, 64)
	if err == nil {
		return &AmountArgResult{
			FixedAmount: fixed,
			IsFixed:     true,
		}, nil
	}

	multiplier := float64(0)

	if strings.HasSuffix(part, "%") || strings.HasPrefix(part, "%") {
		numberStr := strings.TrimSuffix(part, "%")
		numberStr = strings.TrimPrefix(numberStr, "%")

		parsed, err := strconv.ParseFloat(numberStr, 64)
		if err != nil {
			return nil, err
		}

		multiplier = parsed / 100
	} else {
		lower := strings.ToLower(part)
		switch lower {
		case "all", "everything", "*":
			multiplier = 1
		case "half":
			multiplier = 0.5
		}
	}

	return &AmountArgResult{
		Multiplier: multiplier,
	}, nil

}

func (ca *AmountArg) ParseFromInteraction(def *dcmd.ArgDef, data *dcmd.Data, options *dcmd.SlashCommandsParseOptions) (val interface{}, err error) {

	any, err := options.ExpectAny(def.Name)
	if err != nil {
		return nil, err
	}

	var v int64
	switch t := any.(type) {
	case string:
		v, err = strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil, err
		}
	case int64:
		v = t
	default:
	}

	// A valid range has been specified
	if ca.Max != ca.Min {
		if ca.Max < v || ca.Min > v {
			return nil, &dcmd.OutOfRangeError{ArgName: def.Name, Got: v, Min: ca.Min, Max: ca.Max}
		}
	}

	return v, nil
}

func (ca *AmountArg) ParseFromMessage(def *dcmd.ArgDef, part string, data *dcmd.Data) (interface{}, error) {
	v, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return nil, &dcmd.InvalidInt{part}
	}

	// A valid range has been specified
	if ca.Max != ca.Min {
		if ca.Max < v || ca.Min > v {
			return nil, &dcmd.OutOfRangeError{ArgName: def.Name, Got: v, Min: ca.Min, Max: ca.Max}
		}
	}

	return v, nil
}

func (ca *AmountArg) HelpName() string {
	return "Amount"
}

func (ca *AmountArg) SlashCommandOptions(def *dcmd.ArgDef) []*discordgo.ApplicationCommandOption {
	if ca.InteractionString {
		return []*discordgo.ApplicationCommandOption{def.StandardSlashCommandOption(discordgo.ApplicationCommandOptionString)}
	}
	return []*discordgo.ApplicationCommandOption{def.StandardSlashCommandOption(discordgo.ApplicationCommandOptionInteger)}
}

type AmountArgResult struct {
	IsFixed     bool
	FixedAmount int64
	Multiplier  float64
}

// returns the amount based on the current wallet size
func (a *AmountArgResult) Apply(total int64) int64 {
	if a.IsFixed {
		return a.FixedAmount
	}

	return int64(float64(total) * a.Multiplier)
}

// same as apply but also has some restrictions built in
func (a *AmountArgResult) ApplyWithRestrictions(total int64, currencySymbol, sourceName string, shouldBeBelowTotal bool, minAmount int64) (v int64, resp string) {
	v = a.Apply(total)

	if v < minAmount {
		return -1, fmt.Sprintf("Amount can't be less than %s%d", currencySymbol, minAmount)
	}

	if v > total {
		return -1, fmt.Sprintf("You don't have **%s%d** in your %s (you have **%s%d**)", currencySymbol, v, sourceName, currencySymbol, total)
	}

	return v, ""
}

type GamblingLockKey struct {
	GuildID int64
	UserID  int64
}

var (
	moneyAlteringLocks   = make(map[GamblingLockKey]string)
	moneyAlteringLocksmu sync.Mutex
)

func TryLockMoneyAltering(guildID, userID int64, msg string) (bool, string) {
	moneyAlteringLocksmu.Lock()
	defer moneyAlteringLocksmu.Unlock()

	key := GamblingLockKey{
		GuildID: guildID,
		UserID:  userID,
	}

	if resp, ok := moneyAlteringLocks[key]; ok {
		return false, resp
	}

	moneyAlteringLocks[key] = msg
	return true, ""
}

func IsMoneyAlteringLocked(guildID, userID int64) (bool, string) {
	moneyAlteringLocksmu.Lock()
	defer moneyAlteringLocksmu.Unlock()

	key := GamblingLockKey{
		GuildID: guildID,
		UserID:  userID,
	}

	if resp, ok := moneyAlteringLocks[key]; ok {
		return true, resp
	}
	return false, ""
}

func UnlockMoneyAltering(guildID, userID int64) {
	moneyAlteringLocksmu.Lock()
	defer moneyAlteringLocksmu.Unlock()

	key := GamblingLockKey{
		GuildID: guildID,
		UserID:  userID,
	}

	delete(moneyAlteringLocks, key)
}

func moneyAlteringMW(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		if locked, resp := IsMoneyAlteringLocked(data.GuildData.GS.ID, data.Author.ID); locked {
			return ErrorEmbed(data.Author, resp), nil
		}

		return inner(data)
	}
}
