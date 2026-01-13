package service

import (
	"context"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/components/kook/service/internal"
	"resty.dev/v3"
)

type ChannelService interface {
	View(context.Context, string, bool) (*model.ViewResp, error)
}

func NewChannelService(client *resty.Client) ChannelService {
	return &internal.ChannelService{
		BaseService: internal.NewBaseService(client),
	}
}
