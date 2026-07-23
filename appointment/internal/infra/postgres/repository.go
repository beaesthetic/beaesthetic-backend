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
	if e.CancelReason != nil {
		v := string(*e.CancelReason)
		cancelReason = &v
	}
	err := queries.New(r.db).SaveAgendaEvent(ctx, queries.SaveAgendaEventParams{
		ID:                  e.ID,
		EventType:           string(e.Type),
		Title:               e.Title,
		Description:         e.Description,
		StartAt:             timestamp(e.Start),
		EndAt:               timestamp(e.End),
		AttendeeID:          e.Attendee.ID,
		AttendeeDisplayName: e.Attendee.DisplayName,
		Services:            services,
		CancelReason:        nullableText(cancelReason),
		ReminderStatus:      string(e.ReminderStatus),
		ReminderSentAt:      nullableTimestamp(e.ReminderSentAt),
		RemindBeforeSeconds: int32(e.RemindBefore.Seconds()),
		CreatedAt:           timestamp(e.CreatedAt),
		UpdatedAt:           timestamp(e.UpdatedAt),
	})
	if err != nil {
		return err
	}
	return r.publishLifecycleEvents(ctx, events)
}

func (r *Repository) FindAgendaEvent(ctx context.Context, id string) (*domain.AgendaEvent, error) {
	row, err := queries.New(r.db).FindAgendaEvent(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return agendaEventFromRow(row)
}

func (r *Repository) SearchAgendaEvents(ctx context.Context, attendeeID string, start, end *time.Time) ([]domain.AgendaEvent, error) {
	params := queries.SearchAgendaEventIDsParams{}
	if attendeeID != "" {
		params.FilterAttendee = true
		params.AttendeeID = attendeeID
	}
	if start != nil && end != nil {
		params.FilterTimeRange = true
		params.StartAt = timestamp(*start)
		params.EndAt = timestamp(*end)
	}
	ids, err := queries.New(r.db).SearchAgendaEventIDs(ctx, params)
	if err != nil {
		return nil, err
	}
	return r.findAgendaEventsByIDs(ctx, ids)
}

func (r *Repository) FindFutureAppointments(ctx context.Context, from time.Time) ([]domain.AgendaEvent, error) {
	ids, err := queries.New(r.db).FindFutureAppointmentIDs(ctx, queries.FindFutureAppointmentIDsParams{
		EventType: string(domain.EventTypeAppointment),
		StartAt:   timestamp(from),
	})
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

func (r *Repository) FindPendingNotification(ctx context.Context, correlationKey string) (*application.PendingNotification, error) {
	row, err := queries.New(r.db).FindPendingNotification(ctx, correlationKey)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	pending := application.PendingNotification{
		CorrelationKey: row.CorrelationKey,
		AgendaEventID:  row.AgendaEventID,
		Type:           row.NotificationType,
	}
	return &pending, nil
}

func (r *Repository) RemovePendingNotification(ctx context.Context, correlationKey string) error {
	return queries.New(r.db).RemovePendingNotification(ctx, correlationKey)
}

func (r *Repository) SavePendingNotification(ctx context.Context, pending application.PendingNotification) error {
	return queries.New(r.db).SavePendingNotification(ctx, queries.SavePendingNotificationParams{
		CorrelationKey:   pending.CorrelationKey,
		AgendaEventID:    pending.AgendaEventID,
		NotificationType: pending.Type,
		ExpiresAt:        timestamp(time.Now().UTC().Add(24 * time.Hour)),
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
