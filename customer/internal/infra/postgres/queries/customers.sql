-- name: SaveCustomer :exec
INSERT INTO customers (id, name, surname, email, phone, note, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    name = $2,
    surname = $3,
    email = $4,
    phone = $5,
    note = $6,
    updated_at = $7;

-- name: FindCustomerByID :one
SELECT id, name, surname, email, phone, note
FROM customers
WHERE id = $1;

-- name: FindCustomers :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY name, surname
LIMIT $1;

-- name: SearchCustomers :many
SELECT id, name, surname, email, phone, note
FROM customers
WHERE search_text ILIKE '%' || $1 || '%' OR search_text % $1
ORDER BY similarity(search_text, $1) DESC, name, surname
LIMIT $2;

-- name: FindCustomerByPhone :one
SELECT id, name, surname, email, phone, note
FROM customers
WHERE phone = $1
LIMIT 1;

-- name: FindCustomersPageByNameAsc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY name ASC, id
LIMIT $1 OFFSET $2;

-- name: FindCustomersPageByNameDesc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY name DESC, id
LIMIT $1 OFFSET $2;

-- name: FindCustomersPageBySurnameAsc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY surname ASC, id
LIMIT $1 OFFSET $2;

-- name: FindCustomersPageBySurnameDesc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY surname DESC, id
LIMIT $1 OFFSET $2;

-- name: FindCustomersPageByUpdatedAtAsc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY updated_at ASC, id
LIMIT $1 OFFSET $2;

-- name: FindCustomersPageByUpdatedAtDesc :many
SELECT id, name, surname, email, phone, note
FROM customers
ORDER BY updated_at DESC, id
LIMIT $1 OFFSET $2;

-- name: ArchiveDeletedCustomer :execrows
INSERT INTO deleted_customers (id, name, surname, email, phone, note)
SELECT c.id, c.name, c.surname, c.email, c.phone, c.note
FROM customers c
WHERE c.id = $1
ON CONFLICT (id) DO NOTHING;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;
