-- name: SaveAgendaEvent :exec
INSERT INTO agenda_events (
    id,
    event_type,
    title,
    description,
    start_at,
    end_at,
    attendee_id,
    attendee_display_name,
    services,
    cancel_reason,
    reminder_status,
    reminder_sent_at,
    remind_before_seconds,
    version,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 1, $14, $15
)
ON CONFLICT (id) DO UPDATE SET
    event_type = $2,
    title = $3,
    description = $4,
    start_at = $5,
    end_at = $6,
    attendee_id = $7,
    attendee_display_name = $8,
    services = $9,
    cancel_reason = $10,
    reminder_status = $11,
    reminder_sent_at = $12,
    remind_before_seconds = $13,
    version = agenda_events.version + 1,
    updated_at = $15;

-- name: FindAgendaEvent :one
SELECT id,
       event_type,
       title,
       description,
       start_at,
       end_at,
       attendee_id,
       attendee_display_name,
       services,
       cancel_reason,
       reminder_status,
       reminder_sent_at,
       remind_before_seconds,
       version,
       created_at,
       updated_at
FROM agenda_events
WHERE id = $1;

-- name: SearchAgendaEventIDs :many
SELECT id
FROM agenda_events
WHERE cancel_reason IS NULL
  AND (@filter_attendee::boolean = false OR attendee_id = @attendee_id::text)
  AND (@filter_time_range::boolean = false OR (start_at >= @start_at::timestamptz AND start_at <= @end_at::timestamptz))
ORDER BY start_at ASC;

-- name: FindFutureAppointmentIDs :many
SELECT id
FROM agenda_events
WHERE cancel_reason IS NULL
  AND event_type = $1
  AND start_at >= $2
ORDER BY start_at ASC;
