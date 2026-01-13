package internal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type MessageService struct {
	*BaseService
}

func (m *MessageService) Create(ctx context.Context, param model.CreateMessageRequest) (*model.CreateMessageResp, error) {
	req := m.getRequest(ctx)
	body, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	req.SetBody(body)

	var result model.CreateMessageResp
	_, err = m.execute(req, http.MethodPost, kook.MessageCreate, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
