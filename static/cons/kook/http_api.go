package kook

const (
	// Channel
	ChannelList           = "/api/v3/channel/list"
	ChannelView           = "/api/v3/channel/view"
	ChannelCreate         = "/api/v3/channel/create"
	ChannelUpdate         = "/api/v3/channel/update"
	ChannelDelete         = "/api/v3/channel/delete"
	ChannelUserList       = "/api/v3/channel/user-list"
	ChannelMoveUser       = "/api/v3/channel/move-user"
	ChannelKickout        = "/api/v3/channel/kickout"
	ChannelUserGetJoined  = "/api/v3/channel-user/get-joined-channel"

	// Message
	MessageList           = "/api/v3/message/list"
	MessageView           = "/api/v3/message/view"
	MessageCreate         = "/api/v3/message/create"
	MessageUpdate         = "/api/v3/message/update"
	MessageDelete         = "/api/v3/message/delete"
	MessageAddReaction    = "/api/v3/message/add-reaction"
	MessageDeleteReaction = "/api/v3/message/delete-reaction"

	// Direct Message
	DirectMessageList           = "/api/v3/direct-message/list"
	DirectMessageView           = "/api/v3/direct-message/view"
	DirectMessageCreate         = "/api/v3/direct-message/create"
	DirectMessageUpdate         = "/api/v3/direct-message/update"
	DirectMessageDelete         = "/api/v3/direct-message/delete"
	DirectMessageAddReaction    = "/api/v3/direct-message/add-reaction"
	DirectMessageDeleteReaction = "/api/v3/direct-message/delete-reaction"

	// Guild
	GuildList     = "/api/v3/guild/list"
	GuildView     = "/api/v3/guild/view"
	GuildUserList = "/api/v3/guild/user-list"
	GuildLeave    = "/api/v3/guild/leave"
	GuildKickout  = "/api/v3/guild/kickout"

	// Guild Role
	GuildRoleList   = "/api/v3/guild-role/list"
	GuildRoleCreate = "/api/v3/guild-role/create"
	GuildRoleUpdate = "/api/v3/guild-role/update"
	GuildRoleDelete = "/api/v3/guild-role/delete"
	GuildRoleGrant  = "/api/v3/guild-role/grant"
	GuildRoleRevoke = "/api/v3/guild-role/revoke"

	// User
	UserMe      = "/api/v3/user/me"
	UserView    = "/api/v3/user/view"
	UserOffline = "/api/v3/user/offline"

	// Asset
	AssetCreate = "/api/v3/asset/create"

	// Voice
	VoiceJoin      = "/api/v3/voice/join"
	VoiceLeave     = "/api/v3/voice/leave"
	VoiceList      = "/api/v3/voice/list"
	VoiceKeepAlive = "/api/v3/voice/keep-alive"

	// Intimacy
	IntimacyUpdate = "/api/v3/intimacy/update"

	// Guild Emoji
	GuildEmojiList   = "/api/v3/guild-emoji/list"
	GuildEmojiCreate = "/api/v3/guild-emoji/create"
	GuildEmojiUpdate = "/api/v3/guild-emoji/update"
	GuildEmojiDelete = "/api/v3/guild-emoji/delete"

	// Invite
	InviteList   = "/api/v3/invite/list"
	InviteCreate = "/api/v3/invite/create"
	InviteDelete = "/api/v3/invite/delete"

	// Blacklist
	BlacklistList   = "/api/v3/blacklist/list"
	BlacklistCreate = "/api/v3/blacklist/create"
	BlacklistDelete = "/api/v3/blacklist/delete"

	// Badge
	BadgeList = "/api/v3/badge/list" // Assumed based on pattern

	// Game
	GameActivity = "/api/v3/game/activity" // Assumed

	// Thread (Post)
	ThreadList = "/api/v3/thread/list" // Assumed, verified create/reply exist
)
