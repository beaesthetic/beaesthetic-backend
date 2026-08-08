package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	notification "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCustomerNotificationSenderPublishesAppointmentReminderContractMessage(t *testing.T) {
	publisher := &publisherStub{}
	sender := NewCustomerNotificationSender(publisher)
	agendaEvent := &domain.AgendaEvent{
		ID:       "event-1",
		Start:    time.Date(2026, time.July, 4, 13, 30, 0, 0, time.UTC),
		End:      time.Date(2026, time.July, 4, 14, 30, 0, 0, time.UTC),
		Attendee: domain.Attendee{ID: "customer-1"},
	}

	id, err := sender.SendAppointmentReminder(context.Background(), agendaEvent)
	if err != nil {
		t.Fatalf("SendAppointmentReminder() error = %v", err)
	}
	if id != "appointment:event-1:appointment_reminder:2026-07-04T13:30:00Z" {
		t.Fatalf("idempotency key = %q", id)
	}
	if len(publisher.messages) != 1 {
		t.Fatalf("published messages = %d, want 1", len(publisher.messages))
	}
	message := publisher.messages[0]
	if message.Channel != outbox.Channel(ChannelCustomerNotifications) {
		t.Fatalf("channel = %q", message.Channel)
	}
	if message.AffinityKey != outbox.AffinityKey("customer-1") {
		t.Fatalf("affinity key = %q", message.AffinityKey)
	}

	var payload notification.CustomerNotificationRequested
	if err := protojson.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("payload is not CustomerNotificationRequested: %v", err)
	}
	if payload.GetNotificationChannel() != notification.NotificationChannel_NOTIFICATION_CHANNEL_SMS {
		t.Fatalf("notification channel = %s", payload.GetNotificationChannel())
	}
	if payload.GetNotificationType() != application.NotificationTypeAppointmentReminder {
		t.Fatalf("notification type = %q", payload.GetNotificationType())
	}
	bodyMap := payload.GetBody().AsMap()
	if got := bodyMap["eventId"]; got != "event-1" {
		t.Fatalf("body eventId = %v", got)
	}
	if got := bodyMap["startAt"]; got != "2026-07-04T13:30:00Z" {
		t.Fatalf("body startAt = %v", got)
	}
	if got := bodyMap["endAt"]; got != "2026-07-04T14:30:00Z" {
		t.Fatalf("body endAt = %v", got)
	}
}

func TestCustomerNotificationSenderPublishesV2CalendarNotification(t *testing.T) {
	publisher := &publisherStub{}
	sender := NewCustomerNotificationSender(publisher)
	now := time.Date(2026, time.July, 4, 13, 30, 0, 0, time.UTC)
	eventRange, err := domainv2.NewTimeRange(now, now.Add(time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	customer, err := domainv2.NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatal(err)
	}
	event, err := domainv2.NewAppointmentEvent(domainv2.AppointmentEventParams{EventID: "event-1", CalendarID: domainv2.DefaultCalendarID, Range: eventRange, Customer: customer, Now: now})
	if err != nil {
		t.Fatal(err)
	}

	id, err := sender.SendCalendarNotification(context.Background(), event, domainv2.NotificationTypeAppointmentReminder, "request-1")
	if err != nil {
		t.Fatalf("SendCalendarNotification() error = %v", err)
	}
	if id != "request-1" || len(publisher.messages) != 1 {
		t.Fatalf("id=%q messages=%d", id, len(publisher.messages))
	}
	var payload notification.CustomerNotificationRequested
	if err := protojson.Unmarshal(publisher.messages[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.GetCustomerIds()[0] != customer.ID || payload.GetNotificationType() != string(domainv2.NotificationTypeAppointmentReminder) {
		t.Fatalf("payload=%#v", payload)
	}
}

type publisherStub struct {
	messages []outbox.Message
}

func (p *publisherStub) Publish(ctx context.Context, msg ...outbox.Message) error {
	p.messages = append(p.messages, msg...)
	return nil
}

func (p *publisherStub) Close() error { return nil }
