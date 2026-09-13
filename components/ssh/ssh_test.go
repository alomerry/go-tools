package ssh

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
)

var (
	options []Option
)

func init() {
	options = append(options,
		WithHost("10.8.0.3"),
		WithPrivateKeyPath("/Users/alomerry/.ssh/id_ed25519"),
		WithTimeout(5*time.Second),
	)
}

// 依赖内网 SSH 主机（10.8.0.3）与本机私钥，集成开关控制
func TestSSHClient_PrivateKeyAuth(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	var (
		ctx     = context.TODO()
		tc, err = NewClient(ctx, options...)
	)

	err = tc.Connect()
	assert.NoError(t, err)
	defer tc.Close()

	// output, err := tc.RunCommand(ctx, "echo hello")
	// assert.NoError(t, err)
	// assert.Equal(t, "hello\n", output)
  //
	// uptime, err := tc.GetUptime(ctx)
	// assert.NoError(t, err)
	// assert.Equal(t, "up 1 hour, 30 minutes", uptime)
}

func TestSSHClient_AuthFail(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
 
  
  var (
    ctx     = context.TODO()
    opts = append([]Option{},
      WithHost("10.8.0.3"),
      WithPassword("wrongpassword"),
      WithTimeout(5*time.Second),
    )
    tc, err = NewClient(ctx, opts...)
  )

	err = tc.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ssh: handshake failed")
}
