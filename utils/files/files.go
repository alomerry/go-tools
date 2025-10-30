package files

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
)

func GetFileName(filePath string) string {
	return strings.Trim(strings.TrimPrefix(filePath, path.Dir(filePath)), "/")
}

func GetFileType(filePath string) string {
	res := strings.Split(GetFileName(filePath), ".")
	if len(res) == 0 {
		return ""
	}
	return res[len(res)-1]
}

func CreateTempFile(ctx context.Context, fileName string, fn func(file *os.File) error) (string, error) {
	file, err := os.CreateTemp("/tmp", fmt.Sprintf("*_%s", fileName))
	if err != nil {
		return "", err
	}
	defer file.Close()

	name := file.Name()
	if fn == nil {
		return name, nil
	}

	err = fn(file)
	if err != nil {
		return "", err
	}

	return name, nil
}
