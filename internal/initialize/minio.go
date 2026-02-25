package initialize

import (
	"WeDrive/internal/config"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

func MinioInit() (*minio.Client, error) {
	minioClient, err := minio.New(config.GlobalConf.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.GlobalConf.Minio.AccessKey, config.GlobalConf.Minio.SecretKey, ""),
		Secure: config.GlobalConf.Minio.UseSSL,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 检查桶是否存在
	exists, err := minioClient.BucketExists(context.Background(), config.GlobalConf.Minio.BucketName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !exists {
		// 创建桶
		err = minioClient.MakeBucket(context.Background(), config.GlobalConf.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return minioClient, nil
}
