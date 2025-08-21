package apollo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	Init("***")
	val, err := Get("***")
	assert.Nil(t, err)
	assert.NotZero(t, val)
}
