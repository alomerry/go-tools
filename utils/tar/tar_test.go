package tar

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// buildTestArchive 构造测试用 tar.gz 归档，entries 为 (归档名, 内容) 对
func buildTestArchive(t *testing.T, entries map[string]string) string {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "test.tar.gz")
	f, err := os.Create(archivePath)
	assert.Nil(t, err)
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for name, content := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		assert.Nil(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(content))
		assert.Nil(t, err)
	}
	return archivePath
}

func TestUntar(t *testing.T) {
	archivePath := buildTestArchive(t, map[string]string{
		"hello.txt":          "hello go-tools",
		"sub/nested/foo.txt": "nested content",
	})

	dst := t.TempDir()
	assert.Nil(t, UnTar(archivePath, dst))

	content, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "hello go-tools", string(content))

	content, err = os.ReadFile(filepath.Join(dst, "sub", "nested", "foo.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "nested content", string(content))
}

func TestUntarRejectsPathTraversal(t *testing.T) {
	// Zip Slip 回归：条目名含 .. 逃逸必须被拒绝，不得落盘
	archivePath := buildTestArchive(t, map[string]string{
		"../../escape.txt": "pwned",
	})

	dst := t.TempDir()
	err := UnTar(archivePath, dst)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "path traversal"))

	// 逃逸目标（dst 上级目录）不得出现文件
	entries, readErr := os.ReadDir(filepath.Dir(dst))
	assert.Nil(t, readErr)
	for _, e := range entries {
		assert.NotEqual(t, "escape.txt", e.Name())
	}
}

func TestUntarUnsupportedType(t *testing.T) {
	assert.NotNil(t, UnTar(filepath.Join(t.TempDir(), "not-an-archive.zip"), t.TempDir()))
}
