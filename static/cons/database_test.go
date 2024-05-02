package cons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDbDsn(t *testing.T) {
	var dsn = "mysql://aaa:bbb@example.com:3306/xxx"
	info, err := ParseDbDsn(dsn)
	assert.Nil(t, err)
	assert.Equal(t, "aaa", info.User)
	assert.Equal(t, "bbb", info.Password)
	assert.Equal(t, "example.com", info.Host)
	assert.Equal(t, "3306", info.Port)
}
