package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIPV4(t *testing.T) {
	t.Run("test ipv4", func(t *testing.T) {
		assert.True(t, IsIPV4("127.0.0.1"))
		assert.False(t, IsIPV4("127.0.0.1/8"))
		assert.True(t, IsIPV4("10.140.148.19"))
	})
}
