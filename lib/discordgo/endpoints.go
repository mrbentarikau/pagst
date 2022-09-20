// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains variables for all known Discord end points.  All functions
// throughout the Discordgo package use these variables for all connections
// to Discord.  These are all exported and you may modify them if needed.

package discordgo

import "strconv"

// APIVersion is the Discord API version used for the REST and Websocket API.
var APIVersion = "10"

// Known Discord API Endpoints.
var (
	EndpointStatus     string
	EndpointSm         string
	EndpointSmActive   string
	EndpointSmUpcoming string

	EndpointDiscord        string
	EndpointAPI            string
	EndpointGuilds         string
	EndpointChannels       string
	EndpointUsers          string
	EndpointGateway        string
	EndpointGatewayBot     string
	EndpointWebhooks       string
	EndpointStickers       string
	EndpointStageInstances string

	EndpointCDN             string
	EndpointCDNAttachments  string
	EndpointCDNAvatars      string
	EndpointCDNIcons        string
	EndpointCDNSplashes     string
	EndpointCDNChannelIcons string
	EndpointCDNBanners      string
	EndpointCDNGuilds       string

	EndpointAuth           string
	EndpointLogin          string
	EndpointLogout         string
	EndpointVerify         string
	EndpointVerifyResend   string
	EndpointForgotPassword string
	EndpointResetPassword  string
	EndpointRegister       string

	EndpointVoice        string
	EndpointVoiceRegions string
	EndpointVoiceIce     string

	EndpointTutorial           string
	EndpointTutorialIndicators string

	EndpointTrack        string
	EndpointSso          string
	EndpointReport       string
	EndpointIntegrations string

	EndpointUser               = func(uID string) string { return "" }
	EndpointUserAvatar         = func(uID int64, aID string) string { return "" }
	EndpointUserAvatarAnimated = func(uID int64, aID string) string { return "" }
	EndpointDefaultUserAvatar  = func(uDiscriminator string) string { return "" }
	EndpointUserBanner         = func(uID int64, hash string) string { return "" }
	EndpointUserBannerAnimated = func(uID int64, hash string) string { return "" }

	EndpointUserGuilds      = func(uID string) string { return "" }
	EndpointUserGuild       = func(uID string, gID int64) string { return "" }
	EndpointUserGuildMember = func(uID string, gID int64) string { return "" }
	EndpointUserChannels    = func(uID string) string { return "" }
	EndpointUserConnections = func(uID string) string { return "" }

	EndpointUserSettings      = func(uID string) string { return "" }
	EndpointUserGuildSettings = func(uID string, gID int64) string { return "" }
	EndpointUserDevices       = func(uID string) string { return "" }
	EndpointUserNotes         = func(uID int64) string { return "" }

	EndpointGuild                    = func(gID int64) string { return "" }
	EndpointGuildAutoModeration      = func(gID int64) string { return "" }
	EndpointGuildAutoModerationRules = func(gID int64) string { return "" }
	EndpointGuildAutoModerationRule  = func(gID, rID int64) string { return "" }
	EndpointGuildThreads             = func(gID int64) string { return "" }
	EndpointGuildActiveThreads       = func(gID int64) string { return "" }
	EndpointGuildPreview             = func(gID int64) string { return "" }
	EndpointGuildChannels            = func(gID int64) string { return "" }
	EndpointGuildMembers             = func(gID int64) string { return "" }
	EndpointGuildMembersSearch       = func(gID int64) string { return "" }
	EndpointGuildMember              = func(gID int64, uID int64) string { return "" }
	EndpointGuildMemberRole          = func(gID, uID, rID int64) string { return "" }
	EndpointGuildMemberMe            = func(gID int64) string { return "" } // prolly defunct

	EndpointGuildBans            = func(gID int64) string { return "" }
	EndpointGuildBan             = func(gID, uID int64) string { return "" }
	EndpointGuildIntegrations    = func(gID int64) string { return "" }
	EndpointGuildIntegration     = func(gID, iID int64) string { return "" }
	EndpointGuildIntegrationSync = func(gID, iID int64) string { return "" } // prolly defunct

	EndpointGuildRoles                = func(gID int64) string { return "" }
	EndpointGuildRole                 = func(gID, rID int64) string { return "" }
	EndpointGuildInvites              = func(gID int64) string { return "" }
	EndpointGuildWidget               = func(gID int64) string { return "" }
	EndpointGuildEmbed                = func(gID int64) string { return "" }
	EndpointGuildPrune                = func(gID int64) string { return "" }
	EndpointGuildIcon                 = func(gID int64, hash string) string { return "" }
	EndpointGuildIconAnimated         = func(gID int64, hash string) string { return "" }
	EndpointGuildSplash               = func(gID int64, hash string) string { return "" }
	EndpointGuildWebhooks             = func(gID int64) string { return "" }
	EndpointGuildAuditLogs            = func(gID int64) string { return "" }
	EndpointGuildEmojis               = func(gID int64) string { return "" }
	EndpointGuildEmoji                = func(gID, eID int64) string { return "" }
	EndpointGuildBanner               = func(gID int64, hash string) string { return "" }
	EndpointGuildStickers             = func(gID int64) string { return "" }
	EndpointGuildSticker              = func(gID, eID int64) string { return "" }
	EndpointStageInstance             = func(cID int64) string { return "" }
	EndpointGuildScheduledEvents      = func(gID int64) string { return "" }
	EndpointGuildScheduledEvent       = func(gID, eID int64) string { return "" }
	EndpointGuildScheduledEventUsers  = func(gID, eID int64) string { return "" }
	EndpointGuildTemplate             = func(tID int64) string { return "" }
	EndpointGuildTemplates            = func(gID int64) string { return "" }
	EndpointGuildTemplateSync         = func(gID, tID int64) string { return "" }
	EndpointGuildMemberAvatar         = func(gID, uID int64, aID string) string { return "" }
	EndpointGuildMemberAvatarAnimated = func(gID, uID int64, aID string) string { return "" }

	EndpointChannel                             = func(cID int64) string { return "" }
	EndpointChannelThreads                      = func(cID int64) string { return "" }
	EndpointChannelActiveThreads                = func(cID int64) string { return "" }
	EndpointChannelPublicArchivedThreads        = func(cID int64) string { return "" }
	EndpointChannelPrivateArchivedThreads       = func(cID int64) string { return "" }
	EndpointChannelJoinedPrivateArchivedThreads = func(cID int64) string { return "" }
	EndpointChannelPermissions                  = func(cID int64) string { return "" }
	EndpointChannelPermission                   = func(cID, tID int64) string { return "" }
	EndpointChannelInvites                      = func(cID int64) string { return "" }
	EndpointChannelTyping                       = func(cID int64) string { return "" }
	EndpointChannelMessages                     = func(cID int64) string { return "" }
	EndpointChannelMessage                      = func(cID, mID int64) string { return "" }
	EndpointChannelMessageAck                   = func(cID, mID int64) string { return "" }
	EndpointChannelMessageThread                = func(cID, mID int64) string { return "" }
	EndpointChannelMessagesBulkDelete           = func(cID int64) string { return "" }
	EndpointChannelMessagesPins                 = func(cID int64) string { return "" }
	EndpointChannelMessagePin                   = func(cID, mID int64) string { return "" }
	EndpointChannelMessageCrosspost             = func(cID, mID int64) string { return "" }
	EndpointChannelFollow                       = func(cID int64) string { return "" }
	EndpointThreadMembers                       = func(tID int64) string { return "" }
	EndpointThreadMember                        = func(tID int64, mID string) string { return "" }

	EndpointGroupIcon = func(cID int64, hash string) string { return "" }

	EndpointSticker            = func(sID int64) string { return "" }
	EndpointNitroStickersPacks string

	EndpointChannelWebhooks = func(cID int64) string { return "" }
	EndpointWebhook         = func(wID int64) string { return "" }
	EndpointWebhookToken    = func(wID int64, token string) string { return "" }
	EndpointWebhookMessage  = func(wID int64, token, messageID string) string { return "" }

	EndpointMessageReactionsAll = func(cID, mID int64) string { return "" }
	EndpointMessageReactions    = func(cID, mID int64, emoji EmojiName) string {
		return ""
	}
	EndpointMessageReaction = func(cID, mID int64, emoji EmojiName, uID string) string {
		return ""
	}

	EndpointApplicationGlobalCommands           = func(aID int64) string { return "" }
	EndpointApplicationGlobalCommand            = func(aID, cID int64) string { return "" }
	EndpointApplicationGuildCommands            = func(aID int64, gID int64) string { return "" }
	EndpointApplicationGuildCommand             = func(aID int64, gID int64, cmdID int64) string { return "" }
	EndpointApplicationGuildCommandsPermissions = func(aID int64, gID int64) string { return "" }
	EndpointApplicationGuildCommandPermissions  = func(aID int64, gID int64, cmdID int64) string { return "" }
	EndpointApplicationCommandPermissions       = func(aID, gID, cID int64) string { return "" }
	EndpointApplicationCommandsGuildPermissions = func(aID, gID int64) string { return "" }

	EndpointInteraction                = func(aID int64, iToken string) string { return "" }
	EndpointInteractionResponse        = func(iID int64, iToken string) string { return "" }
	EndpointInteractionResponseActions = func(aID int64, iToken string) string { return "" }
	EndpointInteractionFollowupMessage = func(applicationID int64, token string, messageID int64) string {
		return ""
	}

	EndpointFollowupMessageActions = func(aID int64, iToken, mID string) string { return "" }

	EndpointGuildCreate = ""

	EndpointInvite = func(iID string) string { return "" }

	EndpointIntegrationsJoin = func(iID string) string { return "" }

	EndpointEmoji         = func(eID int64) string { return "" }
	EndpointEmojiAnimated = func(eID int64) string { return "" }

	EndpointApplications    = ""
	EndpointApplication     = func(aID int64) string { return "" }
	EndpointApplicationMe   = "" // prolly defunct
	EndpointApplicationsBot = func(aID int64) string { return "" }

	EndpointOauth2                  = ""
	EndpointOauth2Applications      = ""
	EndpointOauth2Application       = func(aID int64) string { return "" }
	EndpointOauth2ApplicationsBot   = func(aID int64) string { return "" }
	EndpointOauth2ApplicationAssets = func(aID int64) string { return "" }

	// prolly defunct section
	EndpointRelationships       = func() string { return "" }          // prolly defunct
	EndpointRelationship        = func(uID int64) string { return "" } // prolly defunct
	EndpointRelationshipsMutual = func(uID int64) string { return "" } // prolly defunct

	EndpointApplicationNonOauth2 = func(aID int64) string { return "" }
	EndpointApplicationCommands  = func(aID int64) string { return "" }
	EndpointApplicationCommand   = func(aID int64, cmdID int64) string { return "" }

	EndpointInteractions        = ""
	EndpointInteractionCallback = func(interactionID int64, token string) string {
		return ""
	}

	EndpointWebhookInteraction = func(applicationID int64, token string) string {
		return ""
	}
	EndpointInteractionOriginalMessage = func(applicationID int64, token string) string {
		return ""
	}
)

