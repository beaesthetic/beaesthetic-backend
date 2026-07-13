package messaging

import (
	"testing"

	notification "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCustomerNotificationCommandFromDelivery(t *testing.T) {
	body, err := structpb.NewStruct(map[string]any{"date": "2026-07-20"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := protojson.Marshal(&notification.CustomerNotificationRequested{
		IdempotencyKey:      "external-key",
		CustomerIds:         []string{"customer-1"},
		NotificationChannel: notification.NotificationChannel_NOTIFICATION_CHANNEL_SMS,
		NotificationType:    "appointment_reminder",
		Body:                body,
	})
	if err != nil {
		t.Fatal(err)
	}
	command, err := customerNotificationCommandFromDelivery(amqp.Delivery{Body: payload})
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

func TestCustomerNotificationCommandFromDeliveryMapsEmailEnum(t *testing.T) {
	payload, err := protojson.Marshal(&notification.CustomerNotificationRequested{
		IdempotencyKey:      "external-key",
		CustomerIds:         []string{"customer-1"},
		NotificationChannel: notification.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL,
		NotificationType:    "appointment_reminder",
	})
	if err != nil {
		t.Fatal(err)
	}
	command, err := customerNotificationCommandFromDelivery(amqp.Delivery{Body: payload})
	if err != nil {
		t.Fatalf("customerNotificationCommandFromDelivery() error = %v", err)
	}
	if command.NotificationChannel != "email" {
		t.Fatalf("NotificationChannel = %q, want email", command.NotificationChannel)
	}
}

func TestCustomerNotificationCommandFromDeliveryRejectsInvalidJSON(t *testing.T) {
	_, err := customerNotificationCommandFromDelivery(amqp.Delivery{Body: []byte(`not-json`)})
	if err == nil {
		t.Fatal("expected error")
	}
}
