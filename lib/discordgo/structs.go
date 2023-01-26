// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains all structures for the discordgo package.  These
// may be moved about later into separate files but I find it easier to have
// them all located together.

package discordgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mrbentarikau/pagst/lib/gojay"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

// A Session represents a connection to the Discord API.
type Session struct {
	// General configurable settings.

	// Authentication token for this session
	Token   string
	MFA     bool
	Intents []GatewayIntent

	// Debug for printing JSON request/responses
	Debug    bool // Deprecated, will be removed.
	LogLevel int

	// Should the session reconnect the websocket on errors.
	ShouldReconnectOnError bool

	// Should the session request compressed websocket data.
	Compress bool

	// Sharding
	ShardID    int
	ShardCount int

	// Should state tracking be enabled.
	// State tracking is the best way for getting the the users
	// active guilds and the members of the guilds.
	StateEnabled bool

	// Whether or not to call event handlers synchronously.
	// e.g false = launch event handlers in their own goroutines.
	SyncEvents bool

	// Max number of REST API retries
	MaxRestRetries int

	// Managed state object, updated internally with events when
	// StateEnabled is true.
	State *State

	// The http client used for REST requests
	Client *http.Client

	// Stores the last HeartbeatAck that was recieved (in UTC)
	LastHeartbeatAck time.Time

	// used to deal with rate limits
	Ratelimiter *RateLimiter

	// The gateway websocket connection
	GatewayManager *GatewayConnectionManager

	tokenInvalid *int32

	// Event handlers
	handlersMu   sync.RWMutex
	handlers     map[string][]*eventHandlerInstance
	onceHandlers map[string][]*eventHandlerInstance
}

/*
// Application stores values for a Discord Application
type Application struct {
	ID                  int64    `json:"id,omitempty,string"`
	Name                string   `json:"name"`
	Icon                string   `json:"icon,omitempty"`
	Description         string   `json:"description,omitempty"`
	RPCOrigins          []string `json:"rpc_origins,omitempty"`
	BotPublic           bool     `json:"bot_public,omitempty"`
	BotRequireCodeGrant bool     `json:"bot_require_code_grant,omitempty"`
	TermsOfServiceURL   string   `json:"terms_of_service_url"`
	PrivacyProxyURL     string   `json:"privacy_policy_url"`
	Owner               *User    `json:"owner"`
	Summary             string   `json:"summary"`
	VerifyKey           string   `json:"verify_key"`
	Team                *Team    `json:"team"`
	GuildID             int64    `json:"guild_id,string"`
	PrimarySKUID        int64    `json:"primary_sku_id,string"`
	Slug                string   `json:"slug"`
	CoverImage          string   `json:"cover_image"`
	Flags               int      `json:"flags,omitempty"`
}
*/

// ApplicationRoleConnectionMetadataType represents the type of application role connection metadata.
type ApplicationRoleConnectionMetadataType int

// Application role connection metadata types.
const (
	ApplicationRoleConnectionMetadataIntegerLessThanOrEqual     ApplicationRoleConnectionMetadataType = 1
	ApplicationRoleConnectionMetadataIntegerGreaterThanOrEqual  ApplicationRoleConnectionMetadataType = 2
	ApplicationRoleConnectionMetadataIntegerEqual               ApplicationRoleConnectionMetadataType = 3
	ApplicationRoleConnectionMetadataIntegerNotEqual            ApplicationRoleConnectionMetadataType = 4
	ApplicationRoleConnectionMetadataDatetimeLessThanOrEqual    ApplicationRoleConnectionMetadataType = 5
	ApplicationRoleConnectionMetadataDatetimeGreaterThanOrEqual ApplicationRoleConnectionMetadataType = 6
	ApplicationRoleConnectionMetadataBooleanEqual               ApplicationRoleConnectionMetadataType = 7
	ApplicationRoleConnectionMetadataBooleanNotEqual            ApplicationRoleConnectionMetadataType = 8
)

// ApplicationRoleConnectionMetadata stores application role connection metadata.
type ApplicationRoleConnectionMetadata struct {
	Type                     ApplicationRoleConnectionMetadataType `json:"type"`
	Key                      string                                `json:"key"`
	Name                     string                                `json:"name"`
	NameLocalizations        map[Locale]string                     `json:"name_localizations"`
	Description              string                                `json:"description"`
	DescriptionLocalizations map[Locale]string                     `json:"description_localizations"`
}

// ApplicationRoleConnection represents the role connection that an application has attached to a user.
type ApplicationRoleConnection struct {
	PlatformName     string            `json:"platform_name"`
	PlatformUsername string            `json:"platform_username"`
	Metadata         map[string]string `json:"metadata"`
}

// UserConnection is a Connection returned from the UserConnections endpoint
type UserConnection struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Revoked      bool           `json:"revoked"`
	Integrations []*Integration `json:"integrations"`
}

// Integration stores integration information
type Integration struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Type              string                  `json:"type"`
	Enabled           bool                    `json:"enabled"`
	Syncing           bool                    `json:"syncing"`
	RoleID            string                  `json:"role_id"`
	EnableEmoticons   bool                    `json:"enable_emoticons"`
	ExpireBehavior    ExpireBehavior          `json:"expire_behavior"`
	ExpireGracePeriod int                     `json:"expire_grace_period"`
	User              *User                   `json:"user"`
	Account           IntegrationAccount      `json:"account"`
	SyncedAt          time.Time               `json:"synced_at"`
	SubscriberCount   int                     `json:"subscriber_count"`
	Revoked           bool                    `json:"revoked"`
	Application       *IntegrationApplication `json:"application"`

	//GuildID int64 `json:"guild_id,string,omitempty"` // Sent in the Integration events
}

// ExpireBehavior of Integration
// https://discord.com/developers/docs/resources/guild#integration-object-integration-expire-behaviors
type ExpireBehavior int

// Block of valid ExpireBehaviors
const (
	ExpireBehaviorRemoveRole ExpireBehavior = iota
	ExpireBehaviorKick
)

// IntegrationAccount is integration account information
// sent by the UserConnections endpoint
type IntegrationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IntegrationApplication https://discord.com/developers/docs/resources/guild#integration-application-object
type IntegrationApplication struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
	// Icon        string `json:"icon"`
	// Description string `json:"description"`
	// Summary     string `json:"summary"`
	// Bot         *User  `json:"bot"`
}

// A VoiceRegion stores data for a specific voice region server.
type VoiceRegion struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InviteTargetType indicates the type of target of an invite
// https://discord.com/developers/docs/resources/invite#invite-object-invite-target-types
type InviteTargetType uint8

// Invite target types
const (
	InviteTargetStream              InviteTargetType = 1
	InviteTargetEmbeddedApplication InviteTargetType = 2
)

// A Invite stores all data related to a specific Discord Guild or Channel invite.
type Invite struct {
	Guild             *Guild           `json:"guild"`
	Channel           *Channel         `json:"channel"`
	Inviter           *User            `json:"inviter"`
	Code              string           `json:"code"`
	CreatedAt         Timestamp        `json:"created_at"`
	MaxAge            int              `json:"max_age"`
	Uses              int              `json:"uses"`
	MaxUses           int              `json:"max_uses"`
	Revoked           bool             `json:"revoked"`
	Temporary         bool             `json:"temporary"`
	Unique            bool             `json:"unique"`
	TargetUser        *User            `json:"target_user"`
	TargetType        InviteTargetType `json:"target_type"`
	TargetApplication *Application     `json:"target_application"`

	// will only be filled when using InviteWithCounts
	ApproximatePresenceCount int `json:"approximate_presence_count"`
	ApproximateMemberCount   int `json:"approximate_member_count"`

	ExpiresAt Timestamp `json:"expires_at"`
}

// ChannelType is the type of a Channel
type ChannelType int

// Block contains known ChannelType values
const (
	ChannelTypeGuildText          ChannelType = 0  // a text channel within a server
	ChannelTypeDM                 ChannelType = 1  // a direct message between users
	ChannelTypeGuildVoice         ChannelType = 2  // a voice channel within a server
	ChannelTypeGroupDM            ChannelType = 3  // a direct message between multiple users
	ChannelTypeGuildCategory      ChannelType = 4  // an organizational category that contains up to 50 channels
	ChannelTypeGuildNews          ChannelType = 5  // a channel that users can follow and crosspost into their own server
	ChannelTypeGuildStore         ChannelType = 6  // a channel in which game developers can sell their game on Discord
	ChannelTypeGuildNewsThread    ChannelType = 10 // a temporary sub-channel within a GUILD_NEWS channel
	ChannelTypeGuildPublicThread  ChannelType = 11 // a temporary sub-channel within a GUILD_TEXT channel
	ChannelTypeGuildPrivateThread ChannelType = 12 // a temporary sub-channel within a GUILD_TEXT channel that is only viewable by those invited and those with the MANAGE_THREADS permission
	ChannelTypeGuildStageVoice    ChannelType = 13 // a voice channel for hosting events with an audience
	ChannelTypeGuildForum         ChannelType = 15 // a channel that can only contain threads
)

// ChannelFlags represent flags of a channel/thread.
type ChannelFlags int

// Block containing known ChannelFlags values.
const (
	// ChannelFlagPinned indicates whether the thread is pinned in the forum channel.
	// NOTE: forum threads only.
	ChannelFlagPinned ChannelFlags = 1 << 1
	// ChannelFlagRequireTag indicates whether a tag is required to be specified when creating a thread.
	// NOTE: forum channels only.
	ChannelFlagRequireTag ChannelFlags = 1 << 4
)

// ForumSortOrderType represents sort order of a forum channel.
type ForumSortOrderType int

const (
	// ForumSortOrderLatestActivity sorts posts by activity.
	ForumSortOrderLatestActivity ForumSortOrderType = 0
	// ForumSortOrderCreationDate sorts posts by creation time (from most recent to oldest).
	ForumSortOrderCreationDate ForumSortOrderType = 1
)

// ForumLayout represents layout of a forum channel.
type ForumLayout int

const (
	// ForumLayoutNotSet represents no default layout.
	ForumLayoutNotSet ForumLayout = 0
	// ForumLayoutListView displays forum posts as a list.
	ForumLayoutListView ForumLayout = 1
	// ForumLayoutGalleryView displays forum posts as a collection of tiles.
	ForumLayoutGalleryView ForumLayout = 2
)

// A Channel holds all data related to an individual Discord channel.
type Channel struct {
	// The ID of the channel.
	ID int64 `json:"id,string"`

	// The ID of the guild to which the channel belongs, if it is in a guild.
	// Else, this ID is empty (e.g. DM channels).
	GuildID int64 `json:"guild_id,string"`

	// The name of the channel.
	Name string `json:"name"`

	// The topic of the channel.
	Topic string `json:"topic"`

	// The type of the channel.
	Type ChannelType `json:"type"`

	// The ID of the last message sent in the channel. This is not
	// guaranteed to be an ID of a valid message.
	LastMessageID int64 `json:"last_message_id,string"`

	// The timestamp of the last pinned message in the channel.
	// nil if the channel has no pinned messages.
	LastPinTimestamp *time.Time `json:"last_pin_timestamp"`

	// An approximate count of messages in a thread, stops counting at 50
	MessageCount int `json:"message_count"`
	// An approximate count of users in a thread, stops counting at 50
	MemberCount int `json:"member_count"`

	// Whether the channel is marked as NSFW.
	NSFW bool `json:"nsfw"`

	// Icon of the group DM channel.
	Icon string `json:"icon"`

	// The position of the channel, used for sorting in client.
	Position int `json:"position"`

	// The bitrate of the channel, if it is a voice channel.
	Bitrate int `json:"bitrate"`

	// The recipients of the channel. This is only populated in DM channels.
	Recipients []*User `json:"recipients"`

	// The messages in the channel. This is only present in state-cached channels,
	// and State.MaxMessageCount must be non-zero.
	Messages []*Message `json:"-"`

	// A list of permission overwrites present for the channel.
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites"`

	// The user limit of the voice channel.
	UserLimit int `json:"user_limit"`

	// The ID of the parent channel, if the channel is under a category
	ParentID int64 `json:"parent_id,string"`

	// Amount of seconds a user has to wait before sending another message or creating another thread (0-21600)
	// bots, as well as users with the permission manage_messages or manage_channel, are unaffected
	RateLimitPerUser int `json:"rate_limit_per_user"`

	// ID of the creator of the group DM or thread
	OwnerID int64 `json:"owner_id,string"`

	// ApplicationID of the DM creator Zeroed if guild channel or not a bot user
	ApplicationID int64 `json:"application_id"`

	// Thread-specific fields not needed by other channels
	ThreadMetadata *ThreadMetadata `json:"thread_metadata"`

	// Thread member object for the current user, if they have joined the thread, only included on certain API endpoints
	Member *ThreadMember `json:"thread_member"`

	// All thread members. State channels only.
	Members []*ThreadMember `json:"-"`

	// Channel flags.
	Flags ChannelFlags `json:"flags"`

	// The set of tags that can be used in a forum channel.
	AvailableTags []ForumTag `json:"available_tags"`

	// The IDs of the set of tags that have been applied to a thread in a forum channel.
	AppliedTags []string `json:"applied_tags"`

	// Emoji to use as the default reaction to a forum post.
	DefaultReactionEmoji ForumDefaultReaction `json:"default_reaction_emoji"`

	// The initial RateLimitPerUser to set on newly created threads in a channel.
	// This field is copied to the thread at creation time and does not live update.
	DefaultThreadRateLimitPerUser int `json:"default_thread_rate_limit_per_user"`

	// The default sort order type used to order posts in forum channels.
	// Defaults to null, which indicates a preferred sort order hasn't been set by a channel admin.
	DefaultSortOrder *ForumSortOrderType `json:"default_sort_order"`

	// The default forum layout view used to display posts in forum channels.
	// Defaults to ForumLayoutNotSet, which indicates a layout view has not been set by a channel admin.
	DefaultForumLayout ForumLayout `json:"default_forum_layout"`
}

func (c *Channel) GetChannelID() int64 {
	return c.ID
}

func (c *Channel) GetGuildID() int64 {
	return c.GuildID
}

// Mention returns a string which mentions the channel
func (c *Channel) Mention() string {
	return fmt.Sprintf("<#%d>", c.ID)
}

func (t ChannelType) IsThread() bool {
	return t == ChannelTypeGuildPrivateThread || t == ChannelTypeGuildPublicThread
}

