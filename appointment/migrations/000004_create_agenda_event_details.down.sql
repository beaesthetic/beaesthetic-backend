DROP TABLE IF EXISTS appointment_notifications;
DROP TABLE IF EXISTS appointment_reminders;
DROP TABLE IF EXISTS appointment_service_items;
DROP TABLE IF EXISTS agenda_time_blocks;
DROP TABLE IF EXISTS agenda_manual_events;
DROP TABLE IF EXISTS appointments;

ALTER TABLE agenda_events
    DROP COLUMN IF EXISTS canceled_at,
    DROP COLUMN IF EXISTS visibility,
    DROP COLUMN IF EXISTS display_description,
    DROP COLUMN IF EXISTS display_title,
    DROP COLUMN IF EXISTS all_day,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS calendar_id;
