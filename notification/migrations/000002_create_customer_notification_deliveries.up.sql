CREATE TABLE IF NOT EXISTS customer_notification_deliveries (
    idempotency_key TEXT PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications (id),
    customer_id TEXT NOT NULL,
    notification_type TEXT NOT NULL,
    notification_channel TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_customer_notification_deliveries_customer_id
    ON customer_notification_deliveries (customer_id);
