ALTER TABLE customer_notifications
    DROP COLUMN IF EXISTS failure_message,
    DROP COLUMN IF EXISTS failure_reason,
    DROP COLUMN IF EXISTS dispatched_at;
