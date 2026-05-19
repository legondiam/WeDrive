package oss

import (
	"WeDrive/internal/config"
	hashutil "WeDrive/pkg/utils/hash"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

type Storage struct {
	client *minio.Client
}

type CompletePart struct {
	PartNumber int
	ETag       string
}

func NewStorage(client *minio.Client) *Storage {
	return &Storage{client: client}
}

func (s *Storage) core() minio.Core {
	return minio.Core{Client: s.client}
}

// UploadFile 上传文件
func (s *Storage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, config.GlobalConf.Minio.BucketName, objectName, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UploadFileByPath 上传本地文件
func (s *Storage) UploadFileByPath(ctx context.Context, objectName string, filePath string, size int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	return s.UploadFile(ctx, objectName, file, size)
}

// NewMultipartUpload 初始化分块上传
func (s *Storage) NewMultipartUpload(ctx context.Context, objectName string) (string, error) {
	uploadID, err := s.core().NewMultipartUpload(ctx, config.GlobalConf.Minio.BucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return uploadID, nil
}

// PresignUploadPart 生成分块直传URL
func (s *Storage) PresignUploadPart(ctx context.Context, objectName string, uploadID string, partNumber int, checksumSHA256Base64 string, expires time.Duration) (string, map[string]string, error) {
	reqParams := make(url.Values)
	reqParams.Set("uploadId", uploadID)
	reqParams.Set("partNumber", strconv.Itoa(partNumber))

	extraHeaders := make(http.Header)
	headers := make(map[string]string)
	if checksumSHA256Base64 != "" {
		extraHeaders.Set("x-amz-checksum-sha256", checksumSHA256Base64)
		headers["x-amz-checksum-sha256"] = checksumSHA256Base64
	}

	signedURL, err := s.client.PresignHeader(ctx, http.MethodPut, config.GlobalConf.Minio.BucketName, objectName, expires, reqParams, extraHeaders)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}
	return signedURL.String(), headers, nil
}

// UploadObjectPart 上传对象分块
func (s *Storage) UploadObjectPart(ctx context.Context, objectName string, uploadID string, partNumber int, reader io.Reader, size int64) (minio.ObjectPart, error) {
	part, err := s.core().PutObjectPart(ctx, config.GlobalConf.Minio.BucketName, objectName, uploadID, partNumber, reader, size, minio.PutObjectPartOptions{})
	if err != nil {
		return minio.ObjectPart{}, errors.WithStack(err)
	}
	return part, nil
}

// ListObjectParts 列出对象分块
func (s *Storage) ListObjectParts(ctx context.Context, objectName string, uploadID string) ([]minio.ObjectPart, error) {
	partMarker := 0
	parts := make([]minio.ObjectPart, 0)
	for {
		result, err := s.core().ListObjectParts(ctx, config.GlobalConf.Minio.BucketName, objectName, uploadID, partMarker, 1000)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parts = append(parts, result.ObjectParts...)
		if !result.IsTruncated {
			break
		}
		partMarker = result.NextPartNumberMarker
	}
	return parts, nil
}

// CompleteMultipartUpload 完成分块上传
func (s *Storage) CompleteMultipartUpload(ctx context.Context, objectName string, uploadID string, parts []CompletePart) error {
	completeParts := make([]minio.CompletePart, 0, len(parts))
	for _, part := range parts {
		completeParts = append(completeParts, minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		})
	}
	_, err := s.core().CompleteMultipartUpload(ctx, config.GlobalConf.Minio.BucketName, objectName, uploadID, completeParts, minio.PutObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// AbortMultipartUpload 终止分块上传
func (s *Storage) AbortMultipartUpload(ctx context.Context, objectName string, uploadID string) error {
	err := s.core().AbortMultipartUpload(ctx, config.GlobalConf.Minio.BucketName, objectName, uploadID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetObject 打开对象
func (s *Storage) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, config.GlobalConf.Minio.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return object, nil
}

// ReadFileRange 读取对象指定字节区间，用于所有权证明校验
func (s *Storage) ReadFileRange(ctx context.Context, objectName string, offset int64, length int64) ([]byte, error) {
	if offset < 0 || length <= 0 {
		return nil, errors.New("invalid object range")
	}
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(offset, offset+length-1); err != nil {
		return nil, errors.WithStack(err)
	}
	object, err := s.client.GetObject(ctx, config.GlobalConf.Minio.BucketName, objectName, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer object.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, object); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

// HashObjectWithSamples 计算对象完整哈希与抽样哈希。
func (s *Storage) HashObjectWithSamples(ctx context.Context, objectName string, size int64) (hashutil.FileHashes, int64, error) {
	object, err := s.GetObject(ctx, objectName)
	if err != nil {
		return hashutil.FileHashes{}, 0, err
	}
	defer object.Close()

	fullHash := sha256.New()
	actualSize, err := io.Copy(fullHash, object)
	if err != nil {
		return hashutil.FileHashes{}, actualSize, errors.WithStack(err)
	}

	readSampleHash := func(offset int64) (string, error) {
		if size == 0 {
			return hashutil.HashBytesHex([]byte{}), nil
		}
		if offset < 0 {
			offset = 0
		}
		if offset > size {
			offset = size
		}
		length := int64(1 << 20)
		if offset+length > size {
			length = size - offset
		}
		if length == 0 {
			return hashutil.HashBytesHex([]byte{}), nil
		}
		data, err := s.ReadFileRange(ctx, objectName, offset, length)
		if err != nil {
			return "", err
		}
		return hashutil.HashBytesHex(data), nil
	}

	headHash, err := readSampleHash(0)
	if err != nil {
		return hashutil.FileHashes{}, actualSize, err
	}
	midOffset := int64(0)
	if size > int64(1<<20) {
		midOffset = (size - int64(1<<20)) / 2
	}
	midHash, err := readSampleHash(midOffset)
	if err != nil {
		return hashutil.FileHashes{}, actualSize, err
	}
	tailOffset := int64(0)
	if size > int64(1<<20) {
		tailOffset = size - int64(1<<20)
	}
	tailHash, err := readSampleHash(tailOffset)
	if err != nil {
		return hashutil.FileHashes{}, actualSize, err
	}

	return hashutil.FileHashes{
		Full: hex.EncodeToString(fullHash.Sum(nil)),
		Head: headHash,
		Mid:  midHash,
		Tail: tailHash,
	}, actualSize, nil
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
