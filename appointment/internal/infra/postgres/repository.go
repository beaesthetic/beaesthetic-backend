package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres/queries"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const ChannelAppointmentInternalJob = "beaesthetic.appointments.internal.job"

type Repository struct {
	db        *ContextDB
	publisher outbox.Publisher
}

func NewRepository(db *ContextDB, publisher outbox.Publisher) *Repository {
	return &Repository{db: db, publisher: publisher}
}

func (r *Repository) Tx(ctx context.Context, atomicFn func(ctx context.Context) error) error {
	return r.db.Tx(ctx, atomicFn)
}

func (r *Repository) SaveAgendaEvent(ctx context.Context, e *domain.AgendaEvent) error {
	events := e.PullEvents()
	services, _ := json.Marshal(e.Services)
	var cancelReason *string
	var canceledAt *time.Time
	if e.CancelReason != nil {
		v := string(*e.CancelReason)
		cancelReason = &v
		vAt := e.UpdatedAt
		canceledAt = &vAt
	}
	err := queries.New(r.db).SaveAgendaEvent(ctx, queries.SaveAgendaEventParams{
		ID:                  e.ID,
		EventType:           persistedLegacyEventType(e.Type),
		Title:               e.Title,
		Description:         e.Description,
		StartAt:             timestamp(e.Start),
		EndAt:               timestamp(e.End),
		AttendeeID:          e.Attendee.ID,
		AttendeeDisplayName: e.Attendee.DisplayName,
		Services:            services,
		CancelReason:        nullableText(cancelReason),
		CanceledAt:          nullableTimestamp(canceledAt),
		ReminderStatus:      string(e.ReminderStatus),
		ReminderSentAt:      nullableTimestamp(e.ReminderSentAt),
		RemindBeforeSeconds: int32(e.RemindBefore.Seconds()),
		CreatedAt:           timestamp(e.CreatedAt),
		UpdatedAt:           timestamp(e.UpdatedAt),
	})
	if err != nil {
		return err
	}
	if err := r.saveAgendaEventDetails(ctx, e); err != nil {
		return err
	}
	return r.publishLifecycleEvents(ctx, events)
}

func persistedLegacyEventType(eventType domain.EventType) string {
	if eventType == domain.EventTypeGeneric {
		return string(domainv2.CalendarEventTypeManual)
	}
	return string(eventType)
}

func (r *Repository) SaveAppointmentReminder(ctx context.Context, e *domain.AgendaEvent) error {
	return r.saveAppointmentReminder(ctx, e)
}

