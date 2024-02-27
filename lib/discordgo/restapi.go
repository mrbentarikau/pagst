// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains functions for interacting with the Discord REST/JSON API
// at the lowest level.

package discordgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // For JPEG decoding
	_ "image/png"  // For PNG decoding
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// All error constants
var (
	ErrJSONUnmarshal           = errors.New("json unmarshal")
	ErrStatusOffline           = errors.New("You can't set your Status to offline")
	ErrVerificationLevelBounds = errors.New("VerificationLevel out of bounds, should be between 0 and 3")
	ErrPruneDaysBounds         = errors.New("the number of days should be more than or equal to 1")
	ErrGuildNoIcon             = errors.New("guild does not have an icon set")
	ErrGuildNoSplash           = errors.New("guild does not have a splash set")
	ErrUnauthorized            = errors.New("HTTP request was unauthorized. This could be because the provided token was not a bot token. Please add \"Bot \" to the start of your token. https://discordapp.com/developers/docs/reference#authentication-example-bot-token-authorization-header")
	ErrTokenInvalid            = errors.New("Invalid token provided, it has been marked as invalid")
)

var (
	// Marshal defines function used to encode JSON payloads
	Marshal func(v interface{}) ([]byte, error) = json.Marshal
	// Unmarshal defines function used to decode JSON payloads
	Unmarshal func(src []byte, v interface{}) error = json.Unmarshal
)

// RESTError stores error information about a request with a bad response code.
// Message is not always present, there are cases where api calls can fail
// without returning a json message.
type RESTError struct {
	Request      *http.Request
	Response     *http.Response
	ResponseBody []byte
	Message      *APIErrorMessage // Message may be nil.
}

// newRestError returns a new REST API error.
func newRestError(req *http.Request, resp *http.Response, body []byte) *RESTError {
	restErr := &RESTError{
		Request:      req,
		Response:     resp,
		ResponseBody: body,
	}
	// Attempt to decode the error and assume no message was provided if it fails
	var msg *APIErrorMessage
	err := Unmarshal(body, &msg)
	if err == nil {
		restErr.Message = msg
	}
	return restErr
}

// Error returns a Rest API Error with its status code and body.
func (r RESTError) Error() string {
	return "HTTP " + r.Response.Status + ", " + string(r.ResponseBody)
}

// RateLimitError is returned when a request exceeds a rate limit
// and ShouldRetryOnRateLimit is false. The request may be manually
// retried after waiting the duration specified by RetryAfter.
type RateLimitError struct {
	*RateLimit
}

// Error returns a rate limit error with rate limited endpoint and retry time.
func (e RateLimitError) Error() string {
	return "Rate limit exceeded on " + e.URL + ", retry after " + e.RetryAfter.String()
}

// RequestConfig is an HTTP request configuration.
type RequestConfig struct {
	Request                *http.Request
	ShouldRetryOnRateLimit bool
	MaxRestRetries         int
	Client                 *http.Client
}

// newRequestConfig returns a new HTTP request configuration based on parameters in Session.
func newRequestConfig(s *Session, req *http.Request) *RequestConfig {
	return &RequestConfig{
		ShouldRetryOnRateLimit: s.ShouldRetryOnRateLimit,
		MaxRestRetries:         s.MaxRestRetries,
		Client:                 s.Client,
		Request:                req,
	}
}

// RequestOption is a function which mutates request configuration.
// It can be supplied as an argument to any REST method.
type RequestOption func(cfg *RequestConfig)

// WithClient changes the HTTP client used for the request.
func WithClient(client *http.Client) RequestOption {
	return func(cfg *RequestConfig) {
		if client != nil {
			cfg.Client = client
		}
	}
}

// WithRetryOnRatelimit controls whether session will retry the request on rate limit.
func WithRetryOnRatelimit(retry bool) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.ShouldRetryOnRateLimit = retry
	}
}

// WithRestRetries changes maximum amount of retries if request fails.
func WithRestRetries(max int) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.MaxRestRetries = max
	}
}

// WithHeader sets a header in the request.
func WithHeader(key, value string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Request.Header.Set(key, value)
	}
}

// WithAuditLogReason changes audit log reason associated with the request.
func WithAuditLogReason(reason string) RequestOption {
	return WithHeader("X-Audit-Log-Reason", reason)
}

// WithLocale changes accepted locale of the request.
func WithLocale(locale Locale) RequestOption {
	return WithHeader("X-Discord-Locale", string(locale))
}

// WithContext changes context of the request.
func WithContext(ctx context.Context) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Request = cfg.Request.WithContext(ctx)
	}
}

// Request is the same as RequestWithBucketID but the bucket id is the same as the urlStr
func (s *Session) Request(method, urlStr string, data interface{}, extraHeaders map[string]string, options ...RequestOption) (response []byte, err error) {
	return s.RequestWithBucketID(method, urlStr, data, extraHeaders, strings.SplitN(urlStr, "?", 2)[0], options...)
}

// RequestWithBucketID makes a (GET/POST/...) Requests to Discord REST API with JSON data.
func (s *Session) RequestWithBucketID(method, urlStr string, data interface{}, extraHeaders map[string]string, bucketID string, options ...RequestOption) (response []byte, err error) {
	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return
		}
	}

	return s.request(method, urlStr, "application/json", body, extraHeaders, bucketID, options...)
}

// request makes a (GET/POST/...) Requests to Discord REST API.
// Sequence is the sequence number, if it fails with a 502 it will
// retry with sequence+1 until it either succeeds or sequence >= session.MaxRestRetries
func (s *Session) request(method, urlStr, contentType string, b []byte, extraHeaders map[string]string, bucketID string, options ...RequestOption) (response []byte, err error) {
	if bucketID == "" {
		bucketID = strings.SplitN(urlStr, "?", 2)[0]
	}

	return s.RequestWithBucket(method, urlStr, contentType, b, extraHeaders, s.Ratelimiter.GetBucket(bucketID), options...)
}

type ReaderWithMockClose struct {
	*bytes.Reader
}

func (rwmc *ReaderWithMockClose) Close() error {
	return nil
}

// RequestWithLockedBucket makes a request using a bucket that's already been locked
func (s *Session) RequestWithBucket(method, urlStr, contentType string, b []byte, extraHeaders map[string]string, bucket *Bucket, options ...RequestOption) (response []byte, err error) {

	for i := 0; i < s.MaxRestRetries; i++ {
		var retry bool
		var ratelimited bool
		response, retry, ratelimited, err = s.doRequest(method, urlStr, contentType, b, extraHeaders, bucket, options...)
		if !retry {
			break
		}

		if err != nil {
			s.log(LogError, "Request error, retrying: %v", err)
		}

		if ratelimited {
			i = 0
		} else {
			time.Sleep(time.Second * time.Duration(i))
		}

	}

	return
}

type CtxKey int

const (
	CtxKeyRatelimitBucket CtxKey = iota
)

// doRequest makes a request using a bucket
func (s *Session) doRequest(method, urlStr, contentType string, b []byte, extraHeaders map[string]string, bucket *Bucket, options ...RequestOption) (response []byte, retry bool, ratelimitRetry bool, err error) {

	if atomic.LoadInt32(s.tokenInvalid) != 0 {
		return nil, false, false, ErrTokenInvalid
	}

	req, resp, err := s.innerDoRequest(method, urlStr, contentType, b, extraHeaders, bucket, options...)
	if err != nil {
		return nil, true, false, err
	}

	defer func() {
		err2 := resp.Body.Close()
		if err2 != nil {
			log.Println("error closing resp body")
		}
	}()

	response, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, false, err
	}

	if s.Debug {

		log.Printf("API RESPONSE  STATUS :: %s\n", resp.Status)
		for k, v := range resp.Header {
			log.Printf("API RESPONSE  HEADER :: [%s] = %+v\n", k, v)
		}
		log.Printf("API RESPONSE    BODY :: [%s]\n\n\n", response)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}

	switch resp.StatusCode {
	case http.StatusBadGateway, http.StatusGatewayTimeout:
		// Retry sending request if possible
		err = errors.Errorf("%s Failed (%s)", urlStr, resp.Status)
		s.log(LogWarning, err.Error())
		return nil, true, false, err

	case 429: // TOO MANY REQUESTS - Rate limiting
		rl := TooManyRequests{}
		err = json.Unmarshal(response, &rl)
		if err != nil {
			s.log(LogError, "rate limit unmarshal error, %s, %q, url: %s", err, string(response), urlStr)
			return
		}

		rl.Bucket = bucket.Key

		s.log(LogInformational, "Rate Limiting %s, retry in %s", urlStr, rl.RetryAfterDur())
		s.handleEvent(rateLimitEventType, &RateLimit{TooManyRequests: &rl, URL: urlStr})

		time.Sleep(rl.RetryAfterDur())
		// we can make the above smarter
		// this method can cause longer delays than required
		return nil, true, true, nil

	// case http.StatusUnauthorized:
	// 	if strings.Index(s.Token, "Bot ") != 0 {
	// 		s.log(LogInformational, ErrUnauthorized.Error())
	// 		err = ErrUnauthorized
	// 	} else {
	// 		atomic.StoreInt32(s.tokenInvalid, 1)
	// 		err = ErrTokenInvalid
	// 	}
	default: // Error condition
		if resp.StatusCode >= 500 || resp.StatusCode < 400 {
			// non 400 response code
			retry = true
		}

		err = newRestError(req, resp, response)
	}

	return
}

func (s *Session) innerDoRequest(method, urlStr, contentType string, b []byte, extraHeaders map[string]string, bucket *Bucket, options ...RequestOption) (*http.Request, *http.Response, error) {
	bucketLockID := s.Ratelimiter.LockBucketObject(bucket)
	defer func() {
		err := bucket.Release(nil, bucketLockID)
		if err != nil {
			s.log(LogError, "failed unlocking ratelimit bucket: %v", err)
		}
	}()

	if s.Debug {
		log.Printf("API REQUEST %8s :: %s\n", method, urlStr)
		log.Printf("API REQUEST  PAYLOAD :: [%s]\n", string(b))
	}

	req, err := http.NewRequest(method, urlStr, bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}

	req.GetBody = func() (io.ReadCloser, error) {
		return &ReaderWithMockClose{bytes.NewReader(b)}, nil
	}

	// we may need to send a request with extra headers
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Not used on initial login..
	// TODO: Verify if a login, otherwise complain about no-token
	if s.Token != "" {
		req.Header.Set("authorization", s.Token)
	}

	// Discord's API returns a 400 Bad Request is Content-Type is set, but the
	// request body is empty.
	if b != nil {
		req.Header.Set("Content-Type", contentType)
	}

	// TODO: Make a configurable static variable.
	req.Header.Set("User-Agent", fmt.Sprintf("DiscordBot (https://github.com/mrbentarikau/pagst, v%s)", VERSION))

	cfg := newRequestConfig(s, req)
	for _, opt := range options {
		if opt != nil {
			opt(cfg)
		}
	}
	req = cfg.Request

	for header, value := range extraHeaders {
		req.Header.Set(header, value)
	}

	// for things such as stats collecting in the roundtripper for example
	ctx := context.WithValue(req.Context(), CtxKeyRatelimitBucket, bucket)
	req = req.WithContext(ctx)

	if s.Debug {
		for k, v := range req.Header {
			log.Printf("API REQUEST   HEADER :: [%s] = %+v\n", k, v)
		}
	}

	resp, err := s.Client.Do(req)
	if err == nil {
		err = bucket.Release(resp.Header, bucketLockID)
	}

	return req, resp, err
}

func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}

// RequestWithoutBucket make a request that doesn't bound to rate limit
func (s *Session) RequestWithoutBucket(method, urlStr, contentType string, b []byte, sequence int, options ...RequestOption) (response []byte, err error) {
	if s.Debug {
		log.Printf("API REQUEST %8s :: %s\n", method, urlStr)
		log.Printf("API REQUEST  PAYLOAD :: [%s]\n", string(b))
	}

	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(b))
	if err != nil {
		return
	}

	// Not used on initial login..
	// TODO: Verify if a login, otherwise complain about no-token
	if s.Token != "" {
		req.Header.Set("authorization", s.Token)
	}

	// Discord's API returns a 400 Bad Request is Content-Type is set, but the
	// request body is empty.
	if b != nil {
		req.Header.Set("Content-Type", contentType)
	}

	// TODO: Make a configurable static variable.
	req.Header.Set("User-Agent", s.UserAgent)

	cfg := newRequestConfig(s, req)
	for _, opt := range options {
		opt(cfg)
	}

	if s.Debug {
		for k, v := range req.Header {
			log.Printf("API REQUEST   HEADER :: [%s] = %+v\n", k, v)
		}
	}

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err2 := resp.Body.Close()
		if s.Debug && err2 != nil {
			log.Println("error closing resp body")
		}
	}()

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if s.Debug {
		log.Printf("API RESPONSE  STATUS :: %s\n", resp.Status)
		for k, v := range resp.Header {
			log.Printf("API RESPONSE  HEADER :: [%s] = %+v\n", k, v)
		}
		log.Printf("API RESPONSE    BODY :: [%s]\n\n\n", response)
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusBadGateway:
		// Retry sending request if possible
		if sequence < cfg.MaxRestRetries {

			s.log(LogInformational, "%s Failed (%s), Retrying...", urlStr, resp.Status)
			response, err = s.RequestWithoutBucket(method, urlStr, contentType, b, sequence+1, options...)
		} else {
			err = fmt.Errorf("exceeded Max retries HTTP %s, %s", resp.Status, response)
		}
	case 429: // TOO MANY REQUESTS - Rate limiting
		rl := TooManyRequests{}
		err = Unmarshal(response, &rl)
		if err != nil {
			s.log(LogError, "rate limit unmarshal error, %s", err)
			return
		}

		if cfg.ShouldRetryOnRateLimit {
			s.log(LogInformational, "Rate Limiting %s, retry in %v", urlStr, rl.RetryAfter)
			s.handleEvent(rateLimitEventType, &RateLimit{TooManyRequests: &rl, URL: urlStr})

			time.Sleep(rl.RetryAfter)
			// we can make the above smarter
			// this method can cause longer delays than required

			response, err = s.RequestWithoutBucket(method, urlStr, contentType, b, sequence, options...)
		} else {
			err = &RateLimitError{&RateLimit{TooManyRequests: &rl, URL: urlStr}}
		}
	case http.StatusUnauthorized:
		if strings.Index(s.Token, "Bot ") != 0 {
			s.log(LogInformational, ErrUnauthorized.Error())
			err = ErrUnauthorized
		}
		fallthrough
	default: // Error condition
		err = newRestError(req, resp, response)
	}

	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Users
// ------------------------------------------------------------------------------------------------

