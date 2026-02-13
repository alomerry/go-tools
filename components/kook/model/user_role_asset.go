package model

type UserMeResp = User // Reusing User struct

type UserViewResp = User

type AssetCreateReq struct {
	File []byte `json:"file"` // 需要特殊处理 multipart
}

type AssetCreateResp struct {
	Url string `json:"url"`
}

type GuildRole struct {
	RoleId      int    `json:"role_id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Hoist       int    `json:"hoist"`
	Mentionable int    `json:"mentionable"`
	Permissions int    `json:"permissions"`
}

type GuildRoleListResp struct {
	Items []GuildRole `json:"items"`
}

type CreateGuildRoleReq struct {
	GuildId string `json:"guild_id"`
	Name    string `json:"name,omitempty"`
}

type UpdateGuildRoleReq struct {
	GuildId     string `json:"guild_id"`
	RoleId      int    `json:"role_id"`
	Name        string `json:"name,omitempty"`
	Color       int    `json:"color,omitempty"`
	Position    int    `json:"position,omitempty"`
	Hoist       int    `json:"hoist,omitempty"`
	Mentionable int    `json:"mentionable,omitempty"`
	Permissions int    `json:"permissions,omitempty"`
}

type DeleteGuildRoleReq struct {
	GuildId string `json:"guild_id"`
	RoleId  int    `json:"role_id"`
}

type GrantRoleReq struct {
	GuildId string `json:"guild_id"`
	RoleId  int    `json:"role_id"`
	UserId  string `json:"user_id"`
}

type RevokeRoleReq struct {
	GuildId string `json:"guild_id"`
	RoleId  int    `json:"role_id"`
	UserId  string `json:"user_id"`
}