// A ChannelEdit holds Channel Feild data for a channel edit.
type ChannelEdit struct {
	Name                          string                 `json:"name,omitempty"`
	Topic                         string                 `json:"topic,omitempty"`
	NSFW                          *bool                  `json:"nsfw,omitempty"`
	Position                      *int                   `json:"position,omitempty"`
	Bitrate                       int                    `json:"bitrate,omitempty"`
	UserLimit                     int                    `json:"user_limit,omitempty"`
	PermissionOverwrites          []*PermissionOverwrite `json:"permission_overwrites,omitempty"`
	ParentID                      *null.String           `json:"parent_id,omitempty"`
	RateLimitPerUser              *int                   `json:"rate_limit_per_user,omitempty"`
	Flags                         *ChannelFlags          `json:"flags,omitempty"`
	DefaultThreadRateLimitPerUser *int                   `json:"default_thread_rate_limit_per_user,omitempty"`

	// NOTE: threads only

	Archived            *bool `json:"archived,omitempty"`
	AutoArchiveDuration int   `json:"auto_archive_duration,omitempty"`
	Locked              *bool `json:"locked,omitempty"`
	Invitable           *bool `json:"invitable,omitempty"`

	// NOTE: forum channels only
	AvailableTags        *[]ForumTag           `json:"available_tags,omitempty"`
	DefaultReactionEmoji *ForumDefaultReaction `json:"default_reaction_emoji,omitempty"`
	DefaultSortOrder     *ForumSortOrderType   `json:"default_sort_order,omitempty"` // TODO: null
	DefaultForumLayout   *ForumLayout          `json:"default_forum_layout,omitempty"`

	// NOTE: forum threads only
	AppliedTags *[]string `json:"applied_tags,omitempty"`
}

// A ChannelFollow holds data returned after following a news channel
type ChannelFollow struct {
	ChannelID string `json:"channel_id"`
	WebhookID string `json:"webhook_id"`
}

type RoleCreate struct {
	Name        string `json:"name,omitempty"`
	Permissions int64  `json:"permissions,string,omitempty"`
	Color       int32  `json:"color,omitempty"`
	Hoist       bool   `json:"hoist"`
	Mentionable bool   `json:"mentionable"`
}

type PermissionOverwriteType int

const (
	PermissionOverwriteTypeRole   PermissionOverwriteType = 0
	PermissionOverwriteTypeMember PermissionOverwriteType = 1
)

// A PermissionOverwrite holds permission overwrite data for a Channel
type PermissionOverwrite struct {
	ID    int64                   `json:"id,string"`
	Type  PermissionOverwriteType `json:"type"`
	Deny  int64                   `json:"deny,string"`
	Allow int64                   `json:"allow,string"`
}

// ThreadStart stores all parameters you can use with MessageThreadStartComplex or ThreadStartComplex
type ThreadStart struct {
	Name                string      `json:"name"`
	AutoArchiveDuration int         `json:"auto_archive_duration,omitempty"`
	Type                ChannelType `json:"type,omitempty"`
	Invitable           bool        `json:"invitable"`
	RateLimitPerUser    int         `json:"rate_limit_per_user,omitempty"`

	// NOTE: forum threads only
	AppliedTags []string `json:"applied_tags,omitempty"`
}

// ThreadMetadata contains a number of thread-specific channel fields that are not needed by other channel types.
type ThreadMetadata struct {
	// Whether the thread is archived
	Archived bool `json:"archived"`
	// Duration in minutes to automatically archive the thread after recent activity, can be set to: 60, 1440, 4320, 10080
	AutoArchiveDuration int `json:"auto_archive_duration"`
	// Timestamp when the thread's archive status was last changed, used for calculating recent activity
	ArchiveTimestamp Timestamp `json:"archive_timestamp"`
	// Whether the thread is locked; when a thread is locked, only users with MANAGE_THREADS can unarchive it
	Locked bool `json:"locked"`
	// Whether non-moderators can add other non-moderators to a thread; only available on private threads
	Invitable bool `json:"invitable"`
}

// ThreadMember is used to indicate whether a user has joined a thread or not.
// NOTE: ID and UserID are empty (omitted) on the member sent within each thread in the GUILD_CREATE event.
type ThreadMember struct {
	// The id of the thread
	ID int64 `json:"id,string,omitempty"`
	// The id of the user
	UserID int64 `json:"user_id,string,omitempty"`
	// The time the current user last joined the thread
	JoinTimestamp Timestamp `json:"join_timestamp"`
	// Any user-thread settings, currently only used for notifications
	Flags int `json:"flags"`
}

// ThreadsList represents a list of threads alongisde with thread member objects for the current user.
type ThreadsList struct {
	Threads []*Channel      `json:"threads"`
	Members []*ThreadMember `json:"members"`
	HasMore bool            `json:"has_more"`
}

// AddedThreadMember holds information about the user who was added to the thread
type AddedThreadMember struct {
	*ThreadMember
	Member   *Member   `json:"member"`
	Presence *Presence `json:"presence"`
}

// ForumDefaultReaction specifies emoji to use as the default reaction to a forum post.
// NOTE: Exactly one of EmojiID and EmojiName must be set.
type ForumDefaultReaction struct {
	// The id of a guild's custom emoji.
	EmojiID int64 `json:"emoji_id,omitempty,string"`
	// The unicode character of the emoji.
	EmojiName string `json:"emoji_name,omitempty"`
}

// ForumTag represents a tag that is able to be applied to a thread in a GUILD_FORUM channel.
type ForumTag struct {
	ID        int64  `json:"id,string,omitempty"`
	Name      string `json:"name"`
	Moderated bool   `json:"moderated"`
	EmojiID   int64  `json:"emoji_id,string,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

// Emoji struct holds data related to Emoji's
type Emoji struct {
	ID            int64   `json:"id,string"`
	Name          string  `json:"name"`
	Roles         IDSlice `json:"roles"`
	Managed       bool    `json:"managed"`
	RequireColons bool    `json:"require_colons"`
	Animated      bool    `json:"animated"`
}

// EmojiRegex is the regex used to find and identify emojis in messages
var (
	EmojiRegex = regexp.MustCompile(`<(a|):[A-z0-9_~]+:[0-9]{18,20}>`)
)

// MessageFormat returns a correctly formatted Emoji for use in Message content and embeds
func (e *Emoji) MessageFormat() string {
	if e.ID != 0 && e.Name != "" {
		if e.Animated {
			return "<a:" + e.APIName() + ">"
		}

		return "<:" + e.APIName() + ">"
	}

	return e.APIName()
}

// APIName returns an correctly formatted API name for use in the MessageReactions endpoints.
func (e *Emoji) APIName() string {
	if e.ID != 0 && e.Name != "" {
		return e.Name + ":" + StrID(e.ID)
	}
	if e.Name != "" {
		return e.Name
	}
	return StrID(e.ID)
}

// EmojiParams represents parameters needed to create or update an Emoji.
type EmojiParams struct {
	// Name of the emoji
	Name string `json:"name,omitempty"`
	// A base64 encoded emoji image, has to be smaller than 256KB.
	// NOTE: can be only set on creation.
	Image string `json:"image,omitempty"`
	// Roles for which this emoji will be available.
	Roles []string `json:"roles,omitempty"`
}

// StickerFormat is the file format of the Sticker.
type StickerFormat int

// Defines all known Sticker types.
const (
	StickerFormatTypePNG    StickerFormat = 1
	StickerFormatTypeAPNG   StickerFormat = 2
	StickerFormatTypeLottie StickerFormat = 3
)

// StickerType is the type of sticker.
type StickerType int

// Defines Sticker types.
const (
	StickerTypeStandard StickerType = 1
	StickerTypeGuild    StickerType = 2
)

// Sticker represents a sticker object that can be sent in a Message.
type Sticker struct {
	ID          int64         `json:"id,string"`
	PackID      int64         `json:"pack_id,string"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Tags        string        `json:"tags"`
	Type        StickerType   `json:"type"`
	FormatType  StickerFormat `json:"format_type"`
	Available   bool          `json:"available"`
	GuildID     int64         `json:"guild_id,string"`
	User        *User         `json:"user"`
	SortValue   int           `json:"sort_value"`
}

// StickerPack represents a pack of standard stickers.
type StickerPack struct {
	ID             int64     `json:"id,string"`
	Stickers       []Sticker `json:"stickers"`
	Name           string    `json:"name"`
	SKUID          int64     `json:"sku_id,string"`
	CoverStickerID int64     `json:"cover_sticker_id,string"`
	Description    string    `json:"description"`
	BannerAssetID  int64     `json:"banner_asset_id,string"`
}

// VerificationLevel type definition
type VerificationLevel int

// Constants for VerificationLevel levels from 0 to 3 inclusive
const (
	VerificationLevelNone VerificationLevel = iota
	VerificationLevelLow
	VerificationLevelMedium
	VerificationLevelHigh
)

// ExplicitContentFilterLevel type definition
type ExplicitContentFilterLevel int

// Constants for ExplicitContentFilterLevel levels from 0 to 2 inclusive
const (
	ExplicitContentFilterDisabled ExplicitContentFilterLevel = iota
	ExplicitContentFilterMembersWithoutRoles
	ExplicitContentFilterAllMembers
)

// MfaLevel type definition
type MfaLevel int

// Constants for MfaLevel levels from 0 to 1 inclusive
const (
	MfaLevelNone MfaLevel = iota
	MfaLevelElevated
)

// PremiumTier type definition
type PremiumTier int

// Constants for PremiumTier levels from 0 to 3 inclusive
const (
	PremiumTierNone PremiumTier = 0
	PremiumTier1    PremiumTier = 1
	PremiumTier2    PremiumTier = 2
	PremiumTier3    PremiumTier = 3
)

// A Guild holds all data related to a specific Discord Guild.  Guilds are also
// sometimes referred to as Servers in the Discord client.
type Guild struct {
	// The ID of the guild.
	ID int64 `json:"id,string"`

	// The name of the guild. (2–100 characters)
	Name string `json:"name"`

	// The hash of the guild's icon. Use Session.GuildIcon
	// to retrieve the icon itself.
	Icon string `json:"icon"`

	// The voice region of the guild.
	Region string `json:"region"`

	// The ID of the AFK voice channel.
	AfkChannelID int64 `json:"afk_channel_id,string"`

	// The user ID of the owner of the guild.
	OwnerID int64 `json:"owner_id,string"`

	// If we are the owner of the guild
	Owner bool `json:"owner"`

	// The time at which the current user joined the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	JoinedAt Timestamp `json:"joined_at"`

	// The hash of the guild's discovery splash.
	DiscoverySplash string `json:"discovery_splash"`

	// The hash of the guild's splash.
	Splash string `json:"splash"`

	// The timeout, in seconds, before a user is considered AFK in voice.
	AfkTimeout int `json:"afk_timeout"`

	// The number of members in the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	MemberCount int `json:"member_count"`

	// The verification level required for the guild.
	VerificationLevel VerificationLevel `json:"verification_level"`

	// Whether the guild is considered large. This is
	// determined by a member threshold in the identify packet,
	// and is currently hard-coded at 250 members in the library.
	Large bool `json:"large"`

	// The default message notification setting for the guild.
	// 0 == all messages, 1 == mentions only.
	DefaultMessageNotifications int `json:"default_message_notifications"`

	// A list of roles in the guild.
	Roles []*Role `json:"roles"`

	// A list of the custom emojis present in the guild.
	Emojis []*Emoji `json:"emojis"`

	// A list of the custom stickers present in the guild.
	Stickers []*Sticker `json:"stickers"`

	// A list of the members in the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	Members []*Member `json:"members"`

	// A list of partial presence objects for members in the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	Presences []*Presence `json:"presences"`

	// The maximum number of presences for the guild (the default value, currently 25000, is in effect when null is returned)
	MaxPresences int `json:"max_presences"`

	// The maximum number of members for the guild
	MaxMembers int `json:"max_members"`

	// A list of channels in the guild.
	// This field is only present in GUILD_CREATE events
	Channels []*Channel `json:"channels"`

	// All active threads in the guild that current user has permission to view
	// This field is only present in GUILD_CREATE events
	Threads []*Channel `json:"threads"`

	// A list of voice states for the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	VoiceStates []*VoiceState `json:"voice_states"`

	// Whether this guild is currently unavailable (most likely due to outage).
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	Unavailable bool `json:"unavailable"`

	// The explicit content filter level
	ExplicitContentFilter ExplicitContentFilterLevel `json:"explicit_content_filter"`

	// The list of enabled guild features
	Features []string `json:"features"`

	// Required MFA level for the guild
	MfaLevel MfaLevel `json:"mfa_level"`

	// Whether or not the Server Widget is enabled
	WidgetEnabled bool `json:"widget_enabled"`

	// The Channel ID for the Server Widget
	WidgetChannelID string `json:"widget_channel_id"`

	// The Channel ID to which system messages are sent (eg join and leave messages)
	SystemChannelID string `json:"system_channel_id"`

	// The System channel flags
	SystemChannelFlags SystemChannelFlag `json:"system_channel_flags"`

	// The ID of the rules channel ID, used for rules.
	RulesChannelID string `json:"rules_channel_id"`

	// the vanity url code for the guild
	VanityURLCode string `json:"vanity_url_code"`

	// the description for the guild
	Description string `json:"description"`

	// The hash of the guild's banner
	Banner string `json:"banner"`

	// The premium tier of the guild
	PremiumTier PremiumTier `json:"premium_tier"`

	// The total number of users currently boosting this server
	PremiumSubscriptionCount int `json:"premium_subscription_count"`

	// The preferred locale of a guild with the "PUBLIC" feature; used in server discovery and notices from Discord; defaults to "en-US"
	PreferredLocale string `json:"preferred_locale"`

	// The id of the channel where admins and moderators of guilds with the "PUBLIC" feature receive notices from Discord
	PublicUpdatesChannelID string `json:"public_updates_channel_id"`

	// The maximum amount of users in a video channel
	MaxVideoChannelUsers int `json:"max_video_channel_users"`

	// Approximate number of members in this guild, returned from the GET /guild/<id> endpoint when with_counts is true
	ApproximateMemberCount int `json:"approximate_member_count"`

	// Approximate number of non-offline members in this guild, returned from the GET /guild/<id> endpoint when with_counts is true
	ApproximatePresenceCount int `json:"approximate_presence_count"`

	// Permissions of our user
	Permissions int64 `json:"permissions,string"`

	// Stage instances in the guild
	StageInstances []*StageInstance `json:"stage_instances"`
}

