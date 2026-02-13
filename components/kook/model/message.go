package model

type Message struct {
	Id           string       `json:"id"`
	Type         int          `json:"type"`
	Content      string       `json:"content"`
	Mention      []string     `json:"mention"`
	MentionAll   bool         `json:"mention_all"`
	MentionRoles []int        `json:"mention_roles"` // API docs say array of role IDs
	MentionHere  bool         `json:"mention_here"`
	Embeds       []any        `json:"embeds"` // 复杂结构，暂时使用 any
	Attachments  any          `json:"attachments"`
	CreateAt     int64        `json:"create_at"`
	UpdatedAt    int64        `json:"updated_at"`
	Reactions    []Reaction   `json:"reactions"`
	Author       User         `json:"author"`
	ImageName    string       `json:"image_name"`
	ReadStatus   bool         `json:"read_status"`
	Quote        *Quote       `json:"quote"`
	MentionInfo  *MentionInfo `json:"mention_info"`
}

type User struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	IdentifyNum string `json:"identify_num,omitempty"`
	Online      bool   `json:"online"`
	Avatar      string `json:"avatar"`
	Bot         bool   `json:"bot"`
	Nickname    string `json:"nickname,omitempty"`
}

type Reaction struct {
	Emoji Emoji `json:"emoji"`
	Count int   `json:"count"`
	Me    bool  `json:"me"`
}

type Emoji struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Quote struct {
	Id       string `json:"id"`
	Type     int    `json:"type"`
	Content  string `json:"content"`
	CreateAt int64  `json:"create_at"`
	Author   User   `json:"author"`
}

type MentionInfo struct {
	MentionPart     []MentionUser `json:"mention_part"`
	MentionRolePart []Role        `json:"mention_role_part"`
}

type MentionUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar"`
}

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

type MessageListResp struct {
	Items []Message `json:"items"`
}

type MessageUpdateReq struct {
	MsgId        string `json:"msg_id"`
	Content      string `json:"content"`
	Quote        string `json:"quote,omitempty"`
	TempTargetId string `json:"temp_target_id,omitempty"`
}

type MessageDeleteReq struct {
	MsgId string `json:"msg_id"`
}

type AddReactionReq struct {
	MsgId string `json:"msg_id"`
	Emoji string `json:"emoji"`
}

type DeleteReactionReq struct {
	MsgId  string `json:"msg_id"`
	Emoji  string `json:"emoji"`
	UserId string `json:"user_id,omitempty"`
}
