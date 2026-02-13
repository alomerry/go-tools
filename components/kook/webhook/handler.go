package webhook

import (
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
	"github.com/alomerry/go-tools/utils/crypto"
)

type EventHandler func(ctx context.Context, data *model.WebhookData) error

type Handler struct {
	encryptKey  string
	verifyToken string

	textHandler      EventHandler
	imageHandler     EventHandler
	videoHandler     EventHandler
	fileHandler      EventHandler
	audioHandler     EventHandler
	kmarkdownHandler EventHandler
	cardHandler      EventHandler
	systemHandler    EventHandler
	defaultHandler   EventHandler
}

func NewHandler(encryptKey, verifyToken string) *Handler {
	return &Handler{
		encryptKey:  encryptKey,
		verifyToken: verifyToken,
	}
}

func (h *Handler) Handle(ctx context.Context, req *model.WebhookRequest) (*model.WebhookResponse, error) {
	if req.Encrypt != "" {
		if err := h.decryptReq(req); err != nil {
			return nil, err
		}
	}

	if req.S != 0 {
		return &model.WebhookResponse{}, nil
	}

	if req.D == nil {
		return nil, errors.New("callback empty data")
	}

	// Handle Challenge (System Event with WebhookChallenge type)
	if req.D.Type == kook.EventSystem && req.D.ChannelType == kook.ChannelTypeWebhookChallenge {
		if req.D.VerifyToken == h.verifyToken {
			return &model.WebhookResponse{Challenge: req.D.Challenge}, nil
		}
		return nil, errors.New("verify token mismatch")
	}

	var err error
	switch req.D.Type {
	case kook.EventText:
		if h.textHandler != nil {
			err = h.textHandler(ctx, req.D)
		}
	case kook.EventImage:
		if h.imageHandler != nil {
			err = h.imageHandler(ctx, req.D)
		}
	case kook.EventVideo:
		if h.videoHandler != nil {
			err = h.videoHandler(ctx, req.D)
		}
	case kook.EventFile:
		if h.fileHandler != nil {
			err = h.fileHandler(ctx, req.D)
		}
	case kook.EventAudio:
		if h.audioHandler != nil {
			err = h.audioHandler(ctx, req.D)
		}
	case kook.EventKMarkdown:
		if h.kmarkdownHandler != nil {
			err = h.kmarkdownHandler(ctx, req.D)
		}
	case kook.EventCard:
		if h.cardHandler != nil {
			err = h.cardHandler(ctx, req.D)
		}
	case kook.EventSystem:
		if h.systemHandler != nil {
			err = h.systemHandler(ctx, req.D)
		}
	default:
		if h.defaultHandler != nil {
			err = h.defaultHandler(ctx, req.D)
		}
	}

	if err != nil {
		return nil, err
	}

	return &model.WebhookResponse{}, nil
}

func (h *Handler) OnText(handler EventHandler) {
	h.textHandler = handler
}

func (h *Handler) OnImage(handler EventHandler) {
	h.imageHandler = handler
}

func (h *Handler) OnVideo(handler EventHandler) {
	h.videoHandler = handler
}

func (h *Handler) OnFile(handler EventHandler) {
	h.fileHandler = handler
}

func (h *Handler) OnAudio(handler EventHandler) {
	h.audioHandler = handler
}

func (h *Handler) OnKMarkdown(handler EventHandler) {
	h.kmarkdownHandler = handler
}

func (h *Handler) OnCard(handler EventHandler) {
	h.cardHandler = handler
}

func (h *Handler) OnSystem(handler EventHandler) {
	h.systemHandler = handler
}

func (h *Handler) OnEvent(handler EventHandler) {
	h.defaultHandler = handler
}

func (h *Handler) decryptReq(req *model.WebhookRequest) error {
	res, err := h.decryptKook(req.Encrypt)
	if err != nil {
		return err
	}

	return json.Unmarshal(res, req)
}

func (h *Handler) decryptKook(str string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	if len(raw) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := raw[:aes.BlockSize]
	encryptedBase64 := raw[aes.BlockSize:]
	ciphertext, err := base64.StdEncoding.DecodeString(string(encryptedBase64))
	if err != nil {
		return nil, err
	}

	key := h.encryptKey
	for len(key) < 32 {
		key += "\x00"
	}

	return crypto.DecryptAES256CBC(ciphertext, []byte(key), iv)
}
