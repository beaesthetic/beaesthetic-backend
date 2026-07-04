package application

import (
	"context"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

const (
	NotificationTypeConfirmation = "confirmation"
)

type ReminderScheduler interface {
	ScheduleReminder(ctx context.Context, eventID string, sendAt time.Time) error
	UnscheduleReminder(ctx context.Context, eventID string) error
}

type NotificationSender interface {
	Send(ctx context.Context, title, content, phoneNumber string) (string, error)
}

type AppointmentLifecycleHandler struct {
	service                *AppointmentService
	customers              CustomerRegistry
	scheduler              ReminderScheduler
	notifications          NotificationSender
	clock                  Clock
	noSendThreshold        time.Duration
	immediateSendThreshold time.Duration
}

func NewAppointmentLifecycleHandler(service *AppointmentService, customers CustomerRegistry, scheduler ReminderScheduler, notifications NotificationSender, clock Clock, noSendThreshold time.Duration, immediateSendThreshold time.Duration) *AppointmentLifecycleHandler {
	return &AppointmentLifecycleHandler{
		service:                service,
		customers:              customers,
		scheduler:              scheduler,
		notifications:          notifications,
		clock:                  clock,
		noSendThreshold:        noSendThreshold,
		immediateSendThreshold: immediateSendThreshold,
	}
}

func (h *AppointmentLifecycleHandler) Handle(ctx context.Context, eventType, eventID string) error {
	if eventID == "" {
		return nil
	}
	agendaEvent, err := h.service.GetAgenda(ctx, eventID)
	if err != nil || agendaEvent == nil {
		return err
	}

	switch eventType {
	case "AgendaEventScheduled":
		return h.handleScheduled(ctx, agendaEvent, false)
	case "AgendaEventRescheduled":
		return h.handleScheduled(ctx, agendaEvent, true)
	case "AgendaEventDeleted":
		return h.scheduler.UnscheduleReminder(ctx, agendaEvent.ID)
	default:
		return nil
	}
}

func (h *AppointmentLifecycleHandler) handleScheduled(ctx context.Context, agendaEvent *domain.AgendaEvent, isRescheduled bool) error {
	reminderErr := h.scheduleReminder(ctx, agendaEvent)
	confirmationErr := h.sendConfirmationNotification(ctx, agendaEvent, isRescheduled)
	if reminderErr != nil {
		return reminderErr
	}
	return confirmationErr
}

func (h *AppointmentLifecycleHandler) scheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	now := h.clock.Now().UTC()
	sendAt, ok := computeReminderSendAt(now, agendaEvent.Start, agendaEvent.RemindBefore, h.noSendThreshold, h.immediateSendThreshold)
	if !ok {
		agendaEvent.MarkReminderUnprocessable(now)
		return h.service.SaveAgendaEvent(ctx, agendaEvent)
	}
	if err := h.scheduler.ScheduleReminder(ctx, agendaEvent.ID, *sendAt); err != nil {
		return err
	}
	agendaEvent.MarkReminderScheduled(now)
	return h.service.SaveAgendaEvent(ctx, agendaEvent)
}

func (h *AppointmentLifecycleHandler) sendConfirmationNotification(ctx context.Context, agendaEvent *domain.AgendaEvent, isRescheduled bool) error {
	customer, err := h.customers.FindByCustomerID(ctx, agendaEvent.Attendee.ID)
	if err != nil {
		return err
	}
	if customer == nil || customer.PhoneNumber == nil {
		return nil
	}

	title, content := confirmationNotificationPayload(agendaEvent, isRescheduled)
	notificationID, err := h.notifications.Send(ctx, title, content, *customer.PhoneNumber)
	if err != nil {
		return err
	}
	return h.service.TrackPendingNotification(ctx, notificationID, agendaEvent.ID, NotificationTypeConfirmation)
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
