package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const sampleChunkSize int64 = 1 << 20
const ChunkIdentitySize int64 = 5 << 20

type FileHashes struct {
	Full string
	Head string
	Mid  string
	Tail string
}

type ChunkHash struct {
	PartNumber int
	Hash       string
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(hash), nil
}

// CheckPassword 检查密码是否正确
func CheckPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) { //密码不匹配
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

// hashBytes 计算字节的sha256值
func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashFile 计算文件的sha256值
func HashFile(fileHeader *multipart.FileHeader) (string, error) {
	hashes, err := HashFileWithSamples(fileHeader)
	if err != nil {
		return "", err
	}
	return hashes.Full, nil
}

// HashFileWithSamples 计算文件完整sha256与抽样sha256
func HashFileWithSamples(fileHeader *multipart.FileHeader) (FileHashes, error) {
	stream, err := fileHeader.Open()
	if err != nil {
		return FileHashes{}, errors.WithStack(err)
	}
	defer stream.Close()

	readerAt, ok := stream.(io.ReaderAt)
	if !ok {
		return FileHashes{}, errors.New("文件流不支持随机读取")
	}

	return HashReaderWithSamples(stream, readerAt, fileHeader.Size)
}

// HashPathWithSamples 计算本地文件完整sha256与抽样sha256
func HashPathWithSamples(filePath string) (FileHashes, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileHashes{}, errors.WithStack(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return FileHashes{}, errors.WithStack(err)
	}

	return HashReaderWithSamples(file, file, info.Size())
}

// HashReaderWithSamples 计算支持随机读取对象的完整sha256与抽样sha256
func HashReaderWithSamples(reader io.Reader, readerAt io.ReaderAt, size int64) (FileHashes, error) {
	fullHash := sha256.New()
	if _, err := io.Copy(fullHash, reader); err != nil {
		return FileHashes{}, errors.WithStack(err)
	}

	//抽样函数
	readSample := func(offset int64) ([]byte, error) {
		if size == 0 {
			return []byte{}, nil
		}
		if offset < 0 {
			offset = 0
		}
		if offset > size {
			offset = size
		}
		length := sampleChunkSize
		if offset+length > size {
			length = size - offset
		}
		buf := make([]byte, length)
		if length == 0 {
			return buf, nil
		}
		if _, err := readerAt.ReadAt(buf, offset); err != nil && err != io.EOF {
			return nil, errors.WithStack(err)
		}
		return buf, nil
	}

	//计算抽样哈希
	headBytes, err := readSample(0)
	if err != nil {
		return FileHashes{}, err
	}

	midOffset := int64(0)
	if size > sampleChunkSize {
		midOffset = (size - sampleChunkSize) / 2
	}
	midBytes, err := readSample(midOffset)
	if err != nil {
		return FileHashes{}, err
	}

	tailOffset := int64(0)
	if size > sampleChunkSize {
		tailOffset = size - sampleChunkSize
	}
	tailBytes, err := readSample(tailOffset)
	if err != nil {
		return FileHashes{}, err
	}

	return FileHashes{
		Full: hex.EncodeToString(fullHash.Sum(nil)),
		Head: hashBytes(headBytes),
		Mid:  hashBytes(midBytes),
		Tail: hashBytes(tailBytes),
	}, nil
}

// ChunkHashesFromReaderAt 计算固定分块哈希
func ChunkHashesFromReaderAt(readerAt io.ReaderAt, size int64, chunkSize int64) ([]ChunkHash, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunk size must be positive")
	}
	if size == 0 {
		return []ChunkHash{{PartNumber: 1, Hash: hashBytes([]byte{})}}, nil
	}

	partCount := int((size + chunkSize - 1) / chunkSize)
	parts := make([]ChunkHash, 0, partCount)
	for partNumber := 1; partNumber <= partCount; partNumber++ {
		offset := int64(partNumber-1) * chunkSize
		length := chunkSize
		if offset+length > size {
			length = size - offset
		}
		buf := make([]byte, length)
		if _, err := readerAt.ReadAt(buf, offset); err != nil && err != io.EOF {
			return nil, errors.WithStack(err)
		}
		parts = append(parts, ChunkHash{
			PartNumber: partNumber,
			Hash:       hashBytes(buf),
		})
	}
	return parts, nil
}

// ChunkHashesFromFileHeader 计算上传文件分块哈希
func ChunkHashesFromFileHeader(fileHeader *multipart.FileHeader, chunkSize int64) ([]ChunkHash, error) {
	stream, err := fileHeader.Open()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer stream.Close()

	readerAt, ok := stream.(io.ReaderAt)
	if !ok {
		return nil, errors.New("文件流不支持随机读取")
	}
	return ChunkHashesFromReaderAt(readerAt, fileHeader.Size, chunkSize)
}

// AggregateChunkHash 计算分块聚合哈希
func AggregateChunkHash(parts []ChunkHash, fileSize int64) (string, error) {
	sum := sha256.New()
	for index, part := range parts {
		if part.PartNumber != index+1 {
			return "", errors.New("分块顺序不连续")
		}
		raw, err := hex.DecodeString(part.Hash)
		if err != nil {
			return "", errors.WithStack(err)
		}
		if _, err := sum.Write(raw); err != nil {
			return "", errors.WithStack(err)
		}
	}
	var sizeBuf [8]byte
	binary.BigEndian.PutUint64(sizeBuf[:], uint64(fileSize))
	if _, err := sum.Write(sizeBuf[:]); err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// HashBytesHex 计算字节切片的 sha256 十六进制值
func HashBytesHex(data []byte) string {
	return hashBytes(data)
}
