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

CREATE TABLE appointment_services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price DOUBLE PRECISION NOT NULL,
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
