package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type SchedulerQueueConsumer struct {
	service *application.AppointmentService
	log     *zap.Logger
}

func NewSchedulerQueueConsumer(service *application.AppointmentService, log *zap.Logger) *SchedulerQueueConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &SchedulerQueueConsumer{service: service, log: log.Named("scheduler_queue_consumer")}
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
	agendaEvent, err := consumer.service.ProcessReminderTimesUp(ctx, event.EventID)
	if err != nil {
		return err
	}
	consumer.log.Info("processed reminder times up", zap.String("event_id", agendaEvent.ID))
	return nil
}

type reminderTimesUpEvent struct {
	EventID string `json:"eventId"`
}
