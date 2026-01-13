package service

import (
	"context"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/components/kook/service/internal"
	"resty.dev/v3"
)

type ChannelService interface {
	View(context.Context, string, bool) (*model.ViewChannelResp, error)
}

type MessageService interface {
	Create(context.Context, model.CreateMessageRequest) (*model.CreateMessageResp, error)
}

func NewChannelService(client *resty.Client) ChannelService {
	return &internal.ChannelService{
		BaseService: internal.NewBaseService(client),
	}
}

func NewMessageService(client *resty.Client) MessageService {
	return &internal.MessageService{
		BaseService: internal.NewBaseService(client),
	}
}
