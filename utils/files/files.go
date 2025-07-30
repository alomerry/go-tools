package files

import (
	"context"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alomerry/go-tools/utils/random"
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

func CreateTempFile(ctx context.Context, fileName string) string {
	fullPath := path.Join("/tmp", fileName+"_"+time.Now().Format("20060102150405")+"_"+random.String(10))
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err := os.Create(fileName)
		if err != nil {
			log.Panicf("create temp file failed, err %v", err)
		}
	}
	return fullPath
}
