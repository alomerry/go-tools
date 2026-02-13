package model

import "encoding/json"

type WebhookRequest struct {
	S       int          `json:"s"` // 信令类型
	D       *WebhookData `json:"d"` // 数据
	Encrypt string       `json:"encrypt,omitempty"`
}

type WebhookResponse struct {
	Challenge string `json:"challenge"`
}

type WebhookData struct {
	Type         int             `json:"type"`
	ChannelType  string          `json:"channel_type"`
	TargetId     string          `json:"target_id"`
	AuthorId     string          `json:"author_id"`
	Content      string          `json:"content"`
	MsgId        string          `json:"msg_id"`
	MsgTimestamp int64           `json:"msg_timestamp"`
	Nonce        string          `json:"nonce"`
	Extra        json.RawMessage `json:"extra"`        // 不同的消息类型，结构不一致
	VerifyToken  string          `json:"verify_token"` // 机器人的 verify token
	Challenge    string          `json:"challenge"`    // 客户端需要原样返回的 Challenge 值
}
