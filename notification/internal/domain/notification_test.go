package domain

import (
	"testing"
	"time"
)

func TestNotificationLifecycleEvents(t *testing.T) {
	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	notification, err := NewNotification("notification-id", "Welcome", "Hello", Channel{Type: ChannelSMS, Phone: "+393331234567"}, now)
	if err != nil {
		t.Fatalf("NewNotification() error = %v", err)
	}
	events := notification.PullEvents()
	if len(events) != 1 {
		t.Fatalf("created events = %d, want 1", len(events))
	}
	if _, ok := events[0].(NotificationCreated); !ok {
		t.Fatalf("created event = %T, want NotificationCreated", events[0])
	}

	notification.MarkSent(ChannelMetadata{ProviderResourceID: "provider-id"}, now.Add(time.Minute))
	events = notification.PullEvents()
	if len(events) != 1 {
		t.Fatalf("sent events = %d, want 1", len(events))
	}
	if _, ok := events[0].(NotificationSent); !ok {
		t.Fatalf("sent event = %T, want NotificationSent", events[0])
	}
	if !notification.IsSent {
		t.Fatal("notification should be sent")
	}

	notification.ConfirmSent(now.Add(2 * time.Minute))
	events = notification.PullEvents()
	if len(events) != 1 {
		t.Fatalf("confirmed events = %d, want 1", len(events))
	}
	if _, ok := events[0].(NotificationSentConfirmed); !ok {
		t.Fatalf("confirmed event = %T, want NotificationSentConfirmed", events[0])
	}
	if !notification.IsSentConfirmed {
		t.Fatal("notification should be confirmed")
	}
}

func TestChannelValidation(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		wantErr bool
	}{
		{name: "sms", channel: Channel{Type: ChannelSMS, Phone: "+393331234567"}},
		{name: "email", channel: Channel{Type: ChannelEmail, Email: "user@example.com"}},
		{name: "whatsapp", channel: Channel{Type: ChannelWhatsApp, Phone: "+393331234567"}},
		{name: "missing phone", channel: Channel{Type: ChannelSMS}, wantErr: true},
		{name: "missing email", channel: Channel{Type: ChannelEmail}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.channel.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
