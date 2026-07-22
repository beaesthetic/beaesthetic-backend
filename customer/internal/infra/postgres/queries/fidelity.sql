-- name: SaveFidelityCard :exec
INSERT INTO fidelity_cards (id, customer_id, solarium_purchases, vouchers, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (id) DO UPDATE SET
    customer_id = $2,
    solarium_purchases = $3,
    vouchers = $4,
    updated_at = now();

-- name: FindFidelityCards :many
SELECT id, customer_id, solarium_purchases, vouchers
FROM fidelity_cards
ORDER BY created_at DESC;

-- name: FindFidelityCardByID :one
SELECT id, customer_id, solarium_purchases, vouchers
FROM fidelity_cards
WHERE id = $1;

-- name: FindFidelityCardsByCustomerID :many
SELECT id, customer_id, solarium_purchases, vouchers
FROM fidelity_cards
WHERE customer_id = $1;

-- name: FindFidelityCardByVoucherID :one
SELECT id, customer_id, solarium_purchases, vouchers
FROM fidelity_cards
WHERE vouchers @> $1::jsonb
LIMIT 1;
