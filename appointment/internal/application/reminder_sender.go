package application

import (
	"context"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	"go.uber.org/zap"
)

const pendingNotificationTypeReminder = "reminder"

type ReminderSender struct {
	service       *AppointmentService
	notifications NotificationSender
	log           *zap.Logger
}

func NewReminderSender(service *AppointmentService, notifications NotificationSender, log *zap.Logger) *ReminderSender {
	if log == nil {
		log = zap.NewNop()
	}
	return &ReminderSender{
		service:       service,
		notifications: notifications,
		log:           log.Named("reminder_sender"),
	}
}

func (s *ReminderSender) SendDueReminder(ctx context.Context, eventID string, expectedStartAt *time.Time) error {
	if eventID == "" {
		s.log.Debug("skipping reminder without event id")
		return nil
	}

	agendaEvent, err := s.service.GetAgenda(ctx, eventID)
	if err != nil {
		s.log.Error("failed to load agenda event for reminder", zap.String("event_id", eventID), zap.Error(err))
		return err
	}
	if agendaEvent == nil {
		s.log.Info("reminder event has no appointment", zap.String("event_id", eventID))
		return nil
	}
	if !isReminderSendable(agendaEvent, expectedStartAt) {
		s.log.Info(
			"skipping stale reminder",
			zap.String("event_id", agendaEvent.ID),
			zap.String("reminder_status", string(agendaEvent.ReminderStatus)),
			zap.Time("start_at", agendaEvent.Start),
			zap.Timep("expected_start_at", expectedStartAt),
			zap.Bool("canceled", agendaEvent.CancelReason != nil),
		)
		return nil
	}

	correlationKey, err := s.notifications.SendAppointmentReminder(ctx, agendaEvent)
	if err != nil {
		s.log.Error("failed to send reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("attendee_id", agendaEvent.Attendee.ID), zap.Error(err))
		return err
	}
	if err := s.service.TrackPendingNotification(ctx, correlationKey, agendaEvent.ID, pendingNotificationTypeReminder); err != nil {
		s.log.Error("failed to track reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey), zap.Error(err))
		return err
	}
	if _, err := s.service.ProcessReminderTimesUp(ctx, eventID); err != nil {
		s.log.Error("failed to mark reminder as processed", zap.String("event_id", eventID), zap.String("correlation_key", correlationKey), zap.Error(err))
		return err
	}

	s.log.Info("sent reminder notification", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey))
	return nil
}

func isReminderSendable(agendaEvent *domain.AgendaEvent, expectedStartAt *time.Time) bool {
	if agendaEvent.CancelReason != nil {
		return false
	}
	if agendaEvent.ReminderStatus != domain.ReminderScheduled {
		return false
	}
	if expectedStartAt != nil && !agendaEvent.Start.UTC().Equal(expectedStartAt.UTC()) {
		return false
	}
	return true
}
