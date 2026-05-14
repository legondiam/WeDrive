package mq

import (
	"WeDrive/internal/repository"
	"WeDrive/pkg/logger"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	cacheInvalidationExchange       = "wedrive.cache.invalidate"
	cacheInvalidationRoutingKey     = "cache.invalidate"
	cacheInvalidationQueue          = "wedrive.cache.invalidate.queue"
	cacheInvalidationDelayQueue     = "wedrive.cache.invalidate.delay"
	cacheInvalidationDeadQueue      = "wedrive.cache.invalidate.dead"
	cacheInvalidationConsumerTag    = "wedrive-cache-invalidation"
	cacheInvalidationTypeUserInfo   = "user_info"
	cacheInvalidationTypeFileList   = "file_list"
	cacheInvalidationTypeRecycle    = "recycle_list"
	cacheInvalidationTypeFileMeta   = "file_meta"
	cacheInvalidationContentType    = "application/json"
	cacheInvalidationDeliveryMode   = amqp.Persistent
	cacheInvalidationPrefetchCount  = 10
	cacheInvalidationPrefetchSize   = 0
	cacheInvalidationPrefetchGlobal = false
	cacheInvalidationRetryDelay     = 5 * time.Second
	cacheInvalidationMaxRetryCount  = 5
)

type CacheInvalidationPublisher struct {
	conn *amqp.Connection
}

type CacheInvalidationMessage struct {
	Type       string `json:"type"`
	UserID     uint   `json:"user_id"`
	ParentID   uint   `json:"parent_id,omitempty"`
	UserFileID uint   `json:"user_file_id,omitempty"`
	RetryCount int    `json:"retry_count"`
}

func NewCacheInvalidationPublisher(conn *amqp.Connection) *CacheInvalidationPublisher {
	return &CacheInvalidationPublisher{conn: conn}
}

// PublishUserInfoRetry 发布用户信息缓存删除重试消息。
func (p *CacheInvalidationPublisher) PublishUserInfoRetry(ctx context.Context, userID uint) error {
	return p.publishRetry(ctx, CacheInvalidationMessage{
		Type:       cacheInvalidationTypeUserInfo,
		UserID:     userID,
		RetryCount: 1,
	})
}

// PublishFileListRetry 发布用户目录列表缓存删除重试消息。
func (p *CacheInvalidationPublisher) PublishFileListRetry(ctx context.Context, userID uint, parentID uint) error {
	return p.publishRetry(ctx, CacheInvalidationMessage{
		Type:       cacheInvalidationTypeFileList,
		UserID:     userID,
		ParentID:   parentID,
		RetryCount: 1,
	})
}

// PublishRecycleListRetry 发布用户回收站列表缓存删除重试消息。
func (p *CacheInvalidationPublisher) PublishRecycleListRetry(ctx context.Context, userID uint) error {
	return p.publishRetry(ctx, CacheInvalidationMessage{
		Type:       cacheInvalidationTypeRecycle,
		UserID:     userID,
		RetryCount: 1,
	})
}

// PublishFileMetaRetry 发布下载文件元数据缓存删除重试消息。
func (p *CacheInvalidationPublisher) PublishFileMetaRetry(ctx context.Context, userID uint, userFileID uint) error {
	return p.publishRetry(ctx, CacheInvalidationMessage{
		Type:       cacheInvalidationTypeFileMeta,
		UserID:     userID,
		UserFileID: userFileID,
		RetryCount: 1,
	})
}

