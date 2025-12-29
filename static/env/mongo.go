package env

import (
	"os"

	"github.com/alomerry/go-tools/static/cons"
)

func GetMongoDSN(defaultVal ...string) string {
	if Local() {
		return os.Getenv(cons.MongoDSN)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}
	return os.Getenv(cons.MongoDSN)
}
