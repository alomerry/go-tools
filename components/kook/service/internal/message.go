package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type MessageService struct {
	*BaseService
}

func (s *MessageService) List(ctx context.Context, targetId string) (*model.MessageListResp, error) {
	// TODO: 如果需要，支持其他列表参数，如 msg_id, pin, flag, page_size
	params := map[string]string{
		"target_id": targetId,
	}
	resp, err := s.client.Get(ctx, kook.MessageList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.MessageListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MessageService) View(ctx context.Context, msgId string) (*model.Message, error) {
	params := map[string]string{
		"msg_id": msgId,
	}
	resp, err := s.client.Get(ctx, kook.MessageView, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.Message
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MessageService) Create(ctx context.Context, req model.CreateMessageRequest) (*model.CreateMessageResp, error) {
	resp, err := s.client.Post(ctx, kook.MessageCreate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.CreateMessageResp
	err = s.after(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MessageService) Update(ctx context.Context, req model.MessageUpdateReq) error {
	resp, err := s.client.Post(ctx, kook.MessageUpdate, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *MessageService) Delete(ctx context.Context, msgId string) error {
	req := model.MessageDeleteReq{
		MsgId: msgId,
	}
	resp, err := s.client.Post(ctx, kook.MessageDelete, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *MessageService) AddReaction(ctx context.Context, msgId, emoji string) error {
	req := model.AddReactionReq{
		MsgId: msgId,
		Emoji: emoji,
	}
	resp, err := s.client.Post(ctx, kook.MessageAddReaction, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *MessageService) DeleteReaction(ctx context.Context, msgId, emoji, userId string) error {
	req := model.DeleteReactionReq{
		MsgId:  msgId,
		Emoji:  emoji,
		UserId: userId,
	}
	resp, err := s.client.Post(ctx, kook.MessageDeleteReaction, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
