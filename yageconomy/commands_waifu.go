package yageconomy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/lib/pq"
	"github.com/mediocregopher/radix/v3"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	WaifuCmdTop = &commands.YAGCommand{
		CmdCategory: CategoryWaifu,
		Name:        "waifuTop",
		Description: "Shows top waifus",
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Page", Type: dcmd.Int, Default: 1},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			conf := CtxConfig(parsed.Context())

			_, err := paginatedmessages.CreatePaginatedMessage(parsed.GuildData.GS.ID, parsed.GuildData.CS.ID, parsed.Args[0].Int(), 0,
				func(p *paginatedmessages.PaginatedMessage, newPage int) (*discordgo.MessageEmbed, error) {

					offset := (newPage - 1) * 10
					items, err := models.EconomyUsers(
						models.EconomyUserWhere.GuildID.EQ(parsed.GuildData.GS.ID),
						qm.OrderBy("waifu_item_worth + waifu_last_claim_amount + waifu_extra_worth desc"),
						qm.Limit(10),
						qm.Offset(offset),
					).AllG(context.Background())

					if err != nil {
						return nil, err
					}

					ids := make([]int64, len(items))
					for i, v := range items {
						ids[i] = v.UserID
					}
					users := bot.GetUsers(parsed.GuildData.GS.ID, ids...)

					embed := SimpleEmbedResponse(parsed.Author, "")
					embed.Title = "Waifu Leaderboard"

					for i, v := range items {
						user := users[i]
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:  fmt.Sprintf("#%d %s", i+offset+1, user.Username),
							Value: fmt.Sprintf("%s%d", conf.CurrencySymbol, WaifuWorth(v)),
						})

					}

					return embed, nil
				})

			return nil, err
		},
	}
	WaifuCmdInfo = &commands.YAGCommand{
		CmdCategory: CategoryWaifu,
		Name:        "waifuInfo",
		Description: "Shows waifu stats of you or your targets",
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Target", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			target := parsed.Author
			account := CtxUser(parsed.Context())
			originAccount := account
			conf := CtxConfig(parsed.Context())

			if parsed.Args[0].Value != nil {
				target = parsed.Args[0].User()
				var err error
				account, _, err = GetCreateAccount(parsed.Context(), target.ID, parsed.GuildData.GS.ID, conf.StartBalance)
				if err != nil {
					return nil, err
				}
			}

			embed := &discordgo.MessageEmbed{
				Author: UserEmebdAuthor(target),
				Color:  ColorBlue,
				Title:  "Waifu stats",
			}

			var usersToFetch []int64

			if account.WaifudBy != 0 {
				usersToFetch = append(usersToFetch, account.WaifudBy)
			}

			if account.WaifuAffinityTowards != 0 {
				usersToFetch = append(usersToFetch, account.WaifuAffinityTowards)
			}

			usersToFetch = append(usersToFetch, account.Waifus...)

			claimedByStr := "No one :("
			affinityStr := "No one"

			var waifus []*discordgo.User
			if len(usersToFetch) > 0 {
				waifus = bot.GetUsers(parsed.GuildData.GS.ID, usersToFetch...)
				if account.WaifudBy != 0 {
					claimedByStr = waifus[0].Username
					waifus = waifus[1:]
				}

				if account.WaifuAffinityTowards != 0 {
					affinityStr = waifus[0].Username
					waifus = waifus[1:]
				}
			}

			var claimedBuf strings.Builder
			if len(account.Waifus) > 0 {
				for _, v := range waifus {
					claimedBuf.WriteString(v.Username + "\n")
				}
			} else {
				claimedBuf.WriteString("No one...")
			}

			items, err := GetWaifuItems(parsed.GuildData.GS.ID, account.UserID)
			if err != nil {
				return nil, err
			}
			var itemsBuf strings.Builder
			for _, v := range items {
				if v.Item.Icon != "" {
					itemsBuf.WriteString(v.Item.Icon)
				} else {
					itemsBuf.WriteString(v.Item.Name)
				}

				itemsBuf.WriteString("x" + strconv.Itoa(v.Quantity) + " ")
			}
			if len(items) < 1 {
				itemsBuf.WriteString("None")
			}

			embed.Fields = []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:  "Price (for you)",
					Value: conf.CurrencySymbol + strconv.FormatInt(WaifuCost(originAccount, account), 10),
				},
				&discordgo.MessageEmbedField{
					Name:  "Claimed By",
					Value: claimedByStr,
				},
				&discordgo.MessageEmbedField{
					Name:  "Likes",
					Value: affinityStr,
				},
				&discordgo.MessageEmbedField{
					Name:  "Changes of hearth",
					Value: strconv.Itoa(account.WaifuAffinityChanges),
				},
				&discordgo.MessageEmbedField{
					Name:  "Divorces",
					Value: strconv.Itoa(account.WaifuDivorces),
				},
				&discordgo.MessageEmbedField{
					Name:  "Gifts",
					Value: itemsBuf.String(),
				},
				&discordgo.MessageEmbedField{
					Name:  "Waifus (" + strconv.Itoa(len(account.Waifus)) + ")",
					Value: claimedBuf.String(),
				},
			}

			return embed, nil
		},
	}
	WaifuCmdClaim = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "waifuClaim",
		Description:  "Claims the target as your waifu, using your wallet money, if no amount is specified it will use the lowest",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Target", Type: dcmd.AdvUserNoMember},
			&dcmd.ArgDef{Name: "Money", Type: &AmountArg{}},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())

			target := parsed.Args[0].User()
			if target.ID == u.ID {
				return ErrorEmbed(u, "You can't claim yourself, silly..."), nil
			}

			// pre-generate the account since its simpler with race conditions and whatnot in the transactions
			targetAccount, _, err := GetCreateAccount(parsed.Context(), target.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			if targetAccount.WaifudBy == u.ID {
				return ErrorEmbed(u, "You have already claimed this waifu"), nil
			}

			// safety checks
			cost := WaifuCost(account, targetAccount)
			claimAmount := cost
			if parsed.Args[1].Value != nil {

				claimAmount = parsed.Args[1].Value.(*AmountArgResult).Apply(account.MoneyWallet)
				if claimAmount < cost {
					return ErrorEmbed(u, "That waifu costs more than **%s%d** that to claim (%s%d)", conf.CurrencySymbol, claimAmount, conf.CurrencySymbol, cost), nil
				}
			}

			if account.MoneyWallet < claimAmount {
				return ErrorEmbed(u, "You don't have that much money in your wallet"), nil
			}

			forcedErrorResp := ""
			err = common.SqlTX(func(tx *sql.Tx) error {

				if targetAccount.WaifudBy != 0 {
					// update old owner account
					result, err := tx.Exec("UPDATE economy_users SET waifus = array_remove(waifus, $3) WHERE guild_id = $1 AND user_id = $2 AND $4 <@ waifus",
						parsed.GuildData.GS.ID, targetAccount.WaifudBy, target.ID, pq.Int64Array([]int64{target.ID}))
					if err != nil {
						return err
					}

					rows, err := result.RowsAffected()
					if err != nil {
						return err
					}

					if rows < 1 {
						return errors.New("failed updating tables, no rows, most likely a race condition")
					}
				}

				// update waifu
				numRows, err := models.EconomyUsers(qm.Where("guild_id = ? AND user_id = ? AND waifud_by = ?", parsed.GuildData.GS.ID, target.ID, targetAccount.WaifudBy)).UpdateAll(
					parsed.Context(), tx, models.M{"waifud_by": u.ID, "waifu_last_claim_amount": claimAmount})

				if err != nil {
					return err
				}

				if numRows < 1 {
					// forcedErrorResp = "That waifu is already claimed by soemone else :("
					return errors.New("Race condition, waifud by changed?")
				}

				_, err = tx.Exec("UPDATE economy_users SET waifus = waifus || $4, money_wallet = money_wallet - $3 WHERE guild_id = $1 AND user_id = $2",
					parsed.GuildData.GS.ID, u.ID, claimAmount, pq.Int64Array([]int64{target.ID}))
				return errors.Wrap(err, "update_waifus")
			})

			if err != nil {
				return nil, err
			}

			if forcedErrorResp != "" {
				return ErrorEmbed(u, forcedErrorResp), nil
			}

			return SimpleEmbedResponse(u, "Claimed **%s** as your waifu using **%s%d**!", target.Username, conf.CurrencySymbol, claimAmount), nil
		},
	}
	WaifuCmdReset = &commands.YAGCommand{
		CmdCategory: CategoryWaifu,
		Name:        "waifuReset",
		Description: "Resets your waifu stats and items, keeping your current waifus",
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())

			account.WaifuItemWorth = 0
			account.WaifuExtraWorth = 0

			_, err := account.UpdateG(parsed.Context(), boil.Whitelist("Waifu_extra_worth", "waifu_item_worth"))
			if err != nil {
				return nil, err
			}

			_, err = common.PQ.Exec("DELETE FROM economy_users_waifu_items WHERE guild_id = $1 AND user_id = $2", parsed.GuildData.GS.ID, account.UserID)
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(parsed.Author, "Reset your waifu stats, keeping your waifus"), nil
		},
	}
	WaifuCmdTransfer = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "WaifuTransfer",
		Description:  "Transfer the ownership of one of your waifus to another user. You must pay 10% of your waifu's value.",
		RequiredArgs: 2,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Waifu", Type: dcmd.AdvUserNoMember},
			&dcmd.ArgDef{Name: "New-Owner", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author
			account := CtxUser(parsed.Context())

			waifu := parsed.Args[0].User()
			newOwner := parsed.Args[1].User()

			conf := CtxConfig(parsed.Context())
			_, _, err := GetCreateAccount(parsed.Context(), newOwner.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			waifuAccount, _, err := GetCreateAccount(parsed.Context(), waifu.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			if !common.ContainsInt64Slice(account.Waifus, waifu.ID) {
				return ErrorEmbed(u, "That person is not your waifu >:u"), nil
			}

			if newOwner.ID == waifu.ID {
				return ErrorEmbed(u, "Can't transfer the waifu to itself!?"), nil
			}

			if newOwner.ID == u.ID {
				return ErrorEmbed(u, "Can't transfer the waifu to yourself!?"), nil
			}

			worth := WaifuWorth(waifuAccount)
			transferFee := int64(float64(worth) * 0.1)
			if account.MoneyWallet < transferFee {
				return ErrorEmbed(u, "Not enough money in your wallet to transfer this waifu (costs %s%d)", conf.CurrencySymbol, transferFee), nil
			}

			err = common.SqlTX(func(tx *sql.Tx) error {
				// update old owner account
				result, err := tx.Exec("UPDATE economy_users SET money_wallet = money_wallet - $3, waifus = array_remove(waifus, $4) WHERE guild_id = $1 AND user_id = $2 AND $5 <@ waifus",
					parsed.GuildData.GS.ID, u.ID, transferFee, waifu.ID, pq.Int64Array([]int64{waifu.ID}))
				if err != nil {
					return err
				}

				rows, err := result.RowsAffected()
				if err != nil {
					return err
				}

				if rows < 1 {
					return errors.New("failed updating tables, no rows, most likely a race condition")
				}

				// update new owner account
				_, err = tx.Exec("UPDATE economy_users SET waifus = waifus || $3 WHERE guild_id = $1 AND user_id = $2", parsed.GuildData.GS.ID, newOwner.ID, pq.Int64Array([]int64{waifu.ID}))
				if err != nil {
					return err
				}

				// update waifu
				_, err = tx.Exec("UPDATE economy_users SET waifud_by = $3 WHERE guild_id = $1 AND user_id = $2", parsed.GuildData.GS.ID, waifu.ID, newOwner.ID)
				return err
			})
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "Transferred **%s** to **%s** for **%s%d**", waifu.Username, newOwner.Username, conf.CurrencySymbol, transferFee), nil
		},
	}
	WaifuCmdDivorce = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "waifuDivorce",
		Description:  "Releases your claim on a specific waifu. You will get some of the money you've spent back unless that waifu has an affinity towards you. 6 hours cooldown.",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Waifu", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author

			waifu := parsed.Args[0].User()

			conf := CtxConfig(parsed.Context())

			waifuAccount, _, err := GetCreateAccount(parsed.Context(), waifu.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			if waifuAccount.WaifudBy != u.ID {
				return ErrorEmbed(u, "This person is not your waifu..."), nil
			}

			moneyBack := int64(float64(waifuAccount.WaifuLastClaimAmount) * 0.5)
			if waifuAccount.WaifuAffinityTowards == u.ID {
				// you get no money back >:u
				moneyBack = 0
			}

			err = common.SqlTX(func(tx *sql.Tx) error {
				// update old owner account
				result, err := tx.Exec("UPDATE economy_users SET money_wallet = money_wallet + $3, waifus = array_remove(waifus, $4), waifu_divorces = waifu_divorces+1 WHERE guild_id = $1 AND user_id = $2 AND $5 <@ waifus",
					parsed.GuildData.GS.ID, u.ID, moneyBack, waifu.ID, pq.Int64Array([]int64{waifu.ID}))
				if err != nil {
					return err
				}

				rows, err := result.RowsAffected()
				if err != nil {
					return err
				}

				if rows < 1 {
					return errors.New("failed updating tables, no rows, most likely a race condition")
				}

				// update waifu
				_, err = tx.Exec("UPDATE economy_users SET waifud_by = 0 WHERE guild_id = $1 AND user_id = $2", parsed.GuildData.GS.ID, waifu.ID)
				return err
			})

			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "You're now divorced with **%s**, you got back **%s%d**", waifu.Username, conf.CurrencySymbol, moneyBack), nil
		},
	}
	WaifuCmdAffinity = &commands.YAGCommand{
		CmdCategory: CategoryWaifu,
		Name:        "waifuAffinity",
		Description: "Sets your affinity towards someone you want to be claimed by. Setting affinity will reduce their claim on you by 20%. Provide no parameters to clear your affinity.",
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Target", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())
			u := parsed.Author

			resp := ""
			affinityMod := 0
			if account.WaifuAffinityTowards != 0 {
				affinityMod = 1
			}

			changed := false
			if parsed.Args[0].Value == nil {
				if account.WaifuAffinityTowards != 0 {
					changed = true
					account.WaifuAffinityTowards = 0
					resp = "Reset your affinity to no-one."
				} else {
					resp = "You already like no one..."
				}
			} else {
				target := parsed.Args[0].User()
				if account.WaifuAffinityTowards != target.ID {
					changed = true
					resp = "Set your affinity towards " + target.Username
					account.WaifuAffinityTowards = target.ID
				} else {
					resp = "You already like that person, try giving them gift to give them more affection?"
				}
			}

			if !changed {
				return ErrorEmbed(u, "%s", resp), nil
			}

			// check cooldown
			var cdResp string
			err := common.RedisPool.Do(radix.Cmd(&cdResp, "SET", fmt.Sprintf("economy_affinity_cd:%d:%d", parsed.GuildData.GS.ID, u.ID), "1", "EX", "1800", "NX"))
			if err != nil {
				return nil, err
			}

			if cdResp != "OK" {
				return ErrorEmbed(u, "This command is still on cooldown"), nil
			}

			_, err = common.PQ.Exec("UPDATE economy_users SET waifu_affinity_towards = $3, waifu_affinity_changes = waifu_affinity_changes + $4 WHERE user_id = $2 AND guild_id = $1",
				parsed.GuildData.GS.ID, u.ID, account.WaifuAffinityTowards, affinityMod)
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, resp), nil
		},
	}
	WaifuCmdGift = &commands.YAGCommand{
		CmdCategory: CategoryWaifu,
		Name:        "waifuGift",
		Description: "Gift an item to someone. This will increase their waifu value by 50% of the gifted item's value if you are not their waifu, or 95% if you are. Provide no parameters to see a list of items that you can gift.",
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Item", Type: dcmd.String},
			&dcmd.ArgDef{Name: "Target", Type: dcmd.AdvUserNoMember},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())
			u := parsed.Author

			if parsed.Args[0].Value == nil {
				return ListWaifuItems(parsed.GuildData.GS.ID, parsed.GuildData.CS.ID, u, account.MoneyWallet, conf.CurrencySymbol)
			}

			if parsed.Args[1].Value == nil {
				return ErrorEmbed(u, "Re-run the command with the user you wanna gift it to included"), nil
			}

			// itemToBuy, err := models.FindEconomyWaifuItemG(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int64())
			itemToBuy, err := FindWaifuItemByName(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Str(), nil)
			if err != nil {
				return nil, err
			}

			if itemToBuy == nil {
				return ErrorEmbed(u, "Unknown item"), nil
			}

			if int64(itemToBuy.Price) > account.MoneyWallet {
				return ErrorEmbed(u, "You don't have enough money in your wallet to gift this item"), nil
			}

			// ensure the target has a wallet
			target := parsed.Args[1].User()
			_, _, err = GetCreateAccount(parsed.Context(), target.ID, parsed.GuildData.GS.ID, conf.StartBalance)
			if err != nil {
				return nil, err
			}

			worthIncreaseModifier := 0.5
			if account.WaifudBy == target.ID {
				worthIncreaseModifier = 0.95
			}

			err = common.SqlTX(func(tx *sql.Tx) error {
				// deduct money from our account
				err = TransferMoneyWallet(parsed.Context(), tx, conf, false, u.ID, common.BotUser.ID, int64(itemToBuy.Price), int64(itemToBuy.Price))
				// _, err := tx.Exec("UPDATE economy_users SET money_wallet = money_wallet - $3 WHERE guild_id = $1 AND user_id = $2", parsed.GuildData.GS.ID, u.ID, itemToBuy.Price)
				if err != nil {
					return err
				}

				worthIncrease := int64(float64(itemToBuy.Price) * worthIncreaseModifier)

				// add the item
				_, err := tx.Exec(`
INSERT INTO economy_users_waifu_items  (guild_id, user_id, item_id, quantity)
VALUES ($1, $2, $3, 1)
ON CONFLICT (guild_id, user_id, item_id)
DO UPDATE SET 
quantity = economy_users_waifu_items.quantity + 1`, parsed.GuildData.GS.ID, target.ID, itemToBuy.LocalID)
				if err != nil {
					return err
				}

				// increase their worth
				_, err = tx.Exec("UPDATE economy_users SET waifu_item_worth = waifu_item_worth + $3, gambling_boost_percentage=gambling_boost_percentage+$4 WHERE guild_id = $1 AND user_id = $2",
					parsed.GuildData.GS.ID, target.ID, worthIncrease, itemToBuy.GamblingBoost)
				return err
			})

			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "Gifted **%s** to **%s** for **%s%d**!", itemToBuy.Name, target.Username, conf.CurrencySymbol, itemToBuy.Price), nil
		},
	}

	/*
		Shop
	*/

	WaifuShopAdd = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "waifuItemAdd",
		Description:  "Adds an item to the waifu shop, only economy adins can use this",
		RequiredArgs: 3,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Price", Type: dcmd.Int},
			&dcmd.ArgDef{Name: "Icon", Type: dcmd.String},
			&dcmd.ArgDef{Name: "Name", Type: dcmd.String},
		},
		ArgSwitches: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "boost", Help: "Gambling boost", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			localID, err := common.GenLocalIncrID(parsed.GuildData.GS.ID, "economy_item")
			if err != nil {
				return nil, err
			}

			m := &models.EconomyWaifuItem{
				GuildID:       parsed.GuildData.GS.ID,
				LocalID:       localID,
				Price:         parsed.Args[0].Int(),
				Icon:          parsed.Args[1].Str(),
				Name:          parsed.Args[2].Str(),
				GamblingBoost: parsed.Switch("boost").Int(),
			}

			err = m.InsertG(parsed.Context(), boil.Infer())
			if err != nil {
				return nil, err
			}

			extraStr := ""
			if m.GamblingBoost > 0 {
				extraStr = fmt.Sprintf(" with a gambling profit boost of +%d%%", m.GamblingBoost)
			}

			conf := CtxConfig(parsed.Context())
			return SimpleEmbedResponse(parsed.Author, "Added **%s** to the shop at the price of **%s%d**%s",
				m.Name, conf.CurrencySymbol, m.Price, extraStr), nil
		},
	}
	WaifuShopEdit = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "waifuItemEdit",
		Description:  "Edits an item in the waifu shop",
		RequiredArgs: 4,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Item", Type: dcmd.String},
			&dcmd.ArgDef{Name: "Price", Type: dcmd.Int},
			&dcmd.ArgDef{Name: "Icon", Type: dcmd.String},
			&dcmd.ArgDef{Name: "Name", Type: dcmd.String},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {

			// item, err := models.FindEconomyWaifuItemG(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int64())
			item, err := FindWaifuItemByName(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Str(), nil)
			if err != nil {
				return nil, err
			}

			if item == nil {
				return ErrorEmbed(parsed.Author, "No item by that id"), nil
			}

			item.Price = parsed.Args[1].Int()
			item.Icon = parsed.Args[2].Str()
			item.Name = parsed.Args[3].Str()

			_, err = item.UpdateG(parsed.Context(), boil.Whitelist("price", "icon", "name"))
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(parsed.Author, "Updated **%s**", item.Name), nil
		},
	}
	WaifuCmdDel = &commands.YAGCommand{
		CmdCategory:  CategoryWaifu,
		Name:         "waifuItemDel",
		Description:  "Removes a item from the waifu shop",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Item", Type: dcmd.String},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {

			item, err := FindWaifuItemByName(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Str(), nil)
			if err != nil {
				return nil, err
			}

			if item == nil {
				return ErrorEmbed(parsed.Author, "No item by that ID"), nil
			}

			_, err = item.DeleteG(parsed.Context())
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(parsed.Author, "Deleted item ID %d", parsed.Args[0].Int()), nil
		},
	}
)

