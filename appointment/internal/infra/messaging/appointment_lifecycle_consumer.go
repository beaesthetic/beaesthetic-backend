package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const channelAppointmentLifecycle = "appointment.lifecycle"

type AppointmentLifecycleConsumer struct {
	handler *application.AppointmentLifecycleHandler
	log     *zap.Logger
}

func NewAppointmentLifecycleConsumer(handler *application.AppointmentLifecycleHandler, log *zap.Logger) *AppointmentLifecycleConsumer {
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
	if event.AgendaEventID == "" {
		consumer.log.Debug("unhandled lifecycle event without event id")
		return nil
	}

	if err := consumer.handler.Handle(ctx, event.Type, event.AgendaEventID); err != nil {
		consumer.log.Error("failed to handle appointment lifecycle event", zap.String("event_id", event.AgendaEventID), zap.String("type", event.Type), zap.Error(err))
	}
	return nil
}

func appointmentLifecycleEventFromDelivery(delivery amqp.Delivery) (appointmentLifecycleEvent, error) {
	message := outboxamqp.MessageFromDelivery(delivery)
	if message.Channel != "" && string(message.Channel) != channelAppointmentLifecycle {
		return appointmentLifecycleEvent{}, fmt.Errorf("unsupported outbox channel %q", message.Channel)
	}
	var event appointmentLifecycleEvent
	if err := json.Unmarshal(message.Payload, &event); err != nil {
		return appointmentLifecycleEvent{}, fmt.Errorf("parse appointment lifecycle event: %w", err)
	}
	return event, nil
}

type appointmentLifecycleEvent struct {
	Type          string `json:"type"`
	AgendaEventID string `json:"agendaEventId"`
}
