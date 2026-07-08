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

DROP INDEX IF EXISTS idx_appointment_services_search_grams_trgm;
DROP INDEX IF EXISTS appointment_services_search_grams_trgm_idx;
DROP INDEX IF EXISTS idx_appointment_services_search_grams;
DROP INDEX IF EXISTS idx_appointment_services_search_text_trgm;

ALTER TABLE appointment_services
    DROP COLUMN IF EXISTS search_grams;

ALTER TABLE appointment_services
    DROP COLUMN IF EXISTS search_text;

ALTER TABLE appointment_services
    ADD COLUMN search_text TEXT GENERATED ALWAYS AS (
        lower(
            coalesce(name, '') || ' ' ||
            appointment_service_tags_search_text(tags)
        )
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_appointment_services_search_text_trgm
    ON appointment_services
    USING GIN (search_text gin_trgm_ops);