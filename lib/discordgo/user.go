package discordgo

import (
	"fmt"
	"strconv"

	"github.com/mrbentarikau/pagst/lib/gojay"
)

// UserFlags is the flags of "user" (see UserFlags* consts)
// https://discord.com/developers/docs/resources/user#user-object-user-flags
type UserFlags int

// Valid UserFlags values
const (
	UserFlagDiscordEmployee           UserFlags = 1 << 0
	UserFlagDiscordPartner            UserFlags = 1 << 1
	UserFlagHypeSquadEvents           UserFlags = 1 << 2
	UserFlagBugHunterLevel1           UserFlags = 1 << 3
	UserFlagHouseBravery              UserFlags = 1 << 6
	UserFlagHouseBrilliance           UserFlags = 1 << 7
	UserFlagHouseBalance              UserFlags = 1 << 8
	UserFlagEarlySupporter            UserFlags = 1 << 9
	UserFlagTeamUser                  UserFlags = 1 << 10
	UserFlagSystem                    UserFlags = 1 << 12
	UserFlagBugHunterLevel2           UserFlags = 1 << 14
	UserFlagVerifiedBot               UserFlags = 1 << 16
	UserFlagVerifiedBotDeveloper      UserFlags = 1 << 17
	UserFlagDiscordCertifiedModerator UserFlags = 1 << 18
	UserFlagBotHTTPInteractions       UserFlags = 1 << 19
	UserFlagSpammer                   UserFlags = 1 << 20
	UserFlagActiveBotDeveloper        UserFlags = 1 << 22
)

// UserPremiumType is the premium type of a user (see UserPremiumType* consts)
// https://discord.com/developers/docs/resources/user#user-object-premium-types
type UserPremiumType int

// Valid UserPremiumType values
const (
	UserPremiumTypeNone         UserPremiumType = 0
	UserPremiumTypeNitroClassic UserPremiumType = 1
	UserPremiumTypeNitro        UserPremiumType = 2
	UserPremiumTypeNitroBasic   UserPremiumType = 3
)

// A User stores all data for an individual Discord user.
type User struct {
	// The ID of the user.
	ID int64 `json:"id,string"`

	// The email of the user. This is only present when
	// the application possesses the email scope for the user.
	Email string `json:"email"`

	// The user's username.
	Username string `json:"username"`

	// The user's display name, if it is set.
	// For bots, this is the application name.
	GlobalName string `json:"global_name"`

	// The hash of the user's avatar. Use Session.UserAvatar
	// to retrieve the avatar itself.
	Avatar string `json:"avatar"`

	// The user's chosen language option.
	Locale string `json:"locale"`

	// The discriminator of the user (4 numbers after name).
	Discriminator string `json:"discriminator"`

	// The token of the user. This is only present for
	// the user represented by the current session.
	Token string `json:"token"`

	// Whether the user's email is verified.
	Verified bool `json:"verified"`

	// Whether the user has multi-factor authentication enabled.
	MFAEnabled bool `json:"mfa_enabled"`

	// The hash of the user's banner image.
	Banner string `json:"banner"`

	// User's banner color, encoded as an integer representation of hexadecimal color code
	AccentColor int `json:"accent_color"`

	BannerColor string `json:"banner_color"`

	// Whether the user is a bot.
	Bot bool `json:"bot"`

	// User's banner color, encoded as an integer representation of hexadecimal color code
	// BannerColor int `json:"banner_color"`

	// The public flags on a user's account.
	// This is a combination of bit masks; the presence of a certain flag can
	// be checked by performing a bitwise AND between this int and the flag.
	PublicFlags UserFlags `json:"public_flags"`

	// The type of Nitro subscription on a user's account.
	// Only available when the request is authorized via a Bearer token.
	PremiumType UserPremiumType `json:"premium_type"`

	// Whether the user is an Official Discord System user (part of the urgent message system).
	System bool `json:"system"`

	// The flags on a user's account.
	// Only available when the request is authorized via a Bearer token.
	Flags int `json:"flags"`
}

// String returns a unique identifier of the form username#discriminator
// or username only if the discriminator is "0"
func (u *User) String() string {
	// If the user has been migrated from the legacy username system, their discriminator is "0".
	// See https://support-dev.discord.com/hc/en-us/articles/13667755828631
	if u.Discriminator == "0" {
		return u.Username
	}

	// The code below handles applications and users without a migrated username.
	// https://support-dev.discord.com/hc/en-us/articles/13667755828631
	return u.Username + "#" + u.Discriminator
}

// Mention return a string which mentions the user
func (u *User) Mention() string {
	return fmt.Sprintf("<@%d>", u.ID)
}

// implement gojay.UnmarshalerJSONObject
func (u *User) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "id":
		return DecodeSnowflake(&u.ID, dec)
	case "username":
		return dec.String(&u.Username)
	case "global_name":
		return dec.String(&u.GlobalName)
	case "avatar":
		return dec.String(&u.Avatar)
	case "banner":
		return dec.String(&u.Banner)
	case "locale":
		return dec.String(&u.Locale)
	case "discriminator":
		return dec.String(&u.Discriminator)
	case "bot":
		return dec.Bool(&u.Bot)
	case "mfa_enabled":
		return dec.Bool(&u.MFAEnabled)
	}

	return nil
}

func (u *User) NKeys() int {
	return 0
}

// AvatarURL returns a URL to the user's avatar.
//
//	size:    The size of the user's avatar as a power of two
//	         if size is an empty string, no size parameter will
//	         be added to the URL.
func (u *User) AvatarURL(sizeArg ...string) string {
	size := "256"
	if len(sizeArg) > 0 {
		size = sizeArg[0]
	}

	return avatarURL(
		u.Avatar,
		EndpointDefaultUserAvatar(u.DefaultAvatarIndex()),
		EndpointUserAvatar(u.ID, u.Avatar),
		EndpointUserAvatarAnimated(u.ID, u.Avatar),
		size,
	)
}

// BannerURL returns the URL of the users's banner image.
//
//		size:    The size of the desired banner image as a power of two
//	        Image size can be any power of two between 16 and 4096.
func (u *User) BannerURL(sizeArg ...string) string {
	size := "256"
	if len(sizeArg) > 0 {
		size = sizeArg[0]
	}

	return bannerURL(u.Banner, EndpointUserBanner(u.ID, u.Banner), EndpointUserBannerAnimated(u.ID, u.Banner), size)
}

// A SelfUser stores user data about the token owner.
// Includes a few extra fields than a normal user struct.
type SelfUser struct {
	*User
	Token string `json:"token"`
}

// AvatarIndex returns the index of the user's avatar
func (u *User) DefaultAvatarIndex() int64 {
	if u.Discriminator == "0" {
		return (u.ID >> 22) % 6
	}

	id, _ := strconv.Atoi(u.Discriminator)
	return int64(id % 5)
}
