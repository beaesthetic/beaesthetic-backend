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
	reminders *application.ReminderSender
	log       *zap.Logger
}

func NewSchedulerQueueConsumer(reminders *application.ReminderSender, log *zap.Logger) *SchedulerQueueConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &SchedulerQueueConsumer{reminders: reminders, log: log.Named("scheduler_queue_consumer")}
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
	return consumer.reminders.SendDueReminder(ctx, event.EventID, nil)
}

type reminderTimesUpEvent struct {
	EventID string `json:"eventId"`
}
