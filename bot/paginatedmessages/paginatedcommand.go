package paginatedmessages

import (
	"github.com/mrbentarikau/pagst/lib/dcmd"
)

type CtxKey int

const CtxKeyNoPagination CtxKey = 1

type PaginatedCommandFunc func(data *dcmd.Data, p *PaginatedMessage, page int) (interface{}, error)

func PaginatedCommand(pageArgIndex int, cb PaginatedCommandFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		page := 1
		if pageArgIndex > -1 {
			page = data.Args[pageArgIndex].Int()
		}

		if page < 1 {
			page = 1
		}

		if data.Context().Value(CtxKeyNoPagination) != nil {
			return cb(data, nil, page)
		}

		pm, err := CreatePaginatedMessage(data.GuildData.GS.ID, data.GuildData.CS.ID, page, 0, func(p *PaginatedMessage, page int) (interface{}, error) {
			return cb(data, p, page)
		})

		return pm, err
	}
}
