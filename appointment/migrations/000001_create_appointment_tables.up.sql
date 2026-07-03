CREATE TABLE IF NOT EXISTS agenda_events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    attendee_id TEXT NOT NULL,
    attendee_display_name TEXT NOT NULL,
    services JSONB NOT NULL DEFAULT '[]'::jsonb,
    cancel_reason TEXT NULL,
    reminder_status TEXT NOT NULL,
    reminder_sent_at TIMESTAMPTZ NULL,
    remind_before_seconds INTEGER NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agenda_events_time ON agenda_events (start_at, end_at);
CREATE INDEX IF NOT EXISTS idx_agenda_events_attendee ON agenda_events (attendee_id);
CREATE INDEX IF NOT EXISTS idx_agenda_events_reminder_status ON agenda_events (reminder_status);

CREATE TABLE IF NOT EXISTS appointment_services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price DOUBLE PRECISION NOT NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    color_hex TEXT NULL,
    search_grams TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS pending_notifications (
    notification_id UUID PRIMARY KEY,
    agenda_event_id UUID NOT NULL,
    notification_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