func FindWaifuItemByName(ctx context.Context, guildID int64, name string, items []*models.EconomyWaifuItem) (item *models.EconomyWaifuItem, err error) {
	if items == nil {
		items, err = models.EconomyWaifuItems(models.EconomyWaifuItemWhere.GuildID.EQ(guildID), qm.OrderBy("local_id asc")).AllG(ctx)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range items {
		if strings.EqualFold(name, v.Name) {
			return v, nil
		}
	}

	return nil, nil
}

func WaifuWorth(target *models.EconomyUser) int64 {
	const base = 100

	worth := base + target.WaifuItemWorth + target.WaifuExtraWorth + target.WaifuLastClaimAmount
	return worth
}

func WaifuCost(from, target *models.EconomyUser) int64 {
	worth := WaifuWorth(target)

	cost := int64(float64(worth) * 1.1)

	if from != nil {
		if target.WaifuAffinityTowards == from.UserID {
			cost = int64(float64(cost) * 0.8)
		}
	}

	if target.WaifudBy != 0 {
		// 10% more expensive if they're claimed already
		cost = int64(float64(cost) * 1.1)
	}

	return cost

}

func ListWaifuItems(guildID, channelID int64, u *discordgo.User, currentMoney int64, currencySymbol string) (*discordgo.MessageEmbed, error) {
	_, err := paginatedmessages.CreatePaginatedMessage(guildID, channelID, 1, 0, func(p *paginatedmessages.PaginatedMessage, newPage int) (*discordgo.MessageEmbed, error) {

		offset := (newPage - 1) * 12

		items, err := models.EconomyWaifuItems(models.EconomyWaifuItemWhere.GuildID.EQ(guildID), qm.OrderBy("local_id asc"), qm.Limit(12), qm.Offset(offset)).AllG(context.Background())
		if err != nil {
			return nil, err
		}

		embed := SimpleEmbedResponse(u, "")
		if len(items) < 1 {
			embed.Description = "No items :("
		}

		embed.Title = "Waifu gift shot!"

		for _, v := range items {
			extraVal := ""
			if v.GamblingBoost > 0 {
				extraVal = fmt.Sprintf(" (+%d%% Gambling)", v.GamblingBoost)
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s %s", v.Icon, v.Name),
				Value:  fmt.Sprintf("%d%s", v.Price, extraVal),
				Inline: true,
			})
		}

		return embed, nil
	})

	return nil, err

}

type WaifuItemsResult struct {
	Item     *models.EconomyWaifuItem
	Quantity int
}

func GetWaifuItems(guildID, waifuID int64) ([]*WaifuItemsResult, error) {
	rows, err := common.PQ.Query(`
SELECT economy_users_waifu_items.item_id,
economy_users_waifu_items.quantity,
economy_waifu_items.name,
economy_waifu_items.icon,
economy_waifu_items.price FROM economy_users_waifu_items
INNER JOIN economy_waifu_items ON 
economy_waifu_items.local_id = economy_users_waifu_items.item_id AND economy_waifu_items.guild_id = economy_users_waifu_items.guild_id
WHERE economy_users_waifu_items.guild_id = $1 AND economy_users_waifu_items.user_id = $2`, guildID, waifuID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*WaifuItemsResult, 0)

	for rows.Next() {

		dst := &models.EconomyWaifuItem{
			GuildID: guildID,
		}
		quantity := 0

		if err := rows.Scan(&dst.LocalID, &quantity, &dst.Name, &dst.Icon, &dst.Price); err != nil {
			return nil, err
		}

		results = append(results, &WaifuItemsResult{
			Item:     dst,
			Quantity: quantity,
		})
	}

	return results, nil
}
