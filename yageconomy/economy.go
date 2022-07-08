package yageconomy

//go:generate sqlboiler --no-hooks psql

import (
	"context"
	"database/sql"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

var logger = common.GetPluginLogger(&Plugin{})

func RegisterPlugin() {
	plugin := &Plugin{}
	common.InitSchemas("economy", DBSchemas...)
	common.RegisterPlugin(plugin)
}

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {

	return &common.PluginInfo{
		Name:     "Economy",
		SysName:  "economy",
		Category: common.PluginCategoryMisc,
	}
}

const (
	DefaultCurrencyName   = "YAGBuck"
	DefaultCurrencySymbol = "$"
)

func DefaultConfig(g int64) *models.EconomyConfig {
	return &models.EconomyConfig{
		GuildID:            g,
		CurrencyName:       DefaultCurrencyName,
		CurrencyNamePlural: DefaultCurrencyName + "s",
		CurrencySymbol:     DefaultCurrencySymbol,
		StartBalance:       1000,

		DailyFrequency: 1440,
		DailyAmount:    250,

		ChatmoneyFrequency: 100,
		ChatmoneyAmountMin: 10,
		ChatmoneyAmountMax: 50,

		FishingMaxWinAmount: 200,
		FishingMinWinAmount: 50,
		FishingCooldown:     30,

		AutoPlantMin:    50,
		AutoPlantMax:    200,
		AutoPlantChance: types.NewDecimal(decimal.New(2, 2)),

		RobFine:     25,
		RobCooldown: 600,

		HeistServerCooldown:            1,
		HeistFailedGamblingBanDuration: 60,
	}
}

func GuildConfigOrDefault(ctx context.Context, guildID int64) (*models.EconomyConfig, error) {
	conf, err := models.FindEconomyConfigG(ctx, guildID)
	if err != nil {
		if err == sql.ErrNoRows {
			return DefaultConfig(guildID), nil
		}

		return nil, err
	}

	return conf, nil
}
