package model

type ViewChannelOption struct {
	NeedChildren bool
}

type ViewChannelResp struct {
	Id             string `json:"id,omitempty"`              // 频道 id
	GuildId        string `json:"guild_id,omitempty"`        // 服务器 id
	UserId         string `json:"user_id,omitempty"`         // 频道创建者 id
	ParentId       string `json:"parent_id,omitempty"`       // 父分组频道 id
	Name           string `json:"name,omitempty"`            // 频道名称
	Topic          string `json:"topic,omitempty"`           // 频道简介
	ChannelType    int    `json:"type,omitempty"`            // 频道类型，0 分组，1 文字，2 语音
	Level          int    `json:"level,omitempty"`           // 频道排序
	SlowMode       int    `json:"slow_mode,omitempty"`       // 慢速限制，用户发送消息之后再次发送消息的等待时间，单位秒
	HasPassword    bool   `json:"has_password,omitempty"`    // 是否已设置密码
	LimitAmount    int    `json:"limit_amount,omitempty"`    // 人数限制
	IsCategory     bool   `json:"is_category,omitempty"`     // 是否为分组类型
	PermissionSync int    `json:"permission_sync,omitempty"` // 是否与分组频道同步权限
	// PermissionOverwrites array  `json:"permission_overwrites,omitempty"` // 针对角色的频道权限覆盖
	// PermissionUsers      array  `json:"permission_users,omitempty"`      // 针对用户的频道权限覆盖
	VoiceQuality string   `json:"voice_quality,omitempty"` // 语音频道质量级别，1 流畅，2 正常，3 高质量
	ServerUrl    string   `json:"server_url,omitempty"`    // 语音服务器地址，HOST:PORT/PATH 的格式
	Children     []string `json:"children,omitempty"`      // 子频道的 id 列表
}
