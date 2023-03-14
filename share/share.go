package share

import (
	"os"
	"path/filepath"
)

const (
	CACHE_FILE_NAME = ".oss_pusher_hash"

	MODE_PUSHER = "pusher"
	MODE_SYNCER = "syncer"

	OSS_PROVIDER_QI_NIU = "qiniu"
)

var (
	ExPath string
)

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	ExPath = filepath.Dir(ex)
}
