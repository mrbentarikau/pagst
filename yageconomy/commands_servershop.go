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
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	ItemTypeRole          = 0
	ItemTypeList          = 1
	ItemTypeGamblingBoost = 2
)

var ShopCommands = []*commands.YAGCommand{
	&commands.YAGCommand{
		CmdCategory: CategoryEconomy,
		Name:        "Shop",
		Description: "Shows the items available in the server shop",
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())

			_, err := paginatedmessages.CreatePaginatedMessage(parsed.GuildData.GS.ID, parsed.GuildData.CS.ID, 1, 0, func(p *paginatedmessages.PaginatedMessage, newPage int) (*discordgo.MessageEmbed, error) {
				offset := (newPage - 1) * 12

				items, err := models.EconomyShopItems(
					models.EconomyShopItemWhere.GuildID.EQ(parsed.GuildData.GS.ID),
					qm.OrderBy("local_id asc"),
					qm.Limit(12),
					qm.Offset(offset),
				).AllG(context.Background())
				if err != nil {
					return nil, err
				}

				embed := SimpleEmbedResponse(u, "")
				embed.Title = "Server Shop!"

				for i, v := range items {
					name := v.Name
					if v.RoleID != 0 {
						r := parsed.GuildData.GS.GetRole(v.RoleID)
						if r != nil {
							name = r.Name
						} else {
							name += "(deleted-role)"
						}
					}

					canAffordStr := ""
					if account.MoneyWallet < v.Cost {
						canAffordStr = "(you can't afford)"
					}

					typStr := "role"
					if v.Type == ItemTypeList {
						typStr = "list"
					} else if v.Type == ItemTypeGamblingBoost {
						typStr = fmt.Sprintf("+%d%% Gambling", v.GamblingBoostPercentage)
					}

					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("#%d - %s", i+offset+1, name),
						Value:  fmt.Sprintf("%s - %s%d%s", typStr, conf.CurrencySymbol, v.Cost, canAffordStr),
						Inline: true,
					})
				}
				if len(items) < 1 {
					embed.Description = "(no items)"
				}

				return embed, nil
			})

			return nil, err
		},
	},
	&commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "Buy",
		Description:  "Buys an item from the server shop",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "item", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author
			account := CtxUser(parsed.Context())
			conf := CtxConfig(parsed.Context())

			shopItem, err := FindServerShopItemByOrderIndex(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int(), nil)
			if err != nil {
				return nil, err
			}
			if shopItem == nil {
				return ErrorEmbed(parsed.Author, "No shop item with that ID"), nil
			}

			if shopItem.Cost > account.MoneyWallet {
				return ErrorEmbed(u, "Not enough money in your wallet :("), nil
			}

			forcedErrorMsg := ""

			// do this in a transaction so we can rollback the purchase if the delivery failed
			err = common.SqlTX(func(tx *sql.Tx) error {
				listValue := ""

				if shopItem.Type == ItemTypeList {

					query := `UPDATE economy_shop_list_items
SET    purchased_by = $3
WHERE  value = (
	SELECT value
	FROM   economy_shop_list_items
	WHERE  purchased_by = 0 AND guild_id = $1 AND list_id = $2
	LIMIT  1
	FOR    UPDATE
)
RETURNING value;`

					row := tx.QueryRow(query, parsed.GuildData.GS.ID, shopItem.LocalID, u.ID)

					err = row.Scan(&listValue)
					if err != nil {
						if errors.Cause(err) == sql.ErrNoRows {
							forcedErrorMsg = "No more iteu in that list available"
							return nil
						}
						return err
					}
				}

				err = TransferMoneyWallet(parsed.Context(), tx, conf, false, u.ID, common.BotUser.ID, shopItem.Cost, shopItem.Cost)
				if err != nil {
					return err
				}

				_, err = tx.Exec("UPDATE economy_users SET gambling_boost_percentage = gambling_boost_percentage + $3 WHERE guild_id = $1 AND user_id = $2",
					parsed.GuildData.GS.ID, u.ID, shopItem.GamblingBoostPercentage)

				if err != nil {
					return err
				}

				// deliver the item
				if shopItem.Type == ItemTypeList {
					err = bot.SendDMEmbed(u.ID, SimpleEmbedResponse(u, "You purhcased one of **%s**, here it is: ||%s||", shopItem.Name, listValue))
				} else if shopItem.Type == ItemTypeRole {
					ms := parsed.GuildData.MS
					if sm := bot.State.GetMember(parsed.GuildData.GS.ID, ms.User.ID); sm != nil {
						// Prefer state member over the one provided in the message, since it may have presence data
						ms = sm
					}
					err = common.AddRoleDS(ms, shopItem.RoleID)
				}

				return err

			})
			if forcedErrorMsg != "" {
				return ErrorEmbed(u, forcedErrorMsg), err
			}

			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "You purchased **%s** for **%s%d**!", shopItem.Name, conf.CurrencySymbol, shopItem.Cost), nil
		},
	},
}

