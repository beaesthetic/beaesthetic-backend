package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) SaveAgendaEvent(ctx context.Context, e *domain.AgendaEvent) error {
	events := e.PullEvents()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	services, _ := json.Marshal(e.Services)
	var cancelReason *string
	if e.CancelReason != nil {
		v := string(*e.CancelReason)
		cancelReason = &v
	}
	_, err = tx.Exec(ctx, `
INSERT INTO agenda_events (id,event_type,title,description,start_at,end_at,attendee_id,attendee_display_name,services,cancel_reason,reminder_status,reminder_sent_at,remind_before_seconds,version,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,1,$14,$15)
ON CONFLICT (id) DO UPDATE SET event_type=$2,title=$3,description=$4,start_at=$5,end_at=$6,attendee_id=$7,attendee_display_name=$8,services=$9,cancel_reason=$10,reminder_status=$11,reminder_sent_at=$12,remind_before_seconds=$13,version=agenda_events.version+1,updated_at=$15`, e.ID, string(e.Type), e.Title, e.Description, e.Start, e.End, e.Attendee.ID, e.Attendee.DisplayName, services, cancelReason, string(e.ReminderStatus), e.ReminderSentAt, int(e.RemindBefore.Seconds()), e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := insertOutbox(ctx, tx, "appointment.lifecycle", event.AgendaEventID, event); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) FindAgendaEvent(ctx context.Context, id string) (*domain.AgendaEvent, error) {
	var e domain.AgendaEvent
	var eventType, reminderStatus string
	var servicesBytes []byte
	var cancelReason *string
	var remindSeconds int
	err := r.db.QueryRow(ctx, `SELECT id,event_type,title,description,start_at,end_at,attendee_id,attendee_display_name,services,cancel_reason,reminder_status,reminder_sent_at,remind_before_seconds,version,created_at,updated_at FROM agenda_events WHERE id=$1`, id).Scan(&e.ID, &eventType, &e.Title, &e.Description, &e.Start, &e.End, &e.Attendee.ID, &e.Attendee.DisplayName, &servicesBytes, &cancelReason, &reminderStatus, &e.ReminderSentAt, &remindSeconds, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(servicesBytes, &e.Services)
	e.Type = domain.EventType(eventType)
	e.ReminderStatus = domain.ReminderStatus(reminderStatus)
	e.RemindBefore = time.Duration(remindSeconds) * time.Second
	if cancelReason != nil {
		v := domain.CancelReason(*cancelReason)
		e.CancelReason = &v
	}
	return &e, nil
}

func (r *Repository) SearchAgendaEvents(ctx context.Context, attendeeID string, start, end *time.Time) ([]domain.AgendaEvent, error) {
	where := []string{"cancel_reason IS NULL"}
	args := []any{}
	idx := 1
	if attendeeID != "" {
		where = append(where, fmt.Sprintf("attendee_id=$%d", idx))
		args = append(args, attendeeID)
		idx++
	}
	if start != nil && end != nil {
		where = append(where, fmt.Sprintf("start_at >= $%d AND start_at <= $%d", idx, idx+1))
		args = append(args, *start, *end)
	}
	rows, err := r.db.Query(ctx, `SELECT id FROM agenda_events WHERE `+strings.Join(where, " AND ")+` ORDER BY start_at ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AgendaEvent
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		e, err := r.FindAgendaEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		if e != nil {
			out = append(out, *e)
		}
	}
	return out, rows.Err()
}

func (r *Repository) SaveService(ctx context.Context, s domain.AppointmentService) (domain.AppointmentService, error) {
	tags, _ := json.Marshal(s.Tags)
	_, err := r.db.Exec(ctx, `INSERT INTO appointment_services (id,name,price,tags,color_hex,search_grams) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (id) DO UPDATE SET name=$2,price=$3,tags=$4,color_hex=$5,search_grams=$6`, s.ID, s.Name, s.Price, tags, s.Color, strings.ToLower(s.Name+" "+strings.Join(s.Tags, " ")))
	return s, err
}
func (r *Repository) FindServices(ctx context.Context) ([]domain.AppointmentService, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,price,tags,color_hex FROM appointment_services ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServices(rows)
}
func (r *Repository) SearchServices(ctx context.Context, text string, limit int) ([]domain.AppointmentService, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,price,tags,color_hex FROM appointment_services WHERE search_grams ILIKE '%' || $1 || '%' ORDER BY name LIMIT $2`, strings.ToLower(text), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServices(rows)
}
func (r *Repository) FindService(ctx context.Context, id string) (*domain.AppointmentService, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,price,tags,color_hex FROM appointment_services WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	services, err := scanServices(rows)
	if err != nil || len(services) == 0 {
		return nil, err
	}
	return &services[0], nil
}

func scanServices(rows pgx.Rows) ([]domain.AppointmentService, error) {
	var out []domain.AppointmentService
	for rows.Next() {
		var s domain.AppointmentService
		var tags []byte
		if err := rows.Scan(&s.ID, &s.Name, &s.Price, &tags, &s.Color); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tags, &s.Tags)
		out = append(out, s)
	}
	return out, rows.Err()
}

func insertOutbox(ctx context.Context, tx pgx.Tx, channel, affinity string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox_messages (id,channel,affinity_key,payload,metadata,occurred_at) VALUES ($1,$2,$3,$4,$5,$6)`, uuid.NewString(), channel, affinity, b, []byte(`{}`), time.Now().UTC())
	return err
}
