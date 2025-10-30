package files

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
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
