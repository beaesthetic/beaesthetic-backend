-- name: SaveWallet :exec
INSERT INTO wallets (id, owner, available_amount, spent, operations, gift_cards, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    owner = $2,
    available_amount = $3,
    spent = $4,
    operations = $5,
    gift_cards = $6,
    updated_at = $8;

-- name: FindWalletByID :one
SELECT id, owner, available_amount, spent, operations, gift_cards, created_at, updated_at
FROM wallets
WHERE id = $1;

-- name: FindWalletByCustomerID :one
SELECT id, owner, available_amount, spent, operations, gift_cards, created_at, updated_at
FROM wallets
WHERE owner = $1;

-- name: FindWalletReadModelByID :one
SELECT
    w.id,
    w.owner,
    w.available_amount,
    w.spent,
    w.operations,
    w.gift_cards,
    w.created_at,
    w.updated_at,
    c.id AS customer_id,
    c.name,
    c.surname,
    c.email,
    c.phone,
    c.note
FROM wallets w
JOIN customers c ON c.id = w.owner
WHERE w.id = $1;

-- name: FindWalletReadModels :many
SELECT
    w.id,
    w.owner,
    w.available_amount,
    w.spent,
    w.operations,
    w.gift_cards,
    w.created_at,
    w.updated_at,
    c.id AS customer_id,
    c.name,
    c.surname,
    c.email,
    c.phone,
    c.note
FROM wallets w
JOIN customers c ON c.id = w.owner
ORDER BY w.created_at DESC;

-- name: SearchWalletReadModels :many
SELECT
    w.id,
    w.owner,
    w.available_amount,
    w.spent,
    w.operations,
    w.gift_cards,
    w.created_at,
    w.updated_at,
    c.id AS customer_id,
    c.name,
    c.surname,
    c.email,
    c.phone,
    c.note
FROM wallets w
JOIN customers c ON c.id = w.owner
WHERE c.search_text ILIKE '%' || $1 || '%' OR c.search_text % $1
ORDER BY w.created_at DESC;
