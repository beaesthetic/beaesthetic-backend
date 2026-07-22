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

-- name: MarkCustomerNotificationSentBySMSGatewayMessageID :one
UPDATE customer_notifications
SET status = 'sent', sent_at = $2
WHERE id = (
    SELECT customer_notification_id
    FROM customer_notification_sms_gateway_messages
    WHERE sms_gateway_message_id = $1
)
RETURNING correlation_key;

-- name: MarkCustomerNotificationFailedBySMSGatewayMessageID :execrows
UPDATE customer_notifications
SET status = 'failed', failed_at = $2
WHERE id = (
    SELECT customer_notification_id
    FROM customer_notification_sms_gateway_messages
    WHERE sms_gateway_message_id = $1
);
