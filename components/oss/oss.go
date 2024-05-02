package oss

import "context"

type OSS interface {
	UploadFromLocal(ctx context.Context, bucket, filePath, ossPath string) (any, error)
}
