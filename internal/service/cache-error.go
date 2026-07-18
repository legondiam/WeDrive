package service

import "github.com/pkg/errors"

var ErrCacheUnavailable = errors.New("缓存服务暂不可用")
