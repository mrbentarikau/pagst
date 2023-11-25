package discordgo

// Constants for the different bit offsets of text channel permissions
const (
	// Deprecated: PermissionReadMessages has been replaced with PermissionViewChannel for text and voice channels
	PermissionReadMessages int64 = 1 << 10

	// Allows for sending messages in a channel and creating threads in a forum (does not allow sending messages in threads).
	PermissionSendMessages int64 = 1 << 11

	// Allows for sending of /tts messages.
	PermissionSendTTSMessages int64 = 1 << 12

	// Allows for deletion of other users messages.
	PermissionManageMessages int64 = 1 << 13

	// Links sent by users with this permission will be auto-embedded.
	PermissionEmbedLinks int64 = 1 << 14

	// Allows for uploading images and files.
	PermissionAttachFiles int64 = 1 << 15

	// Allows for reading of message history.
	PermissionReadMessageHistory int64 = 1 << 16

	// Allows for using the @everyone tag to notify all users in a channel, and the @here tag to notify all online users in a channel.
	PermissionMentionEveryone int64 = 1 << 17

	// Allows the usage of custom emojis from other servers.
	PermissionUseExternalEmojis int64 = 1 << 18

	// Deprecated: PermissionUseSlashCommands has been replaced by PermissionUseApplicationCommands
	PermissionUseSlashCommands int64 = 1 << 31

	// Allows members to use application commands, including slash commands and context menu commands.
	PermissionUseApplicationCommands int64 = 1 << 31

	// Allows for deleting and archiving threads, and viewing all private threads.
	PermissionManageThreads int64 = 1 << 34

	// Allows for creating public and announcement threads.
	PermissionCreatePublicThreads int64 = 1 << 35
	PermissionUsePublicThreads    int64 = PermissionCreatePublicThreads

	// Allows for creating private threads.
	PermissionCreatePrivateThreads int64 = 1 << 36
	PermissionUsePrivateThreads    int64 = PermissionCreatePrivateThreads

	// Allows the usage of custom stickers from other servers.
	PermissionUseExternalStickers int64 = 1 << 37

	// Allows for sending messages in threads.
	PermissionSendMessagesInThreads int64 = 1 << 38

	// Allows sending voice messages.
	PermissionSendVoiceMessages int64 = 1 << 46
)

// Constants for the different bit offsets of voice permissions
const (
	// Allows for using priority speaker in a voice channel.
	PermissionVoicePrioritySpeaker int64 = 1 << 8
	PermissionPrioritySpeaker      int64 = PermissionVoicePrioritySpeaker

	// Allows the user to go live.
	PermissionVoiceStreamVideo int64 = 1 << 9
	PermissionStream           int64 = PermissionVoiceStreamVideo

	// Allows for joining of a voice channel.
	PermissionVoiceConnect int64 = 1 << 20

	// Allows for speaking in a voice channel.
	PermissionVoiceSpeak int64 = 1 << 21

	// Allows for muting members in a voice channel.
	PermissionVoiceMuteMembers int64 = 1 << 22

	// Allows for deafening of members in a voice channel.
	PermissionVoiceDeafenMembers int64 = 1 << 23

	// Allows for moving of members between voice channels.
	PermissionVoiceMoveMembers int64 = 1 << 24

	// Allows for using voice-activity-detection in a voice channel.
	PermissionVoiceUseVAD int64 = 1 << 25

	// Allows for requesting to speak in stage channels.
	PermissionVoiceRequestToSpeak int64 = 1 << 32
	PermissionRequestToSpeak      int64 = PermissionVoiceRequestToSpeak

	// Deprecated: PermissionUseActivities has been replaced by PermissionUseEmbeddedActivities.
	PermissionUseActivities int64 = 1 << 39

	// Allows for using Activities (applications with the EMBEDDED flag) in a voice channel.
	PermissionUseEmbeddedActivities int64 = 1 << 39

	// Allows for using soundboard in a voice channel.
	PermissionUseSoundboard int64 = 1 << 42

	// Allows the usage of custom soundboard sounds from other servers.
	PermissionUseExternalSounds int64 = 1 << 45
)

// Constants for general management.
const (
	// Allows for modification of own nickname.
	PermissionChangeNickname int64 = 1 << 26

	// Allows for modification of other users nicknames.
	PermissionManageNicknames int64 = 1 << 27

	// Allows management and editing of roles.
	PermissionManageRoles int64 = 1 << 28

	// Allows management and editing of webhooks.
	PermissionManageWebhooks int64 = 1 << 29

	// Deprecated: PermissionManageEmojis has been replaced by PermissionManageGuildExpressions.
	PermissionManageEmojis            int64 = 1 << 30
	PermissionManageEmojisAndStickers int64 = PermissionManageEmojis

	// Allows for editing and deleting emojis, stickers, and soundboard sounds created by all users.
	PermissionManageGuildExpressions int64 = 1 << 30

	// Allows for editing and deleting scheduled events created by all users.
	PermissionManageEvents int64 = 1 << 33

	// Allows for viewing role subscription insights.
	PermissionViewCreatorMonetizationAnalytics int64 = 1 << 41

	// Allows for creating emojis, stickers, and soundboard sounds, and editing and deleting those created by the current user.
	PermissionCreateGuildExpressions int64 = 1 << 43

	// Allows for creating scheduled events, and editing and deleting those created by the current user.
	PermissionCreateEvents int64 = 1 << 44
)

// Constants for the different bit offsets of general permissions
const (
	// Allows creation of instant invites.
	PermissionCreateInstantInvite int64 = 1 << 0

	// Allows kicking members.
	PermissionKickMembers int64 = 1 << 1

	// Allows banning members.
	PermissionBanMembers int64 = 1 << 2

	// Allows all permissions and bypasses channel permission overwrites.
	PermissionAdministrator int64 = 1 << 3

	// Allows management and editing of channels.
	PermissionManageChannels int64 = 1 << 4

	// Allows management and editing of the guild.
	PermissionManageServer int64 = 1 << 5
	PermissionManageGuild  int64 = PermissionManageServer

	// Allows for the addition of reactions to messages.
	PermissionAddReactions int64 = 1 << 6

	// Allows for viewing of audit logs.
	PermissionViewAuditLogs int64 = 1 << 7

	// Allows guild members to view a channel, which includes reading messages in text channels and joining voice channels.
	PermissionViewChannel int64 = 1 << 10

	// Allows for viewing guild insights.
	PermissionViewGuildInsights int64 = 1 << 19

	// Allows for timing out users to prevent them from sending or reacting to messages in chat and threads, and from speaking in voice and stage channels.
	PermissionModerateMembers int64 = 1 << 40

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
