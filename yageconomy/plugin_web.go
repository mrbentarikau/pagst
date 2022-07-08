package yageconomy

import (
	"bytes"
	"database/sql"
	_ "embed"
	"image"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/web"
	"github.com/mrbentarikau/pagst/yageconomy/models"
	"github.com/ericlagergren/decimal"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"goji.io"
	"goji.io/pat"
)

//go:embed assets/economy.html
var PageHTML string

type PostConfigForm struct {
	Enabled                        bool
	Admins                         []int64 `valid:"role,true"`
	EnabledChannels                []int64 `valid:"channel,true"`
	CurrencyName                   string  `valid:",1,50"`
	CurrencyNamePlural             string  `valid:",1,50"`
	CurrencySymbol                 string  `valid:",1,50"`
	DailyFrequency                 int64
	DailyAmount                    int64
	ChatmoneyFrequency             int64
	ChatmoneyAmountMin             int64
	ChatmoneyAmountMax             int64
	StartBalance                   int64
	AutoPlantChannels              []int64 `valid:"channel,true"`
	AutoPlantMin                   int64
	AutoPlantMax                   int64
	AutoPlantChance                float64
	RobFine                        int
	RobCooldown                    int
	FishingCooldown                int
	FishingMaxwinAmount            int64
	FishingMinWinAmount            int64
	HeistServerCooldown            int
	HeistFailedGamblingBanDuration int
	HeistFixedPayout               int
}

func (p PostConfigForm) DBModel() *models.EconomyConfig {
	return &models.EconomyConfig{
		Enabled:                        p.Enabled,
		Admins:                         p.Admins,
		EnabledChannels:                p.EnabledChannels,
		CurrencyName:                   p.CurrencyName,
		CurrencyNamePlural:             p.CurrencyNamePlural,
		CurrencySymbol:                 p.CurrencySymbol,
		DailyFrequency:                 p.DailyFrequency,
		DailyAmount:                    p.DailyAmount,
		ChatmoneyFrequency:             p.ChatmoneyFrequency,
		ChatmoneyAmountMin:             p.ChatmoneyAmountMin,
		ChatmoneyAmountMax:             p.ChatmoneyAmountMax,
		StartBalance:                   p.StartBalance,
		AutoPlantChannels:              p.AutoPlantChannels,
		AutoPlantMin:                   p.AutoPlantMin,
		AutoPlantMax:                   p.AutoPlantMax,
		AutoPlantChance:                types.NewDecimal(decimal.New(int64(p.AutoPlantChance*100), 4)),
		RobFine:                        p.RobFine,
		RobCooldown:                    p.RobCooldown,
		FishingCooldown:                p.FishingCooldown,
		FishingMaxWinAmount:            p.FishingMaxwinAmount,
		FishingMinWinAmount:            p.FishingMinWinAmount,
		HeistServerCooldown:            p.HeistServerCooldown,
		HeistFailedGamblingBanDuration: p.HeistFailedGamblingBanDuration,
		HeistFixedPayout:               p.HeistFixedPayout,
	}
}

func (p *Plugin) InitWeb() {
	web.AddHTMLTemplate("yageconomy/assets/economy.html", PageHTML)
	web.AddSidebarItem(web.SidebarCategoryFun, &web.SidebarItem{
		Name: "Economy",
		URL:  "economy",
	})

	subMux := goji.SubMux()

	web.CPMux.Handle(pat.New("/economy"), subMux)
	web.CPMux.Handle(pat.New("/economy/*"), subMux)

	//subMux.Use(web.RequireGuildChannelsMiddleware)

	mainGetHandler := web.ControllerHandler(handleGetEconomy, "cp_economy_settings")

	subMux.Handle(pat.Get(""), mainGetHandler)
	subMux.Handle(pat.Get("/"), mainGetHandler)

	subMux.Handle(pat.Get("/pick_image/:image_id"), http.HandlerFunc(handleGetPickImage))

	subMux.Handle(pat.Post("/delete_pick_image/:image_id"), web.ControllerPostHandler(handleDeleteImage, mainGetHandler, nil))
	subMux.Handle(pat.Post("/pick_image"), web.ControllerPostHandler(handleUploadImage, mainGetHandler, nil))
	subMux.Handle(pat.Post(""), web.ControllerPostHandler(handlePostEconomy, mainGetHandler, PostConfigForm{}))
	subMux.Handle(pat.Post("/"), web.ControllerPostHandler(handlePostEconomy, mainGetHandler, PostConfigForm{}))
}

