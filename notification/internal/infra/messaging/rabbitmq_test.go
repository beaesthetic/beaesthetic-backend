package messaging

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNotificationEventFromDelivery(t *testing.T) {
	tests := []struct {
		name     string
		delivery amqp.Delivery
		wantID   string
		wantErr  bool
	}{
		{
			name: "outbox message",
			delivery: amqp.Delivery{
				Type: "notifications",
				Body: []byte(`{"notificationId":"notification-id"}`),
			},
			wantID: "notification-id",
		},
		{
			name: "legacy message without type",
			delivery: amqp.Delivery{
				Body: []byte(`{"notificationId":"legacy-id"}`),
			},
			wantID: "legacy-id",
		},
		{
			name: "unsupported channel",
			delivery: amqp.Delivery{
				Type: "notifications.confirmed",
				Body: []byte(`{"notificationId":"notification-id"}`),
			},
			wantErr: true,
		},
		{
			name: "invalid payload",
			delivery: amqp.Delivery{
				Type: "notifications",
				Body: []byte(`not-json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := notificationEventFromDelivery(tt.delivery)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if event.NotificationID != tt.wantID {
				t.Fatalf("NotificationID = %q, want %q", event.NotificationID, tt.wantID)
			}
		})
	}
}