func (g *Guild) GetGuildID() int64 {
	return g.ID
}

func (g *Guild) Role(id int64) *Role {
	for _, v := range g.Roles {
		if v.ID == id {
			return v
		}
	}

	return nil
}

func (g *Guild) Channel(id int64) *Channel {
	for _, v := range g.Channels {
		if v.ID == id {
			return v
		}
	}

	return nil
}

// GuildScheduledEvent is a representation of a scheduled event in a guild. Only for retrieval of the data.
// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event
type GuildScheduledEvent struct {
	// The ID of the scheduled event
	ID int64 `json:"id,string"`
	// The guild id which the scheduled event belongs to
	GuildID int64 `json:"guild_id,string"`
	// The channel id in which the scheduled event will be hosted, or null if scheduled entity type is EXTERNAL
	ChannelID int64 `json:"channel_id,string"`
	// The id of the user that created the scheduled event
	CreatorID int64 `json:"creator_id,string"`
	// The name of the scheduled event (1-100 characters)
	Name string `json:"name"`
	// The description of the scheduled event (1-1000 characters)
	Description string `json:"description"`
	// The time the scheduled event will start
	ScheduledStartTime Timestamp `json:"scheduled_start_time"`
	// The time the scheduled event will end, required only when entity_type is EXTERNAL
	ScheduledEndTime *Timestamp `json:"scheduled_end_time"`
	// The privacy level of the scheduled event
	PrivacyLevel GuildScheduledEventPrivacyLevel `json:"privacy_level"`
	// The status of the scheduled event
	Status GuildScheduledEventStatus `json:"status"`
	// Type of the entity where event would be hosted
	// See field requirements
	// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-field-requirements-by-entity-type
	EntityType GuildScheduledEventEntityType `json:"entity_type"`
	// The id of an entity associated with a guild scheduled event
	EntityID string `json:"entity_id"`
	// Additional metadata for the guild scheduled event
	EntityMetadata GuildScheduledEventEntityMetadata `json:"entity_metadata"`
	// The user that created the scheduled event
	Creator *User `json:"creator"`
	// The number of users subscribed to the scheduled event
	UserCount int `json:"user_count"`
	// The cover image hash of the scheduled event
	// see https://discord.com/developers/docs/reference#image-formatting for more
	// information about image formatting
	Image string `json:"image"`
}

// GuildScheduledEventParams are the parameters allowed for creating or updating a scheduled event
// https://discord.com/developers/docs/resources/guild-scheduled-event#create-guild-scheduled-event
type GuildScheduledEventParams struct {
	// The channel id in which the scheduled event will be hosted, or null if scheduled entity type is EXTERNAL
	ChannelID int64 `json:"channel_id,string,omitempty"`
	// The name of the scheduled event (1-100 characters)
	Name string `json:"name,omitempty"`
	// The description of the scheduled event (1-1000 characters)
	Description string `json:"description,omitempty"`
	// The time the scheduled event will start
	ScheduledStartTime *Timestamp `json:"scheduled_start_time,omitempty"`
	// The time the scheduled event will end, required only when entity_type is EXTERNAL
	ScheduledEndTime *Timestamp `json:"scheduled_end_time,omitempty"`
	// The privacy level of the scheduled event
	PrivacyLevel GuildScheduledEventPrivacyLevel `json:"privacy_level,omitempty"`
	// The status of the scheduled event
	Status GuildScheduledEventStatus `json:"status,omitempty"`
	// Type of the entity where event would be hosted
	// See field requirements
	// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-field-requirements-by-entity-type
	EntityType GuildScheduledEventEntityType `json:"entity_type,omitempty"`
	// Additional metadata for the guild scheduled event
	EntityMetadata *GuildScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	// The cover image hash of the scheduled event
	// see https://discord.com/developers/docs/reference#image-formatting for more
	// information about image formatting
	Image string `json:"image,omitempty"`
}

/*
// MarshalJSON is a helper function to marshal GuildScheduledEventParams
func (p GuildScheduledEventParams) MarshalJSON() ([]byte, error) {
	type guildScheduledEventParams GuildScheduledEventParams

	if p.EntityType == GuildScheduledEventEntityTypeExternal && p.ChannelID == "" {
		return Marshal(struct {
			guildScheduledEventParams
			ChannelID json.RawMessage `json:"channel_id"`
		}{
			guildScheduledEventParams: guildScheduledEventParams(p),
			ChannelID:                 json.RawMessage("null"),
		})
	}

	return Marshal(guildScheduledEventParams(p))
}*/

// GuildScheduledEventEntityMetadata holds additional metadata for guild scheduled event.
type GuildScheduledEventEntityMetadata struct {
	// location of the event (1-100 characters)
	// required for events with 'entity_type': EXTERNAL
	Location string `json:"location"`
}

// GuildScheduledEventPrivacyLevel is the privacy level of a scheduled event.
// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-privacy-level
type GuildScheduledEventPrivacyLevel int

const (
	// GuildScheduledEventPrivacyLevelGuildOnly makes the scheduled
	// event is only accessible to guild members
	GuildScheduledEventPrivacyLevelGuildOnly GuildScheduledEventPrivacyLevel = 2
)

// GuildScheduledEventStatus is the status of a scheduled event
// Valid Guild Scheduled Event Status Transitions :
// SCHEDULED --> ACTIVE --> COMPLETED
// SCHEDULED --> CANCELED
// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-status
type GuildScheduledEventStatus int

const (
	// GuildScheduledEventStatusScheduled represents the current event is in scheduled state
	GuildScheduledEventStatusScheduled GuildScheduledEventStatus = 1
	// GuildScheduledEventStatusActive represents the current event is in active state
	GuildScheduledEventStatusActive GuildScheduledEventStatus = 2
	// GuildScheduledEventStatusCompleted represents the current event is in completed state
	GuildScheduledEventStatusCompleted GuildScheduledEventStatus = 3
	// GuildScheduledEventStatusCanceled represents the current event is in canceled state
	GuildScheduledEventStatusCanceled GuildScheduledEventStatus = 4
)

// GuildScheduledEventEntityType is the type of entity associated with a guild scheduled event.
// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-entity-types
type GuildScheduledEventEntityType int

const (
	// GuildScheduledEventEntityTypeStageInstance represents a stage channel
	GuildScheduledEventEntityTypeStageInstance GuildScheduledEventEntityType = 1
	// GuildScheduledEventEntityTypeVoice represents a voice channel
	GuildScheduledEventEntityTypeVoice GuildScheduledEventEntityType = 2
	// GuildScheduledEventEntityTypeExternal represents an external event
	GuildScheduledEventEntityTypeExternal GuildScheduledEventEntityType = 3
)

// GuildScheduledEventUser is a user subscribed to a scheduled event.
// https://discord.com/developers/docs/resources/guild-scheduled-event#guild-scheduled-event-user-object
type GuildScheduledEventUser struct {
	GuildScheduledEventID string  `json:"guild_scheduled_event_id"`
	User                  *User   `json:"user"`
	Member                *Member `json:"member"`
}

// A GuildTemplate represents
type GuildTemplate struct {
	// The unique code for the guild template
	Code string `json:"code"`

	// The name of the template
	Name string `json:"name"`

	// The description for the template
	Description string `json:"description"`

	// The number of times this template has been used
	UsageCount string `json:"usage_count"`

	// The ID of the user who created the template
	CreatorID int64 `json:"creator_id,string"`

	// The user who created the template
	Creator *User `json:"creator"`

	// The timestamp of when the template was created
	CreatedAt Timestamp `json:"created_at"`

	// The timestamp of when the template was last synced
	UpdatedAt Timestamp `json:"updated_at"`

	// The ID of the guild the template was based on
	SourceGuildID int64 `json:"source_guild_id,string"`

	// The guild 'snapshot' this template contains
	SerializedSourceGuild *Guild `json:"serialized_source_guild"`

	// Whether the template has unsynced changes
	IsDirty bool `json:"is_dirty"`
}

// GuildTemplateParams stores the data needed to create or update a GuildTemplate.
type GuildTemplateParams struct {
	// The name of the template (1-100 characters)
	Name string `json:"name,omitempty"`
	// The description of the template (0-120 characters)
	Description string `json:"description,omitempty"`
}

// SystemChannelFlag is the type of flags in the system channel (see SystemChannelFlag* consts)
// https://discord.com/developers/docs/resources/guild#guild-object-system-channel-flags
type SystemChannelFlag int

// Block containing known SystemChannelFlag values
const (
	SystemChannelFlagsSuppressJoinNotifications          SystemChannelFlag = 1 << 0
	SystemChannelFlagsSuppressPremium                    SystemChannelFlag = 1 << 1
	SystemChannelFlagsSuppressGuildReminderNotifications SystemChannelFlag = 1 << 2
	SystemChannelFlagsSuppressJoinNotificationReplies    SystemChannelFlag = 1 << 3
)

/*const (
	SystemChannelFlagsSuppressJoin         SystemChannelFlag = 1 << 0
	SystemChannelFlagsSuppressPremium      SystemChannelFlag = 1 << 1
	SystemChannelFlagsSupressGuildReminder SystemChannelFlag = 1 << 2
	SystemChannelFlagsSupressJoinReplies   SystemChannelFlag = 1 << 3
)*/

// A UserGuild holds a brief version of a Guild
type UserGuild struct {
	ID          int64          `json:"id,string"`
	Name        string         `json:"name"`
	Icon        string         `json:"icon"`
	Owner       bool           `json:"owner"`
	Permissions int64          `json:"permissions,string"`
	Features    []GuildFeature `json:"features"`
}

// IconURL returns a URL to the guild's icon.
//
//	size:    The size of the desired icon image as a power of two
//	         Image size can be any power of two between 16 and 4096.
func (g *Guild) IconURL(size string) string {
	return iconURL(g.Icon, EndpointGuildIcon(g.ID, g.Icon), EndpointGuildIconAnimated(g.ID, g.Icon), size)
	/*if g.Icon == "" {
		return ""
	}

	if strings.HasPrefix(g.Icon, "a_") {
		return EndpointGuildIconAnimated(g.ID, g.Icon)
	}

	return EndpointGuildIcon(g.ID, g.Icon)*/
}

// BannerURL returns a URL to the guild's banner.
//
//	size:    The size of the desired banner image as a power of two
//	         Image size can be any power of two between 16 and 4096.
func (g *Guild) BannerURL(size string) string {
	return bannerURL(g.Banner, EndpointGuildBanner(g.ID, g.Banner), EndpointGuildBannerAnimated(g.ID, g.Banner), size)
	/*if g.Banner == "" {
		return ""
	}
	return EndpointGuildBanner(g.ID, g.Banner)*/
}

// A Guild feature indicates the presence of a feature in a guild
type GuildFeature string

// Constants for GuildFeature
const (
	GuildFeatureAnimatedBanner                GuildFeature = "ANIMATED_BANNER"
	GuildFeatureAnimatedIcon                  GuildFeature = "ANIMATED_ICON"
	GuildFeatureAutoModeration                GuildFeature = "AUTO_MODERATION"
	GuildFeatureBanner                        GuildFeature = "BANNER"
	GuildFeatureCommunity                     GuildFeature = "COMMUNITY"
	GuildFeatureDiscoverable                  GuildFeature = "DISCOVERABLE"
	GuildFeatureFeaturable                    GuildFeature = "FEATURABLE"
	GuildFeatureInviteSplash                  GuildFeature = "INVITE_SPLASH"
	GuildFeatureMemberVerificationGateEnabled GuildFeature = "MEMBER_VERIFICATION_GATE_ENABLED"
	GuildFeatureMonetizationEnabled           GuildFeature = "MONETIZATION_ENABLED"
	GuildFeatureMoreStickers                  GuildFeature = "MORE_STICKERS"
	GuildFeatureNews                          GuildFeature = "NEWS"
	GuildFeaturePartnered                     GuildFeature = "PARTNERED"
	GuildFeaturePreviewEnabled                GuildFeature = "PREVIEW_ENABLED"
	GuildFeaturePrivateThreads                GuildFeature = "PRIVATE_THREADS"
	GuildFeatureRoleIcons                     GuildFeature = "ROLE_ICONS"
	GuildFeatureTicketedEventsEnabled         GuildFeature = "TICKETED_EVENTS_ENABLED"
	GuildFeatureVanityUrl                     GuildFeature = "VANITY_URL"
	GuildFeatureVerified                      GuildFeature = "VERIFIED"
	GuildFeatureVipRegions                    GuildFeature = "VIP_REGIONS"
	GuildFeatureWelcomeScreenEnabled          GuildFeature = "WELCOME_SCREEN_ENABLED"
)

// A GuildParams stores all the data needed to update discord guild settings
type GuildParams struct {
	Name                        string             `json:"name,omitempty"`
	Region                      string             `json:"region,omitempty"`
	VerificationLevel           *VerificationLevel `json:"verification_level,omitempty"`
	DefaultMessageNotifications int                `json:"default_message_notifications,omitempty"` // TODO: Separate type?
	ExplicitContentFilter       int                `json:"explicit_content_filter,omitempty"`
	AfkChannelID                string             `json:"afk_channel_id,omitempty"`
	AfkTimeout                  int                `json:"afk_timeout,omitempty"`
	Icon                        string             `json:"icon,omitempty"`
	OwnerID                     string             `json:"owner_id,omitempty"`
	Splash                      string             `json:"splash,omitempty"`
	DiscoverySplash             string             `json:"discovery_splash,omitempty"`
	Banner                      string             `json:"banner,omitempty"`
	SystemChannelID             string             `json:"system_channel_id,omitempty"`
	SystemChannelFlags          SystemChannelFlag  `json:"system_channel_flags,omitempty"`
	RulesChannelID              string             `json:"rules_channel_id,omitempty"`
	PublicUpdatesChannelID      string             `json:"public_updates_channel_id,omitempty"`
	PreferredLocale             Locale             `json:"preferred_locale,omitempty"`
	Features                    []GuildFeature     `json:"features,omitempty"`
	Description                 string             `json:"description,omitempty"`
	PremiumProgressBarEnabled   *bool              `json:"premium_progress_bar_enabled,omitempty"`
}

