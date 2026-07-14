CREATE TABLE IF NOT EXISTS customer_notifications (
    id UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    notification_type TEXT NOT NULL,
    notification_channel TEXT NOT NULL,
    template_values JSONB NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ NULL,
    failed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_customer_notifications_customer_id
    ON customer_notifications (customer_id);

CREATE INDEX IF NOT EXISTS idx_customer_notifications_type_channel
    ON customer_notifications (notification_type, notification_channel);

CREATE TABLE IF NOT EXISTS customer_notification_sms_gateway_messages (
    id UUID PRIMARY KEY,
    customer_notification_id UUID NOT NULL REFERENCES customer_notifications (id) ON DELETE CASCADE,
    sms_gateway_message_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_customer_notification_sms_gateway_messages_notification_id
    ON customer_notification_sms_gateway_messages (customer_notification_id);
