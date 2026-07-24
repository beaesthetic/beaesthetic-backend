package messaging

import (
	"context"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	notificationcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type NotificationOutcomeQueueConsumer struct {
	service *application.AppointmentService
	log     *zap.Logger
}

func NewNotificationOutcomeQueueConsumer(service *application.AppointmentService, log *zap.Logger) *NotificationOutcomeQueueConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &NotificationOutcomeQueueConsumer{service: service, log: log.Named("notification_outcome_queue_consumer")}
}

func (consumer *NotificationOutcomeQueueConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	var event notificationcontracts.CustomerNotificationOutcome
	if err := protojson.Unmarshal(delivery.Body, &event); err != nil {
		return fmt.Errorf("parse notification outcome event: %w", err)
	}
	if event.GetNotificationId() == "" {
		consumer.log.Warn("notification outcome message does not contain notificationId")
		return nil
	}
	handle, status, err := consumer.notificationOutcomeHandler(event.GetStatus())
	if err != nil {
		consumer.log.Warn("notification outcome message has unsupported status", zap.String("notification_id", event.GetNotificationId()), zap.String("status", event.GetStatus().String()))
		return nil
	}

	consumer.log.Info(
		"received notification outcome",
		zap.String("notification_id", event.GetNotificationId()),
		zap.String("idempotency_key", event.GetIdempotencyKey()),
		zap.String("customer_id", event.GetCustomerId()),
		zap.String("status", status),
		zap.String("reason", event.GetReason()),
	)
	agendaEvent, err := handle(ctx, event.GetNotificationId())
	if err != nil {
		consumer.log.Error("failed to handle notification outcome", zap.String("notification_id", event.GetNotificationId()), zap.String("idempotency_key", event.GetIdempotencyKey()), zap.String("customer_id", event.GetCustomerId()), zap.String("status", status), zap.String("reason", event.GetReason()), zap.Error(err))
		return err
	}
	if agendaEvent == nil {
		consumer.log.Info("notification outcome has no pending appointment", zap.String("notification_id", event.GetNotificationId()), zap.String("idempotency_key", event.GetIdempotencyKey()), zap.String("customer_id", event.GetCustomerId()), zap.String("status", status))
		return nil
	}
	consumer.log.Info("handled notification outcome", zap.String("event_id", agendaEvent.ID), zap.String("notification_id", event.GetNotificationId()), zap.String("idempotency_key", event.GetIdempotencyKey()), zap.String("customer_id", event.GetCustomerId()), zap.String("status", status))
	return nil
}

type notificationOutcomeHandler func(context.Context, string) (*domain.AgendaEvent, error)

func (consumer *NotificationOutcomeQueueConsumer) notificationOutcomeHandler(status notificationcontracts.CustomerNotificationOutcomeStatus) (notificationOutcomeHandler, string, error) {
	switch status {
	case notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_SENT:
		return consumer.service.ConfirmNotification, "sent", nil
	case notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_FAILED:
		return consumer.service.FailNotification, "failed", nil
	default:
		return nil, "", fmt.Errorf("unsupported notification outcome status %s", status.String())
	}
}
