ALTER TABLE pending_notifications
    ALTER COLUMN notification_id TYPE TEXT;

ALTER TABLE pending_notifications
    RENAME COLUMN notification_id TO correlation_key;
