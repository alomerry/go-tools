package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

// EmojiService
type EmojiService struct {
	*BaseService
}

func (s *EmojiService) List(ctx context.Context, guildId string) (*model.GuildEmojiListResp, error) {
	params := map[string]string{"guild_id": guildId}
	resp, err := s.client.Get(ctx, kook.GuildEmojiList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}
	var result model.GuildEmojiListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InviteService
type InviteService struct {
	*BaseService
}

func (s *InviteService) List(ctx context.Context, guildId, channelId string) (*model.InviteListResp, error) {
	params := map[string]string{}
	if guildId != "" {
		params["guild_id"] = guildId
	}
	if channelId != "" {
		params["channel_id"] = channelId
	}
	resp, err := s.client.Get(ctx, kook.InviteList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}
	var result model.InviteListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BlacklistService
type BlacklistService struct {
	*BaseService
}

func (s *BlacklistService) List(ctx context.Context, guildId string) (*model.BlacklistListResp, error) {
	params := map[string]string{"guild_id": guildId}
	resp, err := s.client.Get(ctx, kook.BlacklistList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}
	var result model.BlacklistListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntimacyService
type IntimacyService struct {
	*BaseService
}

func (s *IntimacyService) Update(ctx context.Context, req model.IntimacyUpdateReq) error {
	resp, err := s.client.Post(ctx, kook.IntimacyUpdate, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

// BadgeService
type BadgeService struct {
	*BaseService
}

func (s *BadgeService) List(ctx context.Context, guildId string) (*model.BadgeListResp, error) {
	params := map[string]string{"guild_id": guildId} // 假设 guild_id 是必须的或可选的
	resp, err := s.client.Get(ctx, kook.BadgeList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}
	var result model.BadgeListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GameService
type GameService struct {
	*BaseService
}

func (s *GameService) UpdateActivity(ctx context.Context, req model.GameActivityReq) error {
	resp, err := s.client.Post(ctx, kook.GameActivity, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
