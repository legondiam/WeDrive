package oss

import (
	"WeDrive/internal/config"
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

type Storage struct {
	client *minio.Client
}

func NewStorage(client *minio.Client) *Storage {
	return &Storage{client: client}
}

// UploadFile 上传文件
func (s *Storage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, config.GlobalConf.Minio.BucketName, objectName, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteFile 删除文件
func (s *Storage) DeleteFile(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, config.GlobalConf.Minio.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DownloadFile 下载文件
func (s *Storage) DownloadFile(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, config.GlobalConf.Minio.BucketName, objectName, expiration, nil)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return url.String(), nil
}
