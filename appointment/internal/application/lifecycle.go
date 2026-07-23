package application

import (
	"context"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	"go.uber.org/zap"
)

const (
	NotificationTypeAppointmentConfirmation = "appointment_confirmation"
	NotificationTypeAppointmentReminder     = "appointment_reminder"
	NotificationTypeAppointmentRescheduled  = "appointment_rescheduled"
)

type ReminderScheduler interface {
	ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error
	UnscheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error
}

type NotificationSender interface {
	SendAppointmentReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error)
	SendAppointmentConfirmation(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error)
	SendAppointmentRescheduled(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error)
}

type TransactionRunner interface {
	Tx(ctx context.Context, atomicFn func(ctx context.Context) error) error
}

type AppointmentLifecycleHandler struct {
	service                *AppointmentService
	transactions           TransactionRunner
	customers              CustomerRegistry
	scheduler              ReminderScheduler
	notifications          NotificationSender
	clock                  Clock
	noSendThreshold        time.Duration
	immediateSendThreshold time.Duration
	log                    *zap.Logger
}

func NewAppointmentLifecycleHandler(service *AppointmentService, transactions TransactionRunner, customers CustomerRegistry, scheduler ReminderScheduler, notifications NotificationSender, clock Clock, noSendThreshold time.Duration, immediateSendThreshold time.Duration, log *zap.Logger) *AppointmentLifecycleHandler {
	if log == nil {
		log = zap.NewNop()
	}
	return &AppointmentLifecycleHandler{
		service:                service,
		transactions:           transactions,
		customers:              customers,
		scheduler:              scheduler,
		notifications:          notifications,
		clock:                  clock,
		noSendThreshold:        noSendThreshold,
		immediateSendThreshold: immediateSendThreshold,
		log:                    log.Named("appointment_lifecycle_handler"),
	}
}

func (h *AppointmentLifecycleHandler) Handle(ctx context.Context, eventType, eventID string) error {
	if eventID == "" {
		h.log.Debug("skipping appointment lifecycle event without event id", zap.String("type", eventType))
		return nil
	}

	h.log.Info("handling appointment lifecycle event", zap.String("type", eventType), zap.String("event_id", eventID))
	agendaEvent, err := h.service.GetAgenda(ctx, eventID)
	if err != nil {
		h.log.Error("failed to load agenda event for lifecycle event", zap.String("type", eventType), zap.String("event_id", eventID), zap.Error(err))
		return err
	}
	if agendaEvent == nil {
		h.log.Warn("appointment lifecycle event references missing agenda event", zap.String("type", eventType), zap.String("event_id", eventID))
		return nil
	}

	switch eventType {
	case "AgendaEventScheduled":
		return h.handleScheduled(ctx, agendaEvent, false)
	case "AgendaEventRescheduled":
		return h.handleScheduled(ctx, agendaEvent, true)
	case "AgendaEventDeleted":
		return h.handleDeleted(ctx, agendaEvent)
	default:
		h.log.Debug("ignoring unsupported appointment lifecycle event", zap.String("type", eventType), zap.String("event_id", eventID))
		return nil
	}
}

func (h *AppointmentLifecycleHandler) ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	if agendaEvent == nil {
		return nil
	}
	return h.transactions.Tx(ctx, func(ctx context.Context) error {
		return h.scheduleReminder(ctx, agendaEvent)
	})
}
func (h *AppointmentLifecycleHandler) handleScheduled(ctx context.Context, agendaEvent *domain.AgendaEvent, isRescheduled bool) error {
	h.log.Info("processing scheduled appointment lifecycle", zap.String("event_id", agendaEvent.ID), zap.Bool("rescheduled", isRescheduled))
	return h.transactions.Tx(ctx, func(ctx context.Context) error {
		reminderErr := h.scheduleReminder(ctx, agendaEvent)
		confirmationErr := h.sendConfirmationNotification(ctx, agendaEvent, isRescheduled)
		if reminderErr != nil {
			return reminderErr
		}
		return confirmationErr
	})
}

