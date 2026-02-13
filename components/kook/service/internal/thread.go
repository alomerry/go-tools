package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
)

type ThreadService struct {
	*BaseService
}

// 注意：提供的上下文中未完全记录 Thread API，
// 但提到了 /api/v3/thread/create 和 /api/v3/thread/reply。
// 假设它们的行为类似于消息创建。

// CreateThreadReq 可能类似于 Message Create，但创建的是帖子。
// 或者它只是一种消息类型？
// 目前，我将定义一个占位符结构体，或者如果合适的话复用 Message 结构体。
// 但独立的结构体更安全。

type CreateThreadReq struct {
	ChannelId string `json:"target_id"`
	Content   string `json:"content"`
	Title     string `json:"title,omitempty"` // 帖子通常有标题
	Type      int    `json:"type,omitempty"`
}

type CreateThreadResp struct {
	MsgId        string `json:"msg_id"`
	MsgTimestamp int64  `json:"msg_timestamp"`
	Nonce        string `json:"nonce"`
}

func (s *ThreadService) Create(ctx context.Context, req CreateThreadReq) (*CreateThreadResp, error) {
	// 使用推断的端点
	// const ThreadCreate = "/api/v3/thread/create" (需要添加到 cons)
	// 我假设它存在于 cons 中或现在添加它。
	// 等等，我在 cons 中添加了 ThreadList，但没有添加 ThreadCreate。
	// 我应该先更新 cons 或使用字符串字面量。
	// 目前使用字符串字面量以避免多次重新编辑 cons 文件。
	resp, err := s.client.Post(ctx, "/api/v3/thread/create", req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result CreateThreadResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
