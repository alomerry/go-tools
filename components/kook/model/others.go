package model

// Emoji
type GuildEmoji struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	User User   `json:"user"`
}

type GuildEmojiListResp struct {
	Items []GuildEmoji `json:"items"`
}

// Invite
type Invite struct {
	UrlCode    string `json:"url_code"`
	Url        string `json:"url"`
	GuildId    string `json:"guild_id"`
	ChannelId  string `json:"channel_id"`
	ExpireTime int    `json:"expire_time"`
	PeopleUses int    `json:"people_uses"`
}

type InviteListResp struct {
	Items []Invite `json:"items"`
	Meta  Meta     `json:"meta"`
}

// Blacklist
type Blacklist struct {
	UserId   string `json:"user_id"`
	CreateAt int64  `json:"create_at"`
}

type BlacklistListResp struct {
	Items []Blacklist `json:"items"`
	Meta  Meta        `json:"meta"`
}

// Intimacy
type IntimacyUpdateReq struct {
	UserId      string `json:"user_id"`
	Score       int    `json:"score,omitempty"`
	SocialInfo  string `json:"social_info,omitempty"`
	ImgUrl      string `json:"img_url,omitempty"`
}

// Badge
type Badge struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type BadgeListResp struct {
	Items []Badge `json:"items"`
}

// Game
type GameActivityReq struct {
	Id       int    `json:"id"`
	DataType int    `json:"data_type"` // 1: 进程名称, 2: 游戏名称
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
}