// User returns the user details of the given userID
// userID    : A user ID
func (s *Session) User(userID int64, options ...RequestOption) (st *User, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointUser(StrID(userID)), nil, nil, EndpointUsers, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserMe returns the user details of the current user
func (s *Session) UserMe(options ...RequestOption) (st *User, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointUser("@me"), nil, nil, EndpointUsers, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserAvatar is deprecated. Please use UserAvatarDecode
// userID    : A user ID or "@me" which is a shortcut of current user ID
func (s *Session) UserAvatar(userID int64, options ...RequestOption) (img image.Image, err error) {
	u, err := s.User(userID)
	if err != nil {
		return
	}
	img, err = s.UserAvatarDecode(u, options...)
	return
}

// UserAvatarDecode returns an image.Image of a user's Avatar
// user : The user which avatar should be retrieved
func (s *Session) UserAvatarDecode(u *User, options ...RequestOption) (img image.Image, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointUserAvatar(u.ID, u.Avatar), nil, nil, EndpointUserAvatar(0, ""), options...)
	if err != nil {
		return
	}

	img, _, err = image.Decode(bytes.NewReader(body))
	return
}

// UserUpdate updates a users settings.
func (s *Session) UserUpdate(email, password, username, avatar, newPassword string, options ...RequestOption) (st *User, err error) {

	// NOTE: Avatar must be either the hash/id of existing Avatar or
	// data:image/png;base64,BASE64_STRING_OF_NEW_AVATAR_PNG
	// to set a new avatar.
	// If left blank, avatar will be set to null/blank

	data := struct {
		Email       string `json:"email,omitempty"`
		Password    string `json:"password,omitempty"`
		Username    string `json:"username,omitempty"`
		Avatar      string `json:"avatar,omitempty"`
		NewPassword string `json:"new_password,omitempty"`
	}{email, password, username, avatar, newPassword}

	body, err := s.RequestWithBucketID("PATCH", EndpointUser("@me"), data, nil, EndpointUsers, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserConnections returns the user's connections
func (s *Session) UserConnections(options ...RequestOption) (conn []*UserConnection, err error) {
	response, err := s.RequestWithBucketID("GET", EndpointUserConnections("@me"), nil, nil, EndpointUserConnections("@me"), options...)
	if err != nil {
		return nil, err
	}

	err = unmarshal(response, &conn)
	if err != nil {
		return
	}

	return
}

// UserChannels returns an array of Channel structures for all private
// channels.
func (s *Session) UserChannels(options ...RequestOption) (st []*Channel, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointUserChannels("@me"), nil, nil, EndpointUserChannels(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserChannelCreate creates a new User (Private) Channel with another User
// recipientID : A user ID for the user to which this channel is opened with.
func (s *Session) UserChannelCreate(recipientID int64, options ...RequestOption) (st *Channel, err error) {

	data := struct {
		RecipientID int64 `json:"recipient_id,string"`
	}{recipientID}

	body, err := s.RequestWithBucketID("POST", EndpointUserChannels("@me"), data, nil, EndpointUserChannels(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserGuildMember returns a guild member object for the current user in the given guildID
// guildID : ID of the guild
func (s *Session) UserGuildMember(guildID int64, options ...RequestOption) (st *Member, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointUserGuildMember("@me", guildID), nil, nil, EndpointUserGuildMember("@me", guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserGuilds returns an array of UserGuild structures for all guilds.
// limit       : The number guilds that can be returned. (max 200)
// beforeID    : If provided all guilds returned will be before given ID.
// afterID     : If provided all guilds returned will be after given ID.
// withCounts  : Whether to include approximate member and presence counts or not
func (s *Session) UserGuilds(limit int, beforeID, afterID int64, withCounts bool, options ...RequestOption) (st []*UserGuild, err error) {

	v := url.Values{}

	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if afterID != 0 {
		v.Set("after", StrID(afterID))
	}
	if beforeID != 0 {
		v.Set("before", StrID(beforeID))
	}
	if withCounts {
		v.Set("with_counts", "true")
	}

	uri := EndpointUserGuilds("@me")

	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointUserGuilds(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserChannelPermissions returns the permission of a user in a channel.
// userID    : The ID of the user to calculate permissions for.
// channelID : The ID of the channel to calculate permission for.
//
// NOTE: This function is now deprecated and will be removed in the future.
// Please see the same function inside state.go
func (s *Session) UserChannelPermissions(userID, channelID int64, fetchOptions ...RequestOption) (apermissions int64, err error) {
	// Try to just get permissions from state.
	apermissions, err = s.State.UserChannelPermissions(userID, channelID)
	if err == nil {
		return
	}

	// Otherwise try get as much data from state as possible, falling back to the network.
	channel, err := s.State.Channel(channelID)
	if err != nil || channel == nil {
		channel, err = s.Channel(channelID, fetchOptions...)
		if err != nil {
			return
		}
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil || guild == nil {
		guild, err = s.Guild(channel.GuildID, fetchOptions...)
		if err != nil {
			return
		}
	}

	if userID == guild.OwnerID {
		apermissions = PermissionAll
		return
	}

	member, err := s.State.Member(guild.ID, userID)
	if err != nil || member == nil {
		member, err = s.GuildMember(guild.ID, userID, fetchOptions...)
		if err != nil {
			return
		}
	}

	// return MemberPermissions(guild, channel, member), nil
	return memberPermissions(guild, channel, userID, member.Roles), nil
}

// Calculates the permissions for a member.
// https://support.discordapp.com/hc/en-us/articles/206141927-How-is-the-permission-hierarchy-structured-
func MemberPermissions(guild *Guild, channel *Channel, member *Member, options ...RequestOption) (apermissions int64) {
	userID := member.User.ID

	if userID == guild.OwnerID {
		apermissions = PermissionAll
		return
	}

	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			apermissions |= role.Permissions
			break
		}
	}

	for _, role := range guild.Roles {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				apermissions |= role.Permissions
				break
			}
		}
	}

	if apermissions&PermissionAdministrator == PermissionAdministrator {
		apermissions |= PermissionAll
		// Administrator overwrites everything, so no point in checking further
		return
	}

	if channel != nil {
		// Apply @everyone overrides from the channel.
		for _, overwrite := range channel.PermissionOverwrites {
			if guild.ID == overwrite.ID {
				apermissions &= ^overwrite.Deny
				apermissions |= overwrite.Allow
				break
			}
		}

		denies := int64(0)
		allows := int64(0)

		// Member overwrites can override role overrides, so do two passes
		for _, overwrite := range channel.PermissionOverwrites {
			for _, roleID := range member.Roles {
				if overwrite.Type == PermissionOverwriteTypeRole && roleID == overwrite.ID {
					denies |= overwrite.Deny
					allows |= overwrite.Allow
					break
				}
			}
		}

		apermissions &= ^denies
		apermissions |= allows

		for _, overwrite := range channel.PermissionOverwrites {
			if overwrite.Type == PermissionOverwriteTypeMember && overwrite.ID == userID {
				apermissions &= ^overwrite.Deny
				apermissions |= overwrite.Allow
				break
			}
		}
	}

	return apermissions
}

// from dgo
func memberPermissions(guild *Guild, channel *Channel, userID int64, roles IDSlice) (apermissions int64) {
	if userID == guild.OwnerID {
		apermissions = PermissionAll
		return
	}

	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			apermissions |= role.Permissions
			break
		}
	}

	for _, role := range guild.Roles {
		for _, roleID := range roles {
			if role.ID == roleID {
				apermissions |= role.Permissions
				break
			}
		}
	}

	if apermissions&PermissionAdministrator == PermissionAdministrator {
		apermissions |= PermissionAll
		return
	}

	// Apply @everyone overrides from the channel.
	for _, overwrite := range channel.PermissionOverwrites {
		if guild.ID == overwrite.ID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	var denies, allows int64
	// Member overwrites can override role overrides, so do two passes
	for _, overwrite := range channel.PermissionOverwrites {
		for _, roleID := range roles {
			if overwrite.Type == PermissionOverwriteTypeRole && roleID == overwrite.ID {
				denies |= overwrite.Deny
				allows |= overwrite.Allow
				break
			}
		}
	}

	apermissions &= ^denies
	apermissions |= allows

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == PermissionOverwriteTypeMember && overwrite.ID == userID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	if apermissions&PermissionAdministrator == PermissionAdministrator {
		apermissions |= PermissionAllChannel
	}

	return apermissions
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Guilds
// ------------------------------------------------------------------------------------------------

// Guild returns a Guild structure of a specific Guild.
// guildID   : The ID of a Guild
func (s *Session) Guild(guildID int64, options ...RequestOption) (st *Guild, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuild(guildID), nil, nil, EndpointGuild(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// Guild returns a Guild structure of a specific Guild.
// guildID   : The ID of a Guild
func (s *Session) GuildWithCounts(guildID int64, options ...RequestOption) (st *Guild, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuild(guildID)+"?with_counts=true", nil, nil, EndpointGuild(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildPreview returns a GuildPreview structure of a specific public Guild.
// guildID   : The ID of a Guild
func (s *Session) GuildPreview(guildID int64, options ...RequestOption) (st *GuildPreview, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuildPreview(guildID), nil, nil, EndpointGuildPreview(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildCreate creates a new Guild
// name      : A name for the Guild (2-100 characters)
func (s *Session) GuildCreate(name string, options ...RequestOption) (st *Guild, err error) {

	data := struct {
		Name string `json:"name"`
	}{name}

	body, err := s.RequestWithBucketID("POST", EndpointGuildCreate, data, nil, EndpointGuildCreate, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildEdit edits a new Guild
// guildID   : The ID of a Guild
// g 		 : A GuildParams struct with the values Name, Region and VerificationLevel defined.
func (s *Session) GuildEdit(guildID int64, g *GuildParams, options ...RequestOption) (st *Guild, err error) {

	// Bounds checking for VerificationLevel, interval: [0, 3]
	if g.VerificationLevel != nil {
		val := *g.VerificationLevel
		if val < 0 || val > 3 {
			err = ErrVerificationLevelBounds
			return
		}
	}

	//Bounds checking for regions
	if g.Region != "" {
		isValid := false
		regions, _ := s.VoiceRegions(options...)
		for _, r := range regions {
			if g.Region == r.ID {
				isValid = true
			}
		}
		if !isValid {
			var valid []string
			for _, r := range regions {
				valid = append(valid, r.ID)
			}
			err = fmt.Errorf("region not a valid region (%q)", valid)
			return
		}
	}

	body, err := s.RequestWithBucketID("PATCH", EndpointGuild(guildID), g, nil, EndpointGuild(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildDelete deletes a Guild.
// guildID   : The ID of a Guild
func (s *Session) GuildDelete(guildID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointGuild(guildID), nil, nil, EndpointGuild(guildID), options...)
	return
}

// GuildLeave leaves a Guild.
// guildID   : The ID of a Guild
func (s *Session) GuildLeave(guildID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointUserGuild("@me", guildID), nil, nil, EndpointUserGuild("", guildID), options...)
	return
}

// GuildBans returns an array of User structures for all bans of a
// given guild.
// guildID   : The ID of a Guild.
func (s *Session) GuildBans(guildID int64, options ...RequestOption) (st []*GuildBan, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuildBans(guildID), nil, nil, EndpointGuildBans(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildBan returns a ban object for the given user or a 404 not found if the ban cannot be found. Requires the BAN_MEMBERS permission.
// guildID   : The ID of a Guild.
func (s *Session) GuildBan(guildID, userID int64, options ...RequestOption) (st *GuildBan, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuildBan(guildID, userID), nil, nil, EndpointGuildBan(guildID, 0)+"/", options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildBanCreate bans the given user from the given guild.
// guildID   : The ID of a Guild.
// userID    : The ID of a User
// days      : The number of days of previous comments to delete.
func (s *Session) GuildBanCreate(guildID, userID int64, days int, options ...RequestOption) (err error) {
	return s.GuildBanCreateWithReason(guildID, userID, "", days, options...)
}

// GuildBanCreateWithReason bans the given user from the given guild also providing a reason.
// guildID   : The ID of a Guild.
// userID    : The ID of a User
// reason    : The reason for this ban
// days      : The number of days of previous comments to delete.
func (s *Session) GuildBanCreateWithReason(guildID, userID int64, reason string, days int, options ...RequestOption) (err error) {

	uri := EndpointGuildBan(guildID, userID)

	data := make(map[string]interface{})
	if days > 0 {
		data["delete_message_days"] = days
	}

	var extraHeaders map[string]string
	if reason != "" {
		extraHeaders = map[string]string{"X-Audit-Log-Reason": url.PathEscape(reason)}
	}

	_, err = s.RequestWithBucketID("PUT", uri, data, extraHeaders, EndpointGuildBan(guildID, 0), options...)
	return
}

// GuildBanCreateWithReasonDeleteBySeconds bans the given user from the given guild also providing a reason, with deleting messages by seconds.
// guildID   : The ID of a Guild.
// userID    : The ID of a User
// reason    : The reason for this ban
// seconds      : The number of seconds of previous comments to delete, between 0 and 604800 (7 days).
func (s *Session) GuildBanCreateWithReasonDeleteBySeconds(guildID, userID int64, reason string, seconds int, options ...RequestOption) (err error) {
	uri := EndpointGuildBan(guildID, userID)

	//queryParams := url.Values{}
	data := make(map[string]interface{})
	if seconds > 0 {
		data["delete_message_days"] = seconds
		//queryParams.Set("delete_message_seconds", strconv.Itoa(seconds))
	}
	var extraHeaders map[string]string
	if reason != "" {
		extraHeaders = map[string]string{"X-Audit-Log-Reason": url.PathEscape(reason)}
		//queryParams.Set("reason", reason)
	}

	/*if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}*/

	_, err = s.RequestWithBucketID("PUT", uri, data, extraHeaders, EndpointGuildBan(guildID, 0), options...)
	return
}

// GuildBanDelete removes the given user from the guild bans
// guildID   : The ID of a Guild.
// userID    : The ID of a User
func (s *Session) GuildBanDelete(guildID, userID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildBan(guildID, userID), nil, nil, EndpointGuildBan(guildID, 0), options...)
	return
}

// GuildMembers returns a list of members for a guild.
//
//	guildID  : The ID of a Guild.
//	after    : The id of the member to return members after
//	limit    : max number of members to return (max 1000)
func (s *Session) GuildMembers(guildID int64, after int64, limit int, options ...RequestOption) (st []*Member, err error) {

	uri := EndpointGuildMembers(guildID)

	v := url.Values{}

	if after != 0 {
		v.Set("after", StrID(after))
	}

	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}

	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildMembers(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildMembersSearch returns a list of guild member objects whose username or nickname starts with a provided string
// guildID  : The ID of a Guild
// query    : Query string to match username(s) and nickname(s) against
// limit    : Max number of members to return (default 1, min 1, max 1000)
func (s *Session) GuildMembersSearch(guildID int64, query string, limit int, options ...RequestOption) (st []*Member, err error) {

	uri := EndpointGuildMembersSearch(guildID)

	queryParams := url.Values{}
	queryParams.Set("query", query)
	if limit > 1 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	body, err := s.RequestWithBucketID("GET", uri+"?"+queryParams.Encode(), nil, nil, uri, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildMember returns a member of a guild.
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User
func (s *Session) GuildMember(guildID, userID int64, options ...RequestOption) (st *Member, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildMember(guildID, userID), nil, nil, EndpointGuildMember(guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

func (s *Session) GuildMemberAdd(guildID, userID int64, data *GuildMemberAddParams, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("PUT", EndpointGuildMember(guildID, userID), data, nil, EndpointGuildMember(guildID, 0), options...)
	if err != nil {
		return err
	}

	return err
}

// GuildMemberDelete removes the given user from the given guild.
// guildID   : The ID of a Guild.
// userID    : The ID of a User
func (s *Session) GuildMemberDelete(guildID, userID int64, options ...RequestOption) (err error) {

	return s.GuildMemberDeleteWithReason(guildID, userID, "", options...)
}

// GuildMemberDeleteWithReason removes the given user from the given guild.
// guildID   : The ID of a Guild.
// userID    : The ID of a User
// reason    : The reason for the kick
func (s *Session) GuildMemberDeleteWithReason(guildID, userID int64, reason string, options ...RequestOption) (err error) {
	var extraHeaders map[string]string

	if reason != "" {
		extraHeaders = map[string]string{"X-Audit-Log-Reason": url.PathEscape(reason)}
	}

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildMember(guildID, userID), nil, extraHeaders, EndpointGuildMember(guildID, 0), options...)
	return
}

// GuildMemberEdit edits the roles of a member.
// guildID  : The ID of a Guild.
// userID   : The ID of a User.
// roles    : A list of role ID's to set on the member.
func (s *Session) GuildMemberEdit(guildID, userID int64, data *GuildMemberParams, options ...RequestOption) (st *Member, err error) {
	var body []byte
	body, err = s.RequestWithBucketID("PATCH", EndpointGuildMember(guildID, userID), data, nil, EndpointGuildMember(guildID, 0), options...)
	if err != nil {
		return nil, err
	}

	err = unmarshal(body, &st)
	return
}

// GuildMemberEditComplex edits the nickname and roles of a member.
// NOTE: deprecated, use GuildMemberEdit instead.
// guildID  : The ID of a Guild.
// userID   : The ID of a User.
// data     : A GuildMemberEditData struct with the new nickname and roles
func (s *Session) GuildMemberEditComplex(guildID, userID int64, data *GuildMemberParams, options ...RequestOption) (st *Member, err error) {
	return s.GuildMemberEdit(guildID, userID, data, options...)
}

// GuildMemberMove moves a guild member from one voice channel to another/none
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User.
//	channelID : The ID of a channel to move user to. Use 0 to disconnect the member.
//
// NOTE : I am not entirely set on the name of this function and it may change
// prior to the final 1.0.0 release of Discordgo
func (s *Session) GuildMemberMove(guildID, userID, channelID int64, options ...RequestOption) (err error) {

	data := struct {
		ChannelID NullableID `json:"channel_id,string"`
	}{NullableID(channelID)}

	_, err = s.RequestWithBucketID("PATCH", EndpointGuildMember(guildID, userID), data, nil, EndpointGuildMember(guildID, 0), options...)
	if err != nil {
		return
	}

	return
}

// GuildMemberNickname updates the nickname of a guild member
// guildID   : The ID of a guild
// userID    : The ID of a user or "@me" which is a shortcut of the current user ID
// nickname  : The new nickname
func (s *Session) GuildMemberNickname(guildID, userID int64, nickname string, options ...RequestOption) (err error) {

	data := struct {
		Nick string `json:"nick"`
	}{nickname}

	_, err = s.RequestWithBucketID("PATCH", EndpointGuildMember(guildID, userID), data, nil, EndpointGuildMember(guildID, 0), options...)
	return
}

// GuildMemberNicknameMe updates the nickname the current user
// guildID   : The ID of a guild
// nickname  : The new nickname
func (s *Session) GuildMemberNicknameMe(guildID int64, nickname string, options ...RequestOption) (err error) {

	data := struct {
		Nick string `json:"nick"`
	}{nickname}

	_, err = s.RequestWithBucketID("PATCH", EndpointGuildMemberMe(guildID)+"/nick", data, nil, EndpointGuildMember(guildID, 0), options...)
	return
}

// GuildMemberTimeoutWithReason times out a guild member with a mandatory reason
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User.
//	until     : The timestamp for how long a member should be timed out.
//	            Set to nil to remove timeout.
//
// reason    : The reason for the timeout
func (s *Session) GuildMemberTimeoutWithReason(guildID int64, userID int64, until *time.Time, reason string, options ...RequestOption) (err error) {
	data := struct {
		TimeoutExpiresAt *time.Time `json:"communication_disabled_until"`
	}{until}

	extraHeaders := make(map[string]string)
	if reason != "" {
		extraHeaders["X-Audit-Log-Reason"] = reason
	}
	_, err = s.RequestWithBucketID("PATCH", EndpointGuildMember(guildID, userID), data, extraHeaders, EndpointGuildMember(guildID, 0), options...)
	return
}

// GuildMemberTimeout times out a guild member
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User.
//	until     : The timestamp for how long a member should be timed out.
//	            Set to nil to remove timeout.
func (s *Session) GuildMemberTimeout(guildID int64, userID int64, until *time.Time, reason string, options ...RequestOption) (err error) {
	return s.GuildMemberTimeoutWithReason(guildID, userID, until, reason, options...)
}

// GuildMemberRoleAdd adds the specified role to a given member
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User.
//	roleID 	  : The ID of a Role to be assigned to the user.
func (s *Session) GuildMemberRoleAdd(guildID, userID, roleID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("PUT", EndpointGuildMemberRole(guildID, userID, roleID), nil, nil, EndpointGuildMemberRole(guildID, 0, 0), options...)

	return
}

// GuildMemberRoleRemove removes the specified role to a given member
//
//	guildID   : The ID of a Guild.
//	userID    : The ID of a User.
//	roleID 	  : The ID of a Role to be removed from the user.
func (s *Session) GuildMemberRoleRemove(guildID, userID, roleID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildMemberRole(guildID, userID, roleID), nil, nil, EndpointGuildMemberRole(guildID, 0, 0), options...)

	return
}

// GuildChannels returns an array of Channel structures for all channels of a
// given guild.
// guildID   : The ID of a Guild.
func (s *Session) GuildChannels(guildID int64, options ...RequestOption) (st []*Channel, err error) {

	body, err := s.request("GET", EndpointGuildChannels(guildID), "", nil, nil, EndpointGuildChannels(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildChannelCreate creates a new channel in the given guild
// guildID   : The ID of a Guild.
// name      : Name of the channel (2-100 chars length)
// ctype     : Type of the channel
func (s *Session) GuildChannelCreate(guildID int64, name string, ctype ChannelType, options ...RequestOption) (st *Channel, err error) {

	data := struct {
		Name string      `json:"name"`
		Type ChannelType `json:"type"`
	}{name, ctype}

	body, err := s.RequestWithBucketID("POST", EndpointGuildChannels(guildID), data, nil, EndpointGuildChannels(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildChannelCreateWithOverwrites creates a new channel in the given guild
// guildID     : The ID of a Guild.
// name        : Name of the channel (2-100 chars length)
// ctype       : Type of the channel
// overwrites  : slice of permission overwrites
func (s *Session) GuildChannelCreateWithOverwrites(guildID int64, name string, ctype ChannelType, parentID int64, overwrites []*PermissionOverwrite, options ...RequestOption) (st *Channel, err error) {

	data := struct {
		Name                 string                 `json:"name"`
		Type                 ChannelType            `json:"type"`
		ParentID             int64                  `json:"parent_id,string"`
		PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites"`
	}{name, ctype, parentID, overwrites}

	body, err := s.RequestWithBucketID("POST", EndpointGuildChannels(guildID), data, nil, EndpointGuildChannels(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildChannelsReorder updates the order of channels in a guild
// guildID   : The ID of a Guild.
// channels  : Updated channels.
func (s *Session) GuildChannelsReorder(guildID int64, channels []*Channel, options ...RequestOption) (err error) {

	data := make([]struct {
		ID       int64 `json:"id,string"`
		Position int   `json:"position"`
	}, len(channels))

	for i, c := range channels {
		data[i].ID = c.ID
		data[i].Position = c.Position
	}

	_, err = s.RequestWithBucketID("PATCH", EndpointGuildChannels(guildID), data, nil, EndpointGuildChannels(guildID), options...)
	return
}

// GuildInvites returns an array of Invite structures for the given guild
// guildID   : The ID of a Guild.
func (s *Session) GuildInvites(guildID int64, options ...RequestOption) (st []*Invite, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointGuildInvites(guildID), nil, nil, EndpointGuildInvites(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildRoles returns all roles for a given guild.
// guildID   : The ID of a Guild.
func (s *Session) GuildRoles(guildID int64, options ...RequestOption) (st []*Role, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildRoles(guildID), nil, nil, EndpointGuildRoles(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return // TODO return pointer
}

// GuildRoleCreate creates a new Guild Role and returns it.
// guildID : The ID of a Guild.
// data    : New role parameters.
func (s *Session) GuildRoleCreate(guildID int64, data *RoleParams, options ...RequestOption) (st *Role, err error) {

	body, err := s.RequestWithBucketID("POST", EndpointGuildRoles(guildID), data, nil, EndpointGuildRoles(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildRoleCreateComplex returns a new Guild Role.
// guildID: The ID of a Guild.
func (s *Session) GuildRoleCreateComplex(guildID int64, roleCreate RoleCreate, options ...RequestOption) (st *Role, err error) {

	body, err := s.RequestWithBucketID("POST", EndpointGuildRoles(guildID), roleCreate, nil, EndpointGuildRoles(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildRoleEdit updates an existing Guild Role and updated Role data.
// guildID   : The ID of a Guild.
// roleID    : The ID of a Role.
// data 		 : Updated Role data.
func (s *Session) GuildRoleEdit(guildID, roleID int64, data *RoleParams, options ...RequestOption) (st *Role, err error) {

	// Prevent sending a color int that is too big.
	if data.Color != nil && *data.Color > 0xFFFFFF {
		return nil, fmt.Errorf("colour value cannot be larger than 0xFFFFFF")
	}

	body, err := s.RequestWithBucketID("PATCH", EndpointGuildRole(guildID, roleID), data, nil, EndpointGuildRole(guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildRoleReorder reoders guild roles
// guildID   : The ID of a Guild.
// roles     : A list of ordered roles.
func (s *Session) GuildRoleReorder(guildID int64, roles []*Role, options ...RequestOption) (st []*Role, err error) {

	body, err := s.RequestWithBucketID("PATCH", EndpointGuildRoles(guildID), roles, nil, EndpointGuildRoles(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildRoleDelete deletes an existing role.
// guildID   : The ID of a Guild.
// roleID    : The ID of a Role.
func (s *Session) GuildRoleDelete(guildID, roleID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildRole(guildID, roleID), nil, nil, EndpointGuildRole(guildID, 0), options...)
	return
}

// GuildPruneCount Returns the number of members that would be removed in a prune operation.
// Requires 'KICK_MEMBER' permission.
// guildID	: The ID of a Guild.
// days		: The number of days to count prune for (1 or more).
func (s *Session) GuildPruneCount(guildID int64, days uint32, options ...RequestOption) (count uint32, err error) {
	count = 0

	if days <= 0 {
		err = ErrPruneDaysBounds
		return
	}

	p := struct {
		Pruned uint32 `json:"pruned"`
	}{}

	uri := EndpointGuildPrune(guildID) + fmt.Sprintf("?days=%d", days)
	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildPrune(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &p)
	if err != nil {
		return
	}

	count = p.Pruned

	return
}

// GuildPrune Begin as prune operation. Requires the 'KICK_MEMBERS' permission.
// Returns an object with one 'pruned' key indicating the number of members that were removed in the prune operation.
// guildID	: The ID of a Guild.
// days		: The number of days to count prune for (1 or more).
func (s *Session) GuildPrune(guildID int64, days uint32, options ...RequestOption) (count uint32, err error) {

	count = 0

	if days <= 0 {
		err = ErrPruneDaysBounds
		return
	}

	data := struct {
		days uint32
	}{days}

	p := struct {
		Pruned uint32 `json:"pruned"`
	}{}

	body, err := s.RequestWithBucketID("POST", EndpointGuildPrune(guildID), data, nil, EndpointGuildPrune(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &p)
	if err != nil {
		return
	}

	count = p.Pruned

	return
}

// GuildIntegrations returns an array of Integrations for a guild.
// guildID   : The ID of a Guild.
func (s *Session) GuildIntegrations(guildID int64, options ...RequestOption) (st []*Integration, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildIntegrations(guildID), nil, nil, EndpointGuildIntegrations(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildIntegrationCreate creates a Guild Integration.
// guildID          : The ID of a Guild.
// integrationType  : The Integration type.
// integrationID    : The ID of an integration.
func (s *Session) GuildIntegrationCreate(guildID int64, integrationType string, integrationID int64, options ...RequestOption) (err error) {

	data := struct {
		Type string `json:"type"`
		ID   int64  `json:"id,string"`
	}{integrationType, integrationID}

	_, err = s.RequestWithBucketID("POST", EndpointGuildIntegrations(guildID), data, nil, EndpointGuildIntegrations(guildID), options...)
	return
}

// GuildIntegrationEdit edits a Guild Integration.
// guildID              : The ID of a Guild.
// integrationType      : The Integration type.
// integrationID        : The ID of an integration.
// expireBehavior	      : The behavior when an integration subscription lapses (see the integration object documentation).
// expireGracePeriod    : Period (in seconds) where the integration will ignore lapsed subscriptions.
// enableEmoticons	    : Whether emoticons should be synced for this integration (twitch only currently).
func (s *Session) GuildIntegrationEdit(guildID, integrationID int64, expireBehavior, expireGracePeriod int, enableEmoticons bool, options ...RequestOption) (err error) {

	data := struct {
		ExpireBehavior    int  `json:"expire_behavior"`
		ExpireGracePeriod int  `json:"expire_grace_period"`
		EnableEmoticons   bool `json:"enable_emoticons"`
	}{expireBehavior, expireGracePeriod, enableEmoticons}

	_, err = s.RequestWithBucketID("PATCH", EndpointGuildIntegration(guildID, integrationID), data, nil, EndpointGuildIntegration(guildID, 0), options...)
	return
}

// GuildIntegrationDelete removes the given integration from the Guild.
// guildID          : The ID of a Guild.
// integrationID    : The ID of an integration.
func (s *Session) GuildIntegrationDelete(guildID, integrationID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildIntegration(guildID, integrationID), nil, nil, EndpointGuildIntegration(guildID, 0), options...)
	return
}

// GuildIntegrationSync syncs an integration.
// guildID          : The ID of a Guild.
// integrationID    : The ID of an integration.
func (s *Session) GuildIntegrationSync(guildID, integrationID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("POST", EndpointGuildIntegrationSync(guildID, integrationID), nil, nil, EndpointGuildIntegration(guildID, 0), options...)
	return
}

// GuildIcon returns an image.Image of a guild icon.
// guildID   : The ID of a Guild.
func (s *Session) GuildIcon(guildID int64, options ...RequestOption) (img image.Image, err error) {
	g, err := s.Guild(guildID, options...)
	if err != nil {
		return
	}

	if g.Icon == "" {
		err = ErrGuildNoIcon
		return
	}

	body, err := s.RequestWithBucketID("GET", EndpointGuildIcon(guildID, g.Icon), nil, nil, EndpointGuildIcon(guildID, ""), options...)
	if err != nil {
		return
	}

	img, _, err = image.Decode(bytes.NewReader(body))
	return
}

// GuildSplash returns an image.Image of a guild splash image.
// guildID   : The ID of a Guild.
func (s *Session) GuildSplash(guildID int64, options ...RequestOption) (img image.Image, err error) {
	g, err := s.Guild(guildID, options...)
	if err != nil {
		return
	}

	if g.Splash == "" {
		err = ErrGuildNoSplash
		return
	}

	body, err := s.RequestWithBucketID("GET", EndpointGuildSplash(guildID, g.Splash), nil, nil, EndpointGuildSplash(guildID, ""), options...)
	if err != nil {
		return
	}

	img, _, err = image.Decode(bytes.NewReader(body))
	return
}

// GuildEmbed returns the embed for a Guild.
// guildID   : The ID of a Guild.
func (s *Session) GuildEmbed(guildID int64, options ...RequestOption) (st *GuildEmbed, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildEmbed(guildID), nil, nil, EndpointGuildEmbed(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildEmbedEdit edits the embed of a Guild.
// guildID   : The ID of a Guild.
// data      : New embed data.
func (s *Session) GuildEmbedEdit(guildID int64, data *GuildEmbed, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("PATCH", EndpointGuildEmbed(guildID), data, nil, EndpointGuildEmbed(guildID), options...)
	return
}

// GuildAuditLog returns the audit log for a Guild.
// guildID     : The ID of a Guild.
// userID      : If provided the log will be filtered for the given ID.
// beforeID    : If provided all log entries returned will be before the given ID.
// actionType  : If provided the log will be filtered for the given Action Type.
// limit       : The number messages that can be returned. (default 50, min 1, max 100)
func (s *Session) GuildAuditLog(guildID, userID, beforeID int64, actionType, limit int, options ...RequestOption) (st *GuildAuditLog, err error) {

	uri := EndpointGuildAuditLogs(guildID)

	v := url.Values{}
	if userID != 0 {
		v.Set("user_id", StrID(userID))
	}
	if beforeID != 0 {
		v.Set("before", StrID(beforeID))
	}
	if actionType > 0 {
		v.Set("action_type", strconv.Itoa(int(actionType)))
	}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildAuditLogs(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildEmojiCreate creates a new Emoji.
// guildID : The ID of a Guild.
// data    : New Emoji data.
func (s *Session) GuildEmojiCreate(guildID int64, data *EmojiParams, options ...RequestOption) (emoji *Emoji, err error) {
	body, err := s.RequestWithBucketID("POST", EndpointGuildEmojis(guildID), data, nil, EndpointGuildEmojis(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &emoji)
	return
}

// GuildEmojiEdit modifies and returns updated emoji.
// guildID : The ID of a Guild.
// emojiID : The ID of an Emoji.
// data    : Updated emoji data.
func (s *Session) GuildEmojiEdit(guildID, emojiID int64, data *EmojiParams, options ...RequestOption) (emoji *Emoji, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointGuildEmoji(guildID, emojiID), data, nil, EndpointGuildEmojis(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &emoji)
	return
}

// GuildEmojiDelete deletes an Emoji.
// guildID : The ID of a Guild.
// emojiID : The ID of an Emoji.
func (s *Session) GuildEmojiDelete(guildID, emojiID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildEmoji(guildID, emojiID), nil, nil, EndpointGuildEmojis(guildID), options...)
	return
}

// GuildTemplate returns a GuildTemplate for the given code
// templateCode: The Code of a GuildTemplate
func (s *Session) GuildTemplate(templateCode int64, options ...RequestOption) (st *GuildTemplate, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildTemplate(templateCode), nil, nil, EndpointGuildTemplate(templateCode), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildCreateWithTemplate creates a guild based on a GuildTemplate
// templateCode: The Code of a GuildTemplate
// name: The name of the guild (2-100) characters
// icon: base64 encoded 128x128 image for the guild icon
func (s *Session) GuildCreateWithTemplate(templateCode int64, name, icon string, options ...RequestOption) (st *Guild, err error) {

	data := struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	}{name, icon}

	body, err := s.RequestWithBucketID("POST", EndpointGuildTemplate(templateCode), data, nil, EndpointGuildTemplate(templateCode), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildTemplates returns all of GuildTemplates
// guildID: The ID of the guild
func (s *Session) GuildTemplates(guildID int64, options ...RequestOption) (st []*GuildTemplate, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildTemplates(guildID), nil, nil, EndpointGuildTemplates(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildTemplateCreate creates a template for the guild
// guildID : The ID of the guild
// data    : Template metadata
func (s *Session) GuildTemplateCreate(guildID int64, data *GuildTemplateParams, options ...RequestOption) (st *GuildTemplate) {
	body, err := s.RequestWithBucketID("POST", EndpointGuildTemplates(guildID), data, nil, EndpointGuildTemplates(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildTemplateSync syncs the template to the guild's current state
// guildID: The ID of the guild
// templateCode: The code of the template
func (s *Session) GuildTemplateSync(guildID, templateCode int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("PUT", EndpointGuildTemplateSync(guildID, templateCode), nil, nil, EndpointGuildTemplateSync(guildID, 0), options...)
	return
}

// GuildTemplateEdit modifies the template's metadata
// guildID      : The ID of the guild
// templateCode : The code of the template
// data         : New template metadata
func (s *Session) GuildTemplateEdit(guildID, templateCode int64, data *GuildTemplateParams, options ...RequestOption) (st *GuildTemplate, err error) {

	body, err := s.RequestWithBucketID("PATCH", EndpointGuildTemplateSync(guildID, templateCode), data, nil, EndpointGuildTemplateSync(guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildTemplateDelete deletes the template
// guildID: The ID of the guild
// templateCode: The code of the template
func (s *Session) GuildTemplateDelete(guildID, templateCode int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointGuildTemplateSync(guildID, templateCode), nil, nil, EndpointGuildTemplateSync(guildID, 0), options...)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Channels
// ------------------------------------------------------------------------------------------------

// Channel returns a Channel structure of a specific Channel.
// channelID  : The ID of the Channel you want returned.
func (s *Session) Channel(channelID int64, options ...RequestOption) (st *Channel, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointChannel(channelID), nil, nil, EndpointChannel(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelEdit edits the given channel and returns the updated Channel data.
// channelID  : The ID of a Channel.
// data       : New Channel data.
func (s *Session) ChannelEdit(channelID int64, data *ChannelEdit, options ...RequestOption) (st *Channel, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointChannel(channelID), data, nil, EndpointChannel(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelEditComplex edits an existing channel, replacing the parameters entirely with ChannelEdit struct
// NOTE: deprecated, use ChannelEdit instead
// channelID     : The ID of a Channel
// data          : The channel struct to send
func (s *Session) ChannelEditComplex(channelID int64, data *ChannelEdit, options ...RequestOption) (st *Channel, err error) {
	return s.ChannelEdit(channelID, data, options...)
}

// ChannelDelete deletes the given channel
// channelID  : The ID of a Channel
func (s *Session) ChannelDelete(channelID int64, options ...RequestOption) (st *Channel, err error) {

	body, err := s.RequestWithBucketID("DELETE", EndpointChannel(channelID), nil, nil, EndpointChannel(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelTyping broadcasts to all members that authenticated user is typing in
// the given channel.
// channelID  : The ID of a Channel
func (s *Session) ChannelTyping(channelID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("POST", EndpointChannelTyping(channelID), nil, nil, EndpointChannelTyping(channelID), options...)
	return
}

// ChannelMessages returns an array of Message structures for messages within
// a given channel.
// channelID : The ID of a Channel.
// limit     : The number messages that can be returned. (max 100)
// beforeID  : If provided all messages returned will be before given ID.
// afterID   : If provided all messages returned will be after given ID.
// aroundID  : If provided all messages returned will be around given ID.
func (s *Session) ChannelMessages(channelID int64, limit int, beforeID, afterID, aroundID int64, options ...RequestOption) (st []*Message, err error) {

	uri := EndpointChannelMessages(channelID)

	v := url.Values{}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if afterID != 0 {
		v.Set("after", StrID(afterID))
	}
	if beforeID != 0 {
		v.Set("before", StrID(beforeID))
	}
	if aroundID != 0 {
		v.Set("around", StrID(aroundID))
	}

	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointChannelMessages(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelMessage gets a single message by ID from a given channel.
// channeld  : The ID of a Channel
// messageID : the ID of a Message
func (s *Session) ChannelMessage(channelID, messageID int64, options ...RequestOption) (st *Message, err error) {

	response, err := s.RequestWithBucketID("GET", EndpointChannelMessage(channelID, messageID), nil, nil, EndpointChannelMessage(channelID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(response, &st)
	return
}

// ChannelMessageSend sends a message to the given channel.
// channelID : The ID of a Channel.
// content   : The message to send.
func (s *Session) ChannelMessageSend(channelID int64, content string, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Content:         content,
		AllowedMentions: AllowedMentions{},
	}, options...)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// ChannelMessageSendComplex sends a message to the given channel.
// channelID : The ID of a Channel.
// data      : The message struct to send.
func (s *Session) ChannelMessageSendComplex(channelID int64, msg *MessageSend, options ...RequestOption) (st *Message, err error) {
	msg.Embeds = ValidateComplexMessageEmbeds(msg.Embeds)

	endpoint := EndpointChannelMessages(channelID)

	// TODO: Remove this when compatibility is not required.
	files := msg.Files
	if msg.File != nil {
		if files == nil {
			files = []*File{msg.File}
		} else {
			err = fmt.Errorf("cannot specify both File and Files")
			return
		}
	}

	if msg.StickerIDs != nil {
		if len(msg.StickerIDs) > 3 {
			err = fmt.Errorf("cannot send more than 3 stickers")
			return
		}
	}

	var response []byte
	if len(files) > 0 {
		contentType, body, encodeErr := MultipartBodyWithJSON(msg, files)
		if encodeErr != nil {
			return st, encodeErr
		}

		response, err = s.request("POST", endpoint, contentType, body, nil, endpoint, options...)
	} else {
		response, err = s.RequestWithBucketID("POST", endpoint, msg, nil, endpoint, options...)
	}
	if err != nil {
		return
	}

	err = unmarshal(response, &st)
	return
}

// ChannelMessageSendTTS sends a message to the given channel with Text to Speech.
// channelID : The ID of a Channel.
// content   : The message to send.
func (s *Session) ChannelMessageSendTTS(channelID int64, content string, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Content: content,
		TTS:     true,
	}, options...)
}

// ChannelMessageSendEmbeds sends a message to the given channel with list of embedded data.
// channelID : The ID of a Channel.
// embed     : The list embed data to send.
func (s *Session) ChannelMessageSendEmbedList(channelID int64, embeds []*MessageEmbed, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Embeds: embeds,
	}, options...)
}

// ChannelMessageSendEmbed sends a message to the given channel with embedded data.
// channelID : The ID of a Channel.
// embed     : The embed data to send.
func (s *Session) ChannelMessageSendEmbed(channelID int64, embed *MessageEmbed, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Embeds: []*MessageEmbed{embed},
	}, options...)
}

// ChannelMessageSendEmbeds sends a message to the given channel with multiple embedded data.
// channelID : The ID of a Channel.
// embeds    : The embeds data to send.
func (s *Session) ChannelMessageSendEmbeds(channelID int64, embeds []*MessageEmbed, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Embeds: embeds,
	}, options...)
}

// ChannelMessageSendReply sends a message to the given channel with reference data.
// channelID : The ID of a Channel.
// content   : The message to send.
// reference : The message reference to send.
func (s *Session) ChannelMessageSendReply(channelID int64, content string, reference *MessageReference, options ...RequestOption) (*Message, error) {
	if reference == nil {
		return nil, fmt.Errorf("reply attempted with nil message reference")
	}
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Content:   content,
		Reference: reference,
	}, options...)
}

// ChannelMessageSendEmbedReply sends a message to the given channel with reference data and embedded data.
// channelID : The ID of a Channel.
// embed   : The embed data to send.
// reference : The message reference to send.
func (s *Session) ChannelMessageSendEmbedReply(channelID int64, embed *MessageEmbed, reference *MessageReference, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendEmbedsReply(channelID, []*MessageEmbed{embed}, reference, options...)
}

// ChannelMessageSendEmbedsReply sends a message to the given channel with reference data and multiple embedded data.
// channelID : The ID of a Channel.
// embeds    : The embeds data to send.
// reference : The message reference to send.
func (s *Session) ChannelMessageSendEmbedsReply(channelID int64, embeds []*MessageEmbed, reference *MessageReference, options ...RequestOption) (*Message, error) {
	if reference == nil {
		return nil, fmt.Errorf("reply attempted with nil message reference")
	}
	return s.ChannelMessageSendComplex(channelID, &MessageSend{
		Embeds:    embeds,
		Reference: reference,
	}, options...)
}

// ChannelMessageEdit edits an existing message, replacing it entirely with
// the given content.
// channelID  : The ID of a Channel
// messageID  : The ID of a Message
// content    : The contents of the message
func (s *Session) ChannelMessageEdit(channelID, messageID int64, content string, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageEditComplex(NewMessageEdit(channelID, messageID).SetContent(content), options...)
}

// ChannelMessageEditComplex edits an existing message, replacing it entirely with
// the given MessageEdit struct
func (s *Session) ChannelMessageEditComplex(msg *MessageEdit, options ...RequestOption) (st *Message, err error) {
	msg.Embeds = ValidateComplexMessageEmbeds(msg.Embeds)

	endpoint := EndpointChannelMessage(msg.Channel, msg.ID)

	var response []byte
	if len(msg.Files) > 0 {
		contentType, body, encodeErr := MultipartBodyWithJSON(msg, msg.Files)
		if encodeErr != nil {
			return st, encodeErr
		}
		response, err = s.request("PATCH", endpoint, contentType, body, nil, EndpointChannelMessage(msg.Channel, 0), options...)
	} else {
		response, err = s.RequestWithBucketID("PATCH", endpoint, msg, nil, EndpointChannelMessage(msg.Channel, 0), options...)
	}
	if err != nil {
		return
	}

	err = unmarshal(response, &st)
	return
}

// ChannelMessageEditEmbed edits an existing message with embedded data.
// channelID : The ID of a Channel
// messageID : The ID of a Message
// embed     : The embed data to send
func (s *Session) ChannelMessageEditEmbed(channelID, messageID int64, embed *MessageEmbed, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageEditComplex(NewMessageEdit(channelID, messageID).SetEmbeds([]*MessageEmbed{embed}), options...)
}

// ChannelMessageEditEmbeds edits an existing message with a list of embedded data.
// channelID : The ID of a Channel
// messageID : The ID of a Message
// embeds     : The list of embed data to send
func (s *Session) ChannelMessageEditEmbedList(channelID, messageID int64, embeds []*MessageEmbed, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageEditComplex(NewMessageEdit(channelID, messageID).SetEmbeds(embeds), options...)
}

// ChannelMessageDelete deletes a message from the Channel.
func (s *Session) ChannelMessageDelete(channelID, messageID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointChannelMessage(channelID, messageID), nil, nil, EndpointChannelMessage(channelID, 0), options...)
	return
}

// ChannelMessagesBulkDelete bulk deletes the messages from the channel for the provided messageIDs.
// If only one messageID is in the slice call channelMessageDelete function.
// If the slice is empty do nothing.
// channelID : The ID of the channel for the messages to delete.
// messages  : The IDs of the messages to be deleted. A slice of message IDs. A maximum of 100 messages.
func (s *Session) ChannelMessagesBulkDelete(channelID int64, messages []int64, options ...RequestOption) (err error) {

	if len(messages) == 0 {
		return
	}

	if len(messages) == 1 {
		err = s.ChannelMessageDelete(channelID, messages[0])
		return
	}

	if len(messages) > 100 {
		messages = messages[:100]
	}

	data := struct {
		Messages IDSlice `json:"messages"`
	}{messages}

	_, err = s.RequestWithBucketID("POST", EndpointChannelMessagesBulkDelete(channelID), data, nil, EndpointChannelMessagesBulkDelete(channelID), options...)
	return
}

// ChannelMessagePin pins a message within a given channel.
// channelID: The ID of a channel.
// messageID: The ID of a message.
func (s *Session) ChannelMessagePin(channelID, messageID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("PUT", EndpointChannelMessagePin(channelID, messageID), nil, nil, EndpointChannelMessagePin(channelID, 0), options...)
	return
}

// ChannelMessageUnpin unpins a message within a given channel.
// channelID: The ID of a channel.
// messageID: The ID of a message.
func (s *Session) ChannelMessageUnpin(channelID, messageID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointChannelMessagePin(channelID, messageID), nil, nil, EndpointChannelMessagePin(channelID, 0), options...)
	return
}

// ChannelMessagesPinned returns an array of Message structures for pinned messages
// within a given channel
// channelID : The ID of a Channel.
func (s *Session) ChannelMessagesPinned(channelID int64, options ...RequestOption) (st []*Message, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointChannelMessagesPins(channelID), nil, nil, EndpointChannelMessagesPins(channelID), options...)

	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelFileSend sends a file to the given channel.
// channelID : The ID of a Channel.
// name: The name of the file.
// io.Reader : A reader for the file contents.
func (s *Session) ChannelFileSend(channelID int64, name string, r io.Reader, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{File: &File{Name: name, Reader: r}}, options...)
}

// ChannelFileSendWithMessage sends a file to the given channel with an message.
// DEPRECATED. Use ChannelMessageSendComplex instead.
// channelID : The ID of a Channel.
// content: Optional Message content.
// name: The name of the file.
// io.Reader : A reader for the file contents.
func (s *Session) ChannelFileSendWithMessage(channelID int64, content string, name string, r io.Reader, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(channelID, &MessageSend{File: &File{Name: name, Reader: r}, Content: content}, options...)
}

// ChannelInvites returns an array of Invite structures for the given channel
// channelID   : The ID of a Channel
func (s *Session) ChannelInvites(channelID int64, options ...RequestOption) (st []*Invite, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointChannelInvites(channelID), nil, nil, EndpointChannelInvites(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelInviteCreate creates a new invite for the given channel.
// channelID   : The ID of a Channel
// i           : An Invite struct with the values MaxAge, MaxUses and Temporary defined.
func (s *Session) ChannelInviteCreate(channelID int64, i Invite, options ...RequestOption) (st *Invite, err error) {

	data := struct {
		MaxAge    int  `json:"max_age"`
		MaxUses   int  `json:"max_uses"`
		Temporary bool `json:"temporary"`
		Unique    bool `json:"unique"`
	}{i.MaxAge, i.MaxUses, i.Temporary, i.Unique}

	body, err := s.RequestWithBucketID("POST", EndpointChannelInvites(channelID), data, nil, EndpointChannelInvites(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelPermissionSet creates a Permission Override for the given channel.
// NOTE: This func name may changed.  Using Set instead of Create because
// you can both create a new override or update an override with this function.
func (s *Session) ChannelPermissionSet(channelID, targetID int64, targetType PermissionOverwriteType, allow, deny int64, options ...RequestOption) (err error) {

	data := struct {
		ID    int64                   `json:"id,string"`
		Type  PermissionOverwriteType `json:"type,string"`
		Allow int64                   `json:"allow"`
		Deny  int64                   `json:"deny"`
	}{targetID, targetType, allow, deny}

	_, err = s.RequestWithBucketID("PUT", EndpointChannelPermission(channelID, targetID), data, nil, EndpointChannelPermission(channelID, 0), options...)
	return
}

// ChannelPermissionDelete deletes a specific permission override for the given channel.
// NOTE: Name of this func may change.
func (s *Session) ChannelPermissionDelete(channelID, targetID int64, options ...RequestOption) (err error) {

	_, err = s.RequestWithBucketID("DELETE", EndpointChannelPermission(channelID, targetID), nil, nil, EndpointChannelPermission(channelID, 0), options...)
	return
}

// ChannelMessageCrosspost cross posts a message in a news channel to followers
// of the channel
// channelID   : The ID of a Channel
// messageID   : The ID of a Message
func (s *Session) ChannelMessageCrosspost(channelID, messageID int64, options ...RequestOption) (st *Message, err error) {

	endpoint := EndpointChannelMessageCrosspost(channelID, messageID)

	body, err := s.RequestWithBucketID("POST", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelNewsFollow follows a news channel in the targetID
// channelID   : The ID of a News Channel
// targetID    : The ID of a Channel where the News Channel should post to
func (s *Session) ChannelNewsFollow(channelID, targetID int64, options ...RequestOption) (st *ChannelFollow, err error) {

	endpoint := EndpointChannelFollow(channelID)

	data := struct {
		WebhookChannelID int64 `json:"webhook_channel_id,string"`
	}{targetID}

	body, err := s.RequestWithBucketID("POST", endpoint, data, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Invites
// ------------------------------------------------------------------------------------------------

// Invite returns an Invite structure of the given invite
// inviteID : The invite code
func (s *Session) Invite(inviteID string, options ...RequestOption) (st *Invite, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointInvite(inviteID), nil, nil, EndpointInvite(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// InviteWithCounts returns an Invite structure of the given invite including approximate member counts
// inviteID : The invite code
func (s *Session) InviteWithCounts(inviteID string, options ...RequestOption) (st *Invite, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointInvite(inviteID)+"?with_counts=true", nil, nil, EndpointInvite(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// InviteComplex returns an Invite structure of the given invite including specified fields.
//
//	inviteID                  : The invite code
//	guildScheduledEventID     : If specified, includes specified guild scheduled event.
//	withCounts                : Whether to include approximate member counts or not
//	withExpiration            : Whether to include expiration time or not
func (s *Session) InviteComplex(inviteID, guildScheduledEventID string, withCounts, withExpiration bool, options ...RequestOption) (st *Invite, err error) {
	endpoint := EndpointInvite(inviteID)
	v := url.Values{}
	if guildScheduledEventID != "" {
		v.Set("guild_scheduled_event_id", guildScheduledEventID)
	}
	if withCounts {
		v.Set("with_counts", "true")
	}
	if withExpiration {
		v.Set("with_expiration", "true")
	}

	if len(v) != 0 {
		endpoint += "?" + v.Encode()
	}

	body, err := s.RequestWithBucketID("GET", endpoint, nil, nil, EndpointInvite(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// InviteDelete deletes an existing invite
// inviteID   : the code of an invite
func (s *Session) InviteDelete(inviteID string, options ...RequestOption) (st *Invite, err error) {

	body, err := s.RequestWithBucketID("DELETE", EndpointInvite(inviteID), nil, nil, EndpointInvite(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// InviteAccept accepts an Invite to a Guild or Channel
// inviteID : The invite code
func (s *Session) InviteAccept(inviteID string, options ...RequestOption) (st *Invite, err error) {

	body, err := s.RequestWithBucketID("POST", EndpointInvite(inviteID), nil, nil, EndpointInvite(""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Voice
// ------------------------------------------------------------------------------------------------

// VoiceRegions returns the voice server regions
func (s *Session) VoiceRegions(options ...RequestOption) (st []*VoiceRegion, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointVoiceRegions, nil, nil, EndpointVoiceRegions, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Discord Websockets
// ------------------------------------------------------------------------------------------------

// Gateway returns the websocket Gateway address
func (s *Session) Gateway(options ...RequestOption) (gateway string, err error) {

	response, err := s.RequestWithBucketID("GET", EndpointGateway, nil, nil, EndpointGateway, options...)
	if err != nil {
		return
	}

	temp := struct {
		URL string `json:"url"`
	}{}

	err = unmarshal(response, &temp)
	if err != nil {
		return
	}

	gateway = temp.URL

	// Ensure the gateway always has a trailing slash.
	// MacOS will fail to connect if we add query params without a trailing slash on the base domain.
	if !strings.HasSuffix(gateway, "/") {
		gateway += "/"
	}

	return
}

// GatewayBot returns the websocket Gateway address and the recommended number of shards
func (s *Session) GatewayBot(options ...RequestOption) (st *GatewayBotResponse, err error) {

	response, err := s.RequestWithBucketID("GET", EndpointGatewayBot, nil, nil, EndpointGatewayBot, options...)
	if err != nil {
		return
	}

	err = unmarshal(response, &st)
	if err != nil {
		return
	}

	// Ensure the gateway always has a trailing slash.
	// MacOS will fail to connect if we add query params without a trailing slash on the base domain.
	if !strings.HasSuffix(st.URL, "/") {
		st.URL += "/"
	}

	return
}

// Functions specific to Webhooks

// WebhookCreate returns a new Webhook.
// channelID: The ID of a Channel.
// name     : The name of the webhook.
// avatar   : The avatar of the webhook.
func (s *Session) WebhookCreate(channelID int64, name, avatar string, options ...RequestOption) (st *Webhook, err error) {

	data := struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar,omitempty"`
	}{name, avatar}

	body, err := s.RequestWithBucketID("POST", EndpointChannelWebhooks(channelID), data, nil, EndpointChannelWebhooks(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// ChannelWebhooks returns all webhooks for a given channel.
// channelID: The ID of a channel.
func (s *Session) ChannelWebhooks(channelID int64, options ...RequestOption) (st []*Webhook, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointChannelWebhooks(channelID), nil, nil, EndpointChannelWebhooks(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildWebhooks returns all webhooks for a given guild.
// guildID: The ID of a Guild.
func (s *Session) GuildWebhooks(guildID int64, options ...RequestOption) (st []*Webhook, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointGuildWebhooks(guildID), nil, nil, EndpointGuildWebhooks(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// Webhook returns a webhook for a given ID
// webhookID: The ID of a webhook.
func (s *Session) Webhook(webhookID int64, options ...RequestOption) (st *Webhook, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointWebhook(webhookID), nil, nil, EndpointWebhooks, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// WebhookWithToken returns a webhook for a given ID
// webhookID: The ID of a webhook.
// token    : The auth token for the webhook.
func (s *Session) WebhookWithToken(webhookID int64, token string, options ...RequestOption) (st *Webhook, err error) {

	body, err := s.RequestWithBucketID("GET", EndpointWebhookToken(webhookID, token), nil, nil, EndpointWebhookToken(0, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// WebhookEdit updates an existing WebHook.
// webhookID: The ID of a WebHook.
// name     : The name of the WebHook.
// avatar   : The avatar of the WebHook.
func (s *Session) WebhookEdit(webhookID int64, name, avatar string, channelID int64, options ...RequestOption) (st *Webhook, err error) {

	data := struct {
		Name      string `json:"name,omitempty"`
		Avatar    string `json:"avatar,omitempty"`
		ChannelID int64  `json:"channel_id,string,omitempty"`
	}{name, avatar, channelID}

	body, err := s.RequestWithBucketID("PATCH", EndpointWebhook(webhookID), data, nil, EndpointWebhooks, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// WebhookEditWithToken updates an existing Webhook with an auth token.
// webhookID: The ID of a webhook.
// token    : The auth token for the webhook.
// name     : The name of the webhook.
// avatar   : The avatar of the webhook.
func (s *Session) WebhookEditWithToken(webhookID int64, token, name, avatar string, options ...RequestOption) (st *Webhook, err error) {

	data := struct {
		Name   string `json:"name,omitempty"`
		Avatar string `json:"avatar,omitempty"`
	}{name, avatar}

	var body []byte
	body, err = s.RequestWithBucketID("PATCH", EndpointWebhookToken(webhookID, token), data, nil, EndpointWebhookToken(0, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// WebhookDelete deletes a webhook for a given ID
// webhookID: The ID of a webhook.
func (s *Session) WebhookDelete(webhookID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointWebhook(webhookID), nil, nil, EndpointWebhooks, options...)

	return
}

// WebhookDeleteWithToken deletes a webhook for a given ID with an auth token.
// webhookID: The ID of a webhook.
// token    : The auth token for the webhook.
func (s *Session) WebhookDeleteWithToken(webhookID int64, token string, options ...RequestOption) (st *Webhook, err error) {

	body, err := s.RequestWithBucketID("DELETE", EndpointWebhookToken(webhookID, token), nil, nil, EndpointWebhookToken(0, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// WebhookExecute executes a webhook.
// webhookID: The ID of a webhook.
// token    : The auth token for the webhook
func (s *Session) WebhookExecute(webhookID int64, token string, wait bool, data *WebhookParams, options ...RequestOption) (err error) {
	uri := EndpointWebhookToken(webhookID, token)

	if wait {
		uri += "?wait=true"
	}

	_, err = s.RequestWithBucketID("POST", uri, data, nil, EndpointWebhookToken(webhookID, token), options...)

	return
}

// WebhookExecuteComplex executes a webhook.
// webhookID: The ID of a webhook.
// token    : The auth token for the webhook
func (s *Session) WebhookExecuteComplex(webhookID int64, token string, wait bool, data *WebhookParams, options ...RequestOption) (m *Message, err error) {
	uri := EndpointWebhookToken(webhookID, token)

	if wait {
		uri += "?wait=true"
	}

	endpoint := uri

	// TODO: Remove this when compatibility is not required.
	var files []*File
	if data.File != nil {
		files = []*File{data.File}
	}

	var response []byte
	if len(files) > 0 {
		body := &bytes.Buffer{}
		bodywriter := multipart.NewWriter(body)

		var payload []byte
		payload, err = json.Marshal(data)
		if err != nil {
			return
		}

		var p io.Writer

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="payload_json"`)
		h.Set("Content-Type", "application/json")

		p, err = bodywriter.CreatePart(h)
		if err != nil {
			return
		}

		if _, err = p.Write(payload); err != nil {
			return
		}

		for i, file := range files {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file%d"; filename="%s"`, i, quoteEscaper.Replace(file.Name)))
			contentType := file.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			h.Set("Content-Type", contentType)

			p, err = bodywriter.CreatePart(h)
			if err != nil {
				return
			}

			if _, err = io.Copy(p, file.Reader); err != nil {
				return
			}
		}

		err = bodywriter.Close()
		if err != nil {
			return
		}

		response, err = s.request("POST", endpoint, bodywriter.FormDataContentType(), body.Bytes(), nil, EndpointWebhookToken(webhookID, token), options...)
	} else {
		response, err = s.RequestWithBucketID("POST", endpoint, data, nil, EndpointWebhookToken(webhookID, token), options...)
	}

	if err != nil {
		return
	}

	if wait {
		err = unmarshal(response, &m)
	}

	return

	// _, err = s.RequestWithBucketID("POST", uri, data, EndpointWebhookToken(0, ""))
	// return
}

// WebhookMessage gets a webhook message.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of message to get
func (s *Session) WebhookMessage(webhookID int64, token, messageID string, options ...RequestOption) (message *Message, err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointWebhookToken(0, ""), options...)
	if err != nil {
		return
	}

	err = Unmarshal(body, &message)

	return
}

// WebhookMessageEdit edits a webhook message and returns a new one.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of message to edit
func (s *Session) WebhookMessageEdit(webhookID int64, token, messageID string, data *WebhookEdit, options ...RequestOption) (st *Message, err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	var response []byte
	if len(data.Files) > 0 {
		contentType, body, err := MultipartBodyWithJSON(data, data.Files)
		if err != nil {
			return nil, err
		}

		response, err = s.request("PATCH", uri, contentType, body, nil, uri, options...)
		if err != nil {
			return nil, err
		}
	} else {
		response, err = s.RequestWithBucketID("PATCH", uri, data, nil, EndpointWebhookToken(0, ""), options...)

		if err != nil {
			return nil, err
		}
	}

	err = unmarshal(response, &st)
	return
}

// WebhookMessageDelete deletes a webhook message.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of a message to edit
func (s *Session) WebhookMessageDelete(webhookID int64, token, messageID string, options ...RequestOption) (err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	_, err = s.RequestWithBucketID("DELETE", uri, nil, nil, EndpointWebhookToken(0, ""), options...)
	return
}

// WebhookThreadMessage gets a webhook message.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of message to get
// threadID  : Get a message in the specified thread within a webhook's channel.
func (s *Session) WebhookThreadMessage(webhookID int64, token, threadID, messageID string, options ...RequestOption) (message *Message, err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	v := url.Values{}
	v.Set("thread_id", threadID)
	uri += "?" + v.Encode()

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointWebhookToken(0, ""), options...)
	if err != nil {
		return
	}

	err = Unmarshal(body, &message)

	return
}

// WebhookThreadMessageEdit edits a webhook message in a thread and returns a new one.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of message to edit
// threadID  : Edits a message in the specified thread within a webhook's channel.
func (s *Session) WebhookThreadMessageEdit(webhookID int64, token, threadID, messageID string, data *WebhookEdit, options ...RequestOption) (st *Message, err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	v := url.Values{}
	v.Set("thread_id", threadID)
	uri += "?" + v.Encode()

	var response []byte
	if len(data.Files) > 0 {
		contentType, body, err := MultipartBodyWithJSON(data, data.Files)
		if err != nil {
			return nil, err
		}

		response, err = s.request("PATCH", uri, contentType, body, nil, uri, options...)
		if err != nil {
			return nil, err
		}
	} else {
		response, err = s.RequestWithBucketID("PATCH", uri, data, nil, EndpointWebhookToken(0, ""), options...)

		if err != nil {
			return nil, err
		}
	}

	err = unmarshal(response, &st)
	return
}

// WebhookThreadMessageDelete deletes a webhook message.
// webhookID : The ID of a webhook
// token     : The auth token for the webhook
// messageID : The ID of a message to edit
// threadID  : Deletes a message in the specified thread within a webhook's channel.
func (s *Session) WebhookThreadMessageDelete(webhookID int64, token, threadID, messageID string, options ...RequestOption) (err error) {
	uri := EndpointWebhookMessage(webhookID, token, messageID)

	v := url.Values{}
	v.Set("thread_id", threadID)
	uri += "?" + v.Encode()

	_, err = s.RequestWithBucketID("DELETE", uri, nil, nil, EndpointWebhookToken(0, ""), options...)
	return
}

// MessageReactionAdd creates an emoji reaction to a message.
// channelID : The channel ID.
// messageID : The message ID.
// emoji     : Either the unicode emoji for the reaction, or a guild emoji identifier.
func (s *Session) MessageReactionAdd(channelID, messageID int64, emoji string, options ...RequestOption) error {

	_, err := s.RequestWithBucketID("PUT", EndpointMessageReaction(channelID, messageID, EmojiName{emoji}, "@me"), nil, nil, EndpointMessageReaction(channelID, 0, EmojiName{""}, ""), options...)

	return err
}

// MessageReactionRemove deletes an emoji reaction to a message.
// channelID : The channel ID.
// messageID : The message ID.
// emoji     : Either the unicode emoji for the reaction, or a guild emoji identifier.
// userID	   : The ID of the user to delete the reaction for.
func (s *Session) MessageReactionRemove(channelID, messageID int64, emoji string, userID int64, options ...RequestOption) error {

	_, err := s.RequestWithBucketID("DELETE", EndpointMessageReaction(channelID, messageID, EmojiName{emoji}, StrID(userID)), nil, nil, EndpointMessageReaction(channelID, 0, EmojiName{""}, ""), options...)

	return err
}

// MessageReactionRemoveMe deletes an emoji reaction to a message the current user made.
// channelID : The channel ID.
// messageID : The message ID.
// emoji     : Either the unicode emoji for the reaction, or a guild emoji identifier.
func (s *Session) MessageReactionRemoveMe(channelID, messageID int64, emoji string, options ...RequestOption) error {

	_, err := s.RequestWithBucketID("DELETE", EndpointMessageReaction(channelID, messageID, EmojiName{emoji}, "@me"), nil, nil, EndpointMessageReaction(channelID, 0, EmojiName{""}, ""), options...)

	return err
}

// MessageReactionRemoveEmoji deletes all emoji reactions in a message.
// channelID : The channel ID.
// messageID : The message ID.
// emoji     : Either the unicode emoji for the reaction, or a guild emoji identifier.
func (s *Session) MessageReactionRemoveEmoji(channelID, messageID int64, emoji string, options ...RequestOption) error {

	_, err := s.RequestWithBucketID("DELETE", EndpointMessageReactions(channelID, messageID, EmojiName{emoji}), nil, nil, EndpointMessageReactions(channelID, 0, EmojiName{""}), options...)

	return err
}

// MessageReactionsRemoveAll deletes all reactions from a message
// channelID : The channel ID
// messageID : The message ID.
func (s *Session) MessageReactionsRemoveAll(channelID, messageID int64, options ...RequestOption) error {

	_, err := s.RequestWithBucketID("DELETE", EndpointMessageReactionsAll(channelID, messageID), nil, nil, EndpointMessageReactionsAll(channelID, messageID), options...)

	return err
}

// MessageReactions gets all the users reactions for a specific emoji.
// channelID : The channel ID.
// messageID : The message ID.
// emoji     : Either the Unicode emoji for the reaction, or a guild emoji identifier.
// limit     : max number of users to return (max 100)
func (s *Session) MessageReactions(channelID, messageID int64, emoji string, limit int, before, after int64, options ...RequestOption) (st []*User, err error) {
	uri := EndpointMessageReactions(channelID, messageID, EmojiName{emoji})

	v := url.Values{}

	if limit > 0 {
		if limit > 100 {
			limit = 100
		}

		v.Set("limit", strconv.Itoa(limit))
	}

	if before != 0 {
		v.Set("before", strconv.FormatInt(before, 10))
	} else if after != 0 {
		v.Set("after", strconv.FormatInt(after, 10))
	}

	if len(v) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointMessageReaction(channelID, 0, EmojiName{""}, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to threads
// ------------------------------------------------------------------------------------------------

// MessageThreadStartComplex creates a new thread from an existing message.
// channelID : Channel to create thread in
// messageID : Message to start thread from
// data : Parameters of the thread
func (s *Session) MessageThreadStartComplex(channelID, messageID int64, data *ThreadStart, options ...RequestOption) (ch *Channel, err error) {
	endpoint := EndpointChannelMessageThread(channelID, messageID)
	var body []byte
	body, err = s.RequestWithBucketID("POST", endpoint, data, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &ch)
	return
}

// MessageThreadStart creates a new thread from an existing message.
// channelID       : Channel to create thread in
// messageID       : Message to start thread from
// name            : Name of the thread
// archiveDuration : Auto archive duration (in minutes)
func (s *Session) MessageThreadStart(channelID, messageID int64, name string, archiveDuration int, options ...RequestOption) (ch *Channel, err error) {
	return s.MessageThreadStartComplex(channelID, messageID, &ThreadStart{
		Name:                name,
		AutoArchiveDuration: archiveDuration,
	}, options...)
}

// ThreadStartComplex creates a new thread.
// channelID : Channel to create thread in
// data : Parameters of the thread
func (s *Session) ThreadStartComplex(channelID int64, data *ThreadStart, options ...RequestOption) (ch *Channel, err error) {
	endpoint := EndpointChannelThreads(channelID)
	var body []byte
	body, err = s.RequestWithBucketID("POST", endpoint, data, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &ch)
	return
}

// ThreadStart creates a new thread.
// channelID       : Channel to create thread in
// name            : Name of the thread
// archiveDuration : Auto archive duration (in minutes)
func (s *Session) ThreadStart(channelID int64, name string, typ ChannelType, archiveDuration int, options ...RequestOption) (ch *Channel, err error) {
	return s.ThreadStartComplex(channelID, &ThreadStart{
		Name:                name,
		Type:                typ,
		AutoArchiveDuration: archiveDuration,
	}, options...)
}

// ForumThreadStartComplex starts a new thread (creates a post) in a forum channel.
// channelID : Channel to create thread in
// threadData : Parameters of the thread
// messageData : Parameters of the starting message
func (s *Session) ForumThreadStartComplex(channelID int64, threadData *ThreadStart, messageData *MessageSend, options ...RequestOption) (th *Channel, err error) {
	endpoint := EndpointChannelThreads(channelID)

	// TODO: Remove this when compatibility is not required.
	if messageData.Embed != nil {
		if messageData.Embeds == nil {
			messageData.Embeds = []*MessageEmbed{messageData.Embed}
		} else {
			err = fmt.Errorf("cannot specify both Embed and Embeds")
			return
		}
	}

	for _, embed := range messageData.Embeds {
		if embed.Type == "" {
			embed.Type = "rich"
		}
	}

	// TODO: Remove this when compatibility is not required.
	files := messageData.Files
	if messageData.File != nil {
		if files == nil {
			files = []*File{messageData.File}
		} else {
			err = fmt.Errorf("cannot specify both File and Files")
			return
		}
	}

	data := struct {
		*ThreadStart
		Message *MessageSend `json:"message"`
	}{ThreadStart: threadData, Message: messageData}

	var response []byte
	if len(files) > 0 {
		contentType, body, encodeErr := MultipartBodyWithJSON(data, files)
		if encodeErr != nil {
			return th, encodeErr
		}

		response, err = s.request("POST", endpoint, contentType, body, nil, endpoint, options...)
	} else {
		response, err = s.RequestWithBucketID("POST", endpoint, data, nil, endpoint, options...)
	}
	if err != nil {
		return
	}

	err = unmarshal(response, &th)
	return
}

// ForumThreadStart starts a new thread (post) in a forum channel.
// channelID       : Channel to create thread in.
// name            : Name of the thread
// archiveDuration : Auto archive duration
// content         : Content of the starting message
func (s *Session) ForumThreadStart(channelID int64, name string, archiveDuration int, content string, options ...RequestOption) (th *Channel, err error) {
	return s.ForumThreadStartComplex(channelID, &ThreadStart{
		Name:                name,
		AutoArchiveDuration: archiveDuration,
	}, &MessageSend{Content: content}, options...)
}

// ForumThreadStartEmbed starts a new thread (post) in a forum channel.
// channelID       : Channel to create thread in.
// name            : Name of the thread
// archiveDuration : Auto archive duration
// embed           : Embed data of the starting message.
func (s *Session) ForumThreadStartEmbed(channelID int64, name string, archiveDuration int, embed *MessageEmbed, options ...RequestOption) (th *Channel, err error) {
	return s.ForumThreadStartComplex(channelID, &ThreadStart{
		Name:                name,
		AutoArchiveDuration: archiveDuration,
	}, &MessageSend{Embeds: []*MessageEmbed{embed}}, options...)
}

// ForumThreadStartEmbeds starts a new thread (post) in a forum channel.
// channelID       : Channel to create thread in.
// name            : Name of the thread
// archiveDuration : Auto archive duration
// embeds           : Embeds data of the starting message.
func (s *Session) ForumThreadStartEmbeds(channelID int64, name string, archiveDuration int, embeds []*MessageEmbed, options ...RequestOption) (th *Channel, err error) {
	return s.ForumThreadStartComplex(channelID, &ThreadStart{
		Name:                name,
		AutoArchiveDuration: archiveDuration,
	}, &MessageSend{Embeds: embeds}, options...)
}

// ThreadJoin adds current user to a thread
func (s *Session) ThreadJoin(id int64, options ...RequestOption) error {
	endpoint := EndpointThreadMember(id, "@me")
	_, err := s.RequestWithBucketID("PUT", endpoint, nil, nil, endpoint, options...)
	return err
}

// ThreadLeave removes current user to a thread
func (s *Session) ThreadLeave(id int64, options ...RequestOption) error {
	endpoint := EndpointThreadMember(id, "@me")
	_, err := s.RequestWithBucketID("DELETE", endpoint, nil, nil, endpoint, options...)
	return err
}

// ThreadMemberAdd adds another member to a thread
func (s *Session) ThreadMemberAdd(threadID int64, memberID string, options ...RequestOption) error {
	endpoint := EndpointThreadMember(threadID, memberID)
	_, err := s.RequestWithBucketID("PUT", endpoint, nil, nil, endpoint, options...)
	return err
}

// ThreadMemberRemove removes another member from a thread
func (s *Session) ThreadMemberRemove(threadID int64, memberID string, options ...RequestOption) error {
	endpoint := EndpointThreadMember(threadID, memberID)
	_, err := s.RequestWithBucketID("DELETE", endpoint, nil, nil, endpoint, options...)
	return err
}

// ThreadMember returns thread member object for the specified member of a thread
// withMember : Whether to include a guild member object for each thread member
func (s *Session) ThreadMember(threadID int64, memberID string, withMember bool, options ...RequestOption) (member *ThreadMember, err error) {
	uri := EndpointThreadMember(threadID, memberID)

	queryParams := url.Values{}
	if withMember {
		queryParams.Set("with_member", "true")
	}

	if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}

	var body []byte
	body, err = s.RequestWithBucketID("GET", uri, nil, nil, uri, options...)

	if err != nil {
		return
	}

	err = unmarshal(body, &member)
	return
}

// ThreadMembers returns all members of specified thread.
// limit      : Max number of thread members to return (1-100). Defaults to 100.
// afterID    : Get thread members after this user ID
// withMember : Whether to include a guild member object for each thread member
func (s *Session) ThreadMembers(threadID int64, limit int, withMember bool, afterID int64, options ...RequestOption) (members []*ThreadMember, err error) {
	uri := EndpointThreadMembers(threadID)

	queryParams := url.Values{}
	if withMember {
		queryParams.Set("with_member", "true")
	}
	if limit > 0 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}
	if afterID != 0 {
		queryParams.Set("after", StrID(afterID))
	}

	if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}

	var body []byte
	body, err = s.RequestWithBucketID("GET", uri, nil, nil, uri, options...)

	if err != nil {
		return
	}

	err = unmarshal(body, &members)
	return
}

// ThreadsActive returns all active threads for specified channel.
func (s *Session) ThreadsActive(channelID int64, options ...RequestOption) (threads *ThreadsList, err error) {
	var body []byte
	body, err = s.RequestWithBucketID("GET", EndpointChannelActiveThreads(channelID), nil, nil, EndpointChannelActiveThreads(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &threads)
	return
}

// GuildThreadsActive returns all active threads for specified guild.
func (s *Session) GuildThreadsActive(guildID int64, options ...RequestOption) (threads *ThreadsList, err error) {
	var body []byte
	body, err = s.RequestWithBucketID("GET", EndpointGuildActiveThreads(guildID), nil, nil, EndpointGuildActiveThreads(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &threads)
	return
}

// ThreadsArchived returns archived threads for specified channel.
// before : If specified returns only threads before the timestamp
// limit  : Optional maximum amount of threads to return.
func (s *Session) ThreadsArchived(channelID int64, before *time.Time, limit int, options ...RequestOption) (threads *ThreadsList, err error) {
	endpoint := EndpointChannelPublicArchivedThreads(channelID)
	v := url.Values{}
	if before != nil {
		v.Set("before", before.Format(time.RFC3339))
	}

	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}

	if len(v) > 0 {
		endpoint += "?" + v.Encode()
	}

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &threads)
	return
}

// ThreadsPrivateArchived returns archived private threads for specified channel.
// before : If specified returns only threads before the timestamp
// limit  : Optional maximum amount of threads to return.
func (s *Session) ThreadsPrivateArchived(channelID int64, before *time.Time, limit int, options ...RequestOption) (threads *ThreadsList, err error) {
	endpoint := EndpointChannelPrivateArchivedThreads(channelID)
	v := url.Values{}
	if before != nil {
		v.Set("before", before.Format(time.RFC3339))
	}

	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}

	if len(v) > 0 {
		endpoint += "?" + v.Encode()
	}
	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &threads)
	return
}

// ThreadsPrivateJoinedArchived returns archived joined private threads for specified channel.
// before : If specified returns only threads before the timestamp
// limit  : Optional maximum amount of threads to return.
func (s *Session) ThreadsPrivateJoinedArchived(channelID int64, before *time.Time, limit int, options ...RequestOption) (threads *ThreadsList, err error) {
	endpoint := EndpointChannelJoinedPrivateArchivedThreads(channelID)
	v := url.Values{}
	if before != nil {
		v.Set("before", before.Format(time.RFC3339))
	}

	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}

	if len(v) > 0 {
		endpoint += "?" + v.Encode()
	}
	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &threads)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to application (slash) commands
// ------------------------------------------------------------------------------------------------

// ApplicationCommandCreate creates a global application command and returns it.
// appID       : The application ID.
// guildID     : Guild ID to create guild-specific application command. If empty - creates global application command.
// cmd         : New application command data.
func (s *Session) ApplicationCommandCreate(appID, guildID int64, cmd *ApplicationCommand, options ...RequestOption) (ccmd *ApplicationCommand, err error) {
	endpoint := EndpointApplicationGlobalCommands(appID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommands(appID, guildID)
	}

	body, err := s.RequestWithBucketID("POST", endpoint, *cmd, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &ccmd)

	return
}

// ApplicationCommandEdit edits application command and returns new command data.
// appID       : The application ID.
// cmdID       : Application command ID to edit.
// guildID     : Guild ID to edit guild-specific application command. If empty - edits global application command.
// cmd         : Updated application command data.
func (s *Session) ApplicationCommandEdit(appID, guildID, cmdID int64, cmd *ApplicationCommand, options ...RequestOption) (updated *ApplicationCommand, err error) {
	endpoint := EndpointApplicationGlobalCommand(appID, cmdID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommand(appID, guildID, cmdID)
	}

	body, err := s.RequestWithBucketID("PATCH", endpoint, *cmd, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &updated)

	return
}

// ApplicationCommandBulkOverwrite Creates commands overwriting existing commands. Returns a list of commands.
// appID    : The application ID.
// commands : The commands to create.
func (s *Session) ApplicationCommandBulkOverwrite(appID, guildID int64, commands []*ApplicationCommand, options ...RequestOption) (createdCommands []*ApplicationCommand, err error) {
	endpoint := EndpointApplicationGlobalCommands(appID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommands(appID, guildID)
	}

	body, err := s.RequestWithBucketID("PUT", endpoint, commands, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &createdCommands)

	return
}

// ApplicationCommandDelete deletes application command by ID.
// appID       : The application ID.
// cmdID       : Application command ID to delete.
// guildID     : Guild ID to delete guild-specific application command. If empty - deletes global application command.
func (s *Session) ApplicationCommandDelete(appID, guildID, cmdID int64, options ...RequestOption) error {
	endpoint := EndpointApplicationGlobalCommand(appID, cmdID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommand(appID, guildID, cmdID)
	}

	_, err := s.RequestWithBucketID("DELETE", endpoint, nil, nil, endpoint, options...)

	return err
}

// ApplicationCommand retrieves an application command by given ID.
// appID       : The application ID.
// cmdID       : Application command ID.
// guildID     : Guild ID to retrieve guild-specific application command. If empty - retrieves global application command.
func (s *Session) ApplicationCommand(appID, guildID, cmdID int64, options ...RequestOption) (cmd *ApplicationCommand, err error) {
	endpoint := EndpointApplicationGlobalCommand(appID, cmdID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommand(appID, guildID, cmdID)
	}

	body, err := s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &cmd)

	return
}

// ApplicationCommands retrieves all commands in application.
// appID       : The application ID.
// guildID     : Guild ID to retrieve all guild-specific application commands. If empty - retrieves global application commands.
func (s *Session) ApplicationCommands(appID, guildID int64, options ...RequestOption) (cmd []*ApplicationCommand, err error) {
	endpoint := EndpointApplicationGlobalCommands(appID)
	if guildID != 0 {
		endpoint = EndpointApplicationGuildCommands(appID, guildID)
	}

	body, err := s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &cmd)

	return
}

// GetGlobalApplicationCommands fetches all of the global commands for your application. Returns an array of ApplicationCommand objects.
// GET /applications/{application.id}/commands
func (s *Session) GetGlobalApplicationCommands(applicationID int64, options ...RequestOption) (st []*ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationCommands(applicationID), nil, nil, EndpointApplicationCommands(0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GuildApplicationCommandsPermissions returns permissions for application commands in a guild.
// appID       : The application ID
// guildID     : Guild ID to retrieve application commands permissions for.
func (s *Session) GuildApplicationCommandsPermissions(appID, guildID int64, options ...RequestOption) (permissions []*GuildApplicationCommandPermissions, err error) {
	endpoint := EndpointApplicationCommandsGuildPermissions(appID, guildID)

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &permissions)
	return
}

// ApplicationCommandPermissions returns all permissions of an application command
// appID       : The Application ID
// guildID     : The guild ID containing the application command
// cmdID       : The command ID to retrieve the permissions of
func (s *Session) ApplicationCommandPermissions(appID, guildID, cmdID int64, options ...RequestOption) (permissions *GuildApplicationCommandPermissions, err error) {
	endpoint := EndpointApplicationCommandPermissions(appID, guildID, cmdID)

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &permissions)
	return
}

// ApplicationCommandPermissionsEdit edits the permissions of an application command
// appID       : The Application ID
// guildID     : The guild ID containing the application command
// cmdID       : The command ID to edit the permissions of
// permissions : An object containing a list of permissions for the application command
//
// NOTE: Requires OAuth2 token with applications.commands.permissions.update scope
func (s *Session) ApplicationCommandPermissionsEdit(appID, guildID, cmdID int64, permissions *ApplicationCommandPermissionsList, options ...RequestOption) (err error) {
	endpoint := EndpointApplicationCommandPermissions(appID, guildID, cmdID)

	_, err = s.RequestWithBucketID("PUT", endpoint, permissions, nil, endpoint, options...)
	return
}

// ApplicationCommandPermissionsBatchEdit edits the permissions of a batch of commands
// appID       : The Application ID
// guildID     : The guild ID to batch edit commands of
// permissions : A list of permissions paired with a command ID, guild ID, and application ID per application command
//
// NOTE: This endpoint has been disabled with updates to command permissions (Permissions v2). Please use ApplicationCommandPermissionsEdit instead.
func (s *Session) ApplicationCommandPermissionsBatchEdit(appID, guildID int64, permissions []*GuildApplicationCommandPermissions, options ...RequestOption) (err error) {
	endpoint := EndpointApplicationCommandsGuildPermissions(appID, guildID)

	_, err = s.RequestWithBucketID("PUT", endpoint, permissions, nil, endpoint, options...)
	return
}

// CreateGlobalApplicationCommand creates a new global command. New global commands will be available in all guilds after 1 hour.
// POST /applications/{application.id}/commands
func (s *Session) CreateGlobalApplicationCommand(applicationID int64, command *CreateApplicationCommandRequest, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("POST", EndpointApplicationCommands(applicationID), command, nil, EndpointApplicationCommands(0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// BulkOverwriteGlobalApplicationCommands Takes a list of application commands, overwriting existing commands that are registered globally for this application. Updates will be available in all guilds after 1 hour.
// PUT /applications/{application.id}/commands
func (s *Session) BulkOverwriteGlobalApplicationCommands(applicationID int64, data []*CreateApplicationCommandRequest, options ...RequestOption) (st []*ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("PUT", EndpointApplicationCommands(applicationID), data, nil, EndpointApplicationCommands(0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GetGlobalApplicationCommand fetches a global command for your application. Returns an ApplicationCommand object.
// GET /applications/{application.id}/commands/{command.id}
func (s *Session) GetGlobalApplicationCommand(applicationID int64, cmdID int64, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationCommand(applicationID, cmdID), nil, nil, EndpointApplicationCommand(0, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// EditGlobalApplicationCommand edits a global command. Updates will be available in all guilds after 1 hour.
// PATCH /applications/{application.id}/commands/{command.id}
func (s *Session) EditGlobalApplicationCommand(applicationID int64, cmdID int64, data *EditApplicationCommandRequest, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointApplicationCommand(applicationID, cmdID), data, nil, EndpointApplicationCommand(0, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// DeleteGlobalApplicationCommand deletes a global command.
// DELETE /applications/{application.id}/commands/{command.id}
func (s *Session) DeleteGlobalApplicationCommand(applicationID int64, cmdID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointApplicationCommand(applicationID, cmdID), nil, nil, EndpointApplicationCommand(0, 0), options...)
	return
}

// GetGuildApplicationCommands fetches all of the guild commands for your application for a specific guild.
// GET /applications/{application.id}/guilds/{guild.id}/commands
func (s *Session) GetGuildApplicationCommands(applicationID int64, guildID int64, options ...RequestOption) (st []*ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationGuildCommands(applicationID, guildID), nil, nil, EndpointApplicationGuildCommands(0, guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// CreateGuildApplicationCommands Create a new guild command. New guild commands will be available in the guild immediately. Returns 201 and an ApplicationCommand object. If the command did not already exist, it will count toward daily application command create limits.
// POST /applications/{application.id}/guilds/{guild.id}/commands
func (s *Session) CreateGuildApplicationCommands(applicationID int64, guildID int64, data *CreateApplicationCommandRequest, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("POST", EndpointApplicationGuildCommands(applicationID, guildID), data, nil, EndpointApplicationGuildCommands(0, guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GetGuildApplicationCommand Fetch a guild command for your application.
// GET /applications/{application.id}/guilds/{guild.id}/commands/{command.id}
func (s *Session) GetGuildApplicationCommand(applicationID int64, guildID int64, cmdID int64, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationGuildCommand(applicationID, guildID, cmdID), nil, nil, EndpointApplicationGuildCommand(0, guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// EditGuildApplicationCommand Edit a guild command. Updates for guild commands will be available immediately.
// PATCH /applications/{application.id}/guilds/{guild.id}/commands/{command.id}
func (s *Session) EditGuildApplicationCommand(applicationID int64, guildID int64, cmdID int64, data *EditApplicationCommandRequest, options ...RequestOption) (st *ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointApplicationGuildCommand(applicationID, guildID, cmdID), data, nil, EndpointApplicationGuildCommand(0, guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// DeleteGuildApplicationCommand Delete a guild command.
// DELETE /applications/{application.id}/guilds/{guild.id}/commands/{command.id}
func (s *Session) DeleteGuildApplicationCommand(applicationID int64, guildID int64, cmdID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointApplicationGuildCommand(applicationID, guildID, cmdID), nil, nil, EndpointApplicationGuildCommand(0, guildID, 0), options...)
	return
}

// BulkOverwriteGuildApplicationCommands Takes a list of application commands, overwriting existing commands for the guild.
// PUT /applications/{application.id}/guilds/{guild.id}/commands
func (s *Session) BulkOverwriteGuildApplicationCommands(applicationID int64, guildID int64, data []*CreateApplicationCommandRequest, options ...RequestOption) (st []*ApplicationCommand, err error) {
	body, err := s.RequestWithBucketID("PUT", EndpointApplicationGuildCommands(applicationID, guildID), data, nil, EndpointApplicationGuildCommands(0, guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GetGuildApplicationCommandPermissions Fetches command permissions for all commands for your application in a guild.
// GET /applications/{application.id}/guilds/{guild.id}/commands/permissions
func (s *Session) GetGuildApplicationCommandsPermissions(applicationID int64, guildID int64, options ...RequestOption) (st []*GuildApplicationCommandPermissions, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationGuildCommandsPermissions(applicationID, guildID), nil, nil, EndpointApplicationGuildCommandsPermissions(0, guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// GetGuildApplicationCommandPermissions Fetches command permissions for a specific command for your application in a guild.
// GET /applications/{application.id}/guilds/{guild.id}/commands/{command.id}/permissions
func (s *Session) GetGuildApplicationCommandPermissions(applicationID int64, guildID int64, cmdID int64, options ...RequestOption) (st *GuildApplicationCommandPermissions, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointApplicationGuildCommandPermissions(applicationID, guildID, cmdID), nil, nil, EndpointApplicationGuildCommandPermissions(0, guildID, 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// EditGuildApplicationCommandPermissions Edits command permissions for a specific command for your application in a guild.
// PUT /applications/{application.id}/guilds/{guild.id}/commands/{command.id}/permissions
// TODO: what does this return? docs doesn't say
func (s *Session) EditGuildApplicationCommandPermissions(applicationID int64, guildID int64, cmdID int64, permissions []*ApplicationCommandPermissions, options ...RequestOption) (err error) {
	data := struct {
		Permissions []*ApplicationCommandPermissions `json:"permissions"`
	}{
		permissions,
	}

	_, err = s.RequestWithBucketID("PUT", EndpointApplicationGuildCommandPermissions(applicationID, guildID, cmdID), data, nil, EndpointApplicationGuildCommandPermissions(0, guildID, 0), options...)
	return
}

// BatchEditGuildApplicationCommandsPermissions Fetches command permissions for a specific command for your application in a guild.
// PUT /applications/{application.id}/guilds/{guild.id}/commands/permissions
func (s *Session) BatchEditGuildApplicationCommandsPermissions(applicationID int64, guildID int64, data []*GuildApplicationCommandPermissions, options ...RequestOption) (st []*GuildApplicationCommandPermissions, err error) {

	body, err := s.RequestWithBucketID("PUT", EndpointApplicationGuildCommandsPermissions(applicationID, guildID), data, nil, EndpointApplicationGuildCommandsPermissions(0, guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)

	return
}

// CreateInteractionResponse Create a response to an Interaction from the gateway. Takes an Interaction response.
// POST /interactions/{interaction.id}/{interaction.token}/callback
func (s *Session) CreateInteractionResponse(interactionID int64, token string, data *InteractionResponse, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("POST", EndpointInteractionCallback(interactionID, token), data, nil, EndpointInteractionCallback(0, ""), options...)
	return
}

// GetOriginalInteractionResponse Returns the initial Interaction response. Functions the same as Get Webhook Message.
// GET /webhooks/{application.id}/{interaction.token}/messages/@original
func (s *Session) GetOriginalInteractionResponse(applicationID int64, token string, options ...RequestOption) (st *Message, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointInteractionOriginalMessage(applicationID, token), nil, nil, EndpointInteractionOriginalMessage(0, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// Edits the initial Interaction response. Functions the same as Edit Webhook Message.
// PATCH /webhooks/{application.id}/{interaction.token}/messages/@original
func (s *Session) EditOriginalInteractionResponse(applicationID int64, token string, data *WebhookParams, options ...RequestOption) (st *Message, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointInteractionOriginalMessage(applicationID, token), data, nil, EndpointInteractionOriginalMessage(0, ""), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// DeleteInteractionResponse Deletes the initial Interaction response.
// DELETE /webhooks/{application.id}/{interaction.token}/messages/@original
func (s *Session) DeleteInteractionResponse(applicationID int64, token string, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointInteractionOriginalMessage(applicationID, token), nil, nil, EndpointInteractionOriginalMessage(0, ""), options...)
	return
}

// CreateFollowupMessage Creates a followup message for an Interaction. Functions the same as Execute WebHook, but wait is always true, and flags can be set to 64 in the body to send an ephemeral message.
// POST /webhooks/{application.id}/{interaction.token}
func (s *Session) CreateFollowupMessage(applicationID int64, token string, data *WebhookParams, options ...RequestOption) (st *Message, err error) {
	body, err := s.WebhookExecuteComplex(applicationID, token, true, data, options...)
	return body, err
}

func (s *Session) FollowupMessageCreate(interaction *Interaction, wait bool, data *WebhookParams, options ...RequestOption) (st *Message, err error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	uri := EndpointWebhookToken(interaction.ApplicationID, interaction.Token)

	v := url.Values{}
	// wait is always true for FollowupMessageCreate as mentioned in
	// https://discord.com/developers/docs/interactions/receiving-and-responding#endpoints
	v.Set("wait", "true")
	uri += "?" + v.Encode()

	var body []byte
	contentType := "application/json"
	if len(data.Files) > 0 {
		if contentType, body, err = MultipartBodyWithJSON(data, data.Files); err != nil {
			return st, err
		}
	} else {
		if body, err = Marshal(data); err != nil {
			return st, err
		}
	}

	var response []byte
	// FollowupMessageCreate not bound to global rate limit
	if response, err = s.RequestWithoutBucket("POST", uri, contentType, body, 0, options...); err != nil {
		return
	}

	err = unmarshal(response, &st)
	return
}

// EditFollowupMessage Edits a followup message for an Interaction. Functions the same as Edit Webhook Message.
// PATCH /webhooks/{application.id}/{interaction.token}/messages/{message.id}
func (s *Session) EditFollowupMessage(applicationID int64, token string, messageID int64, data *WebhookParams, options ...RequestOption) (st *Message, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointInteractionFollowupMessage(applicationID, token, messageID), data, nil, EndpointInteractionFollowupMessage(0, "", 0), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// DeleteFollowupMessage Deletes a followup message for an Interaction.
// DELETE /webhooks/{application.id}/{interaction.token}/messages/{message.id}
func (s *Session) DeleteFollowupMessage(applicationID int64, token string, messageID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointInteractionFollowupMessage(applicationID, token, messageID), nil, nil, EndpointInteractionFollowupMessage(0, "", 0), options...)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to stage instances
// ------------------------------------------------------------------------------------------------

// StageInstanceCreate creates and returns a new Stage instance associated to a Stage channel.
// data : Parameters needed to create a stage instance.
// data : The data of the Stage instance to create
func (s *Session) StageInstanceCreate(data *StageInstanceParams, options ...RequestOption) (si *StageInstance, err error) {
	body, err := s.RequestWithBucketID("POST", EndpointStageInstances, data, nil, EndpointStageInstances, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &si)
	return
}

// StageInstance will retrieve a Stage instance by ID of the Stage channel.
// channelID : The ID of the Stage channel
func (s *Session) StageInstance(channelID int64, options ...RequestOption) (si *StageInstance, err error) {
	body, err := s.RequestWithBucketID("GET", EndpointStageInstance(channelID), nil, nil, EndpointStageInstance(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &si)
	return
}

// StageInstanceEdit will edit a Stage instance by ID of the Stage channel.
// channelID : The ID of the Stage channel
// data : The data to edit the Stage instance
func (s *Session) StageInstanceEdit(channelID int64, data *StageInstanceParams, options ...RequestOption) (si *StageInstance, err error) {

	body, err := s.RequestWithBucketID("PATCH", EndpointStageInstance(channelID), data, nil, EndpointStageInstance(channelID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &si)
	return
}

// StageInstanceDelete will delete a Stage instance by ID of the Stage channel.
// channelID : The ID of the Stage channel
func (s *Session) StageInstanceDelete(channelID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointStageInstance(channelID), nil, nil, EndpointStageInstance(channelID), options...)
	return
}

// ------------------------------------------------------------------------------------------------
// Functions specific to guilds scheduled events
// ------------------------------------------------------------------------------------------------

// GuildScheduledEvents returns an array of GuildScheduledEvent for a guild
// guildID        : The ID of a Guild
// userCount      : Whether to include the user count in the response
func (s *Session) GuildScheduledEvents(guildID int64, userCount bool, options ...RequestOption) (st []*GuildScheduledEvent, err error) {
	uri := EndpointGuildScheduledEvents(guildID)
	if userCount {
		uri += "?with_user_count=true"
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildScheduledEvents(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildScheduledEvent returns a specific GuildScheduledEvent in a guild
// guildID        : The ID of a Guild
// eventID        : The ID of the event
// userCount      : Whether to include the user count in the response
func (s *Session) GuildScheduledEvent(guildID, eventID int64, userCount bool, options ...RequestOption) (st *GuildScheduledEvent, err error) {
	uri := EndpointGuildScheduledEvent(guildID, eventID)
	if userCount {
		uri += "?with_user_count=true"
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildScheduledEvent(guildID, eventID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildScheduledEventCreate creates a GuildScheduledEvent for a guild and returns it
// guildID   : The ID of a Guild
// eventID   : The ID of the event
func (s *Session) GuildScheduledEventCreate(guildID int64, event *GuildScheduledEventParams, options ...RequestOption) (st *GuildScheduledEvent, err error) {
	body, err := s.RequestWithBucketID("POST", EndpointGuildScheduledEvents(guildID), event, nil, EndpointGuildScheduledEvents(guildID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildScheduledEventEdit updates a specific event for a guild and returns it.
// guildID   : The ID of a Guild
// eventID   : The ID of the event
func (s *Session) GuildScheduledEventEdit(guildID, eventID int64, event *GuildScheduledEventParams, options ...RequestOption) (st *GuildScheduledEvent, err error) {
	body, err := s.RequestWithBucketID("PATCH", EndpointGuildScheduledEvent(guildID, eventID), event, nil, EndpointGuildScheduledEvent(guildID, eventID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildScheduledEventDelete deletes a specific GuildScheduledEvent in a guild
// guildID   : The ID of a Guild
// eventID   : The ID of the event
func (s *Session) GuildScheduledEventDelete(guildID, eventID int64, options ...RequestOption) (err error) {
	_, err = s.RequestWithBucketID("DELETE", EndpointGuildScheduledEvent(guildID, eventID), nil, nil, EndpointGuildScheduledEvent(guildID, eventID), options...)
	return
}

// GuildScheduledEventUsers returns an array of GuildScheduledEventUser for a particular event in a guild
// guildID    : The ID of a Guild
// eventID    : The ID of the event
// limit      : The maximum number of users to return (Max 100)
// withMember : Whether to include the member object in the response
// beforeID   : If is not empty all returned users entries will be before the given ID
// afterID    : If is not empty all returned users entries will be after the given ID
func (s *Session) GuildScheduledEventUsers(guildID, eventID int64, limit int, withMember bool, beforeID, afterID int64, options ...RequestOption) (st []*GuildScheduledEventUser, err error) {
	uri := EndpointGuildScheduledEventUsers(guildID, eventID)

	queryParams := url.Values{}
	if withMember {
		queryParams.Set("with_member", "true")
	}
	if limit > 0 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}
	if beforeID != 0 {
		queryParams.Set("before", StrID(beforeID))
	}
	if afterID != 0 {
		queryParams.Set("after", StrID(afterID))
	}

	if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}

	body, err := s.RequestWithBucketID("GET", uri, nil, nil, EndpointGuildScheduledEventUsers(guildID, eventID), options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// GuildOnboarding returns the onboarding flow for a guild
// guildID   : The ID of a Guild
func (s *Session) GuildOnboarding(guildID int64, options ...RequestOption) (onboarding GuildOnboarding, err error) {
	endpoint := EndpointGuildOnboarding(guildID)

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &onboarding)
	return
}

// GuildOnboardingEdit edits Onboarding for a Guild.
// guildID   : The ID of a Guild.
// o 		     : A GuildOnboarding struct.
func (s *Session) GuildOnboardingEdit(guildID int64, o *GuildOnboarding, options ...RequestOption) (onboarding *GuildOnboarding, err error) {
	endpoint := EndpointGuildOnboarding(guildID)

	var body []byte
	body, err = s.RequestWithBucketID("PUT", endpoint, o, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &onboarding)
	return
}

// ----------------------------------------------------------------------
// Functions specific to auto moderation
// ----------------------------------------------------------------------

// AutoModerationRules returns a list of auto moderation rules.
// guildID : ID of the guild
func (s *Session) AutoModerationRules(guildID int64, options ...RequestOption) (st []*AutoModerationRule, err error) {
	endpoint := EndpointGuildAutoModerationRules(guildID)

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// AutoModerationRule returns an auto moderation rule.
// guildID : ID of the guild
// ruleID  : ID of the auto moderation rule
func (s *Session) AutoModerationRule(guildID, ruleID int64, options ...RequestOption) (st *AutoModerationRule, err error) {
	endpoint := EndpointGuildAutoModerationRule(guildID, ruleID)

	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// AutoModerationRuleCreate creates an auto moderation rule with the given data and returns it.
// guildID : ID of the guild
// rule    : Rule data
func (s *Session) AutoModerationRuleCreate(guildID int64, rule *AutoModerationRule, options ...RequestOption) (st *AutoModerationRule, err error) {
	endpoint := EndpointGuildAutoModerationRules(guildID)

	var body []byte
	body, err = s.RequestWithBucketID("POST", endpoint, rule, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// AutoModerationRuleEdit edits and returns the updated auto moderation rule.
// guildID : ID of the guild
// ruleID  : ID of the auto moderation rule
// rule    : New rule data
func (s *Session) AutoModerationRuleEdit(guildID, ruleID int64, rule *AutoModerationRule, options ...RequestOption) (st *AutoModerationRule, err error) {
	endpoint := EndpointGuildAutoModerationRule(guildID, ruleID)

	var body []byte
	body, err = s.RequestWithBucketID("PATCH", endpoint, rule, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// AutoModerationRuleDelete deletes an auto moderation rule.
// guildID : ID of the guild
// ruleID  : ID of the auto moderation rule
func (s *Session) AutoModerationRuleDelete(guildID, ruleID int64, options ...RequestOption) (err error) {
	endpoint := EndpointGuildAutoModerationRule(guildID, ruleID)
	_, err = s.RequestWithBucketID("DELETE", endpoint, nil, nil, endpoint, options...)
	return
}

// ApplicationRoleConnectionMetadata returns application role connection metadata.
// appID : ID of the application
func (s *Session) ApplicationRoleConnectionMetadata(appID int64, options ...RequestOption) (st []*ApplicationRoleConnectionMetadata, err error) {
	endpoint := EndpointApplicationRoleConnectionMetadata(appID)
	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ApplicationRoleConnectionMetadataUpdate updates and returns application role connection metadata.
// appID    : ID of the application
// metadata : New metadata
func (s *Session) ApplicationRoleConnectionMetadataUpdate(appID int64, metadata []*ApplicationRoleConnectionMetadata, options ...RequestOption) (st []*ApplicationRoleConnectionMetadata, err error) {
	endpoint := EndpointApplicationRoleConnectionMetadata(appID)
	var body []byte
	body, err = s.RequestWithBucketID("PUT", endpoint, metadata, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// UserApplicationRoleConnection returns user role connection to the specified application.
// appID : ID of the application
func (s *Session) UserApplicationRoleConnection(appID int64, options ...RequestOption) (st *ApplicationRoleConnection, err error) {
	endpoint := EndpointUserApplicationRoleConnection(appID)
	var body []byte
	body, err = s.RequestWithBucketID("GET", endpoint, nil, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return

}

// UserApplicationRoleConnectionUpdate updates and returns user role connection to the specified application.
// appID      : ID of the application
// connection : New ApplicationRoleConnection data
func (s *Session) UserApplicationRoleConnectionUpdate(appID int64, rconn *ApplicationRoleConnection, options ...RequestOption) (st *ApplicationRoleConnection, err error) {
	endpoint := EndpointUserApplicationRoleConnection(appID)
	var body []byte
	body, err = s.RequestWithBucketID("PUT", endpoint, rconn, nil, endpoint, options...)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}
