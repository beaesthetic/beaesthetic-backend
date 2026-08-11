package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type repositoryStub struct {
	txCalls          int
	ids              []string
	found            *domain.CalendarEvent
	saved            []*domain.CalendarEvent
	appointmentSaved bool
	manualSaved      bool
	timeBlockSaved   bool
	reminders        map[string]domain.AppointmentReminder
	notifications    map[string]domain.AppointmentNotification
	inTx             bool
	writesOutsideTx  int
}

type customerResolverStub struct {
	customer domain.CustomerRef
	calls    int
}

func (r *customerResolverStub) ResolveCustomer(context.Context, string) (domain.CustomerRef, error) {
	r.calls++
	return r.customer, nil
}

func (r *repositoryStub) NextCalendarEventID() string {
	value := r.ids[0]
	r.ids = r.ids[1:]
	return value
}

func (r *repositoryStub) Tx(ctx context.Context, atomicFn func(context.Context) error) error {
	r.txCalls++
	r.inTx = true
	defer func() { r.inTx = false }()
	return atomicFn(ctx)
}

func (r *repositoryStub) FindCalendarEvent(context.Context, string) (*domain.CalendarEvent, error) {
	return r.found, nil
}

func (r *repositoryStub) FindCalendarEventView(_ context.Context, calendarEventID string) (*CalendarEventView, error) {
	if r.found == nil {
		return nil, nil
	}
	reminder, ok := r.reminders[calendarEventID]
	if !ok {
		return &CalendarEventView{Event: *r.found}, nil
	}
	return &CalendarEventView{Event: *r.found, Reminder: &reminder}, nil
}

func (r *repositoryStub) SearchCalendarEventViews(ctx context.Context, _ ListCalendarEventsQuery) ([]CalendarEventView, error) {
	view, err := r.FindCalendarEventView(ctx, "event-1")
	if err != nil || view == nil {
		return nil, err
	}
	return []CalendarEventView{*view}, nil
}

func (r *repositoryStub) SaveCalendarEvent(_ context.Context, event *domain.CalendarEvent) error {
	if !r.inTx {
		r.writesOutsideTx++
	}
	r.saved = append(r.saved, event)
	switch event.Detail.(type) {
	case domain.Appointment:
		r.appointmentSaved = true
	case domain.ManualEvent:
		r.manualSaved = true
	case domain.TimeBlock:
		r.timeBlockSaved = true
	}
	return nil
}

func (r *repositoryStub) SaveAppointmentReminderState(_ context.Context, calendarEventID string, reminder domain.AppointmentReminder) error {
	if !r.inTx {
		r.writesOutsideTx++
	}
	if r.reminders == nil {
		r.reminders = make(map[string]domain.AppointmentReminder)
	}
	r.reminders[calendarEventID] = reminder
	return nil
}

func (r *repositoryStub) FindAppointmentNotification(_ context.Context, correlationKey string) (*domain.AppointmentNotification, error) {
	notification, ok := r.notifications[correlationKey]
	if !ok {
		return nil, nil
	}
	return &notification, nil
}

func (r *repositoryStub) SaveAppointmentNotification(_ context.Context, notification domain.AppointmentNotification) error {
	if !r.inTx {
		r.writesOutsideTx++
	}
	if r.notifications == nil {
		r.notifications = make(map[string]domain.AppointmentNotification)
	}
	r.notifications[notification.CorrelationKey] = notification
	return nil
}

