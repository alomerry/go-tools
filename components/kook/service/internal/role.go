package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type RoleService struct {
	*BaseService
}

func (s *RoleService) List(ctx context.Context, guildId string) (*model.GuildRoleListResp, error) {
	params := map[string]string{
		"guild_id": guildId,
	}
	resp, err := s.client.Get(ctx, kook.GuildRoleList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.GuildRoleListResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RoleService) Create(ctx context.Context, req model.CreateGuildRoleReq) (*model.GuildRole, error) {
	resp, err := s.client.Post(ctx, kook.GuildRoleCreate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.GuildRole
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RoleService) Update(ctx context.Context, req model.UpdateGuildRoleReq) (*model.GuildRole, error) {
	resp, err := s.client.Post(ctx, kook.GuildRoleUpdate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.GuildRole
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RoleService) Delete(ctx context.Context, guildId string, roleId int) error {
	req := model.DeleteGuildRoleReq{
		GuildId: guildId,
		RoleId:  roleId,
	}
	resp, err := s.client.Post(ctx, kook.GuildRoleDelete, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *RoleService) Grant(ctx context.Context, guildId, userId string, roleId int) error {
	req := model.GrantRoleReq{
		GuildId: guildId,
		RoleId:  roleId,
		UserId:  userId,
	}
	resp, err := s.client.Post(ctx, kook.GuildRoleGrant, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}

func (s *RoleService) Revoke(ctx context.Context, guildId, userId string, roleId int) error {
	req := model.RevokeRoleReq{
		GuildId: guildId,
		RoleId:  roleId,
		UserId:  userId,
	}
	resp, err := s.client.Post(ctx, kook.GuildRoleRevoke, req2.WithBody(req))
	if err != nil {
		return err
	}
	return s.after(resp, nil)
}
