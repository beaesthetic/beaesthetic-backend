ALTER TABLE pending_notifications
    RENAME COLUMN correlation_key TO notification_id;

ALTER TABLE pending_notifications
    ALTER COLUMN notification_id TYPE UUID USING notification_id::uuid;
