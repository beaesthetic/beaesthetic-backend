package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

var (
	ErrCalendarEventNotFound    = errors.New("calendar event not found")
	ErrAppointmentNotRemindable = errors.New("appointment cannot receive reminders")
	ErrInvalidReminderRequest   = errors.New("invalid reminder request")
)

type CalendarReminderScheduler interface {
	ScheduleCalendarReminder(ctx context.Context, calendarEventID string, expectedStartAt time.Time, sendAt time.Time) error
	UnscheduleCalendarReminder(ctx context.Context, calendarEventID string) error
}

type CalendarNotificationSender interface {
	SendCalendarNotification(ctx context.Context, event domain.CalendarEvent, notificationType domain.NotificationType, idempotencyKey string) (string, error)
}

type AppointmentLifecycleRepository interface {
	Tx(ctx context.Context, atomicFn func(context.Context) error) error
	FindCalendarEventView(ctx context.Context, calendarEventID string) (*CalendarEventView, error)
	SaveAppointmentReminderState(ctx context.Context, calendarEventID string, reminder domain.AppointmentReminder) error
	FindAppointmentNotification(ctx context.Context, correlationKey string) (*domain.AppointmentNotification, error)
	SaveAppointmentNotification(ctx context.Context, notification domain.AppointmentNotification) error
}

type AppointmentLifecycleService struct {
	repository             AppointmentLifecycleRepository
	scheduler              CalendarReminderScheduler
	notifications          CalendarNotificationSender
	clock                  Clock
	noSendThreshold        time.Duration
	immediateSendThreshold time.Duration
}

func NewAppointmentLifecycleService(repository AppointmentLifecycleRepository, scheduler CalendarReminderScheduler, notifications CalendarNotificationSender, clock Clock, noSendThreshold time.Duration, immediateSendThreshold time.Duration) *AppointmentLifecycleService {
	return &AppointmentLifecycleService{
		repository:             repository,
		scheduler:              scheduler,
		notifications:          notifications,
		clock:                  clock,
		noSendThreshold:        noSendThreshold,
		immediateSendThreshold: immediateSendThreshold,
	}
}

func (s *AppointmentLifecycleService) Handle(ctx context.Context, eventType string, calendarEventID string) error {
	switch eventType {
	case "CalendarEventCreated":
		return s.handleScheduled(ctx, calendarEventID, domain.NotificationKindConfirmation)
	case "CalendarEventRescheduled":
		return s.handleScheduled(ctx, calendarEventID, domain.NotificationKindRescheduled)
	case "CalendarEventCanceled":
		return s.handleCanceled(ctx, calendarEventID)
	default:
		return nil
	}
}

func (s *AppointmentLifecycleService) SendDueReminder(ctx context.Context, calendarEventID string, expectedStartAt *time.Time) error {
	return s.repository.Tx(ctx, func(ctx context.Context) error {
		view, appointment, err := s.findAppointment(ctx, calendarEventID)
		if errors.Is(err, ErrAppointmentNotRemindable) {
			return nil
		}
		if err != nil {
			return err
		}
		if view.Event.IsCanceled() || view.Reminder == nil || view.Reminder.Status != domain.ReminderStatusScheduled {
			return nil
		}
		if expectedStartAt != nil && !view.Event.Range.Start.Equal(expectedStartAt.UTC()) {
			return nil
		}
		if err := s.sendNotification(ctx, view.Event, appointment, domain.NotificationKindReminder, notificationIdempotencyKey(view.Event, domain.NotificationTypeAppointmentReminder)); err != nil {
			return err
		}
		view.Reminder.MarkSendRequested(s.clock.Now())
		return s.repository.SaveAppointmentReminderState(ctx, view.Event.ID, *view.Reminder)
	})
}

func (s *AppointmentLifecycleService) RequestReminderResend(ctx context.Context, calendarEventID string, requestKey string) error {
	if calendarEventID == "" || requestKey == "" {
		return ErrInvalidReminderRequest
	}
	return s.repository.Tx(ctx, func(ctx context.Context) error {
		view, appointment, err := s.findAppointment(ctx, calendarEventID)
		if err != nil {
			return err
		}
		if view.Event.IsCanceled() || view.Reminder == nil {
			return ErrAppointmentNotRemindable
		}
		idempotencyKey := fmt.Sprintf("appointment:%s:reminder:resend:%s", view.Event.ID, requestKey)
		if err := s.sendNotification(ctx, view.Event, appointment, domain.NotificationKindReminder, idempotencyKey); err != nil {
			return err
		}
		view.Reminder.MarkSendRequested(s.clock.Now())
		return s.repository.SaveAppointmentReminderState(ctx, view.Event.ID, *view.Reminder)
	})
}

func (s *AppointmentLifecycleService) HandleNotificationOutcome(ctx context.Context, correlationKey string, sent bool, reason string, message string) (string, error) {
	var calendarEventID string
	err := s.repository.Tx(ctx, func(ctx context.Context) error {
		notification, err := s.repository.FindAppointmentNotification(ctx, correlationKey)
		if err != nil || notification == nil {
			return err
		}
		calendarEventID = notification.CalendarEventID
		now := s.clock.Now()
		if sent {
			notification.MarkSent(now)
		} else {
			notification.MarkFailed(reason, message, now)
		}
		if err := s.repository.SaveAppointmentNotification(ctx, *notification); err != nil {
			return err
		}
		if notification.Kind != domain.NotificationKindReminder {
			return nil
		}
		view, err := s.repository.FindCalendarEventView(ctx, notification.CalendarEventID)
		if err != nil {
			return err
		}
		if view == nil || view.Reminder == nil {
			return ErrCalendarEventNotFound
		}
		if sent {
			view.Reminder.MarkSent(now)
		} else {
			view.Reminder.MarkFailed(reason, now)
		}
		return s.repository.SaveAppointmentReminderState(ctx, view.Event.ID, *view.Reminder)
	})
	return calendarEventID, err
}

