package yageconomy

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var GameCommands = []*commands.YAGCommand{

	&commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "BetFlip",
		Aliases:      []string{"bf"},
		Description:  "Bet on heads or tail, if you guess correct you win 2x your bet",
		RequiredArgs: 2,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Amount", Type: &AmountArg{}},
			&dcmd.ArgDef{Name: "Side", Type: dcmd.String},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())
			u := parsed.Author

			amount, resp := parsed.Args[0].Value.(*AmountArgResult).ApplyWithRestrictions(account.MoneyWallet, conf.CurrencySymbol, "wallet", true, 1)
			if resp != "" {
				return ErrorEmbed(u, resp), nil
			}

			guessedHeads := true
			inLowered := strings.ToLower(parsed.Args[1].Str())
			if inLowered == "h" || inLowered == "head" || inLowered == "heads" {
				guessedHeads = true
			} else if inLowered == "t" || inLowered == "tail" || inLowered == "tails" {
				guessedHeads = false
			} else {
				return ErrorEmbed(u, "You can only pick between `heads` and `tails`"), nil
			}

			moneyIn := amount

			if amount > account.MoneyWallet {
				return ErrorEmbed(u, "You don't have that amount in your wallet"), nil
			}

			isHeads := rand.Intn(2) == 0

			won := false
			winningsLosses := amount
			var err error

			if (isHeads && guessedHeads) || (!isHeads && !guessedHeads) {
				won = true
				winningsLosses = ApplyGamblingBoost(account, amount)
				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, 0, u.ID, 0, winningsLosses)
			} else {
				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, u.ID, common.BotUser.ID, winningsLosses, winningsLosses)
			}

			if err != nil {
				return nil, err
			}

			strResult := "heads"
			if !isHeads {
				strResult = "tails"
			}

			msg := ""
			if won {
				msg = fmt.Sprintf("Result is... **%s**: You won! Awarded with **%s%d**", strResult, conf.CurrencySymbol, amount+winningsLosses)
			} else {
				msg = fmt.Sprintf("Result is... **%s**: You lost... you're now **%s%d** poorer...", strResult, conf.CurrencySymbol, moneyIn)
			}

			return SimpleEmbedResponse(u, msg), nil
		},
	},
	&commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "BetRoll",
		Aliases:      []string{"br"},
		Description:  "Rolls 1-100, Rolling over 66 yields x2 of your bet, over 90 -> x4 and 100 -> x10.",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Bet", Type: &AmountArg{}},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())
			u := parsed.Author

			amount, resp := parsed.Args[0].Value.(*AmountArgResult).ApplyWithRestrictions(account.MoneyWallet, conf.CurrencySymbol, "wallet", true, 1)
			if resp != "" {
				return ErrorEmbed(u, resp), nil
			}

			walletMod := -amount

			roll := rand.Intn(100) + 1

			won := roll > 66
			if roll == 100 {
				walletMod = amount * 9
			} else if roll > 90 {
				walletMod = amount * 3
			} else if roll > 66 {
				walletMod = amount
			}

			var err error
			if won {
				// transfer winnings into our account
				walletMod = ApplyGamblingBoost(account, walletMod)
				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, 0, u.ID, 0, walletMod)
			} else {
				// transfer losses into bot account
				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, u.ID, common.BotUser.ID, amount, amount)
			}
			if err != nil {
				return nil, err
			}

			msg := ""
			if won {
				msg = fmt.Sprintf("Rolled **%d** and won! You have been awarded with **%s%d**", roll, conf.CurrencySymbol, walletMod+amount)
			} else {
				msg = fmt.Sprintf("Rolled **%d** and lost... you're now **%s%d** poorer...", roll, conf.CurrencySymbol, amount)
			}

			return SimpleEmbedResponse(u, msg), nil
		},
	}, &commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "Rob",
		Description:  "Steals money from someone, the chance of suceeding = your networth / (their cash + your networth)",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Target", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			target := parsed.Args[0].User()

			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())
			u := parsed.Author

			if conf.RobFine < 1 {
				return ErrorEmbed(u, "No fine as been set, as a result the rob command has been disabled"), nil
			}

			cooldownLeft := account.LastRobAttempt.Add(time.Second * time.Duration(conf.RobCooldown)).Sub(time.Now())
			if cooldownLeft > 0 {
				return ErrorEmbed(u, "The rob command is still on cooldown for you for another %s", common.HumanizeDuration(common.DurationPrecisionSeconds, cooldownLeft)), nil
			}

			if target.ID == parsed.Author.ID {
				return ErrorEmbed(u, "Can't rob yourself..."), nil
			}

			targetAccount, _, err := GetCreateAccount(parsed.Context(), target.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			if targetAccount.MoneyWallet < 1 {
				return ErrorEmbed(u, "This person has no money left in their wallet :("), nil
			}

			if account.MoneyWallet < int64(conf.RobFine) {
				return ErrorEmbed(u, "You don't have enough money in your wallet to pay the fine if you fail"), nil
			}

			account.LastRobAttempt = time.Now()
			_, err = account.UpdateG(parsed.Context(), boil.Whitelist("last_rob_attempt"))
			if err != nil {
				return nil, err
			}

			sucessChance := float64(account.MoneyWallet+account.MoneyBank) / float64(targetAccount.MoneyWallet+account.MoneyWallet+account.MoneyBank)
			if rand.Float64() < sucessChance {
				// sucessfully robbed them

				amount := targetAccount.MoneyWallet

				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, target.ID, u.ID, amount, ApplyGamblingBoost(account, amount))
				if err != nil {
					return nil, err
				}

				return SimpleEmbedResponse(u, "You sucessfully robbed **%s** for **%s%d**!", target.Username, conf.CurrencySymbol, ApplyGamblingBoost(account, amount)), nil
			} else {
				fine := int64(float64(conf.RobFine) / 100 * float64(account.MoneyWallet+account.MoneyBank))

				err = TransferMoneyWallet(parsed.Context(), nil, conf, false, u.ID, common.BotUser.ID, fine, fine)

				if err != nil {
					return nil, err
				}

				return ErrorEmbed(u, "You failed robbing **%s**, you were fined **%s%d** as a result, hopefully you have learned your lesson now.",
					target.Username, conf.CurrencySymbol, fine), nil
			}

		},
	},
	&commands.YAGCommand{
		CmdCategory: CategoryEconomy,
		Name:        "Fish",
		Description: "Attempts to fish for some easy money",
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())

			ms := parsed.GuildData.MS
			if sm := bot.State.GetMember(parsed.GuildData.GS.ID, ms.User.ID); sm != nil {
				// Prefer state member over the one provided in the message, since it may have presence data
				ms = sm
			}

			u := parsed.Author

			if conf.FishingMaxWinAmount < 1 {
				return ErrorEmbed(u, "Fishing not set up on this server"), nil
			}

			wonAmount := int64(0)
			fishAmount := 0
			if rand.Float64() > 0.25 {
				// 75% chance to catch a fish
				wonAmount = rand.Int63n(conf.FishingMaxWinAmount-conf.FishingMinWinAmount) + conf.FishingMinWinAmount
				wonAmount = ApplyGamblingBoost(account, wonAmount)
				fishAmount = 1
			}

			result, err := common.PQ.Exec(`UPDATE economy_users SET last_fishing = now(), money_wallet = money_wallet + $4, fish_caugth = fish_caugth + $5
			WHERE guild_id = $1 AND user_id = $2 AND EXTRACT(EPOCH FROM (now() - last_fishing)) > $3`, parsed.GuildData.GS.ID, ms.User.ID, conf.FishingCooldown*60, wonAmount, fishAmount)
			if err != nil {
				return nil, err
			}

			rows, err := result.RowsAffected()
			if err != nil {
				return nil, err
			}

			if rows < 1 {
				timeToWait := account.LastFishing.Add(time.Duration(conf.FishingCooldown) * time.Minute).Sub(time.Now())
				return ErrorEmbed(u, "You can't fish again yet, please wait another %s.", common.HumanizeDuration(common.DurationPrecisionSeconds, timeToWait)), nil
			}

			if wonAmount == 0 {
				return SimpleEmbedResponse(u, "Aww man, you let your fish slip away..."), nil
			}

			return SimpleEmbedResponse(u, "Nice! You caught a fish worth **%s%d**!", conf.CurrencySymbol, wonAmount), nil
		},
	},
	&commands.YAGCommand{
		CmdCategory: CategoryEconomy,
		Name:        "Heist",
		Description: "Starts a heist in 1 minute from now",
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			resp, err := NewHeist(CtxConfig(parsed.Context()), parsed.GuildData.GS.ID, parsed.GuildData.CS.ID, parsed.Author, CtxUser(parsed.Context()), time.Minute)
			if resp != "" {
				return ErrorEmbed(parsed.Author, "%s", resp), err
			}
			return nil, err
		},
	},
}

func ApplyGamblingBoost(account *models.EconomyUser, winnings int64) int64 {
	return int64(float64(winnings) * ((float64(account.GamblingBoostPercentage) / 100) + 1))
}

func gamblingCmdMiddleware(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		conf := CtxConfig(data.Context())
		account := CtxUser(data.Context())

		gamblingBanLeft := account.LastFailedHeist.Add(time.Duration(conf.HeistFailedGamblingBanDuration) * time.Minute).Sub(time.Now())

		if gamblingBanLeft > 0 {
			return ErrorEmbed(data.Author, "You're still banned from gambling for another %s", common.HumanizeDuration(common.DurationPrecisionSeconds, gamblingBanLeft)), nil
		}

		return inner(data)
	}
}
