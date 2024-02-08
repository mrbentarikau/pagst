package templates

import (
	"errors"

	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
)

// CtxChannel is almost a 1:1 copy of dstate.ChannelState, its needed because we cant expose all those state methods
// we also cant use discordgo.Channel because that would likely break a lot of custom commands at this point.
type CtxChannel struct {
	// These fields never change
	ID      int64
	GuildID int64

	IsForum         bool
	IsDMChannel     bool
	IsPrivateThread bool
	IsThread        bool

	Name                 string                           `json:"name"`
	Topic                string                           `json:"topic"`
	Type                 discordgo.ChannelType            `json:"type"`
	NSFW                 bool                             `json:"nsfw"`
	Icon                 string                           `json:"icon"`
	Position             int                              `json:"position"`
	Bitrate              int                              `json:"bitrate"`
	UserLimit            int                              `json:"user_limit"`
	PermissionOverwrites []*discordgo.PermissionOverwrite `json:"permission_overwrites"`
	ParentID             int64                            `json:"parent_id"`
	RateLimitPerUser     int                              `json:"rate_limit_per_user"`
	OwnerID              int64                            `json:"owner_id"`
	ThreadMetadata       *discordgo.ThreadMetadata        `json:"thread_metadata"`

	// The set of tags that can be used in a forum channel.
	AvailableTags []discordgo.ForumTag `json:"available_tags"`

	// The IDs of the set of tags that have been applied to a thread in a forum channel.
	AppliedTags []int64 `json:"applied_tags"`

	// Default duration, copied onto newly created threads, in minutes, threads will stop showing in the channel list after the specified period of inactivity, can be set to: 60, 1440, 4320, 10080
	DefaultAutoArchiveDuration int `json:"default_auto_archive_duration"`

	// The default forum layout view used to display posts in forum channels.
	// Defaults to ForumLayoutNotSet, which indicates a layout view has not been set by a channel admin.
	DefaultForumLayout discordgo.ForumLayout `json:"default_forum_layout"`

	// Emoji to use as the default reaction to a forum post.
	DefaultReactionEmoji discordgo.ForumDefaultReaction `json:"default_reaction_emoji"`

	// The default sort order type used to order posts in forum channels.
	// Defaults to null, which indicates a preferred sort order hasn't been set by a channel admin.
	DefaultSortOrder *discordgo.ForumSortOrderType `json:"default_sort_order"`

	// The initial RateLimitPerUser to set on newly created threads in a channel.
	// This field is copied to the thread at creation time and does not live update.
	DefaultThreadRateLimitPerUser int `json:"default_auto_archive_duration"`
}

// CtxThreadStart is almost a 1:1 copy of discordgo.ThreadStart but with some added fields
type CtxThreadStart struct {
	Name                string                `json:"name"`
	AutoArchiveDuration int                   `json:"auto_archive_duration,omitempty"`
	Type                discordgo.ChannelType `json:"type,omitempty"`
	Invitable           bool                  `json:"invitable"`
	RateLimitPerUser    int                   `json:"rate_limit_per_user,omitempty"`

	Content *discordgo.MessageSend `json:"content,omitempty"`

	// NOTE: forum threads only - these are names not ids
	AppliedTagNames []string `json:"applied_tag_names,omitempty"`

	// NOTE: forum threads only - IDs
	AppliedTags []int64 `json:"applied_tags,omitempty"`

	// NOTE: message threads only
	MessageID int64 `json:"message_id,omitempty"`
}

func (c *CtxChannel) Mention() (string, error) {
	if c == nil {
		return "", errors.New("channel not found")
	}
	return "<#" + discordgo.StrID(c.ID) + ">", nil

}

func CtxChannelFromCS(cs *dstate.ChannelState) *CtxChannel {

	cop := make([]*discordgo.PermissionOverwrite, len(cs.PermissionOverwrites))
	for i := 0; i < len(cs.PermissionOverwrites); i++ {
		cop[i] = &cs.PermissionOverwrites[i]
	}

	ctxChannel := &CtxChannel{
		ID:              cs.ID,
		GuildID:         cs.GuildID,
		IsForum:         cs.Type.IsForum(),
		IsDMChannel:     cs.IsDMChannel(),
		IsPrivateThread: cs.IsPrivateThread(),
		IsThread:        cs.Type.IsThread(),

		Name:                 cs.Name,
		Type:                 cs.Type,
		Topic:                cs.Topic,
		NSFW:                 cs.NSFW,
		Icon:                 cs.Icon,
		Position:             cs.Position,
		Bitrate:              cs.Bitrate,
		UserLimit:            cs.UserLimit,
		PermissionOverwrites: cop,
		ParentID:             cs.ParentID,
		RateLimitPerUser:     cs.RateLimitPerUser,
		OwnerID:              cs.OwnerID,
		ThreadMetadata:       cs.ThreadMetadata,

		AvailableTags: cs.AvailableTags,
		AppliedTags:   cs.AppliedTags,

		DefaultAutoArchiveDuration:    cs.DefaultAutoArchiveDuration,
		DefaultForumLayout:            cs.DefaultForumLayout,
		DefaultSortOrder:              cs.DefaultSortOrder,
		DefaultReactionEmoji:          cs.DefaultReactionEmoji,
		DefaultThreadRateLimitPerUser: cs.DefaultThreadRateLimitPerUser,
	}

	return ctxChannel
}

type CtxExecReturn struct {
	//Return   Slice
	Return   []interface{}
	Response *discordgo.MessageSend
}

func (c CtxExecReturn) String() string {
	if c.Response != nil {
		return c.Response.Content
	}
	return ""
}
