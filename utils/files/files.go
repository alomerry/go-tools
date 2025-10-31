package files

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	tmpDirCreateOnce = sync.Once{}
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
	logrus.WithContext(ctx).Infof("create tmp file for [%s]", fileName)
	tmpDirCreateOnce.Do(func() {
		err := os.MkdirAll("/tmp", 0755) // Create directory and its parents if they don't exist
		if err != nil {
			log.Fatalf("Failed to create temporary directory: %v", err)
		}
	})

	file, err := os.CreateTemp("/tmp", fmt.Sprintf("*_%s", fileName))
	if err != nil {
		return "", fmt.Errorf("create temp file failed, %s", err.Error())
	}
	defer file.Close()

	info, err := os.Stat(file.Name())
	if err != nil {
		return "", fmt.Errorf("stat temp file failed, %s", err.Error())
	}

	logrus.WithContext(ctx).Infof("tmp file info: %s, %v", info.Name(), info.IsDir())

	name := file.Name()
	if fn == nil {
		return name, nil
	}

	err = fn(file)
	if err != nil {
		return "", fmt.Errorf("do fn for temp file failed, %s", err.Error())
	}

	return name, nil
}
