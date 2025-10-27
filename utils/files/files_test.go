package files

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCreateFile(t *testing.T) {
	t.Run("test create files", func(t *testing.T) {
		fileName, err := CreateTempFile(context.TODO(), "123.txt", func(file *os.File) error {
			_, err := file.WriteString("666")
			return err
		})
		assert.Nil(t, err)
		assert.FileExists(t, fileName)
		os.Remove(fileName)
	})
}
