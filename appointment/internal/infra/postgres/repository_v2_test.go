package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres/queries"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

func TestNewCalendarLifecycleOutboxMessageUsesInternalJobChannel(t *testing.T) {
	message, err := newCalendarLifecycleOutboxMessage(domainv2.CalendarEventCreated("event-1"))
	if err != nil {
		t.Fatalf("newCalendarLifecycleOutboxMessage() error = %v", err)
	}
	if message.Channel != outbox.Channel(ChannelAppointmentInternalJob) {
		t.Fatalf("channel = %s, want %s", message.Channel, ChannelAppointmentInternalJob)
	}
	if message.AffinityKey != outbox.AffinityKey("event-1") {
		t.Fatalf("affinity key = %s, want event-1", message.AffinityKey)
	}

	var payload struct {
		Type            string `json:"type"`
		CalendarEventID string `json:"calendarEventId"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("payload is not json: %v", err)
	}
	if payload.Type != "CalendarEventCreated" || payload.CalendarEventID != "event-1" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestAppointmentReminderV2FromDetailsReconstitutesFullState(t *testing.T) {
	scheduledAt := time.Date(2026, 8, 8, 9, 0, 0, 0, time.UTC)
	sentRequestedAt := scheduledAt.Add(time.Hour)
	row := queries.FindAgendaEventFromDetailsRow{
		ReminderStatus:          pgtype.Text{String: "send_requested", Valid: true},
		RemindBeforeSeconds:     pgtype.Int4{Int32: 86400, Valid: true},
		ReminderScheduledAt:     pgtype.Timestamptz{Time: scheduledAt, Valid: true},
		ReminderSentRequestedAt: pgtype.Timestamptz{Time: sentRequestedAt, Valid: true},
		ReminderUpdatedAt:       pgtype.Timestamptz{Time: sentRequestedAt, Valid: true},
	}

	reminder, err := appointmentReminderV2FromDetails(row)
	if err != nil {
		t.Fatalf("appointmentReminderV2FromDetails() error = %v", err)
	}
	if reminder == nil || reminder.Status != domainv2.ReminderStatusSendRequested || reminder.RemindBefore != 24*time.Hour {
		t.Fatalf("reminder = %#v", reminder)
	}
	if reminder.ScheduledAt == nil || !reminder.ScheduledAt.Equal(scheduledAt) || reminder.SentRequestedAt == nil || !reminder.SentRequestedAt.Equal(sentRequestedAt) {
		t.Fatalf("reminder timestamps = %#v", reminder)
	}
}

func TestServiceItemsV2FromJSONKeepsServiceFields(t *testing.T) {
	data := `[{"Name":"Haircut","serviceId":"service-1","position":0}]`

	items, err := serviceItemsV2FromJSON(data)
	if err != nil {
		t.Fatalf("serviceItemsV2FromJSON() error = %v", err)
	}
	if len(items) != 1 || items[0].ServiceID == nil || *items[0].ServiceID != "service-1" {
		t.Fatalf("service items = %#v", items)
	}
}
