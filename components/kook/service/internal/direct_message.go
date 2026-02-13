package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type DirectMessageService struct {
	*BaseService
}

func (s *DirectMessageService) List(ctx context.Context, chatCode, targetId string) (*model.DirectMessageListResp, error) {
	params := map[string]string{}
	if chatCode != "" {
		params["chat_code"] = chatCode
	}
	if targetId != "" {
		params["target_id"] = targetId
	}

	resp, err := s.client.Get(ctx, kook.DirectMessageList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.DirectMessageListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DirectMessageService) View(ctx context.Context, msgId, chatCode string) (*model.DirectMessage, error) {
	params := map[string]string{
		"msg_id":    msgId,
		"chat_code": chatCode,
	}
	resp, err := s.client.Get(ctx, kook.DirectMessageView, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.DirectMessage
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DirectMessageService) Create(ctx context.Context, req model.CreateDirectMessageReq) (*model.CreateDirectMessageResp, error) {
	resp, err := s.client.Post(ctx, kook.DirectMessageCreate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.CreateDirectMessageResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DirectMessageService) Update(ctx context.Context, req model.UpdateDirectMessageReq) error {
	resp, err := s.client.Post(ctx, kook.DirectMessageUpdate, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *DirectMessageService) Delete(ctx context.Context, msgId string) error {
	req := model.DeleteDirectMessageReq{
		MsgId: msgId,
	}
	resp, err := s.client.Post(ctx, kook.DirectMessageDelete, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *DirectMessageService) AddReaction(ctx context.Context, msgId, emoji string) error {
	req := model.AddReactionReq{ // Reuse Message AddReactionReq
		MsgId: msgId,
		Emoji: emoji,
	}
	resp, err := s.client.Post(ctx, kook.DirectMessageAddReaction, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *DirectMessageService) DeleteReaction(ctx context.Context, msgId, emoji string) error {
	req := model.DeleteReactionReq{ // 复用 Message DeleteReactionReq
		MsgId: msgId,
		Emoji: emoji,
	}
	resp, err := s.client.Post(ctx, kook.DirectMessageDeleteReaction, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
