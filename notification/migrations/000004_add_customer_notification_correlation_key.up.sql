ALTER TABLE customer_notifications
    ADD COLUMN IF NOT EXISTS correlation_key TEXT;

UPDATE customer_notifications
SET correlation_key = regexp_replace(
    idempotency_key,
    ':' || customer_id || ':' || notification_channel || ':' || notification_type || '$',
    ''
)
WHERE correlation_key IS NULL;

ALTER TABLE customer_notifications
    ALTER COLUMN correlation_key SET NOT NULL;
