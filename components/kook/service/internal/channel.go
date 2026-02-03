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

func (c *ChannelService) View(ctx context.Context, targetId string, needChildren bool) (*model.ViewChannelResp, error) {
	// 设置查询参数
	param := map[string]string{
		"target_id":     targetId,
		"need_children": cast.ToString(needChildren),
	}

	resp, err := c.client.Get(ctx, kook.ChannelView, req2.WithQueryParam(param))
	if err != nil {
		return nil, err
	}

	// 执行请求
	var result model.ViewChannelResp
	err = c.after(resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
