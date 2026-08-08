-- name: SaveAgendaEventV2 :exec
INSERT INTO agenda_events (
    id,
    calendar_id,
    event_type,
    title,
    description,
    start_at,
    end_at,
    timezone,
    all_day,
    display_title,
    display_description,
    visibility,
    attendee_id,
    attendee_display_name,
    services,
    cancel_reason,
    canceled_at,
    reminder_status,
    reminder_sent_at,
    remind_before_seconds,
    version,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, '[]'::jsonb, $15, $16, 'PENDING', NULL, 0, 1, $17, $18
)
ON CONFLICT (id) DO UPDATE SET
    calendar_id = $2,
    event_type = $3,
    title = $4,
    description = $5,
    start_at = $6,
    end_at = $7,
    timezone = $8,
    all_day = $9,
    display_title = $10,
    display_description = $11,
    visibility = $12,
    attendee_id = $13,
    attendee_display_name = $14,
    cancel_reason = $15,
    canceled_at = $16,
    version = agenda_events.version + 1,
    updated_at = $18;

-- name: SaveAppointment :exec
INSERT INTO appointments (
    agenda_event_id,
    customer_id,
    customer_display_name,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (agenda_event_id) DO UPDATE SET
    customer_id = $2,
    customer_display_name = $3,
    updated_at = $5;

-- name: DeleteAppointmentServiceItems :exec
DELETE FROM appointment_service_items
WHERE agenda_event_id = $1;

-- name: SaveAppointmentServiceItem :exec
INSERT INTO appointment_service_items (
    agenda_event_id,
    service_id,
    service_name,
    price,
    position
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (agenda_event_id, position) DO UPDATE SET
    service_id = $2,
    service_name = $3,
    price = $4;

-- name: SaveAppointmentReminder :exec
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
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (agenda_event_id) DO UPDATE SET
    status = $2,
    remind_before_seconds = $3,
    scheduled_at = $4,
    sent_requested_at = $5,
    sent_at = $6,
    failed_at = $7,
    failure_reason = $8,
    updated_at = $9;

-- name: SaveAgendaManualEvent :exec
INSERT INTO agenda_manual_events (
    agenda_event_id,
    title,
    description,
    location,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (agenda_event_id) DO UPDATE SET
    title = $2,
    description = $3,
    location = $4,
    updated_at = $6;

-- name: SaveAgendaTimeBlock :exec
INSERT INTO agenda_time_blocks (
    agenda_event_id,
    reason,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (agenda_event_id) DO UPDATE SET
    reason = $2,
    updated_at = $4;

-- name: SaveAppointmentNotification :exec
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
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (correlation_key) DO UPDATE SET
    agenda_event_id = $2,
    notification_kind = $3,
    notification_type = $4,
    status = $5,
    recipient_type = $6,
    recipient_id = $7,
    notification_idempotency_key = $8,
    failure_reason = $9,
    failure_message = $10,
    completed_at = $12,
    expires_at = $13;

-- name: MarkAppointmentNotificationSent :exec
UPDATE appointment_notifications
SET status = 'sent',
    failure_reason = NULL,
    failure_message = NULL,
    completed_at = $2
WHERE correlation_key = $1;

-- name: FindAppointmentNotification :one
SELECT
    n.correlation_key,
    a.agenda_event_id,
    n.notification_kind,
    n.notification_type,
    n.status,
    n.recipient_type,
    n.recipient_id,
    n.notification_idempotency_key,
    n.failure_reason,
    n.failure_message,
    n.created_at,
    n.completed_at,
    n.expires_at
FROM appointment_notifications n
JOIN appointments a ON a.agenda_event_id = n.agenda_event_id
WHERE n.correlation_key = $1;

-- name: FindAgendaEventFromDetails :one
SELECT
    e.id,
    e.calendar_id,
    e.event_type,
    e.title AS legacy_title,
    e.description AS legacy_description,
    e.start_at,
    e.end_at,
    e.timezone,
    e.all_day,
    e.display_title,
    e.display_description,
    e.visibility,
    e.cancel_reason,
    e.canceled_at,
    e.version,
    e.created_at,
    e.updated_at,
    CASE WHEN a.agenda_event_id IS NULL THEN '' ELSE a.agenda_event_id::text END AS agenda_event_id,
    CASE WHEN a.customer_id IS NULL THEN '' ELSE a.customer_id::text END AS customer_id,
    a.customer_display_name,
    m.title AS manual_title,
    m.description AS manual_description,
    m.location AS manual_location,
    tb.reason AS time_block_reason,
    r.status AS reminder_status,
    r.remind_before_seconds,
    r.scheduled_at AS reminder_scheduled_at,
    r.sent_requested_at AS reminder_sent_requested_at,
    r.sent_at AS reminder_sent_at,
    r.failed_at AS reminder_failed_at,
    r.failure_reason AS reminder_failure_reason,
    r.updated_at AS reminder_updated_at,
    coalesce(
        jsonb_agg(
            jsonb_build_object(
                'Name', si.service_name,
                'serviceId', si.service_id,
                'price', si.price,
                'position', si.position
            )
            ORDER BY si.position
        ) FILTER (WHERE si.agenda_event_id IS NOT NULL),
        '[]'::jsonb
    )::text AS services_json
FROM agenda_events e
LEFT JOIN appointments a ON a.agenda_event_id = e.id
LEFT JOIN agenda_manual_events m ON m.agenda_event_id = e.id
LEFT JOIN agenda_time_blocks tb ON tb.agenda_event_id = e.id
LEFT JOIN appointment_reminders r ON r.agenda_event_id = a.agenda_event_id
LEFT JOIN appointment_service_items si ON si.agenda_event_id = a.agenda_event_id
WHERE e.id = $1
GROUP BY
    e.id,
    e.calendar_id,
    e.event_type,
    e.title,
    e.description,
    e.start_at,
    e.end_at,
    e.timezone,
    e.all_day,
    e.display_title,
    e.display_description,
    e.visibility,
    e.cancel_reason,
    e.canceled_at,
    e.version,
    e.created_at,
    e.updated_at,
    a.agenda_event_id,
    a.customer_id,
    a.customer_display_name,
    m.title,
    m.description,
    m.location,
    tb.reason,
    r.status,
    r.remind_before_seconds,
    r.scheduled_at,
    r.sent_requested_at,
    r.sent_at,
    r.failed_at,
    r.failure_reason,
    r.updated_at;

-- name: SearchAgendaEventIDsFromDetails :many
SELECT e.id
FROM agenda_events e
LEFT JOIN appointments a ON a.agenda_event_id = e.id
WHERE e.canceled_at IS NULL
  AND (@filter_calendar::boolean = false OR e.calendar_id::text = @calendar_id::text)
  AND (@filter_customer::boolean = false OR a.customer_id::text = @customer_id::text)
  AND (@filter_event_types::boolean = false OR e.event_type = ANY(@event_types::text[]))
  AND (@filter_time_range::boolean = false OR (e.start_at < @end_at::timestamptz AND e.end_at > @start_at::timestamptz))
ORDER BY e.start_at ASC, e.end_at ASC;

-- name: FindFutureAppointmentAgendaEventIDsFromDetails :many
SELECT e.id
FROM agenda_events e
JOIN appointments a ON a.agenda_event_id = e.id
WHERE e.canceled_at IS NULL
  AND e.start_at >= $1
ORDER BY e.start_at ASC, e.end_at ASC;

-- name: MarkAppointmentNotificationFailed :exec
UPDATE appointment_notifications
SET status = 'failed',
    failure_reason = $2,
    failure_message = $3,
    completed_at = $4
WHERE correlation_key = $1;
