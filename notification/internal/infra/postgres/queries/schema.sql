CREATE TABLE customer_notifications (
    id UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    correlation_key TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    notification_type TEXT NOT NULL,
    notification_channel TEXT NOT NULL,
    template_values JSONB NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ NULL,
    failed_at TIMESTAMPTZ NULL
);

CREATE TABLE customer_notification_sms_gateway_messages (
    id UUID PRIMARY KEY,
    customer_notification_id UUID NOT NULL REFERENCES customer_notifications (id) ON DELETE CASCADE,
    sms_gateway_message_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
