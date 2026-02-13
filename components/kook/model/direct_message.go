package model

type DirectMessage struct {
	Id         string     `json:"id"`
	Type       int        `json:"type"`
	Content    string     `json:"content"`
	AuthorId   string     `json:"author_id"`
	CreateAt   int64      `json:"create_at"`
	Reactions  []Reaction `json:"reactions"`
	ReadStatus bool       `json:"read_status"`
}

type DirectMessageListResp struct {
	Items []DirectMessage `json:"items"`
}

type CreateDirectMessageReq struct {
	Type     int    `json:"type,omitempty"`
	TargetId string `json:"target_id,omitempty"` // 用户 ID
	ChatCode string `json:"chat_code,omitempty"`
	Content  string `json:"content"`
	Quote    string `json:"quote,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
}

type CreateDirectMessageResp struct {
	MsgId        string `json:"msg_id"`
	MsgTimestamp int64  `json:"msg_timestamp"`
	Nonce        string `json:"nonce"`
}

type UpdateDirectMessageReq struct {
	MsgId   string `json:"msg_id"`
	Content string `json:"content"`
	Quote   string `json:"quote,omitempty"`
}

type DeleteDirectMessageReq struct {
	MsgId string `json:"msg_id"`
}
