package webhook

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

func TestHandler_Handle(t *testing.T) {
	h := NewHandler("test-key", "test-token")
	h.OnText(func(ctx context.Context, data *model.WebhookData) error {
		if data.Content != "hello" {
			t.Errorf("expected content 'hello', got %s", data.Content)
		}
		return nil
	})

	req := &model.WebhookRequest{
		S: 0,
		D: &model.WebhookData{
			Type:    kook.EventText,
			Content: "hello",
		},
	}

	_, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
}

func TestHandler_Challenge(t *testing.T) {
	h := NewHandler("test-key", "test-token")
	req := &model.WebhookRequest{
		S: 0,
		D: &model.WebhookData{
			Type:        kook.EventSystem,
			ChannelType: kook.ChannelTypeWebhookChallenge,
			Challenge:   "challenge-code",
			VerifyToken: "test-token",
		},
	}

	res, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	if res.Challenge != "challenge-code" {
		t.Errorf("expected challenge 'challenge-code', got %s", res.Challenge)
	}
}
