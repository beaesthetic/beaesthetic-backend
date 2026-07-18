package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	notificationTypeReminder = "reminder"
)

type SchedulerQueueConsumer struct {
	service       *application.AppointmentService
	notifications application.NotificationSender
	log           *zap.Logger
}

func NewSchedulerQueueConsumer(service *application.AppointmentService, notifications application.NotificationSender, log *zap.Logger) *SchedulerQueueConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &SchedulerQueueConsumer{service: service, notifications: notifications, log: log.Named("scheduler_queue_consumer")}
}

func (consumer *SchedulerQueueConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	var event reminderTimesUpEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return fmt.Errorf("parse reminder times up event: %w", err)
	}
	if event.EventID == "" {
		consumer.log.Debug("unhandled reminder event without event id")
		return nil
	}

	consumer.log.Info("received reminder times up event", zap.String("event_id", event.EventID))
	agendaEvent, err := consumer.service.GetAgenda(ctx, event.EventID)
	if err != nil {
		consumer.log.Error("failed to load agenda event for reminder", zap.String("event_id", event.EventID), zap.Error(err))
		return err
	}
	if agendaEvent == nil {
		consumer.log.Info("reminder event has no appointment", zap.String("event_id", event.EventID))
		return nil
	}

	correlationKey, err := consumer.notifications.SendAppointmentReminder(ctx, agendaEvent)
	if err != nil {
		consumer.log.Error("failed to send reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("attendee_id", agendaEvent.Attendee.ID), zap.Error(err))
		return err
	}
	if err := consumer.service.TrackPendingNotification(ctx, correlationKey, agendaEvent.ID, notificationTypeReminder); err != nil {
		consumer.log.Error("failed to track reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey), zap.Error(err))
		return err
	}
	if _, err := consumer.service.ProcessReminderTimesUp(ctx, event.EventID); err != nil {
		consumer.log.Error("failed to mark reminder as processed", zap.String("event_id", event.EventID), zap.String("correlation_key", correlationKey), zap.Error(err))
		return err
	}

	consumer.log.Info("sent reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey))
	return nil
}

type reminderTimesUpEvent struct {
	EventID string `json:"eventId"`
}
