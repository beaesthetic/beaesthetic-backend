package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

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
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
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
