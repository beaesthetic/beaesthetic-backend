package messaging

import (
	"context"
	"fmt"

	notification "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type CustomerNotificationConsumer struct {
	service *application.CustomerNotificationService
	log     *zap.Logger
}

func NewCustomerNotificationConsumer(service *application.CustomerNotificationService, log *zap.Logger) *CustomerNotificationConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &CustomerNotificationConsumer{service: service, log: log.Named("customer_notification_consumer")}
}

func (consumer *CustomerNotificationConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	command, err := customerNotificationCommandFromDelivery(delivery)
	if err != nil {
		return err
	}
	if err := consumer.service.Process(ctx, command); err != nil {
		return err
	}
	consumer.log.Info(
		"customer notification processed",
		zap.String("notification_type", command.NotificationType),
		zap.String("notification_channel", command.NotificationChannel),
		zap.Int("customer_count", len(command.CustomerIDs)),
	)
	return nil
}

func customerNotificationCommandFromDelivery(delivery amqp.Delivery) (application.CustomerNotificationCommand, error) {
	var message notification.CustomerNotificationRequested
	if err := protojson.Unmarshal(delivery.Body, &message); err != nil {
		return application.CustomerNotificationCommand{}, fmt.Errorf("decode customer notification command: %w", err)
	}
	channel, err := notificationChannelToDomain(message.GetNotificationChannel())
	if err != nil {
		return application.CustomerNotificationCommand{}, err
	}
	body := map[string]any{}
	if message.GetBody() != nil {
		body = message.GetBody().AsMap()
	}
	return application.CustomerNotificationCommand{
		IdempotencyKey:      message.GetIdempotencyKey(),
		CustomerIDs:         message.GetCustomerIds(),
		NotificationChannel: channel,
		NotificationType:    message.GetNotificationType(),
		Body:                body,
	}, nil
}

func notificationChannelToDomain(channel notification.NotificationChannel) (string, error) {
	switch channel {
	case notification.NotificationChannel_NOTIFICATION_CHANNEL_SMS:
		return "sms", nil
	case notification.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL:
		return "email", nil
	case notification.NotificationChannel_NOTIFICATION_CHANNEL_UNSPECIFIED:
		return "", fmt.Errorf("notificationChannel is required")
	default:
		return "", fmt.Errorf("unsupported notificationChannel enum: %s", channel.String())
	}
}
