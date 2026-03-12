package apollo

import (
  "testing"
  
  "github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
  Init("***", "test")
	val, err := Get("***")
	assert.Nil(t, err)
	assert.NotZero(t, val)
}