func TestUpdateReschedulesAndSavesUniformCalendarEvent(t *testing.T) {
	repository := &repositoryStub{}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	service := NewCalendarService(repository, nil, clockStub{now: now})
	eventRange, err := domain.NewTimeRange(now.Add(time.Hour), now.Add(2*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	manualEvent, err := domain.NewManualCalendarEvent(domain.ManualEventParams{
		EventID:    "event-1",
		CalendarID: domain.DefaultCalendarID,
		Range:      eventRange,
		EventTitle: "Internal work",
		Title:      "Internal work",
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	manualEvent.PullEvents()
	repository.found = &manualEvent

	rescheduled, err := service.Update(context.Background(), UpdateCalendarFieldsCommand{
		CalendarEventID: "event-1",
		Changes: CalendarEventChanges{TimeRange: &TimeRangeUpdate{
			Start:    now.Add(3 * time.Hour),
			End:      now.Add(4 * time.Hour),
			Timezone: "Europe/Rome",
		}},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if rescheduled.Range.Start != now.Add(3*time.Hour) || len(repository.saved) != 1 {
		t.Fatalf("unexpected reschedule result: event=%#v saved=%d", rescheduled, len(repository.saved))
	}
	if repository.txCalls != 1 {
		t.Fatalf("tx calls = %d, want 1", repository.txCalls)
	}
}

func TestCancelEventLoadsAndSavesUniformCalendarEvent(t *testing.T) {
	repository := &repositoryStub{}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	service := NewCalendarService(repository, nil, clockStub{now: now})
	eventRange, err := domain.NewTimeRange(now.Add(time.Hour), now.Add(2*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	timeBlock, err := domain.NewTimeBlockCalendarEvent(domain.TimeBlockEventParams{
		EventID:    "event-1",
		CalendarID: domain.DefaultCalendarID,
		Range:      eventRange,
		Title:      "Internal work",
		Reason:     "closed",
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	timeBlock.PullEvents()
	repository.found = &timeBlock

	canceled, err := service.CancelEvent(context.Background(), CancelEventCommand{
		CalendarEventID: "event-1",
		Reason:          domain.CancelReasonDeleted,
	})
	if err != nil {
		t.Fatalf("CancelEvent() error = %v", err)
	}
	if !canceled.IsCanceled() || canceled.Cancellation.Reason != domain.CancelReasonDeleted || len(repository.saved) != 1 {
		t.Fatalf("unexpected cancel result: event=%#v saved=%d", canceled, len(repository.saved))
	}
}

type clockStub struct {
	now time.Time
}

func (c clockStub) Now() time.Time {
	return c.now
}

func TestCreateAppointmentBuildsAppointmentAggregate(t *testing.T) {
	repository := &repositoryStub{ids: []string{"event-1"}}
	customers := &customerResolverStub{customer: domain.CustomerRef{ID: "customer-1", DisplayName: "Jane Doe"}}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	service := NewCalendarService(repository, customers, clockStub{now: now})

	item, err := domain.NewServiceItem(nil, "Haircut", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	appointmentEvent, err := service.Create(context.Background(), CreateAppointmentCommand{
		CalendarID:   domain.DefaultCalendarID,
		Start:        now.Add(time.Hour),
		End:          now.Add(2 * time.Hour),
		CustomerID:   "customer-1",
		Services:     []domain.ServiceItem{item},
		RemindBefore: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if appointmentEvent.Type != domain.CalendarEventTypeAppointment {
		t.Fatalf("unexpected aggregate: %#v", appointmentEvent)
	}
	if _, ok := appointmentEvent.Detail.(domain.Appointment); !ok {
		t.Fatalf("unexpected appointment detail: %#v", appointmentEvent.Detail)
	}
	if customers.calls != 1 {
		t.Fatalf("customer resolver calls = %d, want 1", customers.calls)
	}
	if !repository.appointmentSaved {
		t.Fatal("expected appointment aggregate to be saved")
	}
	if repository.txCalls != 1 {
		t.Fatalf("tx calls = %d, want 1", repository.txCalls)
	}
	reminder, ok := repository.reminders[appointmentEvent.ID]
	if !ok || reminder.Status != domain.ReminderStatusPending || reminder.RemindBefore != 24*time.Hour {
		t.Fatalf("saved reminder = %#v, want pending 24h reminder", reminder)
	}
	if repository.writesOutsideTx != 0 {
		t.Fatalf("writes outside transaction = %d, want 0", repository.writesOutsideTx)
	}
}

func TestCreateAppointmentRejectsInvalidReminderBeforeWriting(t *testing.T) {
	repository := &repositoryStub{ids: []string{"event-1"}}
	customers := &customerResolverStub{customer: domain.CustomerRef{ID: "customer-1", DisplayName: "Jane Doe"}}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	service := NewCalendarService(repository, customers, clockStub{now: now})

	_, err := service.Create(context.Background(), CreateAppointmentCommand{
		CalendarID: domain.DefaultCalendarID,
		Start:      now.Add(time.Hour),
		End:        now.Add(2 * time.Hour),
		CustomerID: "customer-1",
	})
	if !errors.Is(err, domain.ErrInvalidReminder) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidReminder)
	}
	if repository.txCalls != 0 || len(repository.saved) != 0 || len(repository.reminders) != 0 {
		t.Fatalf("invalid reminder caused writes: tx=%d events=%d reminders=%d", repository.txCalls, len(repository.saved), len(repository.reminders))
	}
}

func TestCreateManualEventBuildsManualDetail(t *testing.T) {
	repository := &repositoryStub{ids: []string{"event-1"}}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	service := NewCalendarService(repository, nil, clockStub{now: now})

	manualEvent, err := service.Create(context.Background(), CreateManualEventCommand{
		CalendarID:  domain.DefaultCalendarID,
		Start:       now.Add(time.Hour),
		End:         now.Add(2 * time.Hour),
		Title:       "Internal work",
		ManualTitle: "Internal work",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	detail, ok := manualEvent.Detail.(domain.ManualEvent)
	if !ok || detail.Title != "Internal work" || manualEvent.Type != domain.CalendarEventTypeManual || !repository.manualSaved {
		t.Fatalf("unexpected manual event: %#v", manualEvent)
	}
}

func TestUpdateDispatchesAppointmentChangesInOneTransaction(t *testing.T) {
	repository := &repositoryStub{}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := domain.NewTimeRange(now.Add(time.Hour), now.Add(2*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	customer, err := domain.NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatal(err)
	}
	event, err := domain.NewAppointmentEvent(domain.AppointmentEventParams{
		EventID:    "event-1",
		CalendarID: domain.DefaultCalendarID,
		Range:      eventRange,
		Title:      "Old title",
		Customer:   customer,
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	event.PullEvents()
	repository.found = &event
	service := NewCalendarService(repository, nil, clockStub{now: now.Add(time.Hour)})
	newTitle := "New title"
	item, err := domain.NewServiceItem(nil, "Haircut", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := service.Update(context.Background(), UpdateAppointmentCommand{
		CalendarEventID: "event-1",
		Changes:         CalendarEventChanges{Title: &newTitle},
		Services:        []domain.ServiceItem{item},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	detail, ok := updated.Detail.(domain.Appointment)
	if !ok || updated.Title != newTitle || len(detail.Services) != 1 || detail.Services[0].ServiceName != "Haircut" {
		t.Fatalf("unexpected updated event: %#v", updated)
	}
	if repository.txCalls != 1 || len(repository.saved) != 1 {
		t.Fatalf("tx calls = %d, saved = %d; want 1 and 1", repository.txCalls, len(repository.saved))
	}
}

func TestUpdateRejectsMismatchedDetailBeforeChangingCommonFields(t *testing.T) {
	repository := &repositoryStub{}
	now := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	eventRange, err := domain.NewTimeRange(now.Add(time.Hour), now.Add(2*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	event, err := domain.NewManualCalendarEvent(domain.ManualEventParams{
		EventID:    "event-1",
		CalendarID: domain.DefaultCalendarID,
		Range:      eventRange,
		EventTitle: "Original title",
		Title:      "Internal work",
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	event.PullEvents()
	repository.found = &event
	service := NewCalendarService(repository, nil, clockStub{now: now.Add(time.Hour)})
	newTitle := "Must not be applied"

	_, err = service.Update(context.Background(), UpdateAppointmentCommand{
		CalendarEventID: "event-1",
		Changes:         CalendarEventChanges{Title: &newTitle},
	})
	if !errors.Is(err, domain.ErrInvalidEventDetail) {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrInvalidEventDetail)
	}
	if event.Title != "Original title" || len(repository.saved) != 0 {
		t.Fatalf("mismatched update changed event: %#v", event)
	}
}
