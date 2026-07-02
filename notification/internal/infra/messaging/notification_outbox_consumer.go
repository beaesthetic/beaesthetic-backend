package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

const channelNotifications = "notifications"

type NotificationOutboxConsumer struct {
	service *application.NotificationService
}

func NewNotificationOutboxConsumer(service *application.NotificationService) *NotificationOutboxConsumer {
	return &NotificationOutboxConsumer{service: service}
}

func (consumer *NotificationOutboxConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	event, err := notificationEventFromDelivery(delivery)
	if err != nil {
		return err
	}
	if event.NotificationID == "" {
		return nil
	}
	return consumer.service.SendNotification(ctx, event.NotificationID)
}

func notificationEventFromDelivery(delivery amqp.Delivery) (notificationEvent, error) {
	message := outboxamqp.MessageFromDelivery(delivery)
	if message.Channel != "" && string(message.Channel) != channelNotifications {
		return notificationEvent{}, fmt.Errorf("unsupported outbox channel %q", message.Channel)
	}
	var event notificationEvent
	if err := json.Unmarshal(message.Payload, &event); err != nil {
		return notificationEvent{}, err
	}
	return event, nil
}

type notificationEvent struct {
	NotificationID string `json:"notificationId"`
}
