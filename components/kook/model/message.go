package model

type CreateMessageRequest struct {
	MessageType  int    `json:"type"`                     // 否 POST 消息类型, 见[type], 不传默认为 9, 代表kmarkdown。 9 代表 kmarkdown 消息, 10 代表卡片消息。
	TargetId     string `json:"target_id"`                // 是 POST 目标频道 id
	Content      string `json:"content"`                  // 是 POST 消息内容
	Quote        string `json:"quote,omitempty"`          // 否 POST 回复某条消息的 msgId
	Nonce        string `json:"nonce,omitempty"`          // 否 POST nonce, 服务端不做处理, 原样返回
	TempTargetId string `json:"temp_target_id,omitempty"` // 否 POST 用户 id,如果传了，代表该消息是临时消息，该消息不会存数据库，但是会在频道内只给该用户推送临时消息。用于在频道内针对用户的操作进行单独的回应通知等。
	TemplateId   string `json:"template_id,omitempty"`    // 否 POST 模板消息id, 如果使用了，content会作为模板消息的input，参见模板消息
	ReplyMsgId   string `json:"reply_msg_id,omitempty"`   // 否 POST 当前消息回复的用户5分钟内发送到相同频道的消息的msg_id，如果是当前开发者的第一次回复，扣减当日发送量按n分之一条计算
}

type CreateMessageResp struct {
	MsgId        string `json:"msg_id"`
	MsgTimestamp int64  `json:"msg_timestamp"`
	Nonce        string `json:"nonce"`
}
