package service

import (
	"context"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/components/kook/service/internal"
)

type ChannelService interface {
	View(context.Context, string, bool) (*model.ViewChannelResp, error)
}

type MessageService interface {
	Create(context.Context, model.CreateMessageRequest) (*model.CreateMessageResp, error)
}

func NewChannelService(client http.Client) ChannelService {
	return &internal.ChannelService{
		BaseService: internal.NewBaseService(client),
	}
}

func NewMessageService(client http.Client) MessageService {
	return &internal.MessageService{
		BaseService: internal.NewBaseService(client),
	}
}
