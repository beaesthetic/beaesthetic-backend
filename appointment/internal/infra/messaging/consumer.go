package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Handler interface {
	Process(ctx context.Context, delivery amqp.Delivery) error
}

type Consumer struct {
	dsn     string
	queue   string
	handler Handler
	log     *zap.Logger
}

func NewConsumer(dsn string, queue string, handler Handler, log *zap.Logger) *Consumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &Consumer{dsn: dsn, queue: queue, handler: handler, log: log.Named("rabbitmq_consumer").With(zap.String("queue", queue))}
}

func (consumer *Consumer) Run(ctx context.Context) error {
	consumer.log.Info("connecting rabbitmq consumer")
	conn, err := amqp.Dial(consumer.dsn)
	if err != nil {
		consumer.log.Error("failed to connect rabbitmq consumer", zap.Error(err))
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		consumer.log.Error("failed to open rabbitmq channel", zap.Error(err))
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		consumer.log.Error("failed to configure rabbitmq qos", zap.Error(err))
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, consumer.queue, "", false, false, false, false, nil)
	if err != nil {
		consumer.log.Error("failed to start rabbitmq consumer", zap.Error(err))
		return fmt.Errorf("consume queue %q: %w", consumer.queue, err)
	}
	consumer.log.Info("started rabbitmq consumer")

	for delivery := range deliveries {
		consumer.log.Debug("received rabbitmq message", zap.Uint64("delivery_tag", delivery.DeliveryTag), zap.String("exchange", delivery.Exchange), zap.String("routing_key", delivery.RoutingKey))
		if err := consumer.handler.Process(ctx, delivery); err != nil {
			consumer.log.Error("failed to process rabbitmq message", zap.Uint64("delivery_tag", delivery.DeliveryTag), zap.Error(err))
			_ = delivery.Nack(false, false)
			continue
		}
		if err := delivery.Ack(false); err != nil {
			consumer.log.Error("failed to ack rabbitmq message", zap.Uint64("delivery_tag", delivery.DeliveryTag), zap.Error(err))
			continue
		}
		consumer.log.Debug("acked rabbitmq message", zap.Uint64("delivery_tag", delivery.DeliveryTag))
	}
	consumer.log.Info("rabbitmq consumer stopped")
	return ctx.Err()
}