func (h *AppointmentLifecycleHandler) handleDeleted(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	return h.transactions.Tx(ctx, func(ctx context.Context) error {
		if err := h.scheduler.UnscheduleReminder(ctx, agendaEvent); err != nil {
			h.log.Error("failed to unschedule appointment reminder", zap.String("event_id", agendaEvent.ID), zap.Error(err))
			return err
		}
		agendaEvent.MarkReminderDeleted(h.clock.Now().UTC())
		if err := h.service.SaveAgendaEvent(ctx, agendaEvent); err != nil {
			return err
		}
		h.log.Info("unscheduled appointment reminder", zap.String("event_id", agendaEvent.ID))
		return nil
	})
}

func (h *AppointmentLifecycleHandler) scheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	now := h.clock.Now().UTC()
	sendAt, ok := computeReminderSendAt(now, agendaEvent.Start, agendaEvent.RemindBefore, h.noSendThreshold, h.immediateSendThreshold)
	if !ok {
		h.log.Info("appointment reminder is not sendable", zap.String("event_id", agendaEvent.ID), zap.Time("start_at", agendaEvent.Start), zap.Duration("remind_before", agendaEvent.RemindBefore))
		agendaEvent.MarkReminderUnprocessable(now)
		return h.service.SaveAgendaEvent(ctx, agendaEvent)
	}
	if err := h.scheduler.ScheduleReminder(ctx, agendaEvent, *sendAt); err != nil {
		h.log.Error("failed to schedule appointment reminder", zap.String("event_id", agendaEvent.ID), zap.Time("send_at", *sendAt), zap.Error(err))
		return err
	}
	h.log.Info("scheduled appointment reminder", zap.String("event_id", agendaEvent.ID), zap.Time("send_at", *sendAt))
	agendaEvent.MarkReminderScheduled(now)
	return h.service.SaveAgendaEvent(ctx, agendaEvent)
}

func (h *AppointmentLifecycleHandler) sendConfirmationNotification(ctx context.Context, agendaEvent *domain.AgendaEvent, isRescheduled bool) error {
	customer, err := h.customers.FindByCustomerID(ctx, agendaEvent.Attendee.ID)
	if err != nil {
		h.log.Error("failed to load attendee for confirmation notification", zap.String("event_id", agendaEvent.ID), zap.String("attendee_id", agendaEvent.Attendee.ID), zap.Error(err))
		return err
	}
	if customer == nil || customer.PhoneNumber == nil {
		h.log.Info("attendee has no valid contacts, not sending confirmation", zap.String("event_id", agendaEvent.ID), zap.String("attendee_id", agendaEvent.Attendee.ID))
		return nil
	}

	notificationType := confirmationNotificationType(isRescheduled)
	var correlationKey string
	if isRescheduled {
		correlationKey, err = h.notifications.SendAppointmentRescheduled(ctx, agendaEvent)
	} else {
		correlationKey, err = h.notifications.SendAppointmentConfirmation(ctx, agendaEvent)
	}
	if err != nil {
		h.log.Error("failed to send appointment confirmation", zap.String("event_id", agendaEvent.ID), zap.String("attendee_id", agendaEvent.Attendee.ID), zap.Bool("rescheduled", isRescheduled), zap.Error(err))
		return err
	}
	if err := h.service.TrackPendingNotification(ctx, correlationKey, agendaEvent.ID, notificationType); err != nil {
		h.log.Error("failed to track appointment confirmation", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey), zap.Error(err))
		return err
	}
	h.log.Info("sent appointment confirmation", zap.String("event_id", agendaEvent.ID), zap.String("correlation_key", correlationKey), zap.Bool("rescheduled", isRescheduled))
	return nil
}

func confirmationNotificationType(isRescheduled bool) string {
	if isRescheduled {
		return NotificationTypeAppointmentRescheduled
	}
	return NotificationTypeAppointmentConfirmation
}

func computeReminderSendAt(now time.Time, eventAt time.Time, sendBefore time.Duration, noSendThreshold time.Duration, immediateSendThreshold time.Duration) (*time.Time, bool) {
	potentialRemindDate := eventAt.UTC().Add(-sendBefore)
	millisFromNow := now.UTC().Sub(potentialRemindDate)
	delta := sendBefore - millisFromNow
	switch {
	case delta < noSendThreshold:
		return nil, false
	case delta >= sendBefore:
		v := potentialRemindDate
		return &v, true
	default:
		v := now.UTC().Add(immediateSendThreshold)
		return &v, true
	}
}
