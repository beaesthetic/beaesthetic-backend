package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/beaesthetic/scheduler/internal/config"
	"github.com/beaesthetic/scheduler/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
}

func NewPublisher(cfg config.RabbitMQConfig) (*Publisher, error) {
	dsn := amqpURL(cfg)
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Publisher{connection: conn, channel: ch, exchange: cfg.Exchange}, nil
}

func amqpURL(cfg config.RabbitMQConfig) string {
	vhost := cfg.VHost
	if vhost == "" || vhost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	}

	return (&url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   vhost,
	}).String()
}

func (p *Publisher) Publish(ctx context.Context, job domain.ScheduleJob) error {
	body, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, p.exchange, job.Route, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *Publisher) Close() error {
	var err error
	if p.channel != nil {
		err = p.channel.Close()
	}
	if p.connection != nil {
		if closeErr := p.connection.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}
