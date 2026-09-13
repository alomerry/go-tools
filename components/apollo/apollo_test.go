package apollo

import (
  "testing"

  "github.com/alomerry/go-tools/test"
  "github.com/stretchr/testify/assert"
)

// TestGet 依赖真实 Apollo 集群（写死的 appId/key），集成开关控制
func TestGet(t *testing.T) {
  if !test.IntegrationEnabled() {
    t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
  }
  Init("***", "test")
	val, err := Get("***")
	assert.Nil(t, err)
	assert.NotZero(t, val)
}
