package kook

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/kook/client"
	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestChannelAPI(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		client := client.NewClient(client.WithToken(env.GetKookToken()))
		resp, err := client.ChannelService.View(context.TODO(), "7505882067043210", false)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
