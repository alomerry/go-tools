package files

import (
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
