package moderation

import (
	"fmt"
	"time"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/templates"
	"github.com/mrbentarikau/pagst/logs"
	"github.com/jinzhu/gorm"
)

func init() {
	templates.RegisterSetupFunc(func(ctx *templates.Context) {
		ctx.ContextFuncs["getWarnings"] = tmplGetWarns(ctx)
	})
}

// getWarns returns a slice of all warnings the target user has.
func tmplGetWarns(ctx *templates.Context) interface{} {
	return func(target interface{}) ([]*WarningModel, error) {
		if ctx.IncreaseCheckGenericAPICall() {
			return nil, nil
		}

		gID := ctx.GS.ID
		var warns []*WarningModel
		targetID := templates.TargetUserID(target)

		if targetID == 0 {
			return nil, fmt.Errorf("unknown user %v to get warnings for", target)
		}

		err := common.GORM.Where("user_id = ? AND guild_id = ?", targetID, gID).Order("id desc").Find(&warns).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		for i := range warns {
			purgedWarnLogs := logs.ConfRestrictLogsThirtyDays.GetBool() && warns[i].CreatedAt.Before(time.Now().AddDate(0, 0, -30))
			if purgedWarnLogs {
				warns[i].LogsLink = ""
			}
		}

		return warns, nil
	}
}
