-- name: FindPendingNotification :one
SELECT correlation_key, agenda_event_id, notification_type
FROM pending_notifications
WHERE correlation_key = $1;

-- name: RemovePendingNotification :exec
DELETE FROM pending_notifications
WHERE correlation_key = $1;

-- name: SavePendingNotification :exec
INSERT INTO pending_notifications (correlation_key, agenda_event_id, notification_type, expires_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (correlation_key) DO UPDATE SET
    agenda_event_id = $2,
    notification_type = $3,
    expires_at = $4;
