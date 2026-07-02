package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer struct {
	config  config.RabbitMQConfig
	service *application.NotificationService
	log     *zap.Logger
}

func NewConsumer(config config.RabbitMQConfig, service *application.NotificationService, log *zap.Logger) *Consumer {
	return &Consumer{config: config, service: service, log: log}
}

func (consumer *Consumer) Run(ctx context.Context) error {
	conn, err := amqp.Dial(consumer.config.URL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, consumer.config.NotificationQueue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification queue: %w", err)
	}
	for delivery := range deliveries {
		if err := consumer.handle(ctx, delivery); err != nil {
			consumer.log.Error("failed to handle notification event", zap.Error(err))
			_ = delivery.Nack(false, false)
			continue
		}
		_ = delivery.Ack(false)
	}
	return ctx.Err()
}

func (consumer *Consumer) handle(ctx context.Context, delivery amqp.Delivery) error {
	var event notificationEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}
	if event.NotificationID == "" {
		return nil
	}
	return consumer.service.SendNotification(ctx, event.NotificationID)
}

func ApplyTopology(config config.RabbitMQConfig) error {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()
	retryQueue := config.NotificationQueue + "-retry"
	retryExchange := retryQueue + "-exchange"
	if err := ch.ExchangeDeclare(retryExchange, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(retryQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": "amq.direct",
		"x-message-ttl":          int32(config.RetryTTL / time.Millisecond),
	}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(config.NotificationQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": retryExchange,
	}); err != nil {
		return err
	}
	if err := ch.QueueBind(config.NotificationQueue, config.NotificationQueue, "amq.direct", false, nil); err != nil {
		return err
	}
	return ch.QueueBind(retryQueue, retryQueue, retryExchange, false, nil)
}

type notificationEvent struct {
	NotificationID string `json:"notificationId"`
}
