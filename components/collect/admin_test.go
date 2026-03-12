package collect

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/ssh"
	"github.com/stretchr/testify/assert"
)

func TestAgentAdmin(t *testing.T) {
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
