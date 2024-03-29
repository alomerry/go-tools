package files

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGetFileName(t *testing.T) {
	t.Run("file name has dot", func(t *testing.T) {
		assert.Equal(t, "staging.test.csv", GetFileName("/root/run/staging.test.csv"))
	})
}

func TestGetFileType(t *testing.T) {
	t.Run("file name has dot", func(t *testing.T) {
		assert.Equal(t, "csv", GetFileType("csv"))
	})
}
