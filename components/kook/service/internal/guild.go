package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
	"github.com/spf13/cast"
)

type GuildService struct {
	*BaseService
}

func (s *GuildService) List(ctx context.Context, page, pageSize int, sort string) (*model.GuildListResp, error) {
	params := map[string]string{}
	if page > 0 {
		params["page"] = cast.ToString(page)
	}
	if pageSize > 0 {
		params["page_size"] = cast.ToString(pageSize)
	}
	if sort != "" {
		params["sort"] = sort
	}

	resp, err := s.client.Get(ctx, kook.GuildList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.GuildListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *GuildService) View(ctx context.Context, guildId string) (*model.Guild, error) {
	params := map[string]string{
		"guild_id": guildId,
	}
	resp, err := s.client.Get(ctx, kook.GuildView, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.Guild
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *GuildService) UserList(ctx context.Context, guildId string, page, pageSize int) (*model.GuildUserListResp, error) {
	params := map[string]string{
		"guild_id": guildId,
	}
	if page > 0 {
		params["page"] = cast.ToString(page)
	}
	if pageSize > 0 {
		params["page_size"] = cast.ToString(pageSize)
	}

	resp, err := s.client.Get(ctx, kook.GuildUserList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.GuildUserListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *GuildService) Leave(ctx context.Context, guildId string) error {
	body := map[string]string{
		"guild_id": guildId,
	}
	resp, err := s.client.Post(ctx, kook.GuildLeave, req2.WithBody(body))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *GuildService) Kickout(ctx context.Context, guildId, targetId string) error {
	body := map[string]string{
		"guild_id":  guildId,
		"target_id": targetId,
	}
	resp, err := s.client.Post(ctx, kook.GuildKickout, req2.WithBody(body))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
