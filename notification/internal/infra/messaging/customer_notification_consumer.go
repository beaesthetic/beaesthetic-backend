package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
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
	var command application.CustomerNotificationCommand
	if err := json.Unmarshal(delivery.Body, &command); err != nil {
		return application.CustomerNotificationCommand{}, fmt.Errorf("decode customer notification command: %w", err)
	}
	if command.Body == nil {
		command.Body = map[string]any{}
	}
	return command, nil
}