// A Role stores information about Discord guild member roles.
type Role struct {
	// The ID of the role.
	ID int64 `json:"id,string"`

	// The name of the role.
	Name string `json:"name"`

	// Whether this role is managed by an integration, and
	// thus cannot be manually added to, or taken from, members.
	Managed bool `json:"managed"`

	// Whether this role is mentionable.
	Mentionable bool `json:"mentionable"`

	// Whether this role is hoisted (shows up separately in member list).
	Hoist bool `json:"hoist"`

	// The hex color of this role.
	Color int `json:"color"`

	// The position of this role in the guild's role hierarchy.
	Position int `json:"position"`

	// The permissions of the role on the guild (doesn't include channel overrides).
	// This is a combination of bit masks; the presence of a certain permission can
	// be checked by performing a bitwise AND between this int and the permission.
	Permissions int64 `json:"permissions,string"`
}

// Mention returns a string which mentions the role
func (r *Role) Mention() string {
	if r == nil {
		return "No such role"
	}
	return fmt.Sprintf("<@&%d>", r.ID)
}

// RoleParams represents the parameters needed to create or update a Role
type RoleParams struct {
	// The role's name
	Name string `json:"name,omitempty"`
	// The color the role should have (as a decimal, not hex)
	Color *int `json:"color,omitempty"`
	// Whether to display the role's users separately
	Hoist *bool `json:"hoist,omitempty"`
	// The overall permissions number of the role
	Permissions *int64 `json:"permissions,omitempty,string"`
	// Whether this role is mentionable
	Mentionable *bool `json:"mentionable,omitempty"`
}

// Roles are a collection of Role
type Roles []*Role

func (r Roles) Len() int {
	return len(r)
}

func (r Roles) Less(i, j int) bool {
	return r[i].Position > r[j].Position
}

func (r Roles) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// A VoiceState stores the voice states of Guilds
type VoiceState struct {
	UserID                  int64      `json:"user_id,string"`
	SessionID               string     `json:"session_id"`
	ChannelID               int64      `json:"channel_id,string"`
	GuildID                 int64      `json:"guild_id,string"`
	Suppress                bool       `json:"suppress"`
	SelfMute                bool       `json:"self_mute"`
	SelfDeaf                bool       `json:"self_deaf"`
	Mute                    bool       `json:"mute"`
	Member                  *Member    `json:"member"`
	Deaf                    bool       `json:"deaf"`
	SelfStream              bool       `json:"self_stream"`
	SelfVideo               bool       `json:"self_video"`
	RequestToSpeakTimestamp *time.Time `json:"request_to_speak_timestamp"`
}

// A Presence stores the online, offline, or idle and game status of Guild members.
type Presence struct {
	User   *User  `json:"user"`
	Status Status `json:"status"`

	Activities   Activities   `json:"activities"`
	Since        *int         `json:"since"`
	ClientStatus ClientStatus `json:"client_status"`
}

// implement gojay.UnmarshalerJSONObject
func (p *Presence) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	var err error
	switch key {
	case "user":
		p.User = &User{}
		err = dec.Object(p.User)
	case "status":
		err = dec.String((*string)(&p.Status))
	case "activities":
		err = dec.DecodeArray(&p.Activities)
	default:
	}

	if err != nil {
		return errors.Wrap(err, key)
	}

	return nil
}

func (p *Presence) NKeys() int {
	return 0
}

// A TimeStamps struct contains start and end times used in the rich presence "playing .." Game
type TimeStamps struct {
	EndTimestamp   int64 `json:"end,omitempty"`
	StartTimestamp int64 `json:"start,omitempty"`
}

func (t *TimeStamps) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "start":
		return dec.Int64(&t.StartTimestamp)
	case "end":
		return dec.Int64(&t.EndTimestamp)
	}

	return nil
}

// UnmarshalJSON unmarshals JSON into TimeStamps struct
func (t *TimeStamps) UnmarshalJSON(b []byte) error {
	temp := struct {
		End   json.Number `json:"end,omitempty"`
		Start json.Number `json:"start,omitempty"`
	}{}
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	var endParsed float64
	if temp.End != "" {
		endParsed, err = temp.End.Float64()
		if err != nil {
			return err
		}
	}

	var startParsed float64
	if temp.Start != "" {
		startParsed, err = temp.Start.Float64()
		if err != nil {
			return err
		}
	}

	t.EndTimestamp = int64(endParsed)
	t.StartTimestamp = int64(startParsed)
	return nil
}

func (t *TimeStamps) NKeys() int {
	return 0
}

// An Assets struct contains assets and labels used in the rich presence "playing .." Game
type Assets struct {
	LargeImageID string `json:"large_image,omitempty"`
	SmallImageID string `json:"small_image,omitempty"`
	LargeText    string `json:"large_text,omitempty"`
	SmallText    string `json:"small_text,omitempty"`
}

// A Member stores user information for Guild members. A guild
// member represents a certain user's presence in a guild.
type Member struct {
	// The guild ID on which the member exists.
	GuildID int64 `json:"guild_id,string"`

	// The time at which the member joined the guild, in ISO8601.
	JoinedAt Timestamp `json:"joined_at"`

	// The nickname of the member, if they have one.
	Nick string `json:"nick"`

	// The guild avatar hash of the member, if they have one.
	Avatar string `json:"avatar"`

	// Whether the member is deafened at a guild level.
	Deaf bool `json:"deaf"`

	// Whether the member is muted at a guild level.
	Mute bool `json:"mute"`

	// The underlying user on which the member is based.
	User *User `json:"user"`

	// A list of IDs of the roles which are possessed by the member.
	Roles IDSlice `json:"roles"`

	// Is true while the member hasn't accepted the membership screen.
	Pending bool `json:"pending"`

	// When the user used their Nitro boost on the server
	PremiumSince *time.Time `json:"premium_since"`

	// Total permissions of the member in the channel, including overrides, returned when in the interaction object.
	Permissions int64 `json:"permissions,string"`

	// The time at which the member's timeout will expire.
	// Time in the past or nil if the user is not timed out.
	TimeoutExpiresAt *time.Time `json:"communication_disabled_until"`
}

func (m *Member) GetGuildID() int64 {
	return m.GuildID
}

// AvatarURL returns the URL of the member's avatar
//
//	size:    The size of the user's avatar as a power of two
//	         if size is an empty string, no size parameter will
//	         be added to the URL.
func (m *Member) AvatarURL(size string) string {
	var URL string

	if m == nil {
		return "Member not found"
	}

	u := m.User

	if m.Avatar == "" {
		return u.AvatarURL(size)
	} else if strings.HasPrefix(m.Avatar, "a_") {
		URL = EndpointGuildMemberAvatarAnimated(m.GuildID, u.ID, m.Avatar)
	} else {
		URL = EndpointGuildMemberAvatar(m.GuildID, u.ID, m.Avatar)
	}

	if size != "" {
		return URL + "?size=" + size
	}
	return URL
}

// ClientStatus stores the online, offline, idle, or dnd status of each device of a Guild member.
type ClientStatus struct {
	Desktop Status `json:"desktop"`
	Mobile  Status `json:"mobile"`
	Web     Status `json:"web"`
}

// Status type definition
type Status string

// Constants for Status with the different current available status
const (
	StatusOnline       Status = "online"
	StatusIdle         Status = "idle"
	StatusDoNotDisturb Status = "dnd"
	StatusInvisible    Status = "invisible"
	StatusOffline      Status = "offline"
)

