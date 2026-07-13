package messaging

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestCustomerNotificationCommandFromDelivery(t *testing.T) {
	command, err := customerNotificationCommandFromDelivery(amqp.Delivery{Body: []byte(`{
		"idempotencyKey":"external-key",
		"customerIds":["customer-1"],
		"notificationChannel":"sms",
		"notificationType":"appointment_reminder",
		"body":{"date":"2026-07-20"}
	}`)})
	if err != nil {
		t.Fatalf("customerNotificationCommandFromDelivery() error = %v", err)
	}
	if command.IdempotencyKey != "external-key" || command.NotificationChannel != "sms" || command.NotificationType != "appointment_reminder" {
		t.Fatalf("unexpected command: %+v", command)
	}
	if got := command.Body["date"]; got != "2026-07-20" {
		t.Fatalf("Body[date] = %v, want 2026-07-20", got)
	}
}

func TestCustomerNotificationCommandFromDeliveryRejectsInvalidJSON(t *testing.T) {
	_, err := customerNotificationCommandFromDelivery(amqp.Delivery{Body: []byte(`not-json`)})
	if err == nil {
		t.Fatal("expected error")
	}
}
