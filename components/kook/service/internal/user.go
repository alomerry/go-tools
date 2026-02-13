package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type UserService struct {
	*BaseService
}

func (s *UserService) Me(ctx context.Context) (*model.UserMeResp, error) {
	resp, err := s.client.Get(ctx, kook.UserMe)
	if err != nil {
		return nil, err
	}

	var result model.UserMeResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *UserService) View(ctx context.Context, userId string, guildId string) (*model.UserViewResp, error) {
	params := map[string]string{
		"user_id": userId,
	}
	if guildId != "" {
		params["guild_id"] = guildId
	}

	resp, err := s.client.Get(ctx, kook.UserView, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.UserViewResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *UserService) Offline(ctx context.Context) error {
	resp, err := s.client.Post(ctx, kook.UserOffline)
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
