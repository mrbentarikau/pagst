package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/cplogs"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
	"github.com/sirupsen/logrus"
	"goji.io/pattern"
)

type redditQuoteStruct struct {
	Data struct {
		Children []struct {
			Data struct {
				Selfttext string `json:"selftext"`
			}
		}
	}
}

type discordStatusStruct struct {
	Page struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		URL       string `json:"url"`
		TimeZone  string `json:"time_zone"`
		UpdatedAt string `json:"updated_at"`
	} `json:"page"`
	Incidents []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"incidents"`
}

var ErrTokenExpired = errors.New("OAUTH2 Token expired")

var panelLogKeyCore = cplogs.RegisterActionFormat(&cplogs.ActionFormat{
	Key:          "save_core_config",
	FormatString: "Updated core config",
})

func getRedditQuote() string {
	var redditQuoteQuery []redditQuoteStruct
	var RedditHost = "https://old.reddit.com/"
	var RedditJSON = "r/caubert/random.json"

	queryURL := RedditHost + RedditJSON
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return fmt.Sprint("Error: ", err)
	}

	req.Header.Set("User-Agent", "PAGST/20.42.6702 /u/caubert")
	//req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprint("Error: ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprint("Error: ", err)
	}

	queryErr := json.Unmarshal(body, &redditQuoteQuery)

	if queryErr != nil {
		return ""
		//return fmt.Sprint("Error: ", queryErr)
	}

	return redditQuoteQuery[0].Data.Children[0].Data.Selfttext
}

func getDiscordStatus() (*discordStatusStruct, error) {
	discordStatus := discordStatusStruct{}
	var DiscordHost = "https://discord.statuspage.io/"
	var DiscordJSON = "api/v2/incidents/unresolved.json"

	queryURL := DiscordHost + DiscordJSON
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", common.BotUserAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryErr := json.Unmarshal(body, &discordStatus)

	if queryErr != nil {
		return nil, nil
		//return fmt.Sprint("Error: ", queryErr)
	}

	return &discordStatus, err
}

func SetContextTemplateData(ctx context.Context, data map[string]interface{}) context.Context {
	// Check for existing data
	if val := ctx.Value(common.ContextKeyTemplateData); val != nil {
		cast := val.(TemplateData)
		for k, v := range data {
			cast[k] = v
		}
		return ctx
	}

	// Fallback
	return context.WithValue(ctx, common.ContextKeyTemplateData, TemplateData(data))
}

func DiscordSessionFromContext(ctx context.Context) *discordgo.Session {
	if val := ctx.Value(common.ContextKeyDiscordSession); val != nil {
		if cast, ok := val.(*discordgo.Session); ok {
			return cast
		}
	}
	return nil
}

func RandBase64(size int) string {
	b := make([]byte, size)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

func GenSessionCookie() *http.Cookie {
	data := RandBase64(32)
	cookie := &http.Cookie{
		Name:   "pagstdb-session",
		Value:  data,
		MaxAge: 86400,
		Path:   "/",
	}
	return cookie
}

func LogIgnoreErr(err error) {
	if err != nil {
		logger.Error("Error:", err)
	}
}

type TemplateData map[string]interface{}

func (t TemplateData) AddAlerts(alerts ...*Alert) TemplateData {
	if t["Alerts"] == nil {
		t["Alerts"] = make([]*Alert, 0)
	}

	t["Alerts"] = append(t["Alerts"].([]*Alert), alerts...)
	return t
}

func (t TemplateData) Alerts() []*Alert {
	if v, ok := t["Alerts"]; ok {
		return v.([]*Alert)
	}

	return nil
}

func GetCreateTemplateData(ctx context.Context) (context.Context, TemplateData) {
	if v := ctx.Value(common.ContextKeyTemplateData); v != nil {
		return ctx, v.(TemplateData)
	}
	tmplData := TemplateData(make(map[string]interface{}))
	ctx = context.WithValue(ctx, common.ContextKeyTemplateData, tmplData)
	return ctx, tmplData
}

type Alert struct {
	Style   string
	Message string
}

const (
	AlertDanger  = "danger"
	AlertSuccess = "success"
	AlertInfo    = "info"
	AlertWarning = "warning"
)

func ErrorAlert(args ...interface{}) *Alert {
	return &Alert{
		Style:   AlertDanger,
		Message: fmt.Sprint(args...),
	}
}

func WarningAlert(args ...interface{}) *Alert {
	return &Alert{
		Style:   AlertWarning,
		Message: fmt.Sprint(args...),
	}
}

func SucessAlert(args ...interface{}) *Alert {
	return &Alert{
		Style:   AlertSuccess,
		Message: fmt.Sprint(args...),
	}
}

func ContextGuild(ctx context.Context) *dstate.GuildSet {
	return ctx.Value(common.ContextKeyCurrentGuild).(*dstate.GuildSet)
}

func ContextIsAdmin(ctx context.Context) bool {
	i := ctx.Value(common.ContextKeyIsAdmin)
	if i == nil {
		return false
	}

	return i.(bool)
}

// Returns base context data for control panel plugins
func GetBaseCPContextData(ctx context.Context) (*dstate.GuildSet, TemplateData) {
	var guild *dstate.GuildSet
	if v := ctx.Value(common.ContextKeyCurrentGuild); v != nil {
		guild = v.(*dstate.GuildSet)
	}

	templateData := ctx.Value(common.ContextKeyTemplateData).(TemplateData)

	return guild, templateData
}

// Checks and error and logs it aswell as adding it to the alerts
// returns true if an error occured
func CheckErr(t TemplateData, err error, errMsg string, logger func(...interface{})) bool {
	if err == nil {
		return false
	}

	if errMsg == "" {
		errMsg = err.Error()
	}

	t.AddAlerts(ErrorAlert("An error occurred: ", errMsg))

	if logger != nil {
		logger("An error occured:", err)
	}

	return true
}

// Checks the context if there is a logged in user and if so if he's and admin or not
func IsAdminRequest(ctx context.Context, r *http.Request) (read bool, write bool, gOwner bool) {

	isReadOnlyReq := strings.EqualFold(r.Method, "GET") || strings.EqualFold(r.Method, "OPTIONS")

	if v := ctx.Value(common.ContextKeyCurrentGuild); v != nil {
		// accessing a server page
		g := v.(*dstate.GuildSet)

		gWithConnected := &common.GuildWithConnected{
			UserGuild: &discordgo.UserGuild{
				ID: g.ID,
			},
			Connected: true,
		}

		coreConf := common.ContextCoreConf(ctx)
		member := ContextMember(ctx)

		userID := int64(0)
		var roles []int64

		if member != nil {
			userID = member.User.ID
			roles = member.Roles

			gWithConnected.Permissions = ContextMemberPerms(ctx)
			gWithConnected.Owner = userID == g.OwnerID
		}

		hasRead, hasWrite := GetUserAccessLevel(userID, gWithConnected, coreConf, StaticRoleProvider(roles))

		if gWithConnected.Owner {
			return true, true, true
		}

		if hasWrite {
			return true, true, false
		}

		if hasRead && isReadOnlyReq {
			return true, false, false
		}
	}

	if user := ctx.Value(common.ContextKeyUser); user != nil {
		// there is a active session, but they're not on the related guild (if any)

		cast := user.(*discordgo.User)
		if common.IsOwner(cast.ID) {
			return true, true, true
		}

		if isReadOnlyReq {
			// allow special read only acces for GET and OPTIONS requests, simple and works well
			if hasAcces, err := bot.HasReadOnlyAccess(cast.ID); hasAcces && err == nil {
				return true, false, false
			}
		}
	}

	return false, false, false
}

func NewLogEntryFromContext(ctx context.Context, action string, params ...*cplogs.Param) *cplogs.LogEntry {
	user, ok := ctx.Value(common.ContextKeyUser).(*discordgo.User)
	if !ok {
		return nil
	}

	g := ctx.Value(common.ContextKeyCurrentGuild).(*dstate.GuildSet)

	return cplogs.NewEntry(g.ID, user.ID, user.Username, action, params...)
}

func StaticRoleProvider(roles []int64) func(guildID, userID int64) []int64 {
	return func(guildID, userID int64) []int64 {
		return roles
	}
}

func HasPermissionCTX(ctx context.Context, aperms int64) bool {
	perms := ContextMemberPerms(ctx)
	// Require manageserver, ownership of guild or ownership of bot
	if perms&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator ||
		perms&discordgo.PermissionManageServer == discordgo.PermissionManageServer || perms&aperms == aperms {
		return true
	}

	return false
}

type APIError struct {
	Message string
}

// CtxLogger Returns an always non nil entry either from the context or standard logger
func CtxLogger(ctx context.Context) *logrus.Entry {
	if inter := ctx.Value(common.ContextKeyLogger); inter != nil {
		return inter.(*logrus.Entry)
	}

	return logger
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, err string, statusCode int) {
	if r.FormValue("partial") != "" {
		w.WriteHeader(statusCode)
		w.Write([]byte(`{"error": "` + err + `"}`))
		return
	}

	http.Redirect(w, r, "/?error="+url.QueryEscape(err), http.StatusTemporaryRedirect)
}

func IsRequestPartial(ctx context.Context) bool {
	if v := ctx.Value(common.ContextKeyIsPartial); v != nil {
		return v.(bool)
	}

	return false
}

func ContextUser(ctx context.Context) *discordgo.User {
	return ctx.Value(common.ContextKeyUser).(*discordgo.User)
}

func ContextMember(ctx context.Context) *discordgo.Member {
	i := ctx.Value(common.ContextKeyUserMember)
	if i == nil {
		return nil
	}

	return i.(*discordgo.Member)
}

func ContextMemberPerms(ctx context.Context) int64 {
	i := ctx.Value(common.ContextKeyMemberPermissions)
	if i == nil {
		return 0
	}

	return i.(int64)
}

func ParamOrEmpty(r *http.Request, key string) string {
	s := r.Context().Value(pattern.Variable(key))
	if s != nil {
		return s.(string)
	}

	return ""
}

func Indicator(enabled bool) string {
	const IndEnabled = `<span class="indicator indicator-success"></span>`
	const IndDisabled = `<span class="indicator indicator-danger"></span>`

	if enabled {
		return IndEnabled
	}

	return IndDisabled
}

func EnabledDisabledSpanStatus(enabled bool) (str string) {
	indicator := Indicator(enabled)

	enabledStr := "disabled"
	enabledClass := "danger"
	if enabled {
		enabledStr = "enabled"
		enabledClass = "success"
	}

	return fmt.Sprintf("<span class=\"text-%s\">%s</span>%s", enabledClass, enabledStr, indicator)
}

func GetRequestIP(r *http.Request) string {
	headerField := confReverseProxyClientIPHeader.GetString()
	if headerField == "" {
		li := strings.LastIndex(r.RemoteAddr, ":")
		if li < 0 {
			return r.RemoteAddr
		}

		return r.RemoteAddr[:li]
	}

	return r.Header.Get(headerField)
}

func GetIsReadOnly(ctx context.Context) bool {
	readOnly := ctx.Value(common.ContextKeyIsReadOnly)
	if readOnly == nil {
		return false
	}

	return readOnly.(bool)
}