// A TooManyRequests struct holds information received from Discord
// when receiving a HTTP 429 response.
type TooManyRequests struct {
	Bucket     string  `json:"bucket"`
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

func (t *TooManyRequests) RetryAfterDur() time.Duration {
	return time.Duration(t.RetryAfter*1000) * time.Millisecond
}

// A ReadState stores data on the read state of channels.
type ReadState struct {
	MentionCount  int   `json:"mention_count"`
	LastMessageID int64 `json:"last_message_id,string"`
	ID            int64 `json:"id,string"`
}

// A GuildRole stores data for guild roles.
type GuildRole struct {
	Role    *Role `json:"role"`
	GuildID int64 `json:"guild_id,string"`
}

func (e *GuildRole) GetGuildID() int64 {
	return e.GuildID
}

// A GuildBan stores data for a guild ban.
type GuildBan struct {
	Reason string `json:"reason"`
	User   *User  `json:"user"`
}

// AutoModerationRule stores data for an auto moderation rule.
type AutoModerationRule struct {
	ID              int64                          `json:"id,omitempty,string"`
	GuildID         int64                          `json:"guild_id,omitempty,string"`
	Name            string                         `json:"name,omitempty"`
	CreatorID       int64                          `json:"creator_id,omitempty,string"`
	EventType       AutoModerationRuleEventType    `json:"event_type,omitempty"`
	TriggerType     AutoModerationRuleTriggerType  `json:"trigger_type,omitempty"`
	TriggerMetadata *AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []AutoModerationAction         `json:"actions,omitempty"`
	Enabled         *bool                          `json:"enabled,omitempty"`
	ExemptRoles     *[]string                      `json:"exempt_roles,omitempty"`
	ExemptChannels  *[]string                      `json:"exempt_channels,omitempty"`
}

// AutoModerationRuleEventType indicates in what event context a rule should be checked.
type AutoModerationRuleEventType int

// Auto moderation rule event types.
const (
	// AutoModerationEventMessageSend is checked when a member sends or edits a message in the guild
	AutoModerationEventMessageSend AutoModerationRuleEventType = 1
)

// AutoModerationRuleTriggerType represents the type of content which can trigger the rule.
type AutoModerationRuleTriggerType int

// Auto moderation rule trigger types.
const (
	AutoModerationEventTriggerKeyword       AutoModerationRuleTriggerType = 1
	AutoModerationEventTriggerHarmfulLink   AutoModerationRuleTriggerType = 2
	AutoModerationEventTriggerSpam          AutoModerationRuleTriggerType = 3
	AutoModerationEventTriggerKeywordPreset AutoModerationRuleTriggerType = 4
)

// AutoModerationKeywordPreset represents an internally pre-defined wordset.
type AutoModerationKeywordPreset uint

// Auto moderation keyword presets.
const (
	AutoModerationKeywordPresetProfanity     AutoModerationKeywordPreset = 1
	AutoModerationKeywordPresetSexualContent AutoModerationKeywordPreset = 2
	AutoModerationKeywordPresetSlurs         AutoModerationKeywordPreset = 3
)

// AutoModerationTriggerMetadata represents additional metadata used to determine whether rule should be triggered.
type AutoModerationTriggerMetadata struct {
	// Substrings which will be searched for in content.
	// NOTE: should be only used with keyword trigger type.
	KeywordFilter []string `json:"keyword_filter,omitempty"`

	// Regular expression patterns which will be matched against content (maximum of 10).
	// NOTE: should be only used with keyword trigger type.
	RegexPatterns []string `json:"regex_patterns,omitempty"`

	// Internally pre-defined wordsets which will be searched for in content.
	// NOTE: should be only used with keyword preset trigger type.
	Presets []AutoModerationKeywordPreset `json:"presets,omitempty"`

	// Substrings which should not trigger the rule.
	// NOTE: should be only used with keyword or keyword preset trigger type.
	AllowList *[]string `json:"allow_list,omitempty"`

	// Total number of unique role and user mentions allowed per message.
	// NOTE: should be only used with mention spam trigger type.
	MentionTotalLimit int `json:"mention_total_limit,omitempty"`
}

// AutoModerationActionType represents an action which will execute whenever a rule is triggered.
type AutoModerationActionType int

// Auto moderation actions types.
const (
	AutoModerationRuleActionBlockMessage     AutoModerationActionType = 1
	AutoModerationRuleActionSendAlertMessage AutoModerationActionType = 2
	AutoModerationRuleActionTimeout          AutoModerationActionType = 3
)

// AutoModerationActionMetadata represents additional metadata needed during execution for a specific action type.
type AutoModerationActionMetadata struct {
	// Channel to which user content should be logged.
	// NOTE: should be only used with send alert message action type.
	ChannelID string `json:"channel_id,omitempty"`

	// Timeout duration in seconds (maximum of 2419200 - 4 weeks).
	// NOTE: should be only used with timeout action type.
	Duration int `json:"duration_seconds,omitempty"`
}

// AutoModerationAction stores data for an auto moderation action.
type AutoModerationAction struct {
	Type     AutoModerationActionType      `json:"type"`
	Metadata *AutoModerationActionMetadata `json:"metadata,omitempty"`
}

// A GuildEmbed stores data for a guild embed.
type GuildEmbed struct {
	Enabled   bool  `json:"enabled,omitempty"`
	ChannelID int64 `json:"channel_id,string,omitempty"`
}

// A GuildAuditLog stores data for a guild audit log.
// https://discord.com/developers/docs/resources/audit-log#audit-log-object-audit-log-structure
type GuildAuditLog struct {
	Webhooks        []*Webhook       `json:"webhooks,omitempty"`
	Users           []*User          `json:"users,omitempty"`
	AuditLogEntries []*AuditLogEntry `json:"audit_log_entries"`
	Integrations    []*Integration   `json:"integrations"`
}

// AuditLogEntry for a GuildAuditLog
// https://discord.com/developers/docs/resources/audit-log#audit-log-entry-object-audit-log-entry-structure
type AuditLogEntry struct {
	TargetID int64             `json:"target_id,string"`
	Changes  []*AuditLogChange `json:"changes"`
	/*Changes []struct {
		NewValue interface{} `json:"new_value"`
		OldValue interface{} `json:"old_value"`
		Key      string      `json:"key"`
	} `json:"changes,omitempty"`*/
	UserID     int64            `json:"user_id,string"`
	ID         int64            `json:"id,string"`
	ActionType *AuditLogAction  `json:"action_type"`
	Options    *AuditLogOptions `json:"options"`
	Reason     string           `json:"reason"`
}

// AuditLogChange for an AuditLogEntry
type AuditLogChange struct {
	NewValue interface{}        `json:"new_value"`
	OldValue interface{}        `json:"old_value"`
	Key      *AuditLogChangeKey `json:"key"`
}

// AuditLogChangeKey value for AuditLogChange
// https://discord.com/developers/docs/resources/audit-log#audit-log-change-object-audit-log-change-key
type AuditLogChangeKey string

// Block of valid AuditLogChangeKey
const (
	// AuditLogChangeKeyAfkChannelID is sent when afk channel changed (snowflake) - guild
	AuditLogChangeKeyAfkChannelID AuditLogChangeKey = "afk_channel_id"
	// AuditLogChangeKeyAfkTimeout is sent when afk timeout duration changed (int) - guild
	AuditLogChangeKeyAfkTimeout AuditLogChangeKey = "afk_timeout"
	// AuditLogChangeKeyAllow is sent when a permission on a text or voice channel was allowed for a role (string) - role
	AuditLogChangeKeyAllow AuditLogChangeKey = "allow"
	// AudirChangeKeyApplicationID is sent when application id of the added or removed webhook or bot (snowflake) - channel
	AuditLogChangeKeyApplicationID AuditLogChangeKey = "application_id"
	// AuditLogChangeKeyArchived is sent when thread was archived/unarchived (bool) - thread
	AuditLogChangeKeyArchived AuditLogChangeKey = "archived"
	// AuditLogChangeKeyAsset is sent when asset is changed (string) - sticker
	AuditLogChangeKeyAsset AuditLogChangeKey = "asset"
	// AuditLogChangeKeyAutoArchiveDuration is sent when auto archive duration changed (int) - thread
	AuditLogChangeKeyAutoArchiveDuration AuditLogChangeKey = "auto_archive_duration"
	// AuditLogChangeKeyAvailable is sent when availability of sticker changed (bool) - sticker
	AuditLogChangeKeyAvailable AuditLogChangeKey = "available"
	// AuditLogChangeKeyAvatarHash is sent when user avatar changed (string) - user
	AuditLogChangeKeyAvatarHash AuditLogChangeKey = "avatar_hash"
	// AuditLogChangeKeyBannerHash is sent when guild banner changed (string) - guild
	AuditLogChangeKeyBannerHash AuditLogChangeKey = "banner_hash"
	// AuditLogChangeKeyBitrate is sent when voice channel bitrate changed (int) - channel
	AuditLogChangeKeyBitrate AuditLogChangeKey = "bitrate"
	// AuditLogChangeKeyChannelID is sent when channel for invite code or guild scheduled event changed (snowflake) - invite or guild scheduled event
	AuditLogChangeKeyChannelID AuditLogChangeKey = "channel_id"
	// AuditLogChangeKeyCode is sent when invite code changed (string) - invite
	AuditLogChangeKeyCode AuditLogChangeKey = "code"
	// AuditLogChangeKeyColor is sent when role color changed (int) - role
	AuditLogChangeKeyColor AuditLogChangeKey = "color"
	// AuditLogChangeKeyCommunicationDisabledUntil is sent when member timeout state changed (ISO8601 timestamp) - member
	AuditLogChangeKeyCommunicationDisabledUntil AuditLogChangeKey = "communication_disabled_until"
	// AuditLogChangeKeyDeaf is sent when user server deafened/undeafened (bool) - member
	AuditLogChangeKeyDeaf AuditLogChangeKey = "deaf"
	// AuditLogChangeKeyDefaultAutoArchiveDuration is sent when default auto archive duration for newly created threads changed (int) - channel
	AuditLogChangeKeyDefaultAutoArchiveDuration AuditLogChangeKey = "default_auto_archive_duration"
	// AuditLogChangeKeyDefaultMessageNotification is sent when default message notification level changed (int) - guild
	AuditLogChangeKeyDefaultMessageNotification AuditLogChangeKey = "default_message_notifications"
	// AuditLogChangeKeyDeny is sent when a permission on a text or voice channel was denied for a role (string) - role
	AuditLogChangeKeyDeny AuditLogChangeKey = "deny"
	// AuditLogChangeKeyDescription is sent when description changed (string) - guild, sticker, or guild scheduled event
	AuditLogChangeKeyDescription AuditLogChangeKey = "description"
	// AuditLogChangeKeyDiscoverySplashHash is sent when discovery splash changed (string) - guild
	AuditLogChangeKeyDiscoverySplashHash AuditLogChangeKey = "discovery_splash_hash"
	// AuditLogChangeKeyEnableEmoticons is sent when integration emoticons enabled/disabled (bool) - integration
	AuditLogChangeKeyEnableEmoticons AuditLogChangeKey = "enable_emoticons"
	// AuditLogChangeKeyEntityType is sent when entity type of guild scheduled event was changed (int) - guild scheduled event
	AuditLogChangeKeyEntityType AuditLogChangeKey = "entity_type"
	// AuditLogChangeKeyExpireBehavior is sent when integration expiring subscriber behavior changed (int) - integration
	AuditLogChangeKeyExpireBehavior AuditLogChangeKey = "expire_behavior"
	// AuditLogChangeKeyExpireGracePeriod is sent when integration expire grace period changed (int) - integration
	AuditLogChangeKeyExpireGracePeriod AuditLogChangeKey = "expire_grace_period"
	// AuditLogChangeKeyExplicitContentFilter is sent when change in whose messages are scanned and deleted for explicit content in the server is made (int) - guild
	AuditLogChangeKeyExplicitContentFilter AuditLogChangeKey = "explicit_content_filter"
	// AuditLogChangeKeyFormatType is sent when format type of sticker changed (int - sticker format type) - sticker
	AuditLogChangeKeyFormatType AuditLogChangeKey = "format_type"
	// AuditLogChangeKeyGuildID is sent when guild sticker is in changed (snowflake) - sticker
	AuditLogChangeKeyGuildID AuditLogChangeKey = "guild_id"
	// AuditLogChangeKeyHoist is sent when role is now displayed/no longer displayed separate from online users (bool) - role
	AuditLogChangeKeyHoist AuditLogChangeKey = "hoist"
	// AuditLogChangeKeyIconHash is sent when icon changed (string) - guild or role
	AuditLogChangeKeyIconHash AuditLogChangeKey = "icon_hash"
	// AuditLogChangeKeyID is sent when the id of the changed entity - sometimes used in conjunction with other keys (snowflake) - any
	AuditLogChangeKeyID AuditLogChangeKey = "id"
	// AuditLogChangeKeyInvitable is sent when private thread is now invitable/uninvitable (bool) - thread
	AuditLogChangeKeyInvitable AuditLogChangeKey = "invitable"
	// AuditLogChangeKeyInviterID is sent when person who created invite code changed (snowflake) - invite
	AuditLogChangeKeyInviterID AuditLogChangeKey = "inviter_id"
	// AuditLogChangeKeyLocation is sent when channel id for guild scheduled event changed (string) - guild scheduled event
	AuditLogChangeKeyLocation AuditLogChangeKey = "location"
	// AuditLogChangeKeyLocked is sent when thread was locked/unlocked (bool) - thread
	AuditLogChangeKeyLocked AuditLogChangeKey = "locked"
	// AuditLogChangeKeyMaxAge is sent when invite code expiration time changed (int) - invite
	AuditLogChangeKeyMaxAge AuditLogChangeKey = "max_age"
	// AuditLogChangeKeyMaxUses is sent when max number of times invite code can be used changed (int) - invite
	AuditLogChangeKeyMaxUses AuditLogChangeKey = "max_uses"
	// AuditLogChangeKeyMentionable is sent when role is now mentionable/unmentionable (bool) - role
	AuditLogChangeKeyMentionable AuditLogChangeKey = "mentionable"
	// AuditLogChangeKeyMfaLevel is sent when two-factor auth requirement changed (int - mfa level) - guild
	AuditLogChangeKeyMfaLevel AuditLogChangeKey = "mfa_level"
	// AuditLogChangeKeyMute is sent when user server muted/unmuted (bool) - member
	AuditLogChangeKeyMute AuditLogChangeKey = "mute"
	// AuditLogChangeKeyName is sent when name changed (string) - any
	AuditLogChangeKeyName AuditLogChangeKey = "name"
	// AuditLogChangeKeyNick is sent when user nickname changed (string) - member
	AuditLogChangeKeyNick AuditLogChangeKey = "nick"
	// AuditLogChangeKeyNSFW is sent when channel nsfw restriction changed (bool) - channel
	AuditLogChangeKeyNSFW AuditLogChangeKey = "nsfw"
	// AuditLogChangeKeyOwnerID is sent when owner changed (snowflake) - guild
	AuditLogChangeKeyOwnerID AuditLogChangeKey = "owner_id"
	// AuditLogChangeKeyPermissionOverwrite is sent when permissions on a channel changed (array of channel overwrite objects) - channel
	AuditLogChangeKeyPermissionOverwrite AuditLogChangeKey = "permission_overwrites"
	// AuditLogChangeKeyPermissions is sent when permissions for a role changed (string) - role
	AuditLogChangeKeyPermissions AuditLogChangeKey = "permissions"
	// AuditLogChangeKeyPosition is sent when text or voice channel position changed (int) - channel
	AuditLogChangeKeyPosition AuditLogChangeKey = "position"
	// AuditLogChangeKeyPreferredLocale is sent when preferred locale changed (string) - guild
	AuditLogChangeKeyPreferredLocale AuditLogChangeKey = "preferred_locale"
	// AuditLogChangeKeyPrivacylevel is sent when privacy level of the stage instance changed (integer - privacy level) - stage instance or guild scheduled event
	AuditLogChangeKeyPrivacylevel AuditLogChangeKey = "privacy_level"
	// AuditLogChangeKeyPruneDeleteDays is sent when number of days after which inactive and role-unassigned members are kicked changed (int) - guild
	AuditLogChangeKeyPruneDeleteDays AuditLogChangeKey = "prune_delete_days"
	// AuditLogChangeKeyPulibUpdatesChannelID is sent when id of the public updates channel changed (snowflake) - guild
	AuditLogChangeKeyPulibUpdatesChannelID AuditLogChangeKey = "public_updates_channel_id"
	// AuditLogChangeKeyRateLimitPerUser is sent when amount of seconds a user has to wait before sending another message changed (int) - channel
	AuditLogChangeKeyRateLimitPerUser AuditLogChangeKey = "rate_limit_per_user"
	// AuditLogChangeKeyRegion is sent when region changed (string) - guild
	AuditLogChangeKeyRegion AuditLogChangeKey = "region"
	// AuditLogChangeKeyRulesChannelID is sent when id of the rules channel changed (snowflake) - guild
	AuditLogChangeKeyRulesChannelID AuditLogChangeKey = "rules_channel_id"
	// AuditLogChangeKeySplashHash is sent when invite splash page artwork changed (string) - guild
	AuditLogChangeKeySplashHash AuditLogChangeKey = "splash_hash"
	// AuditLogChangeKeyStatus is sent when status of guild scheduled event was changed (int - guild scheduled event status) - guild scheduled event
	AuditLogChangeKeyStatus AuditLogChangeKey = "status"
	// AuditLogChangeKeySystemChannelID is sent when id of the system channel changed (snowflake) - guild
	AuditLogChangeKeySystemChannelID AuditLogChangeKey = "system_channel_id"
	// AuditLogChangeKeyTags is sent when related emoji of sticker changed (string) - sticker
	AuditLogChangeKeyTags AuditLogChangeKey = "tags"
	// AuditLogChangeKeyTemporary is sent when invite code is now temporary or never expires (bool) - invite
	AuditLogChangeKeyTemporary AuditLogChangeKey = "temporary"
	// TODO: remove when compatibility is not required
	AuditLogChangeKeyTempoary = AuditLogChangeKeyTemporary
	// AuditLogChangeKeyTopic is sent when text channel topic or stage instance topic changed (string) - channel or stage instance
	AuditLogChangeKeyTopic AuditLogChangeKey = "topic"
	// AuditLogChangeKeyType is sent when type of entity created (int or string) - any
	AuditLogChangeKeyType AuditLogChangeKey = "type"
	// AuditLogChangeKeyUnicodeEmoji is sent when role unicode emoji changed (string) - role
	AuditLogChangeKeyUnicodeEmoji AuditLogChangeKey = "unicode_emoji"
	// AuditLogChangeKeyUserLimit is sent when new user limit in a voice channel set (int) - voice channel
	AuditLogChangeKeyUserLimit AuditLogChangeKey = "user_limit"
	// AuditLogChangeKeyUses is sent when number of times invite code used changed (int) - invite
	AuditLogChangeKeyUses AuditLogChangeKey = "uses"
	// AuditLogChangeKeyVanityURLCode is sent when guild invite vanity url changed (string) - guild
	AuditLogChangeKeyVanityURLCode AuditLogChangeKey = "vanity_url_code"
	// AuditLogChangeKeyVerificationLevel is sent when required verification level changed (int - verification level) - guild
	AuditLogChangeKeyVerificationLevel AuditLogChangeKey = "verification_level"
	// AuditLogChangeKeyWidgetChannelID is sent when channel id of the server widget changed (snowflake) - guild
	AuditLogChangeKeyWidgetChannelID AuditLogChangeKey = "widget_channel_id"
	// AuditLogChangeKeyWidgetEnabled is sent when server widget enabled/disabled (bool) - guild
	AuditLogChangeKeyWidgetEnabled AuditLogChangeKey = "widget_enabled"
	// AuditLogChangeKeyRoleAdd is sent when new role added (array of partial role objects) - guild
	AuditLogChangeKeyRoleAdd AuditLogChangeKey = "$add"
	// AuditLogChangeKeyRoleRemove is sent when role removed (array of partial role objects) - guild
	AuditLogChangeKeyRoleRemove AuditLogChangeKey = "$remove"
)

// AuditLogOptions optional data for the AuditLog
// https://discord.com/developers/docs/resources/audit-log#audit-log-entry-object-optional-audit-entry-info
type AuditLogOptions struct {
	DeleteMemberDays string               `json:"delete_member_days"`
	MembersRemoved   string               `json:"members_removed"`
	ChannelID        int64                `json:"channel_id,string"`
	MessageID        int64                `json:"message_id,string"`
	Count            string               `json:"count"`
	ID               int64                `json:"id,string"`
	Type             *AuditLogOptionsType `json:"type"`
	RoleName         string               `json:"role_name"`
}

// AuditLogOptionsType of the AuditLogOption
// https://discord.com/developers/docs/resources/audit-log#audit-log-entry-object-optional-audit-entry-info
type AuditLogOptionsType string

// Valid Types for AuditLogOptionsType
const (
	AuditLogOptionsTypeMember AuditLogOptionsType = "member"
	AuditLogOptionsTypeRole   AuditLogOptionsType = "role"
)

// AuditLogAction is the Action of the AuditLog (see AuditLogAction* consts)
// https://discord.com/developers/docs/resources/audit-log#audit-log-entry-object-audit-log-events
type AuditLogAction int

// Block contains Discord Audit Log Action Types
const (
	AuditLogActionGuildUpdate AuditLogAction = 1

	AuditLogActionChannelCreate          AuditLogAction = 10
	AuditLogActionChannelUpdate          AuditLogAction = 11
	AuditLogActionChannelDelete          AuditLogAction = 12
	AuditLogActionChannelOverwriteCreate AuditLogAction = 13
	AuditLogActionChannelOverwriteUpdate AuditLogAction = 14
	AuditLogActionChannelOverwriteDelete AuditLogAction = 15

	AuditLogActionMemberKick       AuditLogAction = 20
	AuditLogActionMemberPrune      AuditLogAction = 21
	AuditLogActionMemberBanAdd     AuditLogAction = 22
	AuditLogActionMemberBanRemove  AuditLogAction = 23
	AuditLogActionMemberUpdate     AuditLogAction = 24
	AuditLogActionMemberRoleUpdate AuditLogAction = 25
	AuditLogActionMemberMove       AuditLogAction = 26
	AuditLogActionMemberDisconnect AuditLogAction = 27
	AuditLogActionBotAdd           AuditLogAction = 28

	AuditLogActionRoleCreate AuditLogAction = 30
	AuditLogActionRoleUpdate AuditLogAction = 31
	AuditLogActionRoleDelete AuditLogAction = 32

	AuditLogActionInviteCreate AuditLogAction = 40
	AuditLogActionInviteUpdate AuditLogAction = 41
	AuditLogActionInviteDelete AuditLogAction = 42

	AuditLogActionWebhookCreate AuditLogAction = 50
	AuditLogActionWebhookUpdate AuditLogAction = 51
	AuditLogActionWebhookDelete AuditLogAction = 52

	AuditLogActionEmojiCreate AuditLogAction = 60
	AuditLogActionEmojiUpdate AuditLogAction = 61
	AuditLogActionEmojiDelete AuditLogAction = 62

	AuditLogActionMessageDelete     AuditLogAction = 72
	AuditLogActionMessageBulkDelete AuditLogAction = 73
	AuditLogActionMessagePin        AuditLogAction = 74
	AuditLogActionMessageUnpin      AuditLogAction = 75

	AuditLogActionIntegrationCreate   AuditLogAction = 80
	AuditLogActionIntegrationUpdate   AuditLogAction = 81
	AuditLogActionIntegrationDelete   AuditLogAction = 82
	AuditLogActionStageInstanceCreate AuditLogAction = 83
	AuditLogActionStageInstanceUpdate AuditLogAction = 84
	AuditLogActionStageInstanceDelete AuditLogAction = 85

	AuditLogActionStickerCreate AuditLogAction = 90
	AuditLogActionStickerUpdate AuditLogAction = 91
	AuditLogActionStickerDelete AuditLogAction = 92

	AuditLogGuildScheduledEventCreate AuditLogAction = 100
	AuditLogGuildScheduledEventUpdare AuditLogAction = 101
	AuditLogGuildScheduledEventDelete AuditLogAction = 102

	AuditLogActionThreadCreate AuditLogAction = 110
	AuditLogActionThreadUpdate AuditLogAction = 111
	AuditLogActionThreadDelete AuditLogAction = 112

	AuditLogActionApplicationCommandPermissionUpdate AuditLogAction = 121
)

// GuildMemberParams stores data needed to update a member
// https://discord.com/developers/docs/resources/guild#modify-guild-member
type GuildMemberParams struct {
	// Value to set user's nickname to.
	Nick string `json:"nick,omitempty"`
	// Array of role ids the member is assigned.
	Roles *[]string `json:"roles,omitempty"`
	// ID of channel to move user to (if they are connected to voice).
	// Set to "" to remove user from a voice channel.
	ChannelID *string `json:"channel_id,omitempty"`
	// Whether the user is muted in voice channels.
	Mute *bool `json:"mute,omitempty"`
	// Whether the user is deafened in voice channels.
	Deaf *bool `json:"deaf,omitempty"`
	// When the user's timeout will expire and the user will be able
	// to communicate in the guild again (up to 28 days in the future).
	// Set to time.Time{} to remove timeout.
	//CommunicationDisabledUntil string `json:"communication_disabled_until,omitempty"`
	CommunicationDisabledUntil *time.Time `json:"communication_disabled_until,omitempty"`
}

// MarshalJSON is a helper function to marshal GuildMemberParams.
func (p GuildMemberParams) MarshalJSON() (res []byte, err error) {
	type guildMemberParams GuildMemberParams
	v := struct {
		guildMemberParams
		ChannelID                  json.RawMessage `json:"channel_id,omitempty"`
		CommunicationDisabledUntil json.RawMessage `json:"communication_disabled_until,omitempty"`
	}{guildMemberParams: guildMemberParams(p)}

	if p.ChannelID != nil {
		if *p.ChannelID == "" {
			v.ChannelID = json.RawMessage(`null`)
		} else {
			res, err = json.Marshal(p.ChannelID)
			if err != nil {
				return
			}
			v.ChannelID = res
		}
	}

	if p.CommunicationDisabledUntil != nil {
		if p.CommunicationDisabledUntil.IsZero() {
			v.CommunicationDisabledUntil = json.RawMessage(`null`)
		} else {
			res, err = json.Marshal(p.CommunicationDisabledUntil)
			if err != nil {
				return
			}
			v.CommunicationDisabledUntil = res
		}
	}

	return json.Marshal(v)
}

// GuildMemberAddParams stores data needed to add a user to a guild.
// NOTE: All fields are optional, except AccessToken.
type GuildMemberAddParams struct {
	// Valid access_token for the user.
	AccessToken string `json:"access_token"`
	// Value to set users nickname to.
	Nick string `json:"nick,omitempty"`
	// A list of role ID's to set on the member.
	Roles []string `json:"roles,omitempty"`
	// Whether the user is muted.
	Mute bool `json:"mute,omitempty"`
	// Whether the user is deafened.
	Deaf bool `json:"deaf,omitempty"`
}

// An APIErrorMessage is an api error message returned from discord
type APIErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MessageReaction stores the data for a message reaction.
type MessageReaction struct {
	UserID    int64 `json:"user_id,string"`
	MessageID int64 `json:"message_id,string"`
	Emoji     Emoji `json:"emoji"`
	ChannelID int64 `json:"channel_id,string"`
	GuildID   int64 `json:"guild_id,string,omitempty"`
}

func (mr *MessageReaction) GetGuildID() int64 {
	return mr.GuildID
}

func (mr *MessageReaction) GetChannelID() int64 {
	return mr.ChannelID
}

// GatewayBotResponse stores the data for the gateway/bot response
type GatewayBotResponse struct {
	URL               string            `json:"url"`
	Shards            int               `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

type SessionStartLimit struct {
	Total      int   `json:"total,omitempty"`
	Remaining  int   `json:"remaining,omitempty"`
	ResetAfter int64 `json:"reset_after,omitempty"`
}

// SessionInformation provides the information for max concurrency sharding
type SessionInformation struct {
	Total          int `json:"total,omitempty"`
	Remaining      int `json:"remaining,omitempty"`
	ResetAfter     int `json:"reset_after,omitempty"`
	MaxConcurrency int `json:"max_concurrency,omitempty"`
}

// GatewayStatusUpdate is sent by the client to indicate a presence or status update
// https://discord.com/developers/docs/topics/gateway#update-status-gateway-status-update-structure
type GatewayStatusUpdate struct {
	Since  int      `json:"since"`
	Game   Activity `json:"game"`
	Status string   `json:"status"`
	AFK    bool     `json:"afk"`
}

// Activities and GameType is Jonas' add-on
type Activities []*Activity

func (a *Activities) UnmarshalJSONArray(dec *gojay.Decoder) error {
	instance := Activity{}
	err := dec.Object(&instance)
	if err != nil {
		return err
	}
	*a = append(*a, &instance)
	return nil
}

// GameType is the type of "game" (see GameType* consts) in the Game struct
type GameType int

// Valid GameType values
const (
	GameTypeGame GameType = iota
	GameTypeStreaming
	GameTypeListening
	GameTypeWatching
	GameTypeCustom
	GameTypeCompeting
)

// A Game struct holds the name of the "playing .." game for a user
type Game struct {
	Name          string     `json:"name"`
	Type          GameType   `json:"type"`
	URL           string     `json:"url,omitempty"`
	Details       string     `json:"details,omitempty"`
	State         string     `json:"state,omitempty"`
	TimeStamps    TimeStamps `json:"timestamps,omitempty"`
	Assets        Assets     `json:"assets,omitempty"`
	ApplicationID string     `json:"application_id,omitempty"`
	Instance      int8       `json:"instance,omitempty"`
	// TODO: Party and Secrets (unknown structure)
}

// implement gojay.UnmarshalerJSONObject
func (g *Game) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "name":
		return dec.String(&g.Name)
	case "type":
		return dec.Int((*int)(&g.Type))
	case "url":
		return dec.String(&g.URL)
	case "details":
		return dec.String(&g.Details)
	case "state":
		return dec.String(&g.State)
	case "timestamps":
		return dec.Object(&g.TimeStamps)
	case "assets":
	case "application_id":
		var i interface{}
		err := dec.Interface(&i)
		if err != nil {
			return err
		}
		switch t := i.(type) {
		case int64:
			g.ApplicationID = strconv.FormatInt(t, 10)
		case int32:
			g.ApplicationID = strconv.FormatInt(int64(t), 10)
		case string:
			g.ApplicationID = t
		}
	case "instance":
		return dec.Int8(&g.Instance)
	}

	return nil
}

func (g *Game) NKeys() int {
	return 0
}

// Activity defines the Activity sent with GatewayStatusUpdate
// https://discord.com/developers/docs/topics/gateway#activity-object
type Activity struct {
	Name          string       `json:"name"`
	Type          ActivityType `json:"type"`
	URL           string       `json:"url,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	ApplicationID string       `json:"application_id,omitempty"`
	State         string       `json:"state,omitempty"`
	Details       string       `json:"details,omitempty"`
	Timestamps    TimeStamps   `json:"timestamps,omitempty"`
	Emoji         Emoji        `json:"emoji,omitempty"`
	Party         Party        `json:"party,omitempty"`
	Assets        Assets       `json:"assets,omitempty"`
	Secrets       Secrets      `json:"secrets,omitempty"`
	Instance      bool         `json:"instance,omitempty"`
	Flags         int          `json:"flags,omitempty"`
	// Buttons       []*ActivityButton `json:"buttons,omitempty"`
}

// implement gojay.UnmarshalerJSONObject
func (a *Activity) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "name":
		return dec.String(&a.Name)
	case "type":
		return dec.Int((*int)(&a.Type))
	case "url":
		return dec.String(&a.URL)
	case "details":
		return dec.String(&a.Details)
	case "state":
		return dec.String(&a.State)
	case "timestamps":
		return dec.Object(&a.Timestamps)
	case "assets":
	case "application_id":
		var i interface{}
		err := dec.Interface(&i)
		if err != nil {
			return err
		}
		switch t := i.(type) {
		case int64:
			a.ApplicationID = strconv.FormatInt(t, 10)
		case int32:
			a.ApplicationID = strconv.FormatInt(int64(t), 10)
		case string:
			a.ApplicationID = t
		}
	case "instance":
		return dec.Bool(&a.Instance)
	case "flags":
		return dec.Int(&a.Flags)
	case "buttons":
	}

	return nil
}

