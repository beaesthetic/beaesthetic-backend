CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION appointment_service_tags_search_text(tags jsonb)
RETURNS text
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT coalesce(string_agg(tag.value, ' ' ORDER BY tag.ordinality), '')
    FROM jsonb_array_elements_text(coalesce(tags, '[]'::jsonb)) WITH ORDINALITY AS tag(value, ordinality)
$$;

CREATE TABLE agenda_events (
    id UUID PRIMARY KEY,
    calendar_id UUID NOT NULL DEFAULT 'd2a36e25-4824-4167-a062-a5af96f97703',
    event_type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    all_day BOOLEAN NOT NULL DEFAULT false,
    display_title TEXT NULL,
    display_description TEXT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    attendee_id TEXT NOT NULL,
    attendee_display_name TEXT NOT NULL,
    services JSONB NOT NULL DEFAULT '[]'::jsonb,
    cancel_reason TEXT NULL,
    canceled_at TIMESTAMPTZ NULL,
    reminder_status TEXT NOT NULL,
    reminder_sent_at TIMESTAMPTZ NULL,
    remind_before_seconds INTEGER NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE appointment_services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    color_hex TEXT NULL,
    search_text TEXT GENERATED ALWAYS AS (
        lower(
            coalesce(name, '') || ' ' ||
            appointment_service_tags_search_text(tags)
        )
    ) STORED
);

CREATE TABLE pending_notifications (
    correlation_key TEXT PRIMARY KEY,
    agenda_event_id UUID NOT NULL,
    notification_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE appointments (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    customer_id UUID NOT NULL,
    customer_display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE agenda_manual_events (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    title TEXT NOT NULL,
    description TEXT NULL,
    location TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE agenda_time_blocks (
    agenda_event_id UUID PRIMARY KEY REFERENCES agenda_events(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE appointment_service_items (
    agenda_event_id UUID NOT NULL REFERENCES appointments(agenda_event_id),
    service_id TEXT NULL,
    service_name TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (agenda_event_id, position)
);

CREATE TABLE appointment_reminders (
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

CREATE TABLE appointment_notifications (
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
