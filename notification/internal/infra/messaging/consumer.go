package messaging

import (
	"context"
	"fmt"

	contractsmessaging "github.com/petretiandrea/beaesthetic-backend/core-contracts/runtime/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
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
	return &Consumer{
		dsn:     dsn,
		queue:   queue,
		handler: handler,
		log:     log.Named("rabbitmq_consumer").With(zap.String("queue", queue)),
	}
}

func (consumer *Consumer) Run(ctx context.Context) error {
	conn, err := amqp.Dial(consumer.dsn)
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
	deliveries, err := ch.ConsumeWithContext(ctx, consumer.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", consumer.queue, err)
	}
	for delivery := range deliveries {
		messageCtx, span := contractsmessaging.StartConsumer(ctx, consumer.queue, delivery)
		if err := consumer.handler.Process(messageCtx, delivery); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
			consumer.log.Error("failed to process rabbitmq message", zap.Error(err))
			_ = delivery.Nack(false, false)
			continue
		}
		if err := delivery.Ack(false); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
	return ctx.Err()
}