var ShopAdminCommands = []*commands.YAGCommand{
	&commands.YAGCommand{
		CmdCategory:     CategoryEconomy,
		Name:            "ShopAdd",
		Description:     "Adds an item to the shop, only economy admins can use this command",
		LongDescription: "Types are 'role', 'list' and 'gamblingboostx[percentage]' where percentage is the gambling boost percentage\n\nExample: -shopadd gamblingboostx10 1000 10% gambling income increase\nThat will add a item with 10% gambling income increase and with the name '10% gambling income increase'",
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Type", Type: dcmd.String},
			&dcmd.ArgDef{Name: "Price", Type: dcmd.Int},
			&dcmd.ArgDef{Name: "Name", Type: dcmd.String},
		},
		RequiredArgs: 3,
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author

			tStr := strings.ToLower(parsed.Args[0].Str())

			t := ItemTypeRole
			gamblingBoostPercentage := int64(0)
			if tStr == "list" {
				t = ItemTypeList
			} else if strings.HasPrefix(tStr, "gamblingboostx") {
				t = ItemTypeGamblingBoost
				split := strings.Split(tStr, "x")
				if len(split) < 2 {
					return ErrorEmbed(u, "No boost percentage specified, example: `gamblingboostx10` for 10%%"), nil
				}

				var err error
				gamblingBoostPercentage, err = strconv.ParseInt(split[1], 10, 32)
				if err != nil {
					return nil, err
				}
			}

			lID, err := common.GenLocalIncrID(parsed.GuildData.GS.ID, "economy_shop_item")
			if err != nil {
				return nil, err
			}

			roleID := int64(0)
			name := parsed.Args[2].Str()

			// this is a role
			if t == ItemTypeRole {
				for _, v := range parsed.GuildData.GS.Roles {
					if strings.EqualFold(v.Name, name) {
						roleID = v.ID
						name = v.Name
						break
					}
				}
				if roleID == 0 {
					return ErrorEmbed(u, "Unknown role %q", name), nil
				}
			}

			m := &models.EconomyShopItem{
				GuildID: parsed.GuildData.GS.ID,
				LocalID: lID,

				Type: int16(t),

				Cost:                    int64(parsed.Args[1].Int()),
				Name:                    name,
				RoleID:                  roleID,
				GamblingBoostPercentage: int(gamblingBoostPercentage),
			}

			err = m.InsertG(parsed.Context(), boil.Infer())
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "Added **%s** to the shop.", name), nil
		},
	},
	&commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "ShopListAdd",
		Description:  "Adds a item to a shop list, only economy admin can use this command",
		RequiredArgs: 2,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "List-ID", Type: dcmd.Int},
			&dcmd.ArgDef{Name: "Item", Type: dcmd.String},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			u := parsed.Author

			shopItem, err := FindServerShopItemByOrderIndex(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int(), nil)
			if err != nil {
				return nil, err
			}
			if shopItem == nil {
				return ErrorEmbed(parsed.Author, "No shop item with that ID"), nil
			}

			if shopItem.Type != 1 {
				return ErrorEmbed(u, "That shop item is not a list"), nil
			}

			lID, err := common.GenLocalIncrID(parsed.GuildData.GS.ID, "economy_shop_list_item")
			if err != nil {
				return nil, err
			}

			m := &models.EconomyShopListItem{
				GuildID: parsed.GuildData.GS.ID,
				LocalID: lID,
				ListID:  shopItem.LocalID,
				Value:   parsed.Args[1].Str(),
			}

			err = m.InsertG(parsed.Context(), boil.Infer())
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(u, "Added to the list **%s**", shopItem.Name), nil
		},
	},
	&commands.YAGCommand{
		CmdCategory:  CategoryEconomy,
		Name:         "ShopRem",
		Aliases:      []string{"ShopDel"},
		Description:  "Removes a item from the shop, only economy admins can use this command",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "ID", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {

			// shopItem, err := models.FindEconomyShopItemG(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int64())
			shopItem, err := FindServerShopItemByOrderIndex(parsed.Context(), parsed.GuildData.GS.ID, parsed.Args[0].Int(), nil)
			if err != nil {
				return nil, err
			}
			if shopItem == nil {
				return ErrorEmbed(parsed.Author, "No shop item with that ID"), nil
			}

			_, err = shopItem.DeleteG(parsed.Context())
			if err != nil {
				return nil, err
			}

			return SimpleEmbedResponse(parsed.Author, "Deleted **%s** from the server shop", shopItem.Name), nil
		},
	},
}

func FindServerShopItemByOrderIndex(ctx context.Context, guildID int64, orderIndex int, items []*models.EconomyShopItem) (item *models.EconomyShopItem, err error) {
	if items == nil {
		items, err = models.EconomyShopItems(models.EconomyShopItemWhere.GuildID.EQ(guildID), qm.OrderBy("local_id asc")).AllG(ctx)
		if err != nil {
			return nil, err
		}
	}

	orderIndex -= 1

	if len(items) <= orderIndex || orderIndex < 0 {
		return nil, nil
	}

	return items[orderIndex], nil
}
