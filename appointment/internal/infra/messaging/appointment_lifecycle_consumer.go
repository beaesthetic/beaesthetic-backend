package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const ChannelAppointmentInternalJob = "beaesthetic.appointments.internal.job"

type AppointmentLifecycleConsumer struct {
	handler applicationv2.LifecycleEventHandler
	log     *zap.Logger
}

func NewAppointmentLifecycleConsumer(handler applicationv2.LifecycleEventHandler, log *zap.Logger) *AppointmentLifecycleConsumer {
	if log == nil {
		log = zap.NewNop()
	}
	return &AppointmentLifecycleConsumer{handler: handler, log: log.Named("appointment_lifecycle_consumer")}
}

func (consumer *AppointmentLifecycleConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	event, err := appointmentLifecycleEventFromDelivery(delivery)
	if err != nil {
		return err
	}
	eventID := event.EventID()
	if eventID == "" {
		consumer.log.Debug("unhandled lifecycle event without event id", zap.String("type", event.Type))
		return nil
	}

	consumer.log.Info("received appointment lifecycle event", zap.String("event_id", eventID), zap.String("type", event.Type))
	if err := consumer.handler.Handle(ctx, event.Type, eventID); err != nil {
		consumer.log.Error("failed to handle appointment lifecycle event", zap.String("event_id", eventID), zap.String("type", event.Type), zap.Error(err))
		return nil
	}
	consumer.log.Info("processed appointment lifecycle event", zap.String("event_id", eventID), zap.String("type", event.Type))
	return nil
}

func appointmentLifecycleEventFromDelivery(delivery amqp.Delivery) (appointmentLifecycleEvent, error) {
	message := outboxamqp.MessageFromDelivery(delivery)
	if message.Channel != "" && string(message.Channel) != ChannelAppointmentInternalJob {
		return appointmentLifecycleEvent{}, fmt.Errorf("unsupported outbox channel %q", message.Channel)
	}
	var event appointmentLifecycleEvent
	if err := json.Unmarshal(message.Payload, &event); err != nil {
		return appointmentLifecycleEvent{}, fmt.Errorf("parse appointment lifecycle event: %w", err)
	}
	return event, nil
}

type appointmentLifecycleEvent struct {
	Type                string `json:"type"`
	CalendarEventID     string `json:"calendarEventId"`
	LegacyAgendaEventID string `json:"agendaEventId"`
}

func (event appointmentLifecycleEvent) EventID() string {
	if event.CalendarEventID != "" {
		return event.CalendarEventID
	}
	return event.LegacyAgendaEventID
}
