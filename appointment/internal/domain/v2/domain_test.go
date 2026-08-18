package v2

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeCalendarID(t *testing.T) {
	for _, value := range []string{"", DefaultCalendarID, "D2A36E25-4824-4167-A062-A5AF96F97703"} {
		got, err := NormalizeCalendarID(value)
		if err != nil {
			t.Fatalf("NormalizeCalendarID(%q) error = %v", value, err)
		}
		if got != DefaultCalendarID {
			t.Fatalf("NormalizeCalendarID(%q) = %q, want %q", value, got, DefaultCalendarID)
		}
	}

	if _, err := NormalizeCalendarID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"); !errors.Is(err, ErrInvalidCalendarID) {
		t.Fatalf("NormalizeCalendarID(other) error = %v, want ErrInvalidCalendarID", err)
	}
}

func TestAppointmentEventCreatesCalendarAndDetailTogether(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := NewTimeRange(now, now.Add(time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatalf("NewTimeRange() error = %v", err)
	}
	customer, err := NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatalf("NewCustomerRef() error = %v", err)
	}

	event, err := NewAppointmentEvent(AppointmentEventParams{
		EventID:    "event-1",
		CalendarID: DefaultCalendarID,
		Range:      eventRange,
		Title:      "Jane Doe",
		Customer:   customer,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("NewAppointmentEvent() error = %v", err)
	}
	if event.IsCanceled() {
		t.Fatalf("new event should not be canceled")
	}
	if event.Visibility != VisibilityPrivate {
		t.Fatalf("visibility = %s, want %s", event.Visibility, VisibilityPrivate)
	}
	if _, ok := event.Detail.(Appointment); !ok {
		t.Fatalf("unexpected detail: %#v", event.Detail)
	}
}

func TestAppointmentOwnsCustomerAndServicesOnly(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	customer, err := NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatalf("NewCustomerRef() error = %v", err)
	}
	service, err := NewServiceItem(nil, "Haircut", 99)
	if err != nil {
		t.Fatalf("NewServiceItem() error = %v", err)
	}

	appointment, err := NewAppointment(customer, []ServiceItem{service}, now)
	if err != nil {
		t.Fatalf("NewAppointment() error = %v", err)
	}
	if appointment.Services[0].Position != 0 {
		t.Fatalf("service position = %d, want 0", appointment.Services[0].Position)
	}
}

func TestManualCalendarEventCreatesCalendarAndDetailTogether(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := NewTimeRange(now, now.Add(time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatalf("NewTimeRange() error = %v", err)
	}
	event, err := NewManualCalendarEvent(ManualEventParams{
		EventID:    "event-1",
		CalendarID: DefaultCalendarID,
		Range:      eventRange,
		EventTitle: "Internal work",
		Title:      "Internal work",
		Now:        now,
	})
	if err != nil {
		t.Fatalf("NewManualCalendarEvent() error = %v", err)
	}
	detail, ok := event.Detail.(ManualEvent)
	if !ok || event.Type != CalendarEventTypeManual || detail.Title != "Internal work" {
		t.Fatalf("unexpected calendar aggregate: %#v", event)
	}
}

func TestCalendarEventEmitsLifecycleEvents(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := NewTimeRange(now, now.Add(time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatalf("NewTimeRange() error = %v", err)
	}
	customer, err := NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatalf("NewCustomerRef() error = %v", err)
	}
	event, err := NewAppointmentEvent(AppointmentEventParams{
		EventID:    "event-1",
		CalendarID: DefaultCalendarID,
		Range:      eventRange,
		Title:      "Jane Doe",
		Customer:   customer,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("NewAppointmentEvent() error = %v", err)
	}
	events := event.PullEvents()
	if len(events) != 1 || events[0].Type != "CalendarEventCreated" {
		t.Fatalf("created events = %#v", events)
	}

	newRange, err := NewTimeRange(now.Add(2*time.Hour), now.Add(3*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatalf("NewTimeRange() error = %v", err)
	}
	event.Reschedule(newRange, now)
	events = event.PullEvents()
	if len(events) != 1 || events[0].Type != "CalendarEventRescheduled" {
		t.Fatalf("rescheduled events = %#v", events)
	}
}

func TestCalendarEventRescheduleDoesNothingWhenTimeRangeIsUnchanged(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := NewTimeRange(now, now.Add(time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatalf("NewTimeRange() error = %v", err)
	}
	customer, err := NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatalf("NewCustomerRef() error = %v", err)
	}
	event, err := NewAppointmentEvent(AppointmentEventParams{
		EventID:    "event-1",
		CalendarID: DefaultCalendarID,
		Range:      eventRange,
		Title:      "Jane Doe",
		Customer:   customer,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("NewAppointmentEvent() error = %v", err)
	}
	event.PullEvents()

	event.Reschedule(eventRange, now.Add(time.Hour))

	if !event.UpdatedAt.Equal(now.UTC()) {
		t.Fatalf("updated at = %s, want %s", event.UpdatedAt, now.UTC())
	}
	if events := event.PullEvents(); len(events) != 0 {
		t.Fatalf("reschedule events = %#v, want none", events)
	}
}

func TestAppointmentReminderHasIndependentLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	reminder, err := NewAppointmentReminder(time.Hour, now)
	if err != nil {
		t.Fatalf("NewAppointmentReminder() error = %v", err)
	}

	sendAt := now.Add(30 * time.Minute)
	if err := reminder.Schedule(sendAt, now); err != nil {
		t.Fatalf("ScheduleReminder() error = %v", err)
	}
	if reminder.Status != ReminderStatusScheduled {
		t.Fatalf("reminder status = %s, want %s", reminder.Status, ReminderStatusScheduled)
	}
}

func TestAppointmentNotificationMapsKindToType(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	recipient, err := NewCustomerNotificationRecipient("customer-1")
	if err != nil {
		t.Fatalf("NewCustomerNotificationRecipient() error = %v", err)
	}
	notification, err := NewAppointmentNotification(
		"notification-1",
		"appointment-1",
		NotificationKindRescheduled,
		recipient,
		nil,
		now,
		now.Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("NewAppointmentNotification() error = %v", err)
	}
	if notification.Type != NotificationTypeAppointmentRescheduled {
		t.Fatalf("notification type = %s, want %s", notification.Type, NotificationTypeAppointmentRescheduled)
	}
	if notification.CalendarEventID != "appointment-1" {
		t.Fatalf("calendar event id = %s, want appointment-1", notification.CalendarEventID)
	}

	notification.MarkSent(now.Add(time.Minute))
	if notification.Status != NotificationStatusSent {
		t.Fatalf("notification status = %s, want %s", notification.Status, NotificationStatusSent)
	}
}

func TestReminderCannotBeScheduledInThePast(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	reminder, err := NewAppointmentReminder(time.Hour, now)
	if err != nil {
		t.Fatalf("NewAppointmentReminder() error = %v", err)
	}
	if err := reminder.Schedule(now.Add(-time.Minute), now); err != ErrInvalidReminder {
		t.Fatalf("Schedule() error = %v, want %v", err, ErrInvalidReminder)
	}
}
