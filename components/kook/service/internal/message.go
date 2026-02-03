package internal

import (
	"context"
	"encoding/json"

	"github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type MessageService struct {
	*BaseService
}

func (m *MessageService) Create(ctx context.Context, param model.CreateMessageRequest) (*model.CreateMessageResp, error) {
	body, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	var result model.CreateMessageResp

	res, err := m.client.Post(ctx, kook.MessageCreate, req.WithBody(body))
	if err != nil {
		return nil, err
	}

	err = m.after(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
