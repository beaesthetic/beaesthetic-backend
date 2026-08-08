package server

import (
	"context"
	"testing"
	"time"

	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	appointmentcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/appointment"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestReminderBeforeFromProtoDefaultsTo24Hours(t *testing.T) {
	value, err := reminderBeforeFromProto(nil)
	if err != nil {
		t.Fatalf("reminderBeforeFromProto() error = %v", err)
	}
	if value != 24*time.Hour {
		t.Fatalf("reminder before = %s, want 24h", value)
	}
}

func TestReminderBeforeFromProtoRejectsZero(t *testing.T) {
	zero := int32(0)
	if _, err := reminderBeforeFromProto(&zero); err == nil {
		t.Fatal("reminderBeforeFromProto() error = nil, want invalid reminder error")
	}
}

func TestAppointmentReminderProtoMapsLifecycleState(t *testing.T) {
	scheduledAt := time.Date(2026, 8, 8, 9, 0, 0, 0, time.UTC)
	sentRequestedAt := scheduledAt.Add(time.Hour)
	reminder := &domain.AppointmentReminder{
		Status:          domain.ReminderStatusSendRequested,
		RemindBefore:    24 * time.Hour,
		ScheduledAt:     &scheduledAt,
		SentRequestedAt: &sentRequestedAt,
	}

	mapped := appointmentReminderProto(reminder)
	if mapped.GetStatus() != appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_SEND_REQUESTED {
		t.Fatalf("status = %s", mapped.GetStatus())
	}
	if mapped.GetRemindBeforeSeconds() != 86400 || !mapped.GetScheduledAt().AsTime().Equal(scheduledAt) || !mapped.GetSentRequestedAt().AsTime().Equal(sentRequestedAt) {
		t.Fatalf("mapped reminder = %#v", mapped)
	}
}

func TestUpdateCalendarEventCommandBuildsCommonFieldsCommand(t *testing.T) {
	command, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id:         "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title"}},
	})
	if err != nil {
		t.Fatalf("updateCalendarEventCommand() error = %v", err)
	}
	update, ok := command.(applicationv2.UpdateCalendarFieldsCommand)
	if !ok {
		t.Fatalf("command type = %T, want UpdateCalendarFieldsCommand", command)
	}
	if update.Changes.Title == nil || *update.Changes.Title != "" {
		t.Fatalf("title change = %#v, want an explicit empty value", update.Changes.Title)
	}
}

func TestUpdateCalendarEventCommandRejectsMixedDetailMasks(t *testing.T) {
	_, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id: "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"appointment.services",
			"time_block.reason",
		}},
	})
	if err == nil {
		t.Fatal("updateCalendarEventCommand() error = nil, want mixed detail mask error")
	}
}

func TestUpdateCalendarEventCommandClearsManualLocation(t *testing.T) {
	command, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id:         "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"manual_event.location"}},
		Detail: &appointmentcontracts.UpdateCalendarEventRequest_ManualEvent{
			ManualEvent: &appointmentcontracts.UpdateManualEventDetail{},
		},
	})
	if err != nil {
		t.Fatalf("updateCalendarEventCommand() error = %v", err)
	}
	update, ok := command.(applicationv2.UpdateManualEventCommand)
	if !ok {
		t.Fatalf("command type = %T, want UpdateManualEventCommand", command)
	}
	if update.Location == nil || *update.Location != nil {
		t.Fatalf("location change = %#v, want explicit nil location", update.Location)
	}
}
