CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_sent BOOLEAN NOT NULL DEFAULT FALSE,
    is_sent_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    channel_type TEXT NOT NULL,
    channel_address TEXT NOT NULL,
    provider_resource_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT notifications_channel_type_check CHECK (channel_type IN ('sms', 'email', 'whatsapp'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_channel_type
    ON notifications (channel_type);
