package util

import (
	"github.com/jonas747/dcmd"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
)

func RequireBotAdmin(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		if admin, err := bot.IsBotAdmin(data.Msg.Author.ID); admin && err == nil {
			return inner(data)
		}

		return "", nil
	}
}

func RequireOwner(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		if common.IsOwner(data.Msg.Author.ID) {
			return inner(data)
		}

		return "", nil
	}
}