func tmplFormatPercentage(in *decimal.Big) string {
	result := in.Mul(in, decimal.New(100, 0))
	return result.String()
}

func handleGetEconomy(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	g, templateData := web.GetBaseCPContextData(r.Context())

	imgs, err := models.EconomyPickImages2s(qm.Select("id"), qm.Where("guild_id=?", g.ID)).AllG(r.Context())
	if err != nil {
		return templateData, err
	}

	templateData["PickImages"] = imgs

	templateData["fmtDecimalPercentage"] = tmplFormatPercentage

	if templateData["PluginSettings"] == nil {
		conf, err := models.FindEconomyConfigG(r.Context(), g.ID)
		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				conf = DefaultConfig(g.ID)
			} else {
				return templateData, err
			}
		}

		templateData["PluginSettings"] = conf
	}

	return templateData, nil
}

func handlePostEconomy(w http.ResponseWriter, r *http.Request) (templateData web.TemplateData, err error) {
	g, templateData := web.GetBaseCPContextData(r.Context())

	form := r.Context().Value(common.ContextKeyParsedForm).(*PostConfigForm)
	conf := form.DBModel()
	conf.HeistLastUsage = time.Time{}
	conf.GuildID = g.ID

	templateData["PluginSettings"] = conf

	err = conf.UpsertG(r.Context(), true, []string{"guild_id"}, boil.Blacklist("heist_last_usage"), boil.Infer())
	return templateData, nil
}

func handleGetPickImage(w http.ResponseWriter, r *http.Request) {
	g, _ := web.GetBaseCPContextData(r.Context())

	imageID, _ := strconv.ParseInt(pat.Param(r, "image_id"), 10, 64)

	row, err := models.EconomyPickImages2s(qm.Where("guild_id=?", g.ID), qm.Where("id=?", imageID)).OneG(r.Context())
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			w.WriteHeader(404)
			return
		}

		web.CtxLogger(r.Context()).WithError(err).Error("failed retrieving econ pick image")
		w.WriteHeader(500)
		return
	}

	w.Write(row.Image)
}

func handleDeleteImage(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	g, tmplData := web.GetBaseCPContextData(r.Context())

	imageID, _ := strconv.ParseInt(pat.Param(r, "image_id"), 10, 64)

	_, err := models.EconomyPickImages2s(qm.Where("guild_id=?", g.ID), qm.Where("id=?", imageID)).DeleteAll(r.Context(), common.PQ)
	return tmplData, err
}

func handleUploadImage(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	g, tmpl := web.GetBaseCPContextData(ctx)

	file, header, err := r.FormFile("image")
	if err != nil {
		return tmpl, err
	}

	if header.Size > 250000 {
		return tmpl.AddAlerts(web.ErrorAlert("Max image size is 250KB")), nil
	}

	buf := make([]byte, int(header.Size))
	_, err = io.ReadFull(file, buf)
	if err != nil {
		return tmpl, err
	}

	imgHeader, _, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return tmpl, err
	}

	if imgHeader.Width > 1080 || imgHeader.Height > 1920 {
		return tmpl.AddAlerts(web.ErrorAlert("Max image size is 1080x1920")), nil
	}

	m := models.EconomyPickImages2{
		GuildID: g.ID,
		Image:   buf,
	}

	err = m.InsertG(r.Context(), boil.Infer())
	return tmpl, err
}
