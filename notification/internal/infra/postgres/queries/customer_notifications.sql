-- name: ExistsCustomerNotificationByIdempotencyKey :one
SELECT EXISTS (
    SELECT 1 FROM customer_notifications WHERE idempotency_key = $1
);

-- name: CreatePendingCustomerNotification :execrows
INSERT INTO customer_notifications (
    id,
    idempotency_key,
    correlation_key,
    customer_id,
    notification_type,
    notification_channel,
    template_values,
    status,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (idempotency_key) DO NOTHING;

-- name: SaveSMSGatewayDispatch :exec
INSERT INTO customer_notification_sms_gateway_messages (
    id,
    customer_notification_id,
    sms_gateway_message_id,
    created_at
)
VALUES ($1, $2, $3, $4);

-- name: MarkCustomerNotificationDispatched :execrows
UPDATE customer_notifications
SET status = 'dispatched', dispatched_at = $2
WHERE id = $1
  AND status = 'pending';

-- name: MarkCustomerNotificationSentBySMSGatewayMessageID :one
UPDATE customer_notifications
SET status = 'sent', sent_at = $2
WHERE id = (
    SELECT customer_notification_id
    FROM customer_notification_sms_gateway_messages
    WHERE sms_gateway_message_id = $1
)
RETURNING correlation_key, idempotency_key, customer_id;

-- name: MarkCustomerNotificationFailed :one
UPDATE customer_notifications
SET status = 'failed',
    failed_at = $2,
    failure_reason = $3,
    failure_message = $4
WHERE id = $1
  AND status IN ('pending', 'dispatched')
RETURNING correlation_key, idempotency_key, customer_id;

-- name: MarkCustomerNotificationFailedBySMSGatewayMessageID :one
UPDATE customer_notifications
SET status = 'failed',
    failed_at = $2,
    failure_reason = $3,
    failure_message = $4
WHERE id = (
    SELECT customer_notification_id
    FROM customer_notification_sms_gateway_messages
    WHERE sms_gateway_message_id = $1
)
RETURNING correlation_key, idempotency_key, customer_id;
