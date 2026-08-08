package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres/queries"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

func (r *Repository) NextCalendarEventID() string {
	return uuid.NewString()
}

func (r *Repository) FindCalendarEvent(ctx context.Context, agendaEventID string) (*domainv2.CalendarEvent, error) {
	view, err := r.FindCalendarEventView(ctx, agendaEventID)
	if err != nil || view == nil {
		return nil, err
	}
	return &view.Event, nil
}

func (r *Repository) FindCalendarEventView(ctx context.Context, agendaEventID string) (*applicationv2.CalendarEventView, error) {
	row, err := queries.New(r.db).FindAgendaEventFromDetails(ctx, agendaEventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	event, err := agendaEventV2FromDetails(row)
	if err != nil {
		return nil, err
	}
	switch event.Type {
	case domainv2.CalendarEventTypeAppointment:
		event.Detail, err = appointmentV2FromDetails(row)
		if err != nil {
			return nil, err
		}
	case domainv2.CalendarEventTypeManual:
		event.Detail, err = domainv2.NewManualEvent(row.ManualTitle.String, row.ManualDescription.String, nullableString(row.ManualLocation))
		if err != nil {
			return nil, err
		}
	case domainv2.CalendarEventTypeTimeBlock:
		event.Detail, err = domainv2.NewTimeBlock(row.TimeBlockReason.String)
		if err != nil {
			return nil, err
		}
	default:
		return nil, domainv2.ErrInvalidEventType
	}
	event, err = domainv2.ReconstituteCalendarEvent(event)
	if err != nil {
		return nil, err
	}
	reminder, err := appointmentReminderV2FromDetails(row)
	if err != nil {
		return nil, err
	}
	return &applicationv2.CalendarEventView{Event: event, Reminder: reminder}, nil
}

func (r *Repository) SearchCalendarEventViews(ctx context.Context, query applicationv2.ListCalendarEventsQuery) ([]applicationv2.CalendarEventView, error) {
	params := queries.SearchAgendaEventIDsFromDetailsParams{
		FilterCalendar:   query.CalendarID != "",
		CalendarID:       query.CalendarID,
		FilterCustomer:   query.CustomerID != "",
		CustomerID:       query.CustomerID,
		FilterTimeRange:  query.Start != nil && query.End != nil,
		FilterEventTypes: len(query.EventTypes) > 0,
	}
	if query.Start != nil {
		params.StartAt = timestamp(*query.Start)
	}
	if query.End != nil {
		params.EndAt = timestamp(*query.End)
	}
	if len(query.EventTypes) > 0 {
		params.EventTypes = make([]string, 0, len(query.EventTypes))
		for _, eventType := range query.EventTypes {
			params.EventTypes = append(params.EventTypes, string(eventType))
		}
	}
	ids, err := queries.New(r.db).SearchAgendaEventIDsFromDetails(ctx, params)
	if err != nil {
		return nil, err
	}
	out := make([]applicationv2.CalendarEventView, 0, len(ids))
	for _, id := range ids {
		view, err := r.FindCalendarEventView(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *view)
	}
	return out, nil
}

func appointmentReminderV2FromDetails(row queries.FindAgendaEventFromDetailsRow) (*domainv2.AppointmentReminder, error) {
	if !row.ReminderStatus.Valid {
		return nil, nil
	}
	reminder, err := domainv2.ReconstituteAppointmentReminder(domainv2.AppointmentReminder{
		Status:          domainv2.ReminderStatus(row.ReminderStatus.String),
		RemindBefore:    time.Duration(row.RemindBeforeSeconds.Int32) * time.Second,
		ScheduledAt:     nullableTime(row.ReminderScheduledAt),
		SentRequestedAt: nullableTime(row.ReminderSentRequestedAt),
		SentAt:          nullableTime(row.ReminderSentAt),
		FailedAt:        nullableTime(row.ReminderFailedAt),
		FailureReason:   nullableString(row.ReminderFailureReason),
		UpdatedAt:       row.ReminderUpdatedAt.Time,
	})
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (r *Repository) SaveCalendarEvent(ctx context.Context, event *domainv2.CalendarEvent) error {
	if event == nil {
		return domainv2.ErrMissingRequiredData
	}
	var err error
	switch detail := event.Detail.(type) {
	case domainv2.Appointment:
		err = r.saveAppointment(ctx, *event, detail)
	case domainv2.ManualEvent:
		err = r.saveManualEvent(ctx, *event, detail)
	case domainv2.TimeBlock:
		err = r.saveTimeBlock(ctx, *event, detail)
	default:
		return fmt.Errorf("unsupported calendar event type %T", event)
	}
	if err != nil {
		return err
	}
	return r.publishCalendarLifecycleEvents(ctx, event.PullEvents())
}

func (r *Repository) saveAppointment(ctx context.Context, event domainv2.CalendarEvent, appointment domainv2.Appointment) error {
	if err := queries.New(r.db).SaveAgendaEventV2(ctx, agendaEventV2Params(event, appointment.Customer.ID, appointment.Customer.DisplayName)); err != nil {
		return err
	}
	if err := queries.New(r.db).SaveAppointment(ctx, queries.SaveAppointmentParams{
		AgendaEventID:       event.ID,
		CustomerID:          appointment.Customer.ID,
		CustomerDisplayName: appointment.Customer.DisplayName,
		CreatedAt:           timestamp(appointment.CreatedAt),
		UpdatedAt:           timestamp(appointment.UpdatedAt),
	}); err != nil {
		return err
	}
	if err := queries.New(r.db).DeleteAppointmentServiceItems(ctx, event.ID); err != nil {
		return err
	}
	for _, service := range appointment.Services {
		if err := queries.New(r.db).SaveAppointmentServiceItem(ctx, queries.SaveAppointmentServiceItemParams{
			AgendaEventID: event.ID,
			ServiceID:     nullableText(service.ServiceID),
			ServiceName:   service.ServiceName,
			Price:         nullableFloat8(service.Price),
			Position:      int32(service.Position),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) SaveAppointmentReminderState(ctx context.Context, agendaEventID string, reminder domainv2.AppointmentReminder) error {
	return queries.New(r.db).SaveAppointmentReminder(ctx, appointmentReminderV2Params(agendaEventID, reminder))
}

func (r *Repository) FindAppointmentNotification(ctx context.Context, correlationKey string) (*domainv2.AppointmentNotification, error) {
	row, err := queries.New(r.db).FindAppointmentNotification(ctx, correlationKey)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	recipient, err := domainv2.ReconstituteNotificationRecipient(row.RecipientType, row.RecipientID)
	if err != nil {
		return nil, err
	}
	notification, err := domainv2.ReconstituteAppointmentNotification(domainv2.AppointmentNotification{
		CorrelationKey:  row.CorrelationKey,
		CalendarEventID: row.AgendaEventID,
		Kind:            domainv2.NotificationKind(row.NotificationKind),
		Type:            domainv2.NotificationType(row.NotificationType),
		Status:          domainv2.NotificationStatus(row.Status),
		Recipient:       recipient,
		IdempotencyKey:  nullableString(row.NotificationIdempotencyKey),
		FailureReason:   nullableString(row.FailureReason),
		FailureMessage:  nullableString(row.FailureMessage),
		CreatedAt:       row.CreatedAt.Time,
		CompletedAt:     nullableTime(row.CompletedAt),
		ExpiresAt:       row.ExpiresAt.Time,
	})
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *Repository) SaveAppointmentNotification(ctx context.Context, notification domainv2.AppointmentNotification) error {
	return queries.New(r.db).SaveAppointmentNotification(ctx, appointmentNotificationV2Params(notification))
}

func (r *Repository) saveManualEvent(ctx context.Context, event domainv2.CalendarEvent, manualEvent domainv2.ManualEvent) error {
	if err := queries.New(r.db).SaveAgendaEventV2(ctx, agendaEventV2Params(event, "self", "self")); err != nil {
		return err
	}
	return queries.New(r.db).SaveAgendaManualEvent(ctx, queries.SaveAgendaManualEventParams{
		AgendaEventID: event.ID,
		Title:         manualEvent.Title,
		Description:   nullableText(&manualEvent.Description),
		Location:      nullableText(manualEvent.Location),
		CreatedAt:     timestamp(event.CreatedAt),
		UpdatedAt:     timestamp(event.UpdatedAt),
	})
}

func (r *Repository) saveTimeBlock(ctx context.Context, event domainv2.CalendarEvent, timeBlock domainv2.TimeBlock) error {
	if err := queries.New(r.db).SaveAgendaEventV2(ctx, agendaEventV2Params(event, "self", "self")); err != nil {
		return err
	}
	return queries.New(r.db).SaveAgendaTimeBlock(ctx, queries.SaveAgendaTimeBlockParams{
		AgendaEventID: event.ID,
		Reason:        timeBlock.Reason,
		CreatedAt:     timestamp(event.CreatedAt),
		UpdatedAt:     timestamp(event.UpdatedAt),
	})
}

func agendaEventV2Params(event domainv2.CalendarEvent, attendeeID string, attendeeDisplayName string) queries.SaveAgendaEventV2Params {
	title := event.Title
	description := event.Description
	cancelReason := (*string)(nil)
	canceledAt := (*time.Time)(nil)
	if event.Cancellation != nil {
		value := string(event.Cancellation.Reason)
		cancelReason = &value
		valueAt := event.Cancellation.CanceledAt
		canceledAt = &valueAt
	}
	return queries.SaveAgendaEventV2Params{
		ID:                  event.ID,
		CalendarID:          event.CalendarID,
		EventType:           string(event.Type),
		Title:               title,
		Description:         description,
		StartAt:             timestamp(event.Range.Start),
		EndAt:               timestamp(event.Range.End),
		Timezone:            event.Range.Timezone,
		AllDay:              event.Range.AllDay,
		DisplayTitle:        nullableText(&event.Title),
		DisplayDescription:  nullableText(&event.Description),
		Visibility:          string(event.Visibility),
		AttendeeID:          attendeeID,
		AttendeeDisplayName: attendeeDisplayName,
		CancelReason:        nullableText(cancelReason),
		CanceledAt:          nullableTimestamp(canceledAt),
		CreatedAt:           timestamp(event.CreatedAt),
		UpdatedAt:           timestamp(event.UpdatedAt),
	}
}

func appointmentReminderV2Params(agendaEventID string, reminder domainv2.AppointmentReminder) queries.SaveAppointmentReminderParams {
	return queries.SaveAppointmentReminderParams{
		AgendaEventID:       agendaEventID,
		Status:              string(reminder.Status),
		RemindBeforeSeconds: int32(reminder.RemindBefore.Seconds()),
		ScheduledAt:         nullableTimestamp(reminder.ScheduledAt),
		SentRequestedAt:     nullableTimestamp(reminder.SentRequestedAt),
		SentAt:              nullableTimestamp(reminder.SentAt),
		FailedAt:            nullableTimestamp(reminder.FailedAt),
		FailureReason:       nullableText(reminder.FailureReason),
		UpdatedAt:           timestamp(reminder.UpdatedAt),
	}
}

func appointmentNotificationV2Params(notification domainv2.AppointmentNotification) queries.SaveAppointmentNotificationParams {
	return queries.SaveAppointmentNotificationParams{
		CorrelationKey:             notification.CorrelationKey,
		AgendaEventID:              notification.CalendarEventID,
		NotificationKind:           string(notification.Kind),
		NotificationType:           string(notification.Type),
		Status:                     string(notification.Status),
		RecipientType:              notification.Recipient.Kind(),
		RecipientID:                notification.Recipient.ID(),
		NotificationIdempotencyKey: nullableText(notification.IdempotencyKey),
		FailureReason:              nullableText(notification.FailureReason),
		FailureMessage:             nullableText(notification.FailureMessage),
		CreatedAt:                  timestamp(notification.CreatedAt),
		CompletedAt:                nullableTimestamp(notification.CompletedAt),
		ExpiresAt:                  timestamp(notification.ExpiresAt),
	}
}

func (r *Repository) publishCalendarLifecycleEvents(ctx context.Context, events []domainv2.LifecycleEvent) error {
	if len(events) == 0 {
		return nil
	}
	messages := make([]outbox.Message, 0, len(events))
	for _, event := range events {
		message, err := newCalendarLifecycleOutboxMessage(event)
		if err != nil {
			return err
		}
		messages = append(messages, message)
	}
	if err := r.publisher.Publish(ctx, messages...); err != nil {
		return fmt.Errorf("publish calendar lifecycle events: %w", err)
	}
	return nil
}

func newCalendarLifecycleOutboxMessage(event domainv2.LifecycleEvent) (outbox.Message, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return outbox.Message{}, fmt.Errorf("marshal calendar lifecycle event: %w", err)
	}
	return outbox.Message{
		ID:          uuid.NewString(),
		Channel:     outbox.Channel(ChannelAppointmentInternalJob),
		AffinityKey: outbox.AffinityKey(event.CalendarEventID),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  time.Now().UTC(),
	}, nil
}

func agendaEventV2FromDetails(row queries.FindAgendaEventFromDetailsRow) (domainv2.CalendarEvent, error) {
	eventRange, err := domainv2.NewTimeRange(row.StartAt.Time, row.EndAt.Time, row.Timezone, row.AllDay)
	if err != nil {
		return domainv2.CalendarEvent{}, err
	}
	event := domainv2.CalendarEvent{
		ID:          row.ID,
		CalendarID:  row.CalendarID,
		Type:        domainv2.CalendarEventType(row.EventType),
		Range:       eventRange,
		Title:       row.DisplayTitle.String,
		Description: row.DisplayDescription.String,
		Visibility:  domainv2.Visibility(row.Visibility),
		Version:     row.Version,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		UpdatedAt:   row.UpdatedAt.Time.UTC(),
	}
	if row.CancelReason.Valid {
		canceledAt := row.UpdatedAt.Time.UTC()
		if row.CanceledAt.Valid {
			canceledAt = row.CanceledAt.Time.UTC()
		}
		event.Cancellation = &domainv2.CalendarEventCancellation{
			Reason:     domainv2.CancelReason(row.CancelReason.String),
			CanceledAt: canceledAt,
		}
	}
	return event, nil
}

func appointmentV2FromDetails(row queries.FindAgendaEventFromDetailsRow) (domainv2.Appointment, error) {
	customer, err := domainv2.NewCustomerRef(row.CustomerID, row.CustomerDisplayName.String)
	if err != nil {
		return domainv2.Appointment{}, err
	}
	services, err := serviceItemsV2FromJSON(row.ServicesJson)
	if err != nil {
		return domainv2.Appointment{}, err
	}
	return domainv2.ReconstituteAppointment(customer, services, row.CreatedAt.Time, row.UpdatedAt.Time)
}

func serviceItemsV2FromJSON(data string) ([]domainv2.ServiceItem, error) {
	var rows []struct {
		Name      string   `json:"Name"`
		ServiceID *string  `json:"serviceId"`
		Price     *float64 `json:"price"`
		Position  int      `json:"position"`
	}
	if err := json.Unmarshal([]byte(data), &rows); err != nil {
		return nil, err
	}
	items := make([]domainv2.ServiceItem, 0, len(rows))
	for i, row := range rows {
		position := row.Position
		if position < 0 {
			position = i
		}
		item, err := domainv2.NewServiceItem(row.ServiceID, row.Name, row.Price, position)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func nullableString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func nullableFloat8(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{}
	}
	return pgtype.Float8{Float64: *value, Valid: true}
}
