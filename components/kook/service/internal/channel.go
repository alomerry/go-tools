package internal

import (
	"context"
	"net/http"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
	"github.com/spf13/cast"
)

type ChannelService struct {
	*BaseService
}

func (c *ChannelService) View(ctx context.Context, targetId string, needChildren bool) (*model.ViewChannelResp, error) {
	// 创建请求
	req := c.getRequest(ctx)

	// 设置查询参数
	req.SetQueryParams(map[string]string{
		"target_id":     targetId,
		"need_children": cast.ToString(needChildren),
	})

	// 执行请求
	var result model.ViewChannelResp
	_, err := c.execute(req, http.MethodGet, kook.ChannelView, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
