package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
	"github.com/spf13/cast"
)

type ChannelService struct {
	*BaseService
}

func (c *ChannelService) List(ctx context.Context, guildId string, page, pageSize int, channelType int) (*model.ChannelListResp, error) {
	params := map[string]string{
		"guild_id": guildId,
	}
	if page > 0 {
		params["page"] = cast.ToString(page)
	}
	if pageSize > 0 {
		params["page_size"] = cast.ToString(pageSize)
	}
	if channelType > 0 {
		params["type"] = cast.ToString(channelType)
	}

	resp, err := c.client.Get(ctx, kook.ChannelList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.ChannelListResp
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ChannelService) View(ctx context.Context, targetId string, needChildren bool) (*model.ViewChannelResp, error) {
	param := map[string]string{
		"target_id":     targetId,
		"need_children": cast.ToString(needChildren),
	}

	resp, err := c.client.Get(ctx, kook.ChannelView, req2.WithQueryParam(param))
	if err != nil {
		return nil, err
	}

	var result model.ViewChannelResp
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *ChannelService) Create(ctx context.Context, req model.CreateChannelReq) (*model.Channel, error) {
	resp, err := c.client.Post(ctx, kook.ChannelCreate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.Channel
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ChannelService) Update(ctx context.Context, req model.UpdateChannelReq) (*model.Channel, error) {
	resp, err := c.client.Post(ctx, kook.ChannelUpdate, req2.WithBody(req))
	if err != nil {
		return nil, err
	}

	var result model.Channel
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ChannelService) Delete(ctx context.Context, channelId string) error {
	req := model.DeleteChannelReq{
		ChannelId: channelId,
	}
	resp, err := c.client.Post(ctx, kook.ChannelDelete, req2.WithBody(req))
	if err != nil {
		return err
	}
	return c.after(resp, nil)
}

func (c *ChannelService) UserList(ctx context.Context, channelId string) (*model.ChannelUserListResp, error) {
	params := map[string]string{
		"channel_id": channelId,
	}
	resp, err := c.client.Get(ctx, kook.ChannelUserList, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.ChannelUserListResp
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ChannelService) MoveUser(ctx context.Context, channelId string, userIds ...string) error {
	req := model.MoveUserReq{
		TargetId: channelId,
		UserIds:  userIds,
	}
	resp, err := c.client.Post(ctx, kook.ChannelMoveUser, req2.WithBody(req))
	if err != nil {
		return err
	}
	return c.after(resp, nil)
}

func (c *ChannelService) Kickout(ctx context.Context, channelId string, userIds ...string) error {
	req := model.KickoutUserReq{
		TargetId: channelId,
		UserIds:  userIds,
	}
	resp, err := c.client.Post(ctx, kook.ChannelKickout, req2.WithBody(req))
	if err != nil {
		return err
	}
	return c.after(resp, nil)
}

func (c *ChannelService) GetJoinedChannel(ctx context.Context, guildId, userId string, page, pageSize int) (*model.ChannelListResp, error) {
	params := map[string]string{
		"guild_id": guildId,
		"user_id":  userId,
	}
	if page > 0 {
		params["page"] = cast.ToString(page)
	}
	if pageSize > 0 {
		params["page_size"] = cast.ToString(pageSize)
	}

	resp, err := c.client.Get(ctx, kook.ChannelUserGetJoined, req2.WithQueryParam(params))
	if err != nil {
		return nil, err
	}

	var result model.ChannelListResp
	if err := c.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
