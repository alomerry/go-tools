package oss

import (
	"os"

	"github.com/alomerry/go-tools/static/cons/oss"
)

func MinioEndpoint() string {
	endpoint := os.Getenv(oss.MinioEndpoint)
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	return endpoint
}

func MinioAccessKey() string {
	return os.Getenv(oss.MinioAccessKey)
}

func MinioSecretKey() string {
	return os.Getenv(oss.MinioSecretKey)
}
