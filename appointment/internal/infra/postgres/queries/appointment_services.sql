-- name: SaveAppointmentService :exec
INSERT INTO appointment_services (id, name, tags, color_hex)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    name = $2,
    tags = $3,
    color_hex = $4;

-- name: FindAppointmentServices :many
SELECT id, name, tags, color_hex
FROM appointment_services
ORDER BY name;

-- name: SearchAppointmentServices :many
SELECT id, name, tags, color_hex
FROM appointment_services
WHERE search_text ILIKE '%' || sqlc.arg(query)::text || '%'
   OR search_text % sqlc.arg(query)::text
ORDER BY similarity(search_text, sqlc.arg(query)::text) DESC, name
LIMIT sqlc.arg(limit_count)::int;

-- name: FindAppointmentService :one
SELECT id, name, tags, color_hex
FROM appointment_services
WHERE id = $1;
