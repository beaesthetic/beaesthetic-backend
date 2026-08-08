package v2

import (
	"context"
	"time"

	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type Clock interface {
	Now() time.Time
}

type CustomerResolver interface {
	ResolveCustomer(ctx context.Context, customerID string) (domain.CustomerRef, error)
}

type CalendarEventRepository interface {
	NextCalendarEventID() string
	Tx(ctx context.Context, atomicFn func(context.Context) error) error
	FindCalendarEvent(ctx context.Context, agendaEventID string) (*domain.CalendarEvent, error)
	SaveCalendarEvent(ctx context.Context, event *domain.CalendarEvent) error
}

type AppointmentReminderRepository interface {
	SaveAppointmentReminderState(ctx context.Context, calendarEventID string, reminder domain.AppointmentReminder) error
}

type AppointmentNotificationRepository interface {
	FindAppointmentNotification(ctx context.Context, correlationKey string) (*domain.AppointmentNotification, error)
	SaveAppointmentNotification(ctx context.Context, notification domain.AppointmentNotification) error
}

type CalendarEventReadRepository interface {
	FindCalendarEventView(ctx context.Context, calendarEventID string) (*CalendarEventView, error)
	SearchCalendarEventViews(ctx context.Context, query ListCalendarEventsQuery) ([]CalendarEventView, error)
}

type Repository interface {
	CalendarEventRepository
	AppointmentReminderRepository
	AppointmentNotificationRepository
	CalendarEventReadRepository
}

type CalendarService struct {
	repository   Repository
	appointments *AppointmentEventService
	manualEvents *ManualEventService
	timeBlocks   *TimeBlockService
	clock        Clock
}

func NewCalendarService(repository Repository, customers CustomerResolver, clock Clock) *CalendarService {
	return &CalendarService{
		repository:   repository,
		appointments: NewAppointmentEventService(repository, customers, clock),
		manualEvents: NewManualEventService(repository, clock),
		timeBlocks:   NewTimeBlockService(repository, clock),
		clock:        clock,
	}
}

func (s *CalendarService) Create(ctx context.Context, command CreateEventCommand) (*domain.CalendarEvent, error) {
	var calendarEvent domain.CalendarEvent
	var reminder *domain.AppointmentReminder
	var err error
	switch command := command.(type) {
	case CreateAppointmentCommand:
		calendarEvent, err = s.appointments.Create(ctx, command)
		if err == nil {
			createdReminder, reminderErr := domain.NewAppointmentReminder(command.RemindBefore, calendarEvent.CreatedAt)
			if reminderErr != nil {
				return nil, reminderErr
			}
			reminder = &createdReminder
		}
	case CreateManualEventCommand:
		calendarEvent, err = s.manualEvents.Create(ctx, command)
	case CreateTimeBlockCommand:
		calendarEvent, err = s.timeBlocks.Create(ctx, command)
	default:
		return nil, ErrUnsupportedEventType
	}
	if err != nil {
		return nil, err
	}
	if err := s.repository.Tx(ctx, func(ctx context.Context) error {
		if err := s.repository.SaveCalendarEvent(ctx, &calendarEvent); err != nil {
			return err
		}
		if reminder != nil {
			return s.repository.SaveAppointmentReminderState(ctx, calendarEvent.ID, *reminder)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &calendarEvent, nil
}

type AppointmentEventService struct {
	calendarEvents CalendarEventRepository
	customers      CustomerResolver
	clock          Clock
}

func NewAppointmentEventService(repository CalendarEventRepository, customers CustomerResolver, clock Clock) *AppointmentEventService {
	return &AppointmentEventService{calendarEvents: repository, customers: customers, clock: clock}
}

func (s *AppointmentEventService) Create(ctx context.Context, command CreateAppointmentCommand) (domain.CalendarEvent, error) {
	now := s.clock.Now()
	eventRange, err := domain.NewTimeRange(command.Start, command.End, command.Timezone, command.AllDay)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	if s.customers == nil {
		return domain.CalendarEvent{}, domain.ErrMissingRequiredData
	}
	customer, err := s.customers.ResolveCustomer(ctx, command.CustomerID)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	calendarEvent, err := domain.NewAppointmentEvent(domain.AppointmentEventParams{
		EventID:     s.calendarEvents.NextCalendarEventID(),
		CalendarID:  command.CalendarID,
		Range:       eventRange,
		Title:       command.Title,
		Description: command.Description,
		Visibility:  command.Visibility,
		Customer:    customer,
		Services:    command.Services,
		Now:         now,
	})
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	return calendarEvent, nil
}

func (s *AppointmentEventService) Update(ctx context.Context, command UpdateAppointmentCommand) (*domain.CalendarEvent, error) {
	now := s.clock.Now()
	return changeCalendarEvent(ctx, s.calendarEvents, command.CalendarEventID, func(event *domain.CalendarEvent) error {
		if _, ok := event.Detail.(domain.Appointment); !ok {
			return domain.ErrInvalidEventDetail
		}
		if err := applyCalendarEventChanges(event, command.Changes, now); err != nil {
			return err
		}
		return event.ReplaceAppointmentServices(command.Services, now)
	})
}

type ManualEventService struct {
	calendarEvents CalendarEventRepository
	clock          Clock
}

func NewManualEventService(repository CalendarEventRepository, clock Clock) *ManualEventService {
	return &ManualEventService{calendarEvents: repository, clock: clock}
}

func (s *ManualEventService) Create(ctx context.Context, command CreateManualEventCommand) (domain.CalendarEvent, error) {
	now := s.clock.Now()
	eventRange, err := domain.NewTimeRange(command.Start, command.End, command.Timezone, command.AllDay)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	calendarEvent, err := domain.NewManualCalendarEvent(domain.ManualEventParams{
		EventID:          s.calendarEvents.NextCalendarEventID(),
		CalendarID:       command.CalendarID,
		Range:            eventRange,
		EventTitle:       command.Title,
		EventDescription: command.Description,
		Visibility:       command.Visibility,
		Title:            command.ManualTitle,
		Description:      command.ManualDetails,
		Location:         command.Location,
		Now:              now,
	})
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	return calendarEvent, nil
}

func (s *ManualEventService) Update(ctx context.Context, command UpdateManualEventCommand) (*domain.CalendarEvent, error) {
	now := s.clock.Now()
	return changeCalendarEvent(ctx, s.calendarEvents, command.CalendarEventID, func(event *domain.CalendarEvent) error {
		manualEvent, ok := event.Detail.(domain.ManualEvent)
		if !ok {
			return domain.ErrInvalidEventDetail
		}
		if err := applyCalendarEventChanges(event, command.Changes, now); err != nil {
			return err
		}
		title := manualEvent.Title
		description := manualEvent.Description
		location := manualEvent.Location
		if command.Title != nil {
			title = *command.Title
		}
		if command.Description != nil {
			description = *command.Description
		}
		if command.Location != nil {
			location = *command.Location
		}
		return event.ChangeManualDetails(title, description, location, now)
	})
}

type TimeBlockService struct {
	calendarEvents CalendarEventRepository
	clock          Clock
}

func NewTimeBlockService(repository CalendarEventRepository, clock Clock) *TimeBlockService {
	return &TimeBlockService{calendarEvents: repository, clock: clock}
}

func (s *TimeBlockService) Create(ctx context.Context, command CreateTimeBlockCommand) (domain.CalendarEvent, error) {
	now := s.clock.Now()
	eventRange, err := domain.NewTimeRange(command.Start, command.End, command.Timezone, command.AllDay)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	calendarEvent, err := domain.NewTimeBlockCalendarEvent(domain.TimeBlockEventParams{
		EventID:     s.calendarEvents.NextCalendarEventID(),
		CalendarID:  command.CalendarID,
		Range:       eventRange,
		Title:       command.Title,
		Description: command.Description,
		Visibility:  command.Visibility,
		Reason:      command.Reason,
		Now:         now,
	})
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	return calendarEvent, nil
}

func (s *TimeBlockService) Update(ctx context.Context, command UpdateTimeBlockCommand) (*domain.CalendarEvent, error) {
	now := s.clock.Now()
	return changeCalendarEvent(ctx, s.calendarEvents, command.CalendarEventID, func(event *domain.CalendarEvent) error {
		if _, ok := event.Detail.(domain.TimeBlock); !ok {
			return domain.ErrInvalidEventDetail
		}
		if err := applyCalendarEventChanges(event, command.Changes, now); err != nil {
			return err
		}
		return event.ChangeTimeBlockReason(command.Reason, now)
	})
}

func (s *CalendarService) GetCalendarEventView(ctx context.Context, calendarEventID string) (*CalendarEventView, error) {
	return s.repository.FindCalendarEventView(ctx, calendarEventID)
}

func (s *CalendarService) ListCalendarEventViews(ctx context.Context, query ListCalendarEventsQuery) ([]CalendarEventView, error) {
	return s.repository.SearchCalendarEventViews(ctx, query)
}

func (s *CalendarService) Update(ctx context.Context, command UpdateEventCommand) (*domain.CalendarEvent, error) {
	switch command := command.(type) {
	case UpdateCalendarFieldsCommand:
		now := s.clock.Now()
		return changeCalendarEvent(ctx, s.repository, command.CalendarEventID, func(event *domain.CalendarEvent) error {
			return applyCalendarEventChanges(event, command.Changes, now)
		})
	case UpdateAppointmentCommand:
		return s.appointments.Update(ctx, command)
	case UpdateManualEventCommand:
		return s.manualEvents.Update(ctx, command)
	case UpdateTimeBlockCommand:
		return s.timeBlocks.Update(ctx, command)
	default:
		return nil, ErrUnsupportedEventType
	}
}

func (s *CalendarService) CancelEvent(ctx context.Context, command CancelEventCommand) (*domain.CalendarEvent, error) {
	now := s.clock.Now()
	var calendarEvent *domain.CalendarEvent
	if err := s.repository.Tx(ctx, func(ctx context.Context) error {
		found, err := s.repository.FindCalendarEvent(ctx, command.CalendarEventID)
		if err != nil {
			return err
		}
		found.Cancel(command.Reason, now)
		if err := s.repository.SaveCalendarEvent(ctx, found); err != nil {
			return err
		}
		calendarEvent = found
		return nil
	}); err != nil {
		return nil, err
	}
	return calendarEvent, nil
}

func changeCalendarEvent(ctx context.Context, repository CalendarEventRepository, calendarEventID string, change func(*domain.CalendarEvent) error) (*domain.CalendarEvent, error) {
	var calendarEvent *domain.CalendarEvent
	if err := repository.Tx(ctx, func(ctx context.Context) error {
		found, err := repository.FindCalendarEvent(ctx, calendarEventID)
		if err != nil {
			return err
		}
		if err := change(found); err != nil {
			return err
		}
		if err := repository.SaveCalendarEvent(ctx, found); err != nil {
			return err
		}
		calendarEvent = found
		return nil
	}); err != nil {
		return nil, err
	}
	return calendarEvent, nil
}

func applyCalendarEventChanges(event *domain.CalendarEvent, changes CalendarEventChanges, now time.Time) error {
	if changes.TimeRange != nil {
		eventRange, err := domain.NewTimeRange(changes.TimeRange.Start, changes.TimeRange.End, changes.TimeRange.Timezone, changes.TimeRange.AllDay)
		if err != nil {
			return err
		}
		event.Reschedule(eventRange, now)
	}
	if changes.Title != nil {
		event.ChangeTitle(*changes.Title, now)
	}
	if changes.Description != nil {
		event.ChangeDescription(*changes.Description, now)
	}
	if changes.Visibility != nil {
		return event.ChangeVisibility(*changes.Visibility, now)
	}
	return nil
}