func (s *AppointmentLifecycleService) handleScheduled(ctx context.Context, calendarEventID string, notificationKind domain.NotificationKind) error {
	return s.repository.Tx(ctx, func(ctx context.Context) error {
		view, appointment, err := s.findAppointment(ctx, calendarEventID)
		if errors.Is(err, ErrAppointmentNotRemindable) {
			return nil
		}
		if err != nil {
			return err
		}
		if view.Event.IsCanceled() || view.Reminder == nil {
			return nil
		}
		now := s.clock.Now()
		sendAt, sendable := computeCalendarReminderSendAt(now, view.Event.Range.Start, view.Reminder.RemindBefore, s.noSendThreshold, s.immediateSendThreshold)
		if !sendable {
			view.Reminder.MarkUnprocessable("too_late", now)
		} else {
			if err := s.scheduler.ScheduleCalendarReminder(ctx, view.Event.ID, view.Event.Range.Start, *sendAt); err != nil {
				return err
			}
			if err := view.Reminder.Schedule(*sendAt, now); err != nil {
				return err
			}
		}
		if err := s.repository.SaveAppointmentReminderState(ctx, view.Event.ID, *view.Reminder); err != nil {
			return err
		}
		notificationType, _ := notificationTypeForKind(notificationKind)
		return s.sendNotification(ctx, view.Event, appointment, notificationKind, notificationIdempotencyKey(view.Event, notificationType))
	})
}

func (s *AppointmentLifecycleService) handleCanceled(ctx context.Context, calendarEventID string) error {
	return s.repository.Tx(ctx, func(ctx context.Context) error {
		view, _, err := s.findAppointment(ctx, calendarEventID)
		if errors.Is(err, ErrAppointmentNotRemindable) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.scheduler.UnscheduleCalendarReminder(ctx, view.Event.ID); err != nil {
			return err
		}
		if view.Reminder == nil {
			return nil
		}
		view.Reminder.MarkDeleted(s.clock.Now())
		return s.repository.SaveAppointmentReminderState(ctx, view.Event.ID, *view.Reminder)
	})
}

func (s *AppointmentLifecycleService) findAppointment(ctx context.Context, calendarEventID string) (*CalendarEventView, domain.Appointment, error) {
	view, err := s.repository.FindCalendarEventView(ctx, calendarEventID)
	if err != nil {
		return nil, domain.Appointment{}, err
	}
	if view == nil {
		return nil, domain.Appointment{}, ErrCalendarEventNotFound
	}
	appointment, ok := view.Event.Detail.(domain.Appointment)
	if !ok {
		return view, domain.Appointment{}, ErrAppointmentNotRemindable
	}
	return view, appointment, nil
}

func (s *AppointmentLifecycleService) sendNotification(ctx context.Context, event domain.CalendarEvent, appointment domain.Appointment, kind domain.NotificationKind, idempotencyKey string) error {
	notificationType, ok := notificationTypeForKind(kind)
	if !ok {
		return domain.ErrInvalidNotification
	}
	correlationKey, err := s.notifications.SendCalendarNotification(ctx, event, notificationType, idempotencyKey)
	if err != nil {
		return err
	}
	recipient, err := domain.NewCustomerNotificationRecipient(appointment.Customer.ID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	notification, err := domain.NewAppointmentNotification(correlationKey, event.ID, kind, recipient, &idempotencyKey, now, now.Add(24*time.Hour))
	if err != nil {
		return err
	}
	return s.repository.SaveAppointmentNotification(ctx, notification)
}

func notificationTypeForKind(kind domain.NotificationKind) (domain.NotificationType, bool) {
	switch kind {
	case domain.NotificationKindConfirmation:
		return domain.NotificationTypeAppointmentConfirmation, true
	case domain.NotificationKindRescheduled:
		return domain.NotificationTypeAppointmentRescheduled, true
	case domain.NotificationKindReminder:
		return domain.NotificationTypeAppointmentReminder, true
	default:
		return "", false
	}
}

func notificationIdempotencyKey(event domain.CalendarEvent, notificationType domain.NotificationType) string {
	return fmt.Sprintf("appointment:%s:%s:%s", event.ID, notificationType, event.Range.Start.UTC().Format(time.RFC3339))
}

func computeCalendarReminderSendAt(now time.Time, eventAt time.Time, sendBefore time.Duration, noSendThreshold time.Duration, immediateSendThreshold time.Duration) (*time.Time, bool) {
	potentialRemindDate := eventAt.UTC().Add(-sendBefore)
	millisFromNow := now.UTC().Sub(potentialRemindDate)
	delta := sendBefore - millisFromNow
	switch {
	case delta < noSendThreshold:
		return nil, false
	case delta >= sendBefore:
		value := potentialRemindDate
		return &value, true
	default:
		value := now.UTC().Add(immediateSendThreshold)
		return &value, true
	}
}
