ALTER TABLE agenda_events
    ADD COLUMN IF NOT EXISTS calendar_id UUID NOT NULL DEFAULT 'd2a36e25-4824-4167-a062-a5af96f97703',
    ADD COLUMN IF NOT EXISTS timezone TEXT NOT NULL DEFAULT 'UTC',
    ADD COLUMN IF NOT EXISTS all_day BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS display_title TEXT NULL,
    ADD COLUMN IF NOT EXISTS display_description TEXT NULL,
    ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'private',
    ADD COLUMN IF NOT EXISTS canceled_at TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS appointments (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    customer_id UUID NOT NULL,
    customer_display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_appointments_customer_id ON appointments (customer_id);

CREATE TABLE IF NOT EXISTS agenda_manual_events (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    title TEXT NOT NULL,
    description TEXT NULL,
    location TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agenda_time_blocks (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS appointment_service_items (
    agenda_event_id UUID NOT NULL REFERENCES appointments(agenda_event_id),
    service_id TEXT NULL,
    service_name TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (agenda_event_id, position)
);

CREATE TABLE IF NOT EXISTS appointment_reminders (
    agenda_event_id UUID PRIMARY KEY REFERENCES appointments(agenda_event_id),
    status TEXT NOT NULL,
    remind_before_seconds INTEGER NOT NULL,
    scheduled_at TIMESTAMPTZ NULL,
    sent_requested_at TIMESTAMPTZ NULL,
    sent_at TIMESTAMPTZ NULL,
    failed_at TIMESTAMPTZ NULL,
    failure_reason TEXT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS appointment_notifications (
    correlation_key TEXT PRIMARY KEY,
    agenda_event_id UUID NOT NULL REFERENCES appointments(agenda_event_id),
    notification_kind TEXT NOT NULL,
    notification_type TEXT NOT NULL,
    status TEXT NOT NULL,
    recipient_type TEXT NOT NULL,
    recipient_id TEXT NOT NULL,
    notification_idempotency_key TEXT NULL,
    failure_reason TEXT NULL,
    failure_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_appointment_notifications_agenda_event_id ON appointment_notifications (agenda_event_id);
CREATE INDEX IF NOT EXISTS idx_appointment_notifications_recipient ON appointment_notifications (recipient_type, recipient_id);
CREATE INDEX IF NOT EXISTS idx_appointment_notifications_status_expires_at ON appointment_notifications (status, expires_at);
