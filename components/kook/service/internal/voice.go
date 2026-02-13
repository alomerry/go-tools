package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type VoiceService struct {
	*BaseService
}

func (s *VoiceService) Join(ctx context.Context, channelId, password string) (*model.JoinVoiceResp, error) {
	req := model.JoinVoiceReq{
		ChannelId: channelId,
		Password:  password,
	}
	resp, err := s.client.Post(ctx, kook.VoiceJoin, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.JoinVoiceResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *VoiceService) Leave(ctx context.Context, channelId string) error {
	req := model.LeaveVoiceReq{
		ChannelId: channelId,
	}
	resp, err := s.client.Post(ctx, kook.VoiceLeave, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *VoiceService) List(ctx context.Context) (*model.VoiceListResp, error) {
	resp, err := s.client.Get(ctx, kook.VoiceList)
	if err != nil {
		return nil, err
	}

	var result model.VoiceListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *VoiceService) KeepAlive(ctx context.Context, channelId string) error {
	// KeepAlive 实现通常涉及 WebSocket 或定期 ping，
	// 但如果存在用于保持会话的 HTTP 端点，这里将其映射到该端点。
	// 根据文档，存在 /api/v3/voice/keep-alive
	params := map[string]string{
		"channel_id": channelId,
	}
	resp, err := s.client.Get(ctx, kook.VoiceKeepAlive, req2.WithQueryParam(params))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
