ALTER TABLE appointment_services
    DROP COLUMN IF EXISTS price;

ALTER TABLE appointment_service_items
    DROP COLUMN IF EXISTS price;