func (a *Activity) NKeys() int {
	return 0
}

// UnmarshalJSON is a custom unmarshaljson to make CreatedAt a time.Time instead of an int
func (activity *Activity) UnmarshalJSON(b []byte) error {
	temp := struct {
		Name          string       `json:"name"`
		Type          ActivityType `json:"type"`
		URL           string       `json:"url,omitempty"`
		CreatedAt     int64        `json:"created_at"`
		ApplicationID string       `json:"application_id,omitempty"`
		State         string       `json:"state,omitempty"`
		Details       string       `json:"details,omitempty"`
		Timestamps    TimeStamps   `json:"timestamps,omitempty"`
		Emoji         Emoji        `json:"emoji,omitempty"`
		Party         Party        `json:"party,omitempty"`
		Assets        Assets       `json:"assets,omitempty"`
		Secrets       Secrets      `json:"secrets,omitempty"`
		Instance      bool         `json:"instance,omitempty"`
		Flags         int          `json:"flags,omitempty"`
		// Buttons       []*ActivityButton `json:"buttons,omitempty"`
	}{}
	err := Unmarshal(b, &temp)
	if err != nil {
		return err
	}
	activity.CreatedAt = time.Unix(0, temp.CreatedAt*1000000)
	activity.ApplicationID = temp.ApplicationID
	activity.Assets = temp.Assets
	activity.Details = temp.Details
	activity.Emoji = temp.Emoji
	activity.Flags = temp.Flags
	activity.Instance = temp.Instance
	activity.Name = temp.Name
	activity.Party = temp.Party
	activity.Secrets = temp.Secrets
	activity.State = temp.State
	activity.Timestamps = temp.Timestamps
	activity.Type = temp.Type
	activity.URL = temp.URL
	// activity.Buttons = temp.Buttons
	return nil
}

type ActivityButton struct {
	Label string `json:"label,omitempty"`
	URL   string `json:"url,omitempty"`
}

func (ab *ActivityButton) UnmarshalJSONArray(dec *gojay.Decoder, key string) error {
	switch key {
	case "label":
		return dec.String(&ab.Label)
	case "url":
		return dec.String(&ab.URL)
	}

	return nil
}

func (ab *ActivityButton) NKeys() int {
	return 0
}

// Party defines the Party field in the Activity struct
// https://discord.com/developers/docs/topics/gateway#activity-object
type Party struct {
	ID   string `json:"id,omitempty"`
	Size []int  `json:"size,omitempty"`
}

// Secrets defines the Secrets field for the Activity struct
// https://discord.com/developers/docs/topics/gateway#activity-object
type Secrets struct {
	Join     string `json:"join,omitempty"`
	Spectate string `json:"spectate,omitempty"`
	Match    string `json:"match,omitempty"`
}

// ActivityType is the type of Activity (see ActivityType* consts) in the Activity struct
// https://discord.com/developers/docs/topics/gateway#activity-object-activity-types
type ActivityType int

// Valid ActivityType values
const (
	ActivityTypeGame      ActivityType = 0
	ActivityTypeStreaming ActivityType = 1
	ActivityTypeListening ActivityType = 2
	ActivityTypeWatching  ActivityType = 3
	ActivityTypeCustom    ActivityType = 4
	ActivityTypeCompeting ActivityType = 5
)

// Identify is sent during initial handshake with the discord gateway.
// https://discord.com/developers/docs/topics/gateway#identify
type Identify struct {
	Token          string              `json:"token"`
	Properties     IdentifyProperties  `json:"properties"`
	Compress       bool                `json:"compress"`
	LargeThreshold int                 `json:"large_threshold"`
	Shard          *[2]int             `json:"shard,omitempty"`
	Presence       GatewayStatusUpdate `json:"presence,omitempty"`
	Intents        Intent              `json:"intents"`
}

