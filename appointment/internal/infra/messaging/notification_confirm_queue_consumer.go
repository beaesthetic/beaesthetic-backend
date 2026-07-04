package messaging

import (
	"context"
	"encoding/json"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type NotificationConfirmQueueConsumer struct {
	service *application.AppointmentService
	log     *zap.Logger
}

func NewNotificationConfirmQueueConsumer(service *application.AppointmentService, log *zap.Logger) *NotificationConfirmQueueConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &NotificationConfirmQueueConsumer{service: service, log: log.Named("notification_confirm_queue_consumer")}
}

func (consumer *NotificationConfirmQueueConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	var event notificationConfirmedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}
	if event.NotificationID == "" {
		consumer.log.Warn("notification confirm message does not contain notificationId")
		return nil
	}
	agendaEvent, err := consumer.service.ConfirmNotification(ctx, event.NotificationID)
	if err != nil {
		return err
	}
	if agendaEvent == nil {
		consumer.log.Info("notification confirmation has no pending appointment", zap.String("notification_id", event.NotificationID))
		return nil
	}
	consumer.log.Info("confirmed reminder sent", zap.String("event_id", agendaEvent.ID), zap.String("notification_id", event.NotificationID))
	return nil
}

type notificationConfirmedEvent struct {
	NotificationID string `json:"notificationId"`
}
