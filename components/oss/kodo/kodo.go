package kodo

import (
	"context"
	"io"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env/oss"
	"github.com/alomerry/go-tools/utils/files"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/downloader"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/region"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
)

type client struct {
	cred *credentials.Credentials
	dm   *downloader.DownloadManager
	um   *uploader.UploadManager

	useSSL     bool
	bucketName string
}

type uploadReturnBody struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func NewKodoClient(cfg meta.Config) (meta.OSSClient, error) {
	c := &client{
		useSSL:     cfg.SSL,
		bucketName: "alomerry",
	}

	c.cred = credentials.NewCredentials(cfg.AccessKey, cfg.SecretKey)
	return c, nil
}

func (q *client) UploadManager() *uploader.UploadManager {
	if q.um == nil {
		q.um = uploader.NewUploadManager(&uploader.UploadManagerOptions{
			Options: http_client.Options{
				Credentials: q.cred,
				Regions:     region.GetRegionByID("z0", q.useSSL),
			}})
	}

	return q.um
}

func (q *client) PutObject(ctx context.Context, objectKey string, reader io.Reader, _ int64) error {
	var res uploadReturnBody
	err := q.UploadManager().UploadReader(ctx, reader, &uploader.ObjectOptions{
		BucketName: q.bucketName,
		ObjectName: &objectKey,
		FileName:   files.GetFileName(objectKey),
	}, &res)
	return err
}

func (q *client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (q *client) DownloadToFile(ctx context.Context, objectKey string) (string, error) {
	urlsProvider := downloader.SignURLsProvider(
		downloader.NewStaticDomainBasedURLsProvider([]string{oss.KodoCdnDomain()}),
		downloader.NewCredentialsSigner(q.cred),
		&downloader.SignOptions{TTL: 1 * time.Hour})

	tmpFileName, err := files.CreateTempFile(ctx, cons.EmptyStr, nil)
	if err != nil {
		return "", err
	}

	downloaded, err := q.dm.DownloadToFile(
		ctx,
		objectKey,
		tmpFileName,
		&downloader.ObjectOptions{
			GenerateOptions:      downloader.GenerateOptions{BucketName: oss.KodoBasicBucket()},
			DownloadURLsProvider: urlsProvider,
		})
	if err != nil {
		return "", err
	}

	if downloaded == 0 {
		// TODO
	}
	panic("implement me")
}

func (q *client) RemoveObject(ctx context.Context, objectKey string) error {
	//TODO implement me
	panic("implement me")
}

func (q *client) StatObject(ctx context.Context, objectKey string) (meta.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (q *client) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (q *client) CreateBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	panic("implement me")
}

func (q *client) ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]meta.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (q *client) RemoveBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	panic("implement me")
}

func (q *client) Bucket(ctx context.Context, bucketName string) meta.OSSClient {
	q.bucketName = bucketName
	return q
}
