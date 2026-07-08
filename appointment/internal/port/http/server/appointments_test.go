package server

import (
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	httpserver "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server/generated"
)

func TestToActivityResponseMapsReminderStatusConsistently(t *testing.T) {
	tests := []struct {
		name             string
		status           domain.ReminderStatus
		wantReminderSent bool
		wantStatus       httpserver.AppointmentEventResponseReminderStatus
	}{
		{
			name:             "not sent",
			status:           domain.ReminderPending,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusNOTSENT,
		},
		{
			name:             "send in progress",
			status:           domain.ReminderSentRequested,
			wantReminderSent: false,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusSENDINPROGRESS,
		},
		{
			name:             "sent",
			status:           domain.ReminderSent,
			wantReminderSent: true,
			wantStatus:       httpserver.AppointmentEventResponseReminderStatusSENT,
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