func (r *Repository) FindAgendaEvent(ctx context.Context, id string) (*domain.AgendaEvent, error) {
	row, err := queries.New(r.db).FindAgendaEventFromDetails(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return agendaEventFromDetailsRow(row)
}

func (r *Repository) SearchAgendaEvents(ctx context.Context, attendeeID string, start, end *time.Time) ([]domain.AgendaEvent, error) {
	params := queries.SearchAgendaEventIDsFromDetailsParams{}
	if attendeeID != "" {
		params.FilterCustomer = true
		params.CustomerID = attendeeID
	}
	if start != nil && end != nil {
		params.FilterTimeRange = true
		params.StartAt = timestamp(*start)
		params.EndAt = timestamp(*end)
	}
	ids, err := queries.New(r.db).SearchAgendaEventIDsFromDetails(ctx, params)
	if err != nil {
		return nil, err
	}
	return r.findAgendaEventsByIDs(ctx, ids)
}

func (r *Repository) FindFutureAppointments(ctx context.Context, from time.Time) ([]domain.AgendaEvent, error) {
	ids, err := queries.New(r.db).FindFutureAppointmentAgendaEventIDsFromDetails(ctx, timestamp(from))
	if err != nil {
		return nil, err
	}
	return r.findAgendaEventsByIDs(ctx, ids)
}

func (r *Repository) findAgendaEventsByIDs(ctx context.Context, ids []string) ([]domain.AgendaEvent, error) {
	out := make([]domain.AgendaEvent, 0, len(ids))
	for _, id := range ids {
		e, err := r.FindAgendaEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		if e != nil {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (r *Repository) SaveService(ctx context.Context, s domain.AppointmentService) (domain.AppointmentService, error) {
	tags, _ := json.Marshal(s.Tags)
	err := queries.New(r.db).SaveAppointmentService(ctx, queries.SaveAppointmentServiceParams{
		ID:       s.ID,
		Name:     s.Name,
		Price:    s.Price,
		Tags:     tags,
		ColorHex: nullableText(s.Color),
	})
	return s, err
}

func (r *Repository) FindServices(ctx context.Context) ([]domain.AppointmentService, error) {
	rows, err := queries.New(r.db).FindAppointmentServices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AppointmentService, 0, len(rows))
	for _, row := range rows {
		service, err := appointmentServiceFromFields(row.ID, row.Name, row.Price, row.Tags, row.ColorHex)
		if err != nil {
			return nil, err
		}
		out = append(out, service)
	}
	return out, nil
}

func (r *Repository) SearchServices(ctx context.Context, text string, limit int) ([]domain.AppointmentService, error) {
	query := strings.TrimSpace(strings.ToLower(text))
	if query == "" {
		return r.FindServices(ctx)
	}
	rows, err := queries.New(r.db).SearchAppointmentServices(ctx, queries.SearchAppointmentServicesParams{
		Query:      query,
		LimitCount: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.AppointmentService, 0, len(rows))
	for _, row := range rows {
		service, err := appointmentServiceFromFields(row.ID, row.Name, row.Price, row.Tags, row.ColorHex)
		if err != nil {
			return nil, err
		}
		out = append(out, service)
	}
	return out, nil
}

func (r *Repository) FindService(ctx context.Context, id string) (*domain.AppointmentService, error) {
	row, err := queries.New(r.db).FindAppointmentService(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	service, err := appointmentServiceFromFields(row.ID, row.Name, row.Price, row.Tags, row.ColorHex)
	return &service, err
}

func (r *Repository) FindAppointmentNotificationTracking(ctx context.Context, correlationKey string) (*application.AppointmentNotificationTracking, error) {
	row, err := queries.New(r.db).FindAppointmentNotification(ctx, correlationKey)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	notification := application.AppointmentNotificationTracking{
		CorrelationKey: row.CorrelationKey,
		AgendaEventID:  row.AgendaEventID,
		Kind:           row.NotificationKind,
		Type:           row.NotificationType,
	}
	return &notification, nil
}

func (r *Repository) SaveAppointmentNotificationTracking(ctx context.Context, notification application.AppointmentNotificationTracking) error {
	return r.saveAppointmentNotification(ctx, notification)
}

func (r *Repository) MarkAppointmentNotificationSent(ctx context.Context, correlationKey string, completedAt time.Time) error {
	return queries.New(r.db).MarkAppointmentNotificationSent(ctx, queries.MarkAppointmentNotificationSentParams{
		CorrelationKey: correlationKey,
		CompletedAt:    timestamp(completedAt),
	})
}

func (r *Repository) MarkAppointmentNotificationFailed(ctx context.Context, correlationKey string, reason string, message string, completedAt time.Time) error {
	return queries.New(r.db).MarkAppointmentNotificationFailed(ctx, queries.MarkAppointmentNotificationFailedParams{
		CorrelationKey: correlationKey,
		FailureReason:  nullableText(&reason),
		FailureMessage: nullableText(&message),
		CompletedAt:    timestamp(completedAt),
	})
}

func (r *Repository) saveAgendaEventDetails(ctx context.Context, e *domain.AgendaEvent) error {
	switch e.Type {
	case domain.EventTypeAppointment:
		return r.saveAppointmentDetails(ctx, e)
	case domain.EventTypeGeneric:
		return queries.New(r.db).SaveAgendaManualEvent(ctx, queries.SaveAgendaManualEventParams{
			AgendaEventID: e.ID,
			Title:         e.Title,
			Description:   nullableText(&e.Description),
			Location:      pgtype.Text{},
			CreatedAt:     timestamp(e.CreatedAt),
			UpdatedAt:     timestamp(e.UpdatedAt),
		})
	default:
		return nil
	}
}

func (r *Repository) saveAppointmentDetails(ctx context.Context, e *domain.AgendaEvent) error {
	if err := queries.New(r.db).SaveAppointment(ctx, queries.SaveAppointmentParams{
		AgendaEventID:       e.ID,
		CustomerID:          e.Attendee.ID,
		CustomerDisplayName: e.Attendee.DisplayName,
		CreatedAt:           timestamp(e.CreatedAt),
		UpdatedAt:           timestamp(e.UpdatedAt),
	}); err != nil {
		return err
	}
	if err := queries.New(r.db).DeleteAppointmentServiceItems(ctx, e.ID); err != nil {
		return err
	}
	for i, service := range e.Services {
		if err := queries.New(r.db).SaveAppointmentServiceItem(ctx, queries.SaveAppointmentServiceItemParams{
			AgendaEventID: e.ID,
			ServiceID:     pgtype.Text{},
			ServiceName:   service.Name,
			Price:         pgtype.Float8{},
			Position:      int32(i),
		}); err != nil {
			return err
		}
	}
	return r.saveAppointmentReminder(ctx, e)
}

func (r *Repository) saveAppointmentReminder(ctx context.Context, e *domain.AgendaEvent) error {
	return queries.New(r.db).SaveAppointmentReminder(ctx, queries.SaveAppointmentReminderParams{
		AgendaEventID:       e.ID,
		Status:              appointmentReminderStatus(e.ReminderStatus),
		RemindBeforeSeconds: int32(e.RemindBefore.Seconds()),
		ScheduledAt:         reminderTimestamp(e, domain.ReminderScheduled),
		SentRequestedAt:     reminderTimestamp(e, domain.ReminderSentRequested),
		SentAt:              nullableTimestamp(e.ReminderSentAt),
		FailedAt:            reminderTimestamp(e, domain.ReminderFailToSend),
		FailureReason:       reminderFailureReason(e.ReminderStatus),
		UpdatedAt:           timestamp(e.UpdatedAt),
	})
}

func (r *Repository) saveAppointmentNotification(ctx context.Context, notification application.AppointmentNotificationTracking) error {
	event, err := r.FindAgendaEvent(ctx, notification.AgendaEventID)
	if err != nil || event == nil || event.Type != domain.EventTypeAppointment {
		return err
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return queries.New(r.db).SaveAppointmentNotification(ctx, queries.SaveAppointmentNotificationParams{
		CorrelationKey:             notification.CorrelationKey,
		AgendaEventID:              event.ID,
		NotificationKind:           notificationKind(notification.Type),
		NotificationType:           notificationType(notification.Type),
		Status:                     "pending",
		RecipientType:              "customer",
		RecipientID:                event.Attendee.ID,
		NotificationIdempotencyKey: nullableText(&notification.CorrelationKey),
		FailureReason:              pgtype.Text{},
		FailureMessage:             pgtype.Text{},
		CreatedAt:                  timestamp(time.Now().UTC()),
		CompletedAt:                pgtype.Timestamptz{},
		ExpiresAt:                  timestamp(expiresAt),
	})
}

func agendaEventFromRow(row queries.AgendaEvent) (*domain.AgendaEvent, error) {
	var services []domain.AppointmentServiceRef
	if err := json.Unmarshal(row.Services, &services); err != nil {
		return nil, err
	}

	var cancelReason *domain.CancelReason
	if row.CancelReason.Valid {
		v := domain.CancelReason(row.CancelReason.String)
		cancelReason = &v
	}

	var reminderSentAt *time.Time
	if row.ReminderSentAt.Valid {
		v := row.ReminderSentAt.Time.UTC()
		reminderSentAt = &v
	}

	return &domain.AgendaEvent{
		ID:          row.ID,
		Type:        domain.EventType(row.EventType),
		Title:       row.Title,
		Description: row.Description,
		Start:       row.StartAt.Time.UTC(),
		End:         row.EndAt.Time.UTC(),
		Attendee: domain.Attendee{
			ID:          row.AttendeeID,
			DisplayName: row.AttendeeDisplayName,
		},
		Services:       services,
		CancelReason:   cancelReason,
		ReminderStatus: domain.ReminderStatus(row.ReminderStatus),
		ReminderSentAt: reminderSentAt,
		RemindBefore:   time.Duration(row.RemindBeforeSeconds) * time.Second,
		Version:        row.Version,
		CreatedAt:      row.CreatedAt.Time.UTC(),
		UpdatedAt:      row.UpdatedAt.Time.UTC(),
	}, nil
}

func agendaEventFromDetailsRow(row queries.FindAgendaEventFromDetailsRow) (*domain.AgendaEvent, error) {
	var services []domain.AppointmentServiceRef
	if row.ServicesJson != "" {
		if err := json.Unmarshal([]byte(row.ServicesJson), &services); err != nil {
			return nil, err
		}
	}

	var cancelReason *domain.CancelReason
	if row.CancelReason.Valid {
		v := domain.CancelReason(row.CancelReason.String)
		cancelReason = &v
	}

	var reminderSentAt *time.Time
	if row.ReminderSentAt.Valid {
		v := row.ReminderSentAt.Time.UTC()
		reminderSentAt = &v
	}

	eventType := domain.EventTypeGeneric
	title := row.LegacyTitle
	description := row.LegacyDescription
	attendee := domain.Attendee{ID: "self", DisplayName: "self"}
	reminderStatus := domain.ReminderPending
	remindBefore := time.Duration(0)

	if row.AgendaEventID != "" {
		eventType = domain.EventTypeAppointment
		title = row.LegacyTitle
		description = row.LegacyDescription
		attendee = domain.Attendee{
			ID:          row.CustomerID,
			DisplayName: row.CustomerDisplayName.String,
		}
		reminderStatus = legacyReminderStatus(row.ReminderStatus)
		if row.RemindBeforeSeconds.Valid {
			remindBefore = time.Duration(row.RemindBeforeSeconds.Int32) * time.Second
		}
	} else if row.ManualTitle.Valid {
		title = row.ManualTitle.String
		if row.ManualDescription.Valid {
			description = row.ManualDescription.String
		}
	}

	return &domain.AgendaEvent{
		ID:             row.ID,
		Type:           eventType,
		Title:          title,
		Description:    description,
		Start:          row.StartAt.Time.UTC(),
		End:            row.EndAt.Time.UTC(),
		Attendee:       attendee,
		Services:       services,
		CancelReason:   cancelReason,
		ReminderStatus: reminderStatus,
		ReminderSentAt: reminderSentAt,
		RemindBefore:   remindBefore,
		Version:        row.Version,
		CreatedAt:      row.CreatedAt.Time.UTC(),
		UpdatedAt:      row.UpdatedAt.Time.UTC(),
	}, nil
}

func appointmentServiceFromFields(id string, name string, price float64, tagsBytes []byte, color pgtype.Text) (domain.AppointmentService, error) {
	var tags []string
	if err := json.Unmarshal(tagsBytes, &tags); err != nil {
		return domain.AppointmentService{}, err
	}
	return domain.AppointmentService{
		ID:    id,
		Name:  name,
		Price: price,
		Tags:  tags,
		Color: textPointer(color),
	}, nil
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func nullableTimestamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return timestamp(*value)
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func reminderTimestamp(event *domain.AgendaEvent, status domain.ReminderStatus) pgtype.Timestamptz {
	if event.ReminderStatus != status {
		return pgtype.Timestamptz{}
	}
	return timestamp(event.UpdatedAt)
}

func reminderFailureReason(status domain.ReminderStatus) pgtype.Text {
	switch status {
	case domain.ReminderFailToSend:
		reason := "notification_failed"
		return nullableText(&reason)
	case domain.ReminderUnprocessable:
		reason := "unprocessable"
		return nullableText(&reason)
	default:
		return pgtype.Text{}
	}
}

func appointmentReminderStatus(status domain.ReminderStatus) string {
	switch status {
	case domain.ReminderPending:
		return "pending"
	case domain.ReminderScheduled:
		return "scheduled"
	case domain.ReminderUnprocessable:
		return "unprocessable"
	case domain.ReminderSentRequested:
		return "send_requested"
	case domain.ReminderSent:
		return "sent"
	case domain.ReminderFailToSend:
		return "failed"
	case domain.ReminderDeleted:
		return "deleted"
	default:
		return strings.ToLower(string(status))
	}
}

func legacyReminderStatus(status pgtype.Text) domain.ReminderStatus {
	if !status.Valid {
		return domain.ReminderPending
	}
	switch status.String {
	case "pending":
		return domain.ReminderPending
	case "scheduled":
		return domain.ReminderScheduled
	case "unprocessable":
		return domain.ReminderUnprocessable
	case "send_requested":
		return domain.ReminderSentRequested
	case "sent":
		return domain.ReminderSent
	case "failed":
		return domain.ReminderFailToSend
	case "deleted":
		return domain.ReminderDeleted
	default:
		return domain.ReminderStatus(strings.ToUpper(status.String))
	}
}

func notificationKind(notificationType string) string {
	switch notificationType {
	case application.NotificationTypeAppointmentConfirmation:
		return "confirmation"
	case application.NotificationTypeAppointmentRescheduled:
		return "rescheduled"
	case application.NotificationTypeAppointmentReminder, "reminder", "Reminder":
		return "reminder"
	default:
		return notificationType
	}
}

func notificationType(notificationType string) string {
	switch notificationType {
	case "reminder", "Reminder":
		return application.NotificationTypeAppointmentReminder
	default:
		return notificationType
	}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func (r *Repository) publishLifecycleEvents(ctx context.Context, events []domain.LifecycleEvent) error {
	if len(events) == 0 {
		return nil
	}
	messages := make([]outbox.Message, 0, len(events))
	for _, event := range events {
		message, err := newOutboxMessage(event)
		if err != nil {
			return err
		}
		messages = append(messages, message)
	}
	if err := r.publisher.Publish(ctx, messages...); err != nil {
		return fmt.Errorf("publish outbox appointment events: %w", err)
	}
	return nil
}

func newOutboxMessage(event domain.LifecycleEvent) (outbox.Message, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return outbox.Message{}, fmt.Errorf("marshal appointment lifecycle event: %w", err)
	}
	return outbox.Message{
		ID:          uuid.NewString(),
		Channel:     outbox.Channel(ChannelAppointmentInternalJob),
		AffinityKey: outbox.AffinityKey(event.AgendaEventID),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  time.Now().UTC(),
	}, nil
}
