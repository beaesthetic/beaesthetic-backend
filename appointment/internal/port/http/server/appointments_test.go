package server

import (
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	httpserver "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server/generated"
)

func TestToActivityResponseMapsAppointmentReminderStatusConsistently(t *testing.T) {
	tests := []struct {
		name             string
		status           domain.ReminderStatus
		wantReminderSent bool
		wantStatus       httpserver.AppointmentEventResponseReminderStatus
	}{
		{
			name:             "pending is not sent",
			status:           domain.ReminderPending,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
		{
			name:             "scheduled is not sent",
			status:           domain.ReminderScheduled,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
		{
			name:             "sent requested is send in progress",
			status:           domain.ReminderSentRequested,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusSENDINPROGRESS,
		},
		{
			name:             "sent is sent",
			status:           domain.ReminderSent,
			wantReminderSent: true,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusSENT,
		},
		{
			name:             "deleted is not sent",
			status:           domain.ReminderDeleted,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
		{
			name:             "fail to send is not sent",
			status:           domain.ReminderFailToSend,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
		{
			name:             "unprocessable is not sent",
			status:           domain.ReminderUnprocessable,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := toActivityResponse(domain.AgendaEvent{
				ID:             "f334b5fe-a663-40a4-824c-8f72d8a786a0",
				Type:           domain.EventTypeAppointment,
				Title:          "Appointment",
				Start:          time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
				End:            time.Date(2026, 7, 9, 11, 0, 0, 0, time.UTC),
				Attendee:       domain.Attendee{ID: "52d74c3f-20c5-44b0-92d5-3d451881a7f7", DisplayName: "Mario Rossi"},
				ReminderStatus: tt.status,
				RemindBefore:   24 * time.Hour,
			})

			appointment, err := response.AsAppointmentEventResponse()
			if err != nil {
				t.Fatalf("AsAppointmentEventResponse returned error: %v", err)
			}
			if appointment.ReminderSent != tt.wantReminderSent {
				t.Fatalf("ReminderSent = %v, want %v", appointment.ReminderSent, tt.wantReminderSent)
			}
			if appointment.Reminder.Status == nil {
				t.Fatal("Reminder.Status is nil")
			}
			if *appointment.Reminder.Status != tt.wantStatus {
				t.Fatalf("Reminder.Status = %s, want %s", *appointment.Reminder.Status, tt.wantStatus)
			}
		})
	}
}

func TestToActivityResponseMapsEventReminderStatusConsistently(t *testing.T) {
	tests := []struct {
		name             string
		status           domain.ReminderStatus
		wantReminderSent bool
		wantStatus       httpserver.EventResponseReminderStatus
	}{
		{
			name:             "pending is not sent",
			status:           domain.ReminderPending,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusNOTSENT,
		},
		{
			name:             "scheduled is not sent",
			status:           domain.ReminderScheduled,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusNOTSENT,
		},
		{
			name:             "sent requested is send in progress",
			status:           domain.ReminderSentRequested,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusSENDINPROGRESS,
		},
		{
			name:             "sent is sent",
			status:           domain.ReminderSent,
			wantReminderSent: true,
			wantStatus:       httpserver.EventResponseReminderStatusSENT,
		},
		{
			name:             "deleted is not sent",
			status:           domain.ReminderDeleted,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusNOTSENT,
		},
		{
			name:             "fail to send is not sent",
			status:           domain.ReminderFailToSend,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusNOTSENT,
		},
		{
			name:             "unprocessable is not sent",
			status:           domain.ReminderUnprocessable,
			wantReminderSent: false,
			wantStatus:       httpserver.EventResponseReminderStatusNOTSENT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := toActivityResponse(domain.AgendaEvent{
				ID:             "f334b5fe-a663-40a4-824c-8f72d8a786a0",
				Type:           domain.EventTypeGeneric,
				Title:          "Blocked slot",
				Description:    "Notes",
				Start:          time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
				End:            time.Date(2026, 7, 9, 11, 0, 0, 0, time.UTC),
				Attendee:       domain.Attendee{ID: "52d74c3f-20c5-44b0-92d5-3d451881a7f7", DisplayName: "Mario Rossi"},
				ReminderStatus: tt.status,
				RemindBefore:   24 * time.Hour,
			})

			event, err := response.AsEventResponse()
			if err != nil {
				t.Fatalf("AsEventResponse returned error: %v", err)
			}
			if event.ReminderSent != tt.wantReminderSent {
				t.Fatalf("ReminderSent = %v, want %v", event.ReminderSent, tt.wantReminderSent)
			}
			if event.Reminder.Status == nil {
				t.Fatal("Reminder.Status is nil")
			}
			if *event.Reminder.Status != tt.wantStatus {
				t.Fatalf("Reminder.Status = %s, want %s", *event.Reminder.Status, tt.wantStatus)
			}
		})
	}
}
