package initialize

import (
	"WeDrive/internal/config"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

func RabbitMQInit() (*amqp.Connection, error) {
	mqConf := config.GlobalConf.RabbitMQ
	vhost := mqConf.Vhost
	if vhost == "" {
		vhost = "/"
	}
	dsn := fmt.Sprintf("amqp://%s@%s:%d/%s",
		url.UserPassword(mqConf.User, mqConf.Password).String(),
		mqConf.Host,
		mqConf.Port,
		url.PathEscape(vhost),
	)
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, errors.WithStack(err)
	}
	_ = ch.Close()
	return conn, nil
}
