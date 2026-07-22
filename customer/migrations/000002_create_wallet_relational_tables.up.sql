CREATE TABLE IF NOT EXISTS wallet_credit_lots (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets (id),
    initial_amount DOUBLE PRECISION NOT NULL,
    remaining_amount DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wallet_credit_lots_wallet_id
    ON wallet_credit_lots (wallet_id);

CREATE INDEX IF NOT EXISTS idx_wallet_credit_lots_expires_at
    ON wallet_credit_lots (expires_at);

CREATE TABLE IF NOT EXISTS wallet_operations (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets (id),
    credit_lot_id UUID NULL REFERENCES wallet_credit_lots (id),
    type TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_wallet_operations_wallet_id
    ON wallet_operations (wallet_id);

CREATE INDEX IF NOT EXISTS idx_wallet_operations_credit_lot_id
    ON wallet_operations (credit_lot_id);

CREATE INDEX IF NOT EXISTS idx_wallet_operations_occurred_at
    ON wallet_operations (occurred_at);
