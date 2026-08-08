package postgres

import (
	"context"
	"fmt"
	"io"
)

const uuidPattern = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`

type LegacyAgendaBackfill struct {
	LegacyAppointments          int64
	BackfilledAppointments      int64
	SkippedInvalidCustomers     int64
	BackfilledManualEvents      int64
	BackfilledServiceItems      int64
	BackfilledReminders         int64
	BackfilledNotifications     int64
	SkippedExpiredNotifications int64
}

type LegacyAgendaBackfiller struct {
	db *ContextDB
}

func NewLegacyAgendaBackfiller(db *ContextDB) *LegacyAgendaBackfiller {
	return &LegacyAgendaBackfiller{db: db}
}

func (b *LegacyAgendaBackfiller) Backfill(ctx context.Context, dryRun bool, out io.Writer) (LegacyAgendaBackfill, error) {
	report, err := b.inspectLegacyAgenda(ctx)
	if err != nil {
		return LegacyAgendaBackfill{}, err
	}
	if dryRun {
		writeBackfillReport(out, report, true)
		return report, nil
	}
	if err := validateLegacyAgendaBackfill(report); err != nil {
		return report, err
	}
	err = b.db.Tx(ctx, func(ctx context.Context) error {
		var err error
		if _, err = execRows(ctx, b.db, normalizeCalendarEventBaseSQL); err != nil {
			return fmt.Errorf("normalize calendar event base fields: %w", err)
		}
		if report.BackfilledAppointments, err = execRows(ctx, b.db, backfillAppointmentsSQL, uuidPattern); err != nil {
			return fmt.Errorf("backfill appointments: %w", err)
		}
		if report.BackfilledManualEvents, err = execRows(ctx, b.db, backfillManualEventsSQL); err != nil {
			return fmt.Errorf("backfill manual events: %w", err)
		}
		if _, err = execRows(ctx, b.db, normalizeManualEventTypesSQL); err != nil {
			return fmt.Errorf("normalize manual event types: %w", err)
		}
		if report.BackfilledServiceItems, err = execRows(ctx, b.db, backfillAppointmentServiceItemsSQL); err != nil {
			return fmt.Errorf("backfill appointment service items: %w", err)
		}
		if report.BackfilledReminders, err = execRows(ctx, b.db, backfillAppointmentRemindersSQL); err != nil {
			return fmt.Errorf("backfill appointment reminders: %w", err)
		}
		if report.BackfilledNotifications, err = execRows(ctx, b.db, backfillAppointmentNotificationsSQL); err != nil {
			return fmt.Errorf("backfill appointment notifications: %w", err)
		}
		return nil
	})
	if err != nil {
		return LegacyAgendaBackfill{}, err
	}
	writeBackfillReport(out, report, false)
	return report, nil
}

func validateLegacyAgendaBackfill(report LegacyAgendaBackfill) error {
	if report.SkippedInvalidCustomers > 0 {
		return fmt.Errorf("cannot backfill: %d appointments have a non-UUID customer id", report.SkippedInvalidCustomers)
	}
	return nil
}

func (b *LegacyAgendaBackfiller) inspectLegacyAgenda(ctx context.Context) (LegacyAgendaBackfill, error) {
	var report LegacyAgendaBackfill
	if err := b.db.QueryRow(ctx, `SELECT count(*) FROM agenda_events WHERE event_type = 'appointment'`).Scan(&report.LegacyAppointments); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count legacy appointments: %w", err)
	}
	if err := b.db.QueryRow(ctx, `SELECT count(*) FROM agenda_events WHERE event_type = 'appointment' AND attendee_id !~ $1`, uuidPattern).Scan(&report.SkippedInvalidCustomers); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count invalid appointment customers: %w", err)
	}
	if err := b.db.QueryRow(ctx, `SELECT count(*) FROM agenda_events WHERE event_type = 'appointment' AND attendee_id ~ $1`, uuidPattern).Scan(&report.BackfilledAppointments); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count migratable appointments: %w", err)
	}
	if err := b.db.QueryRow(ctx, `SELECT count(*) FROM agenda_events WHERE event_type = 'event'`).Scan(&report.BackfilledManualEvents); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count migratable manual events: %w", err)
	}
	if err := b.db.QueryRow(ctx, `
		SELECT count(*)
		FROM agenda_events e
		JOIN LATERAL jsonb_array_elements(e.services) AS service(item) ON true
		WHERE e.event_type = 'appointment'
		  AND e.attendee_id ~ $1
		  AND coalesce(nullif(service.item ->> 'Name', ''), nullif(service.item ->> 'name', ''), '') <> ''
	`, uuidPattern).Scan(&report.BackfilledServiceItems); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count migratable appointment service items: %w", err)
	}
	if err := b.db.QueryRow(ctx, `SELECT count(*) FROM agenda_events WHERE event_type = 'appointment' AND attendee_id ~ $1`, uuidPattern).Scan(&report.BackfilledReminders); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count migratable reminders: %w", err)
	}
	if err := b.db.QueryRow(ctx, `
		SELECT count(*)
		FROM pending_notifications p
		JOIN agenda_events e ON e.id = p.agenda_event_id
		WHERE e.event_type = 'appointment'
		  AND e.attendee_id ~ $1
		  AND p.expires_at > now()
	`, uuidPattern).Scan(&report.BackfilledNotifications); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count migratable notifications: %w", err)
	}
	if err := b.db.QueryRow(ctx, `
		SELECT count(*)
		FROM pending_notifications p
		JOIN agenda_events e ON e.id = p.agenda_event_id
		WHERE e.event_type = 'appointment'
		  AND e.attendee_id ~ $1
		  AND p.expires_at <= now()
	`, uuidPattern).Scan(&report.SkippedExpiredNotifications); err != nil {
		return LegacyAgendaBackfill{}, fmt.Errorf("count expired appointment notifications: %w", err)
	}
	return report, nil
}

func execRows(ctx context.Context, db *ContextDB, sql string, args ...any) (int64, error) {
	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func writeBackfillReport(out io.Writer, report LegacyAgendaBackfill, dryRun bool) {
	if out == nil {
		return
	}
	mode := "run"
	if dryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(out, "legacy agenda backfill %s\n", mode)
	fmt.Fprintf(out, "legacy_appointments=%d\n", report.LegacyAppointments)
	fmt.Fprintf(out, "skipped_invalid_customers=%d\n", report.SkippedInvalidCustomers)
	fmt.Fprintf(out, "backfilled_appointments=%d\n", report.BackfilledAppointments)
	fmt.Fprintf(out, "backfilled_manual_events=%d\n", report.BackfilledManualEvents)
	fmt.Fprintf(out, "backfilled_service_items=%d\n", report.BackfilledServiceItems)
	fmt.Fprintf(out, "backfilled_reminders=%d\n", report.BackfilledReminders)
	fmt.Fprintf(out, "backfilled_notifications=%d\n", report.BackfilledNotifications)
	fmt.Fprintf(out, "skipped_expired_notifications=%d\n", report.SkippedExpiredNotifications)
}

const backfillAppointmentsSQL = `
INSERT INTO appointments (
    agenda_event_id,
    customer_id,
    customer_display_name,
    created_at,
    updated_at
)
SELECT
    id,
    attendee_id::uuid,
    attendee_display_name,
    created_at,
    updated_at
