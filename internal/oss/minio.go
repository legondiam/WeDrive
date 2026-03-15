package oss

import (
	"WeDrive/internal/config"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
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
func (s *Storage) DownloadFile(ctx context.Context, objectName string, fileName string, expiration time.Duration, tier string) (string, error) {
	reqParams := make(url.Values)
	if fileName != "" {
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(fileName)))
		reqParams.Set("response-content-type", "application/octet-stream")
	}
	minioURL, err := s.client.PresignedGetObject(ctx, config.GlobalConf.Minio.BucketName, objectName, expiration, reqParams)
	if err != nil {
		return "", errors.WithStack(err)
	}
	base := strings.TrimRight(config.GlobalConf.Download.PublicBaseURL, "/")
	if base == "" {
		return minioURL.String(), nil
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", errors.WithStack(err)
	}
	exp := strconv.FormatInt(time.Now().Add(expiration).Unix(), 10)
	uri := minioURL.EscapedPath()
	signature := s.secureLinkSig(exp, uri, tier, config.GlobalConf.Download.SignSecret)

	wrappedPath := "/WeDrive/" + exp + "/" + tier + "/" + signature + uri
	finalURL := &url.URL{
		Scheme:   baseURL.Scheme,
		Host:     baseURL.Host,
		Path:     wrappedPath,
		RawQuery: minioURL.RawQuery,
	}
	return finalURL.String(), nil
}

// secureLinkSig 生成安全链接签名
func (s *Storage) secureLinkSig(exp string, uri string, tier string, secret string) string {
	sum := md5.Sum([]byte(uri + "|" + exp + "|" + tier + "|" + secret))
	sig := base64.RawURLEncoding.EncodeToString(sum[:])
	return sig
}