// IdentifyProperties contains the "properties" portion of an Identify packet
// https://discord.com/developers/docs/topics/gateway#identify-identify-connection-properties
type IdentifyProperties struct {
	OS              string `json:"$os"`
	Browser         string `json:"$browser"`
	Device          string `json:"$device"`
	Referer         string `json:"$referer"`
	ReferringDomain string `json:"$referring_domain"`
}

// StageInstance holds information about a live stage.
// https://discord.com/developers/docs/resources/stage-instance#stage-instance-resource
type StageInstance struct {
	// The id of this Stage instance
	ID int64 `json:"id,string"`
	// The guild id of the associated Stage channel
	GuildID int64 `json:"guild_id,string"`
	// The id of the associated Stage channel
	ChannelID int64 `json:"channel_id,string"`
	// The topic of the Stage instance (1-120 characters)
	Topic string `json:"topic"`
	// The privacy level of the Stage instance
	// https://discord.com/developers/docs/resources/stage-instance#stage-instance-object-privacy-level
	PrivacyLevel StageInstancePrivacyLevel `json:"privacy_level"`
	// Whether or not Stage Discovery is disabled (deprecated)
	DiscoverableDisabled bool `json:"discoverable_disabled"`
	// The id of the scheduled event for this Stage instance
	GuildScheduledEventID int64 `json:"guild_scheduled_event_id,string"`
}

// StageInstanceParams represents the parameters needed to create or edit a stage instance
type StageInstanceParams struct {
	// ChannelID represents the id of the Stage channel
	ChannelID int64 `json:"channel_id,string,omitempty"`
	// Topic of the Stage instance (1-120 characters)
	Topic string `json:"topic,omitempty"`
	// PrivacyLevel of the Stage instance (default GUILD_ONLY)
	PrivacyLevel StageInstancePrivacyLevel `json:"privacy_level,omitempty"`
	// SendStartNotification will notify @everyone that a Stage instance has started
	SendStartNotification bool `json:"send_start_notification,omitempty"`
}

// StageInstancePrivacyLevel represents the privacy level of a Stage instance
// https://discord.com/developers/docs/resources/stage-instance#stage-instance-object-privacy-level
type StageInstancePrivacyLevel int

const (
	// StageInstancePrivacyLevelPublic The Stage instance is visible publicly. (deprecated)
	StageInstancePrivacyLevelPublic StageInstancePrivacyLevel = 1
	// StageInstancePrivacyLevelGuildOnly The Stage instance is visible to only guild members.
	StageInstancePrivacyLevelGuildOnly StageInstancePrivacyLevel = 2
)

/*
// Block contains Discord JSON Error Response codes
const (
	ErrCodeUnknownAccount     = 10001
	ErrCodeUnknownApplication = 10002
	ErrCodeUnknownChannel     = 10003
	ErrCodeUnknownGuild       = 10004
	ErrCodeUnknownIntegration = 10005
	ErrCodeUnknownInvite      = 10006
	ErrCodeUnknownMember      = 10007
	ErrCodeUnknownMessage     = 10008
	ErrCodeUnknownOverwrite   = 10009
	ErrCodeUnknownProvider    = 10010
	ErrCodeUnknownRole        = 10011
	ErrCodeUnknownToken       = 10012
	ErrCodeUnknownUser        = 10013
	ErrCodeUnknownEmoji       = 10014
	ErrCodeUnknownWebhook     = 10015

	ErrCodeBotsCannotUseEndpoint  = 20001
	ErrCodeOnlyBotsCanUseEndpoint = 20002

	ErrCodeMaximumGuildsReached     = 30001
	ErrCodeMaximumFriendsReached    = 30002
	ErrCodeMaximumPinsReached       = 30003
	ErrCodeMaximumGuildRolesReached = 30005
	ErrCodeTooManyReactions         = 30010

	ErrCodeUnauthorized = 40001

	ErrCodeMissingAccess                             = 50001
	ErrCodeInvalidAccountType                        = 50002
	ErrCodeCannotExecuteActionOnDMChannel            = 50003
	ErrCodeEmbedCisabled                             = 50004
	ErrCodeCannotEditFromAnotherUser                 = 50005
	ErrCodeCannotSendEmptyMessage                    = 50006
	ErrCodeCannotSendMessagesToThisUser              = 50007
	ErrCodeCannotSendMessagesInVoiceChannel          = 50008
	ErrCodeChannelVerificationLevelTooHigh           = 50009
	ErrCodeOAuth2ApplicationDoesNotHaveBot           = 50010
	ErrCodeOAuth2ApplicationLimitReached             = 50011
	ErrCodeInvalidOAuthState                         = 50012
	ErrCodeMissingPermissions                        = 50013
	ErrCodeInvalidAuthenticationToken                = 50014
	ErrCodeNoteTooLong                               = 50015
	ErrCodeTooFewOrTooManyMessagesToDelete           = 50016
	ErrCodeCanOnlyPinMessageToOriginatingChannel     = 50019
	ErrCodeCannotExecuteActionOnSystemMessage        = 50021
	ErrCodeMessageProvidedTooOldForBulkDelete        = 50034
	ErrCodeInvalidFormBody                           = 50035
	ErrCodeInviteAcceptedToGuildApplicationsBotNotIn = 50036

	ErrCodeReactionBlocked = 90001
)*/

// InviteUser is a partial user obejct from the invite event(s)
type InviteUser struct {
	ID            int64  `json:"id,string"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	Username      string `json:"username"`
}

type CreateApplicationCommandRequest struct {
	Name              string                      `json:"name"`        // 1-32 character name matching ^[\w-]{1,32}$
	Description       string                      `json:"description"` // 1-100 character description
	Type              ApplicationCommandType      `json:"type,omitempty"`
	Options           []*ApplicationCommandOption `json:"options"`                      // the parameters for the command
	DefaultPermission *bool                       `json:"default_permission,omitempty"` // (default true)	whether the command is enabled by default when the app is added to a guild
	NSFW              bool                        `json:"nsfw,omitempty"`               // marks a command as age-restricted
}

func (a *ApplicationCommandInteractionDataResolved) UnmarshalJSON(b []byte) error {
	var temp *applicationCommandInteractionDataResolvedTemp
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	*a = ApplicationCommandInteractionDataResolved{
		Users:    make(map[int64]*User),
		Members:  make(map[int64]*Member),
		Roles:    make(map[int64]*Role),
		Channels: make(map[int64]*Channel),
	}

	for k, v := range temp.Channels {
		parsed, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return err
		}
		a.Channels[parsed] = v
	}

	for k, v := range temp.Roles {
		parsed, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return err
		}
		a.Roles[parsed] = v
	}

	for k, v := range temp.Members {
		parsed, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return err
		}
		a.Members[parsed] = v
	}

	for k, v := range temp.Users {
		parsed, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return err
		}
		a.Users[parsed] = v
	}

	return nil
}

type applicationCommandInteractionDataResolvedTemp struct {
	Users    map[string]*User    `json:"users"`
	Members  map[string]*Member  `json:"members"`
	Roles    map[string]*Role    `json:"roles"`
	Channels map[string]*Channel `json:"channels"`
}

type applicationCommandInteractionDataOptionTemporary struct {
	Name    string                                     `json:"name"`    // the name of the parameter
	Type    ApplicationCommandOptionType               `json:"type"`    // value of ApplicationCommandOptionType
	Value   json.RawMessage                            `json:"value"`   // the value of the pair
	Options []*ApplicationCommandInteractionDataOption `json:"options"` // present if this option is a group or subcommand
}

func (a *ApplicationCommandInteractionDataOption) UnmarshalJSON(b []byte) error {
	var temp *applicationCommandInteractionDataOptionTemporary
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	*a = ApplicationCommandInteractionDataOption{
		Name:    temp.Name,
		Type:    temp.Type,
		Options: temp.Options,
	}

	switch temp.Type {
	case ApplicationCommandOptionString:
		v := ""
		err = json.Unmarshal(temp.Value, &v)
		a.Value = v
	case ApplicationCommandOptionInteger:
		v := int64(0)
		err = json.Unmarshal(temp.Value, &v)
		a.Value = v
	case ApplicationCommandOptionBoolean:
		v := false
		err = json.Unmarshal(temp.Value, &v)
		a.Value = v
	case ApplicationCommandOptionUser, ApplicationCommandOptionChannel, ApplicationCommandOptionRole:
		// parse the snowflake
		v := ""
		err = json.Unmarshal(temp.Value, &v)
		if err == nil {
			a.Value, err = strconv.ParseInt(v, 10, 64)
		}
	case ApplicationCommandOptionSubCommand:
	case ApplicationCommandOptionSubCommandGroup:
	}

	return err
}

/*
	type InteractionResponse struct {
		Kind InteractionResponseType                    `json:"type"` // the type of response
		Data *InteractionApplicationCommandCallbackData `json:"data"` // an optional response message
	}

type InteractionResponseType int

const (

	InteractionResponseTypePong                             InteractionResponseType = 1 // ACK a Ping
	InteractionResponseTypeAcknowledge                      InteractionResponseType = 2 // DEPRECATED ACK a command without sending a message, eating the user's input
	InteractionResponseTypeChannelMessage                   InteractionResponseType = 3 // DEPRECATED respond with a message, eating the user's input
	InteractionResponseTypeChannelMessageWithSource         InteractionResponseType = 4 // respond to an interaction with a message
	InteractionResponseTypeDeferredChannelMessageWithSource InteractionResponseType = 5 // ACK an interaction and edit a response later, the user sees a loading state

)
*/
type InteractionApplicationCommandCallbackData struct {
	TTS             bool             `json:"tts,omitempty"`              //	is the response TTS
	Content         *string          `json:"content,omitempty"`          //	message content
	Embeds          []MessageEmbed   `json:"embeds,omitempty"`           // supports up to 10 embeds
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"` // allowed mentions object
	Flags           int              `json:"flags,omitempty"`            //	set to 64 to make your response ephemeral
}

// Constants for the different bit offsets of text channel permissions
const (
	// Deprecated: PermissionReadMessages has been replaced with PermissionViewChannel for text and voice channels
	PermissionReadMessages           int64 = 0x0000000000000400
	PermissionSendMessages           int64 = 0x0000000000000800
	PermissionSendTTSMessages        int64 = 0x0000000000001000
	PermissionManageMessages         int64 = 0x0000000000002000
	PermissionEmbedLinks             int64 = 0x0000000000004000
	PermissionAttachFiles            int64 = 0x0000000000008000
	PermissionReadMessageHistory     int64 = 0x0000000000010000
	PermissionMentionEveryone        int64 = 0x0000000000020000
	PermissionUseExternalEmojis      int64 = 0x0000000000040000
	PermissionUseSlashCommands       int64 = 0x0000000080000000
	PermissionUseApplicationCommands int64 = PermissionUseSlashCommands
	PermissionManageThreads          int64 = 0x0000000400000000
	PermissionCreatePublicThreads    int64 = 0x0000000800000000
	PermissionUsePublicThreads       int64 = PermissionCreatePublicThreads
	PermissionCreatePrivateThreads   int64 = 0x0000001000000000
	PermissionUsePrivateThreads      int64 = PermissionCreatePrivateThreads
	PermissionUseExternalStickers    int64 = 0x0000002000000000
	PermissionSendMessagesInThreads  int64 = 0x0000004000000000
)

// Constants for the different bit offsets of voice permissions
const (
	PermissionVoicePrioritySpeaker  int64 = 0x0000000000000100
	PermissionPrioritySpeaker       int64 = PermissionVoicePrioritySpeaker
	PermissionVoiceStreamVideo      int64 = 0x0000000000000200
	PermissionStream                int64 = PermissionVoiceStreamVideo
	PermissionVoiceConnect          int64 = 0x0000000000100000
	PermissionVoiceSpeak            int64 = 0x0000000000200000
	PermissionVoiceMuteMembers      int64 = 0x0000000000400000
	PermissionVoiceDeafenMembers    int64 = 0x0000000000800000
	PermissionVoiceMoveMembers      int64 = 0x0000000001000000
	PermissionVoiceUseVAD           int64 = 0x0000000002000000
	PermissionVoiceRequestToSpeak   int64 = 0x0000000100000000
	PermissionRequestToSpeak        int64 = PermissionVoiceRequestToSpeak
	PermissionUseActivities         int64 = 0x0000008000000000
	PermissionUseEmbeddedActivities int64 = PermissionUseActivities
)

// Constants for general management.
const (
	PermissionChangeNickname          int64 = 0x0000000004000000
	PermissionManageNicknames         int64 = 0x0000000008000000
	PermissionManageRoles             int64 = 0x0000000010000000
	PermissionManageWebhooks          int64 = 0x0000000020000000
	PermissionManageEmojis            int64 = 0x0000000040000000
	PermissionManageEmojisAndStickers int64 = PermissionManageEmojis
	PermissionManageEvents            int64 = 0x0000000200000000
)

// Constants for the different bit offsets of general permissions
const (
	PermissionCreateInstantInvite int64 = 0x0000000000000001
	PermissionKickMembers         int64 = 0x0000000000000002
	PermissionBanMembers          int64 = 0x0000000000000004
	PermissionAdministrator       int64 = 0x0000000000000008
	PermissionManageChannels      int64 = 0x0000000000000010
	PermissionManageServer        int64 = 0x0000000000000020
	PermissionManageGuild         int64 = PermissionManageServer
	PermissionAddReactions        int64 = 0x0000000000000040
	PermissionViewAuditLogs       int64 = 0x0000000000000080
	PermissionViewChannel         int64 = 0x0000000000000400
	PermissionViewGuildInsights   int64 = 0x0000000000080000
	PermissionModerateMembers     int64 = 0x0000010000000000

	PermissionAllText = PermissionViewChannel |
		PermissionSendMessages |
		PermissionSendTTSMessages |
		PermissionManageMessages |
		PermissionEmbedLinks |
		PermissionAttachFiles |
		PermissionReadMessageHistory |
		PermissionMentionEveryone
	PermissionAllVoice = PermissionViewChannel |
		PermissionVoiceConnect |
		PermissionVoiceSpeak |
		PermissionVoiceMuteMembers |
		PermissionVoiceDeafenMembers |
		PermissionVoiceMoveMembers |
		PermissionVoiceUseVAD |
		PermissionVoicePrioritySpeaker
	PermissionAllChannel = PermissionAllText |
		PermissionAllVoice |
		PermissionCreateInstantInvite |
		PermissionManageRoles |
		PermissionManageChannels |
		PermissionAddReactions |
		PermissionViewAuditLogs
	PermissionAll = PermissionAllChannel |
		PermissionKickMembers |
		PermissionBanMembers |
		PermissionManageServer |
		PermissionAdministrator |
		PermissionManageWebhooks |
		PermissionManageEmojis
)