FROM agenda_events
WHERE event_type = 'appointment'
  AND attendee_id ~ $1
ON CONFLICT (agenda_event_id) DO UPDATE SET
    customer_id = EXCLUDED.customer_id,
    customer_display_name = EXCLUDED.customer_display_name,
    updated_at = EXCLUDED.updated_at
`

const normalizeManualEventTypesSQL = `
UPDATE agenda_events
SET event_type = 'manual'
WHERE event_type = 'event'
`

const normalizeCalendarEventBaseSQL = `
UPDATE agenda_events
SET calendar_id = 'd2a36e25-4824-4167-a062-a5af96f97703',
    display_title = coalesce(display_title, nullif(title, '')),
    display_description = coalesce(display_description, nullif(description, '')),
    canceled_at = CASE
        WHEN cancel_reason IS NOT NULL THEN coalesce(canceled_at, updated_at)
        ELSE canceled_at
    END
WHERE calendar_id <> 'd2a36e25-4824-4167-a062-a5af96f97703'
   OR (display_title IS NULL AND title <> '')
   OR (display_description IS NULL AND description <> '')
   OR (cancel_reason IS NOT NULL AND canceled_at IS NULL)
`

const backfillManualEventsSQL = `
INSERT INTO agenda_manual_events (
    agenda_event_id,
    title,
    description,
    location,
    created_at,
    updated_at
)
SELECT
    id,
    title,
    NULLIF(description, ''),
    NULL,
    created_at,
    updated_at