// publishRetry 发布缓存删除重试消息
func (p *CacheInvalidationPublisher) publishRetry(ctx context.Context, msg CacheInvalidationMessage) error {
	if p == nil || p.conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	defer ch.Close()

	if err := declareCacheInvalidation(ch); err != nil {
		return err
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	//使用默认交换机，发送消息到延迟队列
	return errors.WithStack(ch.PublishWithContext(ctx, "", cacheInvalidationDelayQueue, false, false, amqp.Publishing{
		ContentType:  cacheInvalidationContentType,
		DeliveryMode: cacheInvalidationDeliveryMode,
		Expiration:   strconv.FormatInt(cacheInvalidationRetryDelay.Milliseconds(), 10),
		Body:         body,
	}))
}

// publishCacheInvalidationDead 将超过重试次数的缓存删除消息发送到死信队列。
func publishCacheInvalidationDead(ctx context.Context, ch *amqp.Channel, body []byte) error {
	return errors.WithStack(ch.PublishWithContext(ctx, "", cacheInvalidationDeadQueue, false, false, amqp.Publishing{
		ContentType:  cacheInvalidationContentType,
		DeliveryMode: cacheInvalidationDeliveryMode,
		Body:         body,
	}))
}

// StartCacheInvalidationConsumer 启动缓存失效消息消费者。
func StartCacheInvalidationConsumer(conn *amqp.Connection, userCache *repository.UserCacheRepo, fileCache *repository.FileCacheRepo) error {
	if conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := declareCacheInvalidation(ch); err != nil {
		_ = ch.Close()
		return err
	}
	if err := ch.Qos(cacheInvalidationPrefetchCount, cacheInvalidationPrefetchSize, cacheInvalidationPrefetchGlobal); err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	deliveries, err := ch.Consume(cacheInvalidationQueue, cacheInvalidationConsumerTag, false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveries {
			handleCacheInvalidation(delivery, ch, userCache, fileCache)
		}
	}()
	return nil
}

// declareCacheInvalidation 声明交换机、消费队列和延迟队列
func declareCacheInvalidation(ch *amqp.Channel) error {
	// 声明交换机
	if err := ch.ExchangeDeclare(cacheInvalidationExchange, "direct", true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	// 声明消费队列
	if _, err := ch.QueueDeclare(cacheInvalidationQueue, true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	// 声明死信队列
	if _, err := ch.QueueDeclare(cacheInvalidationDeadQueue, true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	// 绑定消费队列到交换机
	if err := ch.QueueBind(cacheInvalidationQueue, cacheInvalidationRoutingKey, cacheInvalidationExchange, false, nil); err != nil {
		return errors.WithStack(err)
	}
	// 声明延迟队列
	if _, err := ch.QueueDeclare(cacheInvalidationDelayQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    cacheInvalidationExchange,
		"x-dead-letter-routing-key": cacheInvalidationRoutingKey,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// handleCacheInvalidation 处理单条缓存失效消息。
func handleCacheInvalidation(delivery amqp.Delivery, ch *amqp.Channel, userCache *repository.UserCacheRepo, fileCache *repository.FileCacheRepo) {
	var msg CacheInvalidationMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		logger.S.Warnf("缓存失效消息解析失败:%v", err)
		_ = delivery.Ack(false)
		return
	}
	if msg.UserID == 0 {
		logger.S.Warnf("缓存失效消息无效:%s", string(delivery.Body))
		_ = delivery.Ack(false)
		return
	}
	if err := deleteCacheByMessage(context.Background(), msg, userCache, fileCache); err != nil {
		logger.S.Warnf("MQ删除用户信息缓存失败:%v", err)
		//重试次数超过限制后，投递死信队列
		if msg.RetryCount >= cacheInvalidationMaxRetryCount {
			if deadErr := publishCacheInvalidationDead(context.Background(), ch, delivery.Body); deadErr != nil {
				logger.S.Warnf("投递缓存删除死信消息失败:%v", deadErr)
			}
			_ = delivery.Ack(false)
			return
		}
		msg.RetryCount++
		body, marshalErr := json.Marshal(msg)
		if marshalErr != nil {
			logger.S.Warnf("序列化用户信息缓存删除重试消息失败:%v", marshalErr)
			_ = delivery.Ack(false)
			return
		}
		if retryErr := ch.PublishWithContext(context.Background(), "", cacheInvalidationDelayQueue, false, false, amqp.Publishing{
			ContentType:  cacheInvalidationContentType,
			DeliveryMode: cacheInvalidationDeliveryMode,
			Expiration:   strconv.FormatInt(cacheInvalidationRetryDelay.Milliseconds(), 10),
			Body:         body,
		}); retryErr != nil {
			logger.S.Warnf("重新投递缓存删除重试消息失败:%v", retryErr)
		}
		_ = delivery.Ack(false)
		return
	}
	logger.S.Infof("MQ删除缓存成功, type: %s, userID: %d", msg.Type, msg.UserID)
	_ = delivery.Ack(false)
}

// deleteCacheByMessage 根据消息类型删除对应缓存。
func deleteCacheByMessage(ctx context.Context, msg CacheInvalidationMessage, userCache *repository.UserCacheRepo, fileCache *repository.FileCacheRepo) error {
	switch msg.Type {
	case cacheInvalidationTypeUserInfo:
		return userCache.DeleteUserInfo(ctx, msg.UserID)
	case cacheInvalidationTypeFileList:
		return fileCache.DeleteUserFileList(ctx, msg.UserID, msg.ParentID)
	case cacheInvalidationTypeRecycle:
		return fileCache.DeleteRecycleBinList(ctx, msg.UserID)
	case cacheInvalidationTypeFileMeta:
		return fileCache.DeleteDownloadFileMeta(ctx, msg.UserID, msg.UserFileID)
	default:
		return errors.Errorf("不支持的缓存失效类型: %s", msg.Type)
	}
}
