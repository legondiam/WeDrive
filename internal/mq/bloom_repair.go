package mq

import (
	"WeDrive/internal/cache"
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
	bloomRepairExchange       = "wedrive.bloom.repair"
	bloomRepairRoutingKey     = "bloom.repair"
	bloomRepairQueue          = "wedrive.bloom.repair.queue"
	bloomRepairDelayQueue     = "wedrive.bloom.repair.delay"
	bloomRepairDeadQueue      = "wedrive.bloom.repair.dead"
	bloomRepairConsumerTag    = "wedrive-bloom-repair"
	bloomRepairTypeShareToken = "share_token"
	bloomRepairTypeFileID     = "file_identity"
	bloomRepairContentType    = "application/json"
	bloomRepairRetryDelay     = 5 * time.Second
	bloomRepairMaxRetryCount  = 5
)

type BloomRepairPublisher struct {
	conn *amqp.Connection
}

type BloomRepairMessage struct {
	Type       string `json:"type"`
	Token      string `json:"token,omitempty"`
	HashType   string `json:"hash_type,omitempty"`
	FileHash   string `json:"file_hash,omitempty"`
	RetryCount int    `json:"retry_count"`
}

func NewBloomRepairPublisher(conn *amqp.Connection) *BloomRepairPublisher {
	return &BloomRepairPublisher{conn: conn}
}

// PublishShareTokenRepair 发布分享 token 布隆补偿消息。
func (p *BloomRepairPublisher) PublishShareTokenRepair(ctx context.Context, token string) error {
	return p.publishRetry(ctx, BloomRepairMessage{
		Type:       bloomRepairTypeShareToken,
		Token:      token,
		RetryCount: 1,
	})
}

// PublishFileIdentityRepair 发布文件身份布隆补偿消息。
func (p *BloomRepairPublisher) PublishFileIdentityRepair(ctx context.Context, hashType string, fileHash string) error {
	return p.publishRetry(ctx, BloomRepairMessage{
		Type:       bloomRepairTypeFileID,
		HashType:   hashType,
		FileHash:   fileHash,
		RetryCount: 1,
	})
}

// publishRetry 发布布隆补偿重试消息。
func (p *BloomRepairPublisher) publishRetry(ctx context.Context, msg BloomRepairMessage) error {
	if p == nil || p.conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	defer ch.Close()
	if err := declareBloomRepair(ch); err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return errors.WithStack(err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	body, err := json.Marshal(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	return publishWithConfirm(ctx, ch, confirms, "", bloomRepairDelayQueue, amqp.Publishing{
		ContentType:  bloomRepairContentType,
		DeliveryMode: amqp.Persistent,
		Expiration:   strconv.FormatInt(bloomRepairRetryDelay.Milliseconds(), 10),
		Body:         body,
	})
}

// StartBloomRepairConsumer 启动布隆补偿消费者。
func StartBloomRepairConsumer(conn *amqp.Connection, bloomRepo *repository.BloomRepo) error {
	if conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := declareBloomRepair(ch); err != nil {
		_ = ch.Close()
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	deliveries, err := ch.Consume(bloomRepairQueue, bloomRepairConsumerTag, false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	go func() {
		defer ch.Close()
		for delivery := range deliveries {
			handleBloomRepair(delivery, ch, confirms, bloomRepo)
		}
	}()
	return nil
}

// declareBloomRepair 声明布隆补偿交换机和队列。
func declareBloomRepair(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(bloomRepairExchange, "direct", true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(bloomRepairQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": bloomRepairDeadQueue,
	}); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(bloomRepairDeadQueue, true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if err := ch.QueueBind(bloomRepairQueue, bloomRepairRoutingKey, bloomRepairExchange, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(bloomRepairDelayQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    bloomRepairExchange,
		"x-dead-letter-routing-key": bloomRepairRoutingKey,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// handleBloomRepair 处理单条布隆补偿消息。
func handleBloomRepair(delivery amqp.Delivery, ch *amqp.Channel, confirms <-chan amqp.Confirmation, bloomRepo *repository.BloomRepo) {
	var msg BloomRepairMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		logger.S.Warnf("布隆补偿消息解析失败:%v", err)
		_ = delivery.Ack(false)
		return
	}
	if err := repairBloomByMessage(context.Background(), msg, bloomRepo); err != nil {
		logger.S.Warnf("布隆补偿失败:%v", err)
		if msg.RetryCount >= bloomRepairMaxRetryCount {
			_ = setBloomNotReadyByMessage(context.Background(), msg, bloomRepo)
			if deadErr := publishWithConfirm(context.Background(), ch, confirms, "", bloomRepairDeadQueue, amqp.Publishing{
				ContentType:  bloomRepairContentType,
				DeliveryMode: amqp.Persistent,
				Body:         delivery.Body,
			}); deadErr != nil {
				logger.S.Warnf("投递布隆补偿死信失败:%v", deadErr)
				_ = delivery.Nack(false, false)
				return
			}
			_ = delivery.Ack(false)
			return
		}
		msg.RetryCount++
		body, marshalErr := json.Marshal(msg)
		if marshalErr != nil {
			logger.S.Warnf("序列化布隆补偿重试消息失败:%v", marshalErr)
			_ = delivery.Ack(false)
			return
		}
		if retryErr := publishWithConfirm(context.Background(), ch, confirms, "", bloomRepairDelayQueue, amqp.Publishing{
			ContentType:  bloomRepairContentType,
			DeliveryMode: amqp.Persistent,
			Expiration:   strconv.FormatInt(bloomRepairRetryDelay.Milliseconds(), 10),
			Body:         body,
		}); retryErr != nil {
			logger.S.Warnf("重新投递布隆补偿消息失败:%v", retryErr)
			_ = delivery.Nack(false, false)
			return
		}
		_ = delivery.Ack(false)
		return
	}
	_ = delivery.Ack(false)
}

// repairBloomByMessage 根据消息补写布隆过滤器。
func repairBloomByMessage(ctx context.Context, msg BloomRepairMessage, bloomRepo *repository.BloomRepo) error {
	switch msg.Type {
	case bloomRepairTypeShareToken:
		if msg.Token == "" {
			return errors.New("share token bloom repair message invalid")
		}
		return bloomRepo.Add(ctx, cache.BloomShareToken, msg.Token)
	case bloomRepairTypeFileID:
		if msg.HashType == "" || msg.FileHash == "" {
			return errors.New("file identity bloom repair message invalid")
		}
		return bloomRepo.Add(ctx, cache.BloomFileIdentity, cache.FileIdentityBloomItem(msg.HashType, msg.FileHash))
	default:
		return errors.Errorf("不支持的布隆补偿类型: %s", msg.Type)
	}
}

// setBloomNotReadyByMessage 在补偿死信时关闭对应布隆过滤器。
func setBloomNotReadyByMessage(ctx context.Context, msg BloomRepairMessage, bloomRepo *repository.BloomRepo) error {
	switch msg.Type {
	case bloomRepairTypeShareToken:
		return bloomRepo.SetReady(ctx, cache.BloomShareToken, false)
	case bloomRepairTypeFileID:
		return bloomRepo.SetReady(ctx, cache.BloomFileIdentity, false)
	default:
		return nil
	}
}