FROM agenda_events
WHERE event_type = 'event'
ON CONFLICT (agenda_event_id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    location = EXCLUDED.location,
    updated_at = EXCLUDED.updated_at
`

const backfillAppointmentServiceItemsSQL = `
DELETE FROM appointment_service_items
WHERE agenda_event_id IN (
    SELECT agenda_event_id
    FROM appointments
);

INSERT INTO appointment_service_items (
    agenda_event_id,
    service_id,
    service_name,
    price,
    position
)
SELECT
    e.id,
    NULL,
    coalesce(nullif(service.item ->> 'Name', ''), nullif(service.item ->> 'name', ''), ''),
    NULL,
    (service.ordinality - 1)::integer
FROM agenda_events e
JOIN appointments a ON a.agenda_event_id = e.id
JOIN LATERAL jsonb_array_elements(e.services) WITH ORDINALITY AS service(item, ordinality) ON true
WHERE e.event_type = 'appointment'
  AND coalesce(nullif(service.item ->> 'Name', ''), nullif(service.item ->> 'name', ''), '') <> ''
ON CONFLICT (agenda_event_id, position) DO UPDATE SET
    service_id = EXCLUDED.service_id,
    service_name = EXCLUDED.service_name,
    price = EXCLUDED.price
`

const backfillAppointmentRemindersSQL = `
INSERT INTO appointment_reminders (
    agenda_event_id,
    status,
    remind_before_seconds,
    scheduled_at,
    sent_requested_at,
    sent_at,
    failed_at,
    failure_reason,
    updated_at
)
SELECT
    a.agenda_event_id,
    CASE e.reminder_status
        WHEN 'PENDING' THEN 'pending'
        WHEN 'SCHEDULED' THEN 'scheduled'
        WHEN 'UNPROCESSABLE' THEN 'unprocessable'
        WHEN 'SENT_REQUESTED' THEN 'send_requested'
        WHEN 'SENT' THEN 'sent'
        WHEN 'FAIL_TO_SEND' THEN 'failed'
        WHEN 'DELETED' THEN 'deleted'
        ELSE lower(e.reminder_status)
    END,
    e.remind_before_seconds,
    CASE WHEN e.reminder_status = 'SCHEDULED' THEN e.updated_at ELSE NULL END,
    CASE WHEN e.reminder_status = 'SENT_REQUESTED' THEN e.updated_at ELSE NULL END,
    e.reminder_sent_at,
    CASE WHEN e.reminder_status = 'FAIL_TO_SEND' THEN e.updated_at ELSE NULL END,
    CASE
        WHEN e.reminder_status = 'FAIL_TO_SEND' THEN 'notification_failed'
        WHEN e.reminder_status = 'UNPROCESSABLE' THEN 'unprocessable'
        ELSE NULL
    END,
    e.updated_at
FROM agenda_events e
JOIN appointments a ON a.agenda_event_id = e.id
WHERE e.event_type = 'appointment'
ON CONFLICT (agenda_event_id) DO UPDATE SET
    status = EXCLUDED.status,
    remind_before_seconds = EXCLUDED.remind_before_seconds,
    scheduled_at = EXCLUDED.scheduled_at,
    sent_requested_at = EXCLUDED.sent_requested_at,
    sent_at = EXCLUDED.sent_at,
    failed_at = EXCLUDED.failed_at,
    failure_reason = EXCLUDED.failure_reason,
    updated_at = EXCLUDED.updated_at
`

const backfillAppointmentNotificationsSQL = `
INSERT INTO appointment_notifications (
    correlation_key,
    agenda_event_id,
    notification_kind,
    notification_type,
    status,
    recipient_type,
    recipient_id,
    notification_idempotency_key,
    failure_reason,
    failure_message,
    created_at,
    completed_at,
    expires_at
)
SELECT
    p.correlation_key,
    a.agenda_event_id,
    CASE
        WHEN p.notification_type IN ('reminder', 'Reminder', 'appointment_reminder') THEN 'reminder'
        WHEN p.notification_type = 'appointment_rescheduled' THEN 'rescheduled'
        WHEN p.notification_type = 'appointment_confirmation' THEN 'confirmation'
        ELSE p.notification_type
    END,
    CASE
        WHEN p.notification_type IN ('reminder', 'Reminder') THEN 'appointment_reminder'
        ELSE p.notification_type
    END,
    'pending',
    'customer',
    a.customer_id,
    p.correlation_key,
    NULL,
    NULL,
    least(p.expires_at, now()),
    NULL,
    p.expires_at
FROM pending_notifications p
JOIN appointments a ON a.agenda_event_id = p.agenda_event_id
WHERE p.expires_at > now()
ON CONFLICT (correlation_key) DO UPDATE SET
    agenda_event_id = EXCLUDED.agenda_event_id,
    notification_kind = EXCLUDED.notification_kind,
    notification_type = EXCLUDED.notification_type,
    status = EXCLUDED.status,
    recipient_type = EXCLUDED.recipient_type,
    recipient_id = EXCLUDED.recipient_id,
    notification_idempotency_key = EXCLUDED.notification_idempotency_key,
    expires_at = EXCLUDED.expires_at
`