func CreateEndpoints(base string) {
	EndpointStatus = "https://status.discord.com/api/v2/"
	EndpointSm = EndpointStatus + "scheduled-maintenances/"
	EndpointSmActive = EndpointSm + "active.json"
	EndpointSmUpcoming = EndpointSm + "upcoming.json"

	EndpointDiscord = base
	EndpointAPI = EndpointDiscord + "api/v" + APIVersion + "/"
	EndpointGuilds = EndpointAPI + "guilds/"
	EndpointChannels = EndpointAPI + "channels/"
	EndpointUsers = EndpointAPI + "users/"
	EndpointGateway = EndpointAPI + "gateway"
	EndpointGatewayBot = EndpointGateway + "/bot"
	EndpointWebhooks = EndpointAPI + "webhooks/"
	EndpointStickers = EndpointAPI + "stickers/"
	EndpointStageInstances = EndpointAPI + "stage-instances"

	EndpointCDN = "https://cdn.discordapp.com/"
	EndpointCDNAttachments = EndpointCDN + "attachments/"
	EndpointCDNAvatars = EndpointCDN + "avatars/"
	EndpointCDNIcons = EndpointCDN + "icons/"
	EndpointCDNSplashes = EndpointCDN + "splashes/"
	EndpointCDNChannelIcons = EndpointCDN + "channel-icons/"
	EndpointCDNBanners = EndpointCDN + "banners/"
	EndpointCDNGuilds = EndpointCDN + "guilds/"

	EndpointAuth = EndpointAPI + "auth/"
	EndpointLogin = EndpointAuth + "login"
	EndpointLogout = EndpointAuth + "logout"
	EndpointVerify = EndpointAuth + "verify"
	EndpointVerifyResend = EndpointAuth + "verify/resend"
	EndpointForgotPassword = EndpointAuth + "forgot"
	EndpointResetPassword = EndpointAuth + "reset"
	EndpointRegister = EndpointAuth + "register"

	EndpointVoice = EndpointAPI + "/voice/"
	EndpointVoiceRegions = EndpointVoice + "regions"
	EndpointVoiceIce = EndpointVoice + "ice"

	EndpointTutorial = EndpointAPI + "tutorial/"
	EndpointTutorialIndicators = EndpointTutorial + "indicators"

	EndpointTrack = EndpointAPI + "track"
	EndpointSso = EndpointAPI + "sso"
	EndpointReport = EndpointAPI + "report"
	EndpointIntegrations = EndpointAPI + "integrations"

	EndpointUser = func(uID string) string { return EndpointUsers + uID }
	EndpointUserAvatar = func(uID int64, aID string) string { return EndpointCDNAvatars + StrID(uID) + "/" + aID + ".png" }
	EndpointUserAvatarAnimated = func(uID int64, aID string) string { return EndpointCDNAvatars + StrID(uID) + "/" + aID + ".gif" }
	EndpointDefaultUserAvatar = func(uDiscriminator string) string {
		uDiscriminatorInt, _ := strconv.Atoi(uDiscriminator)
		return EndpointCDN + "embed/avatars/" + strconv.Itoa(uDiscriminatorInt%5) + ".png"
	}
	EndpointUserSettings = func(uID string) string { return EndpointUsers + uID + "/settings" }
	EndpointUserGuilds = func(uID string) string { return EndpointUsers + uID + "/guilds" }
	EndpointUserGuild = func(uID string, gID int64) string { return EndpointUsers + uID + "/guilds/" + StrID(gID) }
	EndpointUserGuildMember = func(uID string, gID int64) string { return EndpointUserGuild(uID, gID) + "/member" }
	EndpointUserGuildSettings = func(uID string, gID int64) string { return EndpointUsers + uID + "/guilds/" + StrID(gID) + "/settings" }
	EndpointUserChannels = func(uID string) string { return EndpointUsers + uID + "/channels" }
	EndpointUserDevices = func(uID string) string { return EndpointUsers + uID + "/devices" }
	EndpointUserConnections = func(uID string) string { return EndpointUsers + uID + "/connections" }
	EndpointUserNotes = func(uID int64) string { return EndpointUsers + "@me/notes/" + StrID(uID) }

	EndpointUserBanner = func(uID int64, hash string) string {
		return EndpointCDNBanners + StrID(uID) + "/" + hash + ".png"
	}
	EndpointUserBannerAnimated = func(uID int64, hash string) string {
		return EndpointCDNBanners + StrID(uID) + "/" + hash + ".gif"
	}

	EndpointGuild = func(gID int64) string { return EndpointGuilds + StrID(gID) }
	EndpointGuildAutoModeration = func(gID int64) string { return EndpointGuild(gID) + "/auto-moderation" }
	EndpointGuildAutoModerationRules = func(gID int64) string { return EndpointGuildAutoModeration(gID) + "/rules" }
	EndpointGuildAutoModerationRule = func(gID, rID int64) string { return EndpointGuildAutoModerationRules(gID) + "/" + StrID(rID) }
	EndpointGuildThreads = func(gID int64) string { return EndpointGuild(gID) + "/threads" }
	EndpointGuildActiveThreads = func(gID int64) string { return EndpointGuildThreads(gID) + "/active" }
	EndpointGuildPreview = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/preview" }
	EndpointGuildChannels = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/channels" }
	EndpointGuildMembers = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/members" }
	EndpointGuildMembersSearch = func(gID int64) string { return EndpointGuildMembers(gID) + "/search" }
	EndpointGuildMember = func(gID int64, uID int64) string { return EndpointGuilds + StrID(gID) + "/members/" + StrID(uID) }
	EndpointGuildMemberRole = func(gID, uID, rID int64) string {
		return EndpointGuilds + StrID(gID) + "/members/" + StrID(uID) + "/roles/" + StrID(rID)
	}
	EndpointGuildMemberMe = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/members/@me" }

	EndpointGuildBans = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/bans" }
	EndpointGuildBan = func(gID, uID int64) string { return EndpointGuilds + StrID(gID) + "/bans/" + StrID(uID) }
	EndpointGuildIntegrations = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/integrations" }
	EndpointGuildIntegration = func(gID, iID int64) string { return EndpointGuilds + StrID(gID) + "/integrations/" + StrID(iID) }
	EndpointGuildIntegrationSync = func(gID, iID int64) string {
		return EndpointGuilds + StrID(gID) + "/integrations/" + StrID(iID) + "/sync"
	}
	EndpointGuildRoles = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/roles" }
	EndpointGuildRole = func(gID, rID int64) string { return EndpointGuilds + StrID(gID) + "/roles/" + StrID(rID) }
	EndpointGuildInvites = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/invites" }
	EndpointGuildWidget = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/widget" }
	EndpointGuildEmbed = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/embed" }
	EndpointGuildPrune = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/prune" }
	EndpointGuildIcon = func(gID int64, hash string) string { return EndpointCDNIcons + StrID(gID) + "/" + hash + ".png" }
	EndpointGuildIconAnimated = func(gID int64, hash string) string { return EndpointCDNIcons + StrID(gID) + "/" + hash + ".gif" }
	EndpointGuildSplash = func(gID int64, hash string) string { return EndpointCDNSplashes + StrID(gID) + "/" + hash + ".png" }
	EndpointGuildWebhooks = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/webhooks" }
	EndpointGuildAuditLogs = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/audit-logs" }
	EndpointGuildEmojis = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/emojis" }
	EndpointGuildEmoji = func(gID, eID int64) string { return EndpointGuilds + StrID(gID) + "/emojis/" + StrID(eID) }
	EndpointGuildBanner = func(gID int64, hash string) string { return EndpointCDNBanners + StrID(gID) + "/" + hash + ".png" }
	EndpointGuildStickers = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/stickers" }
	EndpointGuildSticker = func(gID, sID int64) string { return EndpointGuilds + StrID(gID) + "/stickers/" + StrID(sID) }
	EndpointStageInstance = func(cID int64) string { return EndpointStageInstances + "/" + StrID(cID) }
	EndpointGuildScheduledEvents = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/scheduled-events" }
	EndpointGuildScheduledEvent = func(gID, eID int64) string { return EndpointGuilds + StrID(gID) + "/scheduled-events/" + StrID(eID) }
	EndpointGuildScheduledEventUsers = func(gID, eID int64) string { return EndpointGuildScheduledEvent(gID, eID) + "/users" }
	EndpointGuildTemplate = func(tID int64) string { return EndpointGuilds + "/templates/" + StrID(tID) }
	EndpointGuildTemplates = func(gID int64) string { return EndpointGuilds + StrID(gID) + "/templates" }
	EndpointGuildTemplateSync = func(gID, tID int64) string { return EndpointGuilds + StrID(gID) + "/templates/" + StrID(tID) }
	EndpointGuildMemberAvatar = func(gID int64, uID int64, aID string) string {
		return EndpointCDNGuilds + StrID(gID) + "/users/" + StrID(uID) + "/avatars/" + aID + ".png"
	}
	EndpointGuildMemberAvatarAnimated = func(gID int64, uID int64, aID string) string {
		return EndpointCDNGuilds + StrID(gID) + "/users/" + StrID(uID) + "/avatars/" + aID + ".gif"
	}

	EndpointChannel = func(cID int64) string { return EndpointChannels + StrID(cID) }
	EndpointChannelThreads = func(cID int64) string { return EndpointChannel(cID) + "/threads" }
	EndpointChannelActiveThreads = func(cID int64) string { return EndpointChannelThreads(cID) + "/active" }
	EndpointChannelPublicArchivedThreads = func(cID int64) string { return EndpointChannelThreads(cID) + "/archived/public" }
	EndpointChannelPrivateArchivedThreads = func(cID int64) string { return EndpointChannelThreads(cID) + "/archived/private" }
	EndpointChannelJoinedPrivateArchivedThreads = func(cID int64) string { return EndpointChannel(cID) + "/users/@me/threads/archived/private" }
	EndpointChannelPermissions = func(cID int64) string { return EndpointChannels + StrID(cID) + "/permissions" }
	EndpointChannelPermission = func(cID, tID int64) string { return EndpointChannels + StrID(cID) + "/permissions/" + StrID(tID) }
	EndpointChannelInvites = func(cID int64) string { return EndpointChannels + StrID(cID) + "/invites" }
	EndpointChannelTyping = func(cID int64) string { return EndpointChannels + StrID(cID) + "/typing" }
	EndpointChannelMessages = func(cID int64) string { return EndpointChannels + StrID(cID) + "/messages" }
	EndpointChannelMessage = func(cID, mID int64) string { return EndpointChannels + StrID(cID) + "/messages/" + StrID(mID) }
	EndpointChannelMessageThread = func(cID, mID int64) string { return EndpointChannelMessage(cID, mID) + "/threads" }
	EndpointChannelMessageAck = func(cID, mID int64) string { return EndpointChannels + StrID(cID) + "/messages/" + StrID(mID) + "/ack" }
	EndpointChannelMessagesBulkDelete = func(cID int64) string { return EndpointChannel(cID) + "/messages/bulk-delete" }
	EndpointChannelMessagesPins = func(cID int64) string { return EndpointChannel(cID) + "/pins" }
	EndpointChannelMessagePin = func(cID, mID int64) string { return EndpointChannel(cID) + "/pins/" + StrID(mID) }
	EndpointChannelMessageCrosspost = func(cID, mID int64) string { return EndpointChannel(cID) + "/messages/" + StrID(mID) + "/crosspost" }
	EndpointChannelFollow = func(cID int64) string { return EndpointChannel(cID) + "/followers" }
	EndpointThreadMembers = func(tID int64) string { return EndpointChannel(tID) + "/thread-members" }
	EndpointThreadMember = func(tID int64, mID string) string { return EndpointThreadMembers(tID) + "/" + mID }

	EndpointGroupIcon = func(cID int64, hash string) string { return EndpointCDNChannelIcons + StrID(cID) + "/" + hash + ".png" }

	EndpointSticker = func(sID int64) string { return EndpointStickers + StrID(sID) }
	EndpointNitroStickersPacks = EndpointAPI + "/sticker-packs"

	EndpointChannelWebhooks = func(cID int64) string { return EndpointChannel(cID) + "/webhooks" }
	EndpointWebhook = func(wID int64) string { return EndpointWebhooks + StrID(wID) }
	EndpointWebhookToken = func(wID int64, token string) string { return EndpointWebhooks + StrID(wID) + "/" + token }
	EndpointWebhookMessage = func(wID int64, token, messageID string) string {
		return EndpointWebhookToken(wID, token) + "/messages/" + messageID
	}

	EndpointMessageReactionsAll = func(cID, mID int64) string {
		return EndpointChannelMessage(cID, mID) + "/reactions"
	}
	EndpointMessageReactions = func(cID, mID int64, emoji EmojiName) string {
		return EndpointChannelMessage(cID, mID) + "/reactions/" + emoji.String()
	}
	EndpointMessageReaction = func(cID, mID int64, emoji EmojiName, uID string) string {
		return EndpointMessageReactions(cID, mID, emoji) + "/" + uID
	}

	EndpointApplicationGlobalCommands = func(aID int64) string { return EndpointApplication(aID) + "/commands" }
	EndpointApplicationGlobalCommand = func(aID, cID int64) string { return EndpointApplicationGlobalCommands(aID) + "/" + StrID(cID) }
	EndpointApplicationGuildCommands = func(aID int64, gID int64) string {
		return EndpointApplicationNonOauth2(aID) + "/guilds/" + StrID(gID) + "/commands"
	}
	EndpointApplicationGuildCommand = func(aID int64, gID int64, cmdID int64) string {
		return EndpointApplicationGuildCommands(aID, gID) + "/" + StrID(cmdID)
	}
	EndpointApplicationGuildCommandsPermissions = func(aID int64, gID int64) string {
		return EndpointApplicationGuildCommands(aID, gID) + "/permissions"
	}
	EndpointApplicationGuildCommandPermissions = func(aID int64, gID int64, cmdID int64) string {
		return EndpointApplicationGuildCommand(aID, gID, cmdID) + "/permissions"
	}
	EndpointApplicationCommandPermissions = func(aID, gID, cID int64) string {
		return EndpointApplicationGuildCommand(aID, gID, cID) + "/permissions"
	}
	EndpointApplicationCommandsGuildPermissions = func(aID, gID int64) string {
		return EndpointApplicationGuildCommands(aID, gID) + "/permissions"
	}
	EndpointInteraction = func(aID int64, iToken string) string {
		return EndpointAPI + "interactions/" + StrID(aID) + "/" + iToken
	}
	EndpointInteractionResponse = func(iID int64, iToken string) string {
		return EndpointInteraction(iID, iToken) + "/callback"
	}
	EndpointInteractionResponseActions = func(aID int64, iToken string) string {
		return EndpointWebhookMessage(aID, iToken, "@original")
	}
	EndpointInteractionFollowupMessage = func(applicationID int64, token string, messageID int64) string {
		return EndpointWebhookInteraction(applicationID, token) + "/messages/" + StrID(messageID)
	}
	EndpointFollowupMessageActions = func(aID int64, iToken, mID string) string {
		return EndpointWebhookMessage(aID, iToken, mID)
	}

	EndpointGuildCreate = EndpointAPI + "guilds"

	EndpointInvite = func(iID string) string { return EndpointAPI + "invites/" + iID }

	EndpointIntegrationsJoin = func(iID string) string { return EndpointAPI + "integrations/" + iID + "/join" }

	EndpointEmoji = func(eID int64) string { return EndpointAPI + "emojis/" + StrID(eID) + ".png" }
	EndpointEmojiAnimated = func(eID int64) string { return EndpointAPI + "emojis/" + StrID(eID) + ".gif" }

	EndpointOauth2 = EndpointAPI + "oauth2/"
	EndpointOauth2Applications = EndpointAPI + "applications"
	EndpointOauth2Application = func(aID int64) string { return EndpointOauth2Applications + "/" + StrID(aID) }
	EndpointOauth2ApplicationsBot = func(aID int64) string { return EndpointOauth2Applications + "/" + StrID(aID) + "/bot" }
	EndpointOauth2ApplicationAssets = func(aID int64) string { return EndpointOauth2Applications + "/" + StrID(aID) + "/assets" }

	EndpointApplications = EndpointOauth2 + "applications"
	EndpointApplication = func(aID int64) string { return EndpointApplications + "/" + StrID(aID) }
	EndpointApplicationMe = EndpointApplications + "/@me"
	EndpointApplicationsBot = func(aID int64) string { return EndpointApplications + "/" + StrID(aID) + "/bot" }

	// prolly defunct section
	EndpointRelationships = func() string { return EndpointUsers + "@me" + "/relationships" }                     // prolly defunct
	EndpointRelationship = func(uID int64) string { return EndpointRelationships() + "/" + StrID(uID) }           // prolly defunct
	EndpointRelationshipsMutual = func(uID int64) string { return EndpointUsers + StrID(uID) + "/relationships" } // prolly defunct

	EndpointApplicationNonOauth2 = func(aID int64) string { return EndpointAPI + "applications/" + StrID(aID) }
	EndpointApplicationCommands = func(aID int64) string { return EndpointApplicationNonOauth2(aID) + "/commands" }
	EndpointApplicationCommand = func(aID int64, cmdID int64) string {
		return EndpointApplicationNonOauth2(aID) + "/commands/" + StrID(cmdID)
	}

	EndpointInteractions = EndpointAPI + "interactions"
	EndpointInteractionCallback = func(interactionID int64, token string) string {
		return EndpointInteractions + "/" + StrID(interactionID) + "/" + token + "/callback"
	}
	EndpointWebhookInteraction = func(applicationID int64, token string) string {
		return EndpointWebhooks + "/" + StrID(applicationID) + "/" + token
	}

	EndpointInteractionOriginalMessage = func(applicationID int64, token string) string {
		return EndpointWebhookInteraction(applicationID, token) + "/messages/@original"
	}

}

func init() {
	CreateEndpoints("https://discord.com/")
}
