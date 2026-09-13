package collect

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/ssh"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
)

// TestAgentAdmin 依赖内网 SSH 主机（10.8.0.3）与本机私钥，集成开关控制
func TestAgentAdmin(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	var (
		ctx     = context.TODO()
		options = append([]ssh.Option{},
			ssh.WithHost("10.8.0.3"),
			ssh.WithPrivateKeyPath("/Users/alomerry/.ssh/id_ed25519"),
			ssh.WithTimeout(5*time.Second),
		)
	)

	admin, err := NewAgentAdmin(ctx, options...)
	assert.NoError(t, err)
  defer admin.Close(ctx)
  
  _, err = admin.RegisterAgent(ctx)
  assert.NoError(t, err)
}
