package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func FileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func IndexOfArray[T string](item T, array []T) int {
	for i := range array {
		if item == array[i] {
			return i
		}
	}
	return -1
}
