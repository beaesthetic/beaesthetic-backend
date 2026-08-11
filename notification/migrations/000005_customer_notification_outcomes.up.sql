ALTER TABLE customer_notifications
    ADD COLUMN IF NOT EXISTS dispatched_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS failure_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS failure_message TEXT NULL;
