package mq

import (
	"WeDrive/pkg/logger"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	uploadVerificationExchange       = "wedrive.upload.verify"
	uploadVerificationRoutingKey     = "upload.verify"
	uploadVerificationQueue          = "wedrive.upload.verify.queue"
	uploadVerificationDelayQueue     = "wedrive.upload.verify.delay"
	uploadVerificationDeadQueue      = "wedrive.upload.verify.dead"
	uploadVerificationConsumerTag    = "wedrive-upload-verification"
	uploadVerificationContentType    = "application/json"
	uploadVerificationDeliveryMode   = amqp.Persistent
	uploadVerificationPrefetchCount  = 2
	uploadVerificationPrefetchSize   = 0
	uploadVerificationPrefetchGlobal = false
	uploadVerificationRetryDelay     = 10 * time.Second
	uploadVerificationMaxRetryCount  = 5
)

type UploadVerificationPublisher struct {
	conn *amqp.Connection
}

type UploadVerificationMessage struct {
	UploadID   uint `json:"upload_id"`
	RetryCount int  `json:"retry_count"`
}

type UploadVerificationHandler func(ctx context.Context, uploadID uint) error

func NewUploadVerificationPublisher(conn *amqp.Connection) *UploadVerificationPublisher {
	return &UploadVerificationPublisher{conn: conn}
}

// PublishUploadVerification 发布分块上传哈希校验消息。
func (p *UploadVerificationPublisher) PublishUploadVerification(ctx context.Context, uploadID uint) error {
	if p == nil || p.conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	defer ch.Close()

	if err := declareUploadVerification(ch); err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return errors.WithStack(err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	body, err := json.Marshal(UploadVerificationMessage{
		UploadID:   uploadID,
		RetryCount: 1,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return publishWithConfirm(ctx, ch, confirms, uploadVerificationExchange, uploadVerificationRoutingKey, amqp.Publishing{
		ContentType:  uploadVerificationContentType,
		DeliveryMode: uploadVerificationDeliveryMode,
		Body:         body,
	})
}

// StartUploadVerificationConsumer 启动分块上传哈希校验消费者。
func StartUploadVerificationConsumer(conn *amqp.Connection, handler UploadVerificationHandler) error {
	if conn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	ch, err := conn.Channel()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := declareUploadVerification(ch); err != nil {
		_ = ch.Close()
		return err
	}
	if err := ch.Qos(uploadVerificationPrefetchCount, uploadVerificationPrefetchSize, uploadVerificationPrefetchGlobal); err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	deliveries, err := ch.Consume(uploadVerificationQueue, uploadVerificationConsumerTag, false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return errors.WithStack(err)
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveries {
			handleUploadVerification(delivery, ch, confirms, handler)
		}
	}()
	return nil
}

// declareUploadVerification 声明上传校验交换机、队列和延迟队列。
func declareUploadVerification(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(uploadVerificationExchange, "direct", true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(uploadVerificationQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": uploadVerificationDeadQueue,
	}); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(uploadVerificationDeadQueue, true, false, false, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if err := ch.QueueBind(uploadVerificationQueue, uploadVerificationRoutingKey, uploadVerificationExchange, false, nil); err != nil {
		return errors.WithStack(err)
	}
	if _, err := ch.QueueDeclare(uploadVerificationDelayQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    uploadVerificationExchange,
		"x-dead-letter-routing-key": uploadVerificationRoutingKey,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// handleUploadVerification 处理单条上传校验消息。
func handleUploadVerification(delivery amqp.Delivery, ch *amqp.Channel, confirms <-chan amqp.Confirmation, handler UploadVerificationHandler) {
	var msg UploadVerificationMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		logger.S.Warnf("上传校验消息解析失败:%v", err)
		_ = delivery.Ack(false)
		return
	}
	if msg.UploadID == 0 {
		logger.S.Warnf("上传校验消息无效:%s", string(delivery.Body))
		_ = delivery.Ack(false)
		return
	}
	if err := handler(context.Background(), msg.UploadID); err != nil {
		logger.S.Warnf("上传校验失败, uploadID: %d, err: %v", msg.UploadID, err)
		if msg.RetryCount >= uploadVerificationMaxRetryCount {
			if deadErr := publishUploadVerificationDead(context.Background(), ch, confirms, delivery.Body); deadErr != nil {
				logger.S.Warnf("投递上传校验死信消息失败:%v", deadErr)
				_ = delivery.Nack(false, false)
				return
			}
			_ = delivery.Ack(false)
			return
		}
		msg.RetryCount++
		body, marshalErr := json.Marshal(msg)
		if marshalErr != nil {
			logger.S.Warnf("序列化上传校验重试消息失败:%v", marshalErr)
			_ = delivery.Ack(false)
			return
		}
		if retryErr := publishWithConfirm(context.Background(), ch, confirms, "", uploadVerificationDelayQueue, amqp.Publishing{
			ContentType:  uploadVerificationContentType,
			DeliveryMode: uploadVerificationDeliveryMode,
			Expiration:   strconv.FormatInt(uploadVerificationRetryDelay.Milliseconds(), 10),
			Body:         body,
		}); retryErr != nil {
			logger.S.Warnf("重新投递上传校验重试消息失败:%v", retryErr)
			_ = delivery.Nack(false, false)
			return
		}
		_ = delivery.Ack(false)
		return
	}
	logger.S.Infof("上传校验完成, uploadID: %d", msg.UploadID)
	_ = delivery.Ack(false)
}

// publishUploadVerificationDead 将超过重试次数的上传校验消息发送到死信队列。
func publishUploadVerificationDead(ctx context.Context, ch *amqp.Channel, confirms <-chan amqp.Confirmation, body []byte) error {
	return publishWithConfirm(ctx, ch, confirms, "", uploadVerificationDeadQueue, amqp.Publishing{
		ContentType:  uploadVerificationContentType,
		DeliveryMode: uploadVerificationDeliveryMode,
		Body:         body,
	})
}
