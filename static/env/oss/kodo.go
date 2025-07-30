package oss

import (
	"github.com/alomerry/go-tools/static/cons/oss"
	"os"
)

func KodoCdnDomain() string {
	domain := os.Getenv(oss.KodoCdnDomain)
	if domain == "" {
		domain = "cdn.alomerry.cn"
	}
	return domain
}

func KodoBasicBucket() string {
	bucket := os.Getenv(oss.KodoBasicBucket)
	if bucket == "" {
		bucket = "alomerry"
	}
	return bucket
}

func KodoSecretKey() string {
	return os.Getenv(oss.KodoSecretKey)
}

func KodoAccessKey() string {
	return os.Getenv(oss.KodoAccessKey)
}
