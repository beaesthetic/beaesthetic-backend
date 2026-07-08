DROP INDEX IF EXISTS idx_appointment_services_search_text_trgm;

ALTER TABLE appointment_services
    DROP COLUMN IF EXISTS search_text;

DROP FUNCTION IF EXISTS appointment_service_tags_search_text(jsonb);

ALTER TABLE appointment_services
    ADD COLUMN search_text TEXT GENERATED ALWAYS AS (
        lower(
            coalesce(name, '') || ' ' ||
            coalesce(tags::text, '')
        )
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_appointment_services_search_text_trgm
    ON appointment_services
    USING GIN (search_text gin_trgm_ops);