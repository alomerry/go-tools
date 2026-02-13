package kook

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/kook/service"
	"github.com/stretchr/testify/assert"
)

// MockClient 简单的 Mock Client 实现，仅用于编译通过性测试
// 实际单元测试应使用 Mock 框架或更复杂的 Mock 逻辑
type MockClient struct{}

func (m *MockClient) Get(ctx context.Context, url string, opts ...any) (http.Response, error) {
	return nil, nil
}
func (m *MockClient) Post(ctx context.Context, url string, opts ...any) (http.Response, error) {
	return nil, nil
}
func (m *MockClient) Close(ctx context.Context) error { return nil }

func TestServices(t *testing.T) {
	// 仅验证 Service 工厂函数和接口定义是否正确
	// 实际逻辑测试需要 Mock http.Client 的行为
	// 这里主要确保编译通过和基本结构正确
	var client http.Client // 接口类型，通常我们会传递一个模拟实现

	channelSvc := service.NewChannelService(client)
	assert.NotNil(t, channelSvc)

	guildSvc := service.NewGuildService(client)
	assert.NotNil(t, guildSvc)

	messageSvc := service.NewMessageService(client)
	assert.NotNil(t, messageSvc)

	dmSvc := service.NewDirectMessageService(client)
	assert.NotNil(t, dmSvc)

	userSvc := service.NewUserService(client)
	assert.NotNil(t, userSvc)

	roleSvc := service.NewRoleService(client)
	assert.NotNil(t, roleSvc)

	assetSvc := service.NewAssetService(client)
	assert.NotNil(t, assetSvc)

	voiceSvc := service.NewVoiceService(client)
	assert.NotNil(t, voiceSvc)

	emojiSvc := service.NewEmojiService(client)
	assert.NotNil(t, emojiSvc)

	inviteSvc := service.NewInviteService(client)
	assert.NotNil(t, inviteSvc)

	blacklistSvc := service.NewBlacklistService(client)
	assert.NotNil(t, blacklistSvc)

	intimacySvc := service.NewIntimacyService(client)
	assert.NotNil(t, intimacySvc)

	badgeSvc := service.NewBadgeService(client)
	assert.NotNil(t, badgeSvc)

	gameSvc := service.NewGameService(client)
	assert.NotNil(t, gameSvc)
}
