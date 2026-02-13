package model

type Guild struct {
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Topic            string   `json:"topic"`
	UserId           string   `json:"user_id"`
	Icon             string   `json:"icon"`
	NotifyType       int      `json:"notify_type"`
	Region           string   `json:"region"`
	EnableOpen       bool     `json:"enable_open"`
	OpenId           string   `json:"open_id"`
	DefaultChannelId string   `json:"default_channel_id"`
	WelcomeChannelId string   `json:"welcome_channel_id"`
	BoostNum         int      `json:"boost_num"`
	Level            int      `json:"level"`
	Roles            []Role   `json:"roles,omitempty"`
	Channels         []Channel `json:"channels,omitempty"`
}

type Role struct {
	RoleId      int    `json:"role_id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Hoist       int    `json:"hoist"`
	Mentionable int    `json:"mentionable"`
	Permissions int    `json:"permissions"`
}

// Channel struct might be in channel.go, but for now defining minimal here or assuming it will be there.
// If channel.go defines it, I should use that or move it here.
// Let's check channel.go later. For now, I'll assume Channel struct is needed here or I can use map/any if cyclic dependency issues arise, but same package is fine.
// I will assume Channel is defined in channel.go or I will add it here if not.
// Actually, let's put Channel in channel.go, but for GuildView it returns channels.
// Since they are in the same package `model`, it's fine.

type GuildListResp struct {
	Items []Guild `json:"items"`
	Meta  Meta    `json:"meta"`
	Sort  Sort    `json:"sort"`
}

type GuildUser struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	IdentifyNum string `json:"identify_num"`
	Online      bool   `json:"online"`
	Status      int    `json:"status"`
	Bot         bool   `json:"bot"`
	Avatar      string `json:"avatar"`
	VipAvatar   string `json:"vip_avatar"`
	Nickname    string `json:"nickname"`
	Roles       []int  `json:"roles"`
}

type GuildUserListResp struct {
	Items        []GuildUser `json:"items"`
	Meta         Meta        `json:"meta"`
	Sort         Sort        `json:"sort"`
	UserCount    int         `json:"user_count"`
	OnlineCount  int         `json:"online_count"`
	OfflineCount int         `json:"offline_count"`
}