// Block contains Discord JSON Error Response codes
const (
	ErrCodeGeneralError = 0

	ErrCodeUnknownAccount                        = 10001
	ErrCodeUnknownApplication                    = 10002
	ErrCodeUnknownChannel                        = 10003
	ErrCodeUnknownGuild                          = 10004
	ErrCodeUnknownIntegration                    = 10005
	ErrCodeUnknownInvite                         = 10006
	ErrCodeUnknownMember                         = 10007
	ErrCodeUnknownMessage                        = 10008
	ErrCodeUnknownOverwrite                      = 10009
	ErrCodeUnknownProvider                       = 10010
	ErrCodeUnknownRole                           = 10011
	ErrCodeUnknownToken                          = 10012
	ErrCodeUnknownUser                           = 10013
	ErrCodeUnknownEmoji                          = 10014
	ErrCodeUnknownWebhook                        = 10015
	ErrCodeUnknownWebhookService                 = 10016
	ErrCodeUnknownSession                        = 10020
	ErrCodeUnknownBan                            = 10026
	ErrCodeUnknownSKU                            = 10027
	ErrCodeUnknownStoreListing                   = 10028
	ErrCodeUnknownEntitlement                    = 10029
	ErrCodeUnknownBuild                          = 10030
	ErrCodeUnknownLobby                          = 10031
	ErrCodeUnknownBranch                         = 10032
	ErrCodeUnknownStoreDirectoryLayout           = 10033
	ErrCodeUnknownRedistributable                = 10036
	ErrCodeUnknownGiftCode                       = 10038
	ErrCodeUnknownStream                         = 10049
	ErrCodeUnknownPremiumServerSubscribeCooldown = 10050
	ErrCodeUnknownGuildTemplate                  = 10057
	ErrCodeUnknownDiscoveryCategory              = 10059
	ErrCodeUnknownSticker                        = 10060
	ErrCodeUnknownInteraction                    = 10062
	ErrCodeUnknownApplicationCommand             = 10063
	ErrCodeUnknownApplicationCommandPermissions  = 10066
	ErrCodeUnknownStageInstance                  = 10067
	ErrCodeUnknownGuildMemberVerificationForm    = 10068
	ErrCodeUnknownGuildWelcomeScreen             = 10069
	ErrCodeUnknownGuildScheduledEvent            = 10070
	ErrCodeUnknownGuildScheduledEventUser        = 10071
	ErrUnknownTag                                = 10087

	ErrCodeBotsCannotUseEndpoint                                            = 20001
	ErrCodeOnlyBotsCanUseEndpoint                                           = 20002
	ErrCodeExplicitContentCannotBeSentToTheDesiredRecipients                = 20009
	ErrCodeYouAreNotAuthorizedToPerformThisActionOnThisApplication          = 20012
	ErrCodeThisActionCannotBePerformedDueToSlowmodeRateLimit                = 20016
	ErrCodeOnlyTheOwnerOfThisAccountCanPerformThisAction                    = 20018
	ErrCodeMessageCannotBeEditedDueToAnnouncementRateLimits                 = 20022
	ErrCodeChannelHasHitWriteRateLimit                                      = 20028
	ErrCodeTheWriteActionYouArePerformingOnTheServerHasHitTheWriteRateLimit = 20029
	ErrCodeStageTopicContainsNotAllowedWordsForPublicStages                 = 20031
	ErrCodeGuildPremiumSubscriptionLevelTooLow                              = 20035

	ErrCodeMaximumGuildsReached                                     = 30001
	ErrCodeMaximumPinsReached                                       = 30003
	ErrCodeMaximumNumberOfRecipientsReached                         = 30004
	ErrCodeMaximumGuildRolesReached                                 = 30005
	ErrCodeMaximumNumberOfWebhooksReached                           = 30007
	ErrCodeMaximumNumberOfEmojisReached                             = 30008
	ErrCodeTooManyReactions                                         = 30010
	ErrCodeMaximumNumberOfGuildChannelsReached                      = 30013
	ErrCodeMaximumNumberOfAttachmentsInAMessageReached              = 30015
	ErrCodeMaximumNumberOfInvitesReached                            = 30016
	ErrCodeMaximumNumberOfAnimatedEmojisReached                     = 30018
	ErrCodeMaximumNumberOfServerMembersReached                      = 30019
	ErrCodeMaximumNumberOfGuildDiscoverySubcategoriesReached        = 30030
	ErrCodeGuildAlreadyHasATemplate                                 = 30031
	ErrCodeMaximumNumberOfThreadParticipantsReached                 = 30033
	ErrCodeMaximumNumberOfBansForNonGuildMembersHaveBeenExceeded    = 30035
	ErrCodeMaximumNumberOfBansFetchesHasBeenReached                 = 30037
	ErrCodeMaximumNumberOfUncompletedGuildScheduledEventsReached    = 30038
	ErrCodeMaximumNumberOfStickersReached                           = 30039
	ErrCodeMaximumNumberOfPruneRequestsHasBeenReached               = 30040
	ErrCodeMaximumNumberOfGuildWidgetSettingsUpdatesHasBeenReached  = 30042
	ErrCodeMaximumNumberOfEditsToMessagesOlderThanOneHourReached    = 30046
	ErrCodeMaximumNumberOfPinnedThreadsInForumChannelHasBeenReached = 30047
	ErrCodeMaximumNumberOfTagsInForumChannelHasBeenReached          = 30048

	ErrCodeUnauthorized                           = 40001
	ErrCodeActionRequiredVerifiedAccount          = 40002
	ErrCodeOpeningDirectMessagesTooFast           = 40003
	ErrCodeSendMessagesHasBeenTemporarilyDisabled = 40004
	ErrCodeRequestEntityTooLarge                  = 40005
	ErrCodeFeatureTemporarilyDisabledServerSide   = 40006
	ErrCodeUserIsBannedFromThisGuild              = 40007
	ErrCodeTargetIsNotConnectedToVoice            = 40032
	ErrCodeMessageAlreadyCrossposted              = 40033
	ErrCodeAnApplicationWithThatNameAlreadyExists = 40041
	ErrCodeInteractionHasAlreadyBeenAcknowledged  = 40060
	ErrCodeTagNamesMustBeUnique                   = 40061

	ErrCodeMissingAccess                                                = 50001
	ErrCodeInvalidAccountType                                           = 50002
	ErrCodeCannotExecuteActionOnDMChannel                               = 50003
	ErrCodeEmbedDisabled                                                = 50004
	ErrCodeGuildWidgetDisabled                                          = 50004
	ErrCodeCannotEditFromAnotherUser                                    = 50005
	ErrCodeCannotSendEmptyMessage                                       = 50006
	ErrCodeCannotSendMessagesToThisUser                                 = 50007
	ErrCodeCannotSendMessagesInVoiceChannel                             = 50008
	ErrCodeChannelVerificationLevelTooHigh                              = 50009
	ErrCodeOAuth2ApplicationDoesNotHaveBot                              = 50010
	ErrCodeOAuth2ApplicationLimitReached                                = 50011
	ErrCodeInvalidOAuthState                                            = 50012
	ErrCodeMissingPermissions                                           = 50013
	ErrCodeInvalidAuthenticationToken                                   = 50014
	ErrCodeTooFewOrTooManyMessagesToDelete                              = 50016
	ErrCodeCanOnlyPinMessageToOriginatingChannel                        = 50019
	ErrCodeInviteCodeWasEitherInvalidOrTaken                            = 50020
	ErrCodeCannotExecuteActionOnSystemMessage                           = 50021
	ErrCodeCannotExecuteActionOnThisChannelType                         = 50024
	ErrCodeInvalidOAuth2AccessTokenProvided                             = 50025
	ErrCodeMissingRequiredOAuth2Scope                                   = 50026
	ErrCodeInvalidWebhookTokenProvided                                  = 50027
	ErrCodeInvalidRole                                                  = 50028
	ErrCodeInvalidRecipients                                            = 50033
	ErrCodeMessageProvidedTooOldForBulkDelete                           = 50034
	ErrCodeInvalidFormBody                                              = 50035
	ErrCodeInviteAcceptedToGuildApplicationsBotNotIn                    = 50036
	ErrCodeInvalidAPIVersionProvided                                    = 50041
	ErrCodeFileUploadedExceedsTheMaximumSize                            = 50045
	ErrCodeInvalidFileUploaded                                          = 50046
	ErrCodeInvalidGuild                                                 = 50055
	ErrCodeInvalidMessageType                                           = 50068
	ErrCodeCannotDeleteAChannelRequiredForCommunityGuilds               = 50074
	ErrCodeInvalidStickerSent                                           = 50081
	ErrCodePerformedOperationOnArchivedThread                           = 50083
	ErrCodeBeforeValueIsEarlierThanThreadCreationDate                   = 50085
	ErrCodeCommunityServerChannelsMustBeTextChannels                    = 50086
	ErrCodeThisServerIsNotAvailableInYourLocation                       = 50095
	ErrCodeThisServerNeedsMonetizationEnabledInOrderToPerformThisAction = 50097
	ErrCodeThisServerNeedsMoreBoostsToPerformThisAction                 = 50101
	ErrCodeTheRequestBodyContainsInvalidJSON                            = 50109

	ErrCodeNoUsersWithDiscordTagExist = 80004

	ErrCodeReactionBlocked = 90001

	ErrCodeAPIResourceIsCurrentlyOverloaded = 130000

	ErrCodeTheStageIsAlreadyOpen = 150006

	ErrCodeCannotReplyWithoutPermissionToReadMessageHistory = 160002
	ErrCodeThreadAlreadyCreatedForThisMessage               = 160004
	ErrCodeThreadIsLocked                                   = 160005
	ErrCodeMaximumNumberOfActiveThreadsReached              = 160006
	ErrCodeMaximumNumberOfActiveAnnouncementThreadsReached  = 160007

	ErrCodeInvalidJSONForUploadedLottieFile                    = 170001
	ErrCodeUploadedLottiesCannotContainRasterizedImages        = 170002
	ErrCodeStickerMaximumFramerateExceeded                     = 170003
	ErrCodeStickerFrameCountExceedsMaximumOfOneThousandFrames  = 170004
	ErrCodeLottieAnimationMaximumDimensionsExceeded            = 170005
	ErrCodeStickerFrameRateOutOfRange                          = 170006
	ErrCodeStickerAnimationDurationExceedsMaximumOfFiveSeconds = 170007

	ErrCodeCannotUpdateAFinishedEvent             = 180000
	ErrCodeFailedToCreateStageNeededForStageEvent = 180002
)

// Intent is the type of a Gateway Intent
// https://discord.com/developers/docs/topics/gateway#gateway-intents
type Intent int

// Constants for the different bit offsets of intents
const (
	IntentGuilds                      Intent = 1 << 0
	IntentGuildMembers                Intent = 1 << 1
	IntentGuildBans                   Intent = 1 << 2
	IntentGuildEmojis                 Intent = 1 << 3
	IntentGuildIntegrations           Intent = 1 << 4
	IntentGuildWebhooks               Intent = 1 << 5
	IntentGuildInvites                Intent = 1 << 6
	IntentGuildVoiceStates            Intent = 1 << 7
	IntentGuildPresences              Intent = 1 << 8
	IntentGuildMessages               Intent = 1 << 9
	IntentGuildMessageReactions       Intent = 1 << 10
	IntentGuildMessageTyping          Intent = 1 << 11
	IntentDirectMessages              Intent = 1 << 12
	IntentDirectMessageReactions      Intent = 1 << 13
	IntentDirectMessageTyping         Intent = 1 << 14
	IntentMessageContent              Intent = 1 << 15
	IntentGuildScheduledEvents        Intent = 1 << 16
	IntentAutoModerationConfiguration Intent = 1 << 20
	IntentAutoModerationExecution     Intent = 1 << 21

	// TODO: remove when compatibility is not needed

	IntentsGuilds                 Intent = 1 << 0
	IntentsGuildMembers           Intent = 1 << 1
	IntentsGuildModeration        Intent = 1 << 2
	IntentsGuildBans              Intent = IntentsGuildModeration // TODO: remove when compatibility is not needed
	IntentsGuildEmojis            Intent = 1 << 3
	IntentsGuildIntegrations      Intent = 1 << 4
	IntentsGuildWebhooks          Intent = 1 << 5
	IntentsGuildInvites           Intent = 1 << 6
	IntentsGuildVoiceStates       Intent = 1 << 7
	IntentsGuildPresences         Intent = 1 << 8
	IntentsGuildMessages          Intent = 1 << 9
	IntentsGuildMessageReactions  Intent = 1 << 10
	IntentsGuildMessageTyping     Intent = 1 << 11
	IntentsDirectMessages         Intent = 1 << 12
	IntentsDirectMessageReactions Intent = 1 << 13
	IntentsDirectMessageTyping    Intent = 1 << 14
	IntentsMessageContent         Intent = 1 << 15
	IntentsGuildScheduledEvents   Intent = 1 << 16

	IntentsAllWithoutPrivileged = IntentGuilds |
		IntentGuildBans |
		IntentGuildEmojis |
		IntentGuildIntegrations |
		IntentGuildWebhooks |
		IntentGuildInvites |
		IntentGuildVoiceStates |
		IntentGuildMessages |
		IntentGuildMessageReactions |
		IntentGuildMessageTyping |
		IntentDirectMessages |
		IntentDirectMessageReactions |
		IntentDirectMessageTyping |
		IntentGuildScheduledEvents |
		IntentAutoModerationConfiguration |
		IntentAutoModerationExecution

	IntentsAll = IntentsAllWithoutPrivileged |
		IntentGuildMembers |
		IntentGuildPresences |
		IntentMessageContent

	IntentsNone Intent = 0
)

// MakeIntent used to help convert a gateway intent value for use in the Identify structure;
// this was useful to help support the use of a pointer type when intents were optional.
// This is now a no-op, and is not necessary to use.
func MakeIntent(intents Intent) Intent {
	return intents
}
