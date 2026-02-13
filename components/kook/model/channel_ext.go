package model

type ChannelListResp struct {
	Items []Channel `json:"items"`
	Meta  Meta      `json:"meta"`
	Sort  Sort      `json:"sort"`
}

type CreateChannelReq struct {
	GuildId      string `json:"guild_id"`
	Name         string `json:"name"`
	Type         int    `json:"type,omitempty"`
	ParentId     string `json:"parent_id,omitempty"`
	LimitAmount  int    `json:"limit_amount,omitempty"`
	VoiceQuality string `json:"voice_quality,omitempty"`
	IsCategory   int    `json:"is_category,omitempty"` // 1 是, 0 否
}

type UpdateChannelReq struct {
	ChannelId    string `json:"channel_id"`
	Name         string `json:"name,omitempty"`
	Topic        string `json:"topic,omitempty"`
	SlowMode     int    `json:"slow_mode,omitempty"`
	LimitAmount  int    `json:"limit_amount,omitempty"`
	VoiceQuality string `json:"voice_quality,omitempty"`
	Password     string `json:"password,omitempty"`
}

type DeleteChannelReq struct {
	ChannelId string `json:"channel_id"`
}

type ChannelUserListResp struct {
	Items []ChannelUser `json:"items"`
	Meta  Meta          `json:"meta"`
}

type ChannelUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Online   bool   `json:"online"`
}

type MoveUserReq struct {
	TargetId string   `json:"target_id"`
	UserIds  []string `json:"user_ids"`
}

type KickoutUserReq struct {
	TargetId string   `json:"target_id"`
	UserIds  []string `json:"user_ids"`
}
