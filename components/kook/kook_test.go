package kook

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/kook/client"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestChannelAPI(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		cli := client.NewClient(client.WithToken(env.GetKookToken()))
		resp, err := cli.ChannelService.View(context.TODO(), "7505882067043210", false)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestMessageAPI(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		cli := client.NewClient(client.WithToken(env.GetKookToken()))
		resp, err := cli.MessageService.Create(context.TODO(), model.CreateMessageRequest{
			MessageType: 10,
			// GuildId:     "9914224083754636",
			TargetId: "2698426473813722",
			Content:  `[{"type":"card","theme":"danger","modules":[{"type":"section","text":{"type":"kmarkdown","content":"[**(font)严重(font)[danger]**] **account 异常数告警**\n**信息**：(ins)[/backend.account.AccountService/KookHook] failed, err: crypto/aes: invalid key size 33(ins)\n**条件**：当前值 [2] >= 阈值 [2]\n**次数**：Total=2\t[10.42.0.220=2]\n(met)193563334(met)"}},{"type":"context","elements":[{"type":"plain-text","content":"2026-01-13 15:05:06"}]}]}]`,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
