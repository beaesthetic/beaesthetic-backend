CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL DEFAULT '',
    email TEXT NULL,
    phone TEXT NULL,
    note TEXT NOT NULL DEFAULT '',
    search_text TEXT GENERATED ALWAYS AS (
        lower(coalesce(name, '') || ' ' || coalesce(surname, '') || ' ' || coalesce(email, '') || ' ' || coalesce(phone, '') || ' ' || coalesce(note, ''))
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_customers_search_text_trgm ON customers USING GIN (search_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers (phone);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers (name, surname, id);

CREATE TABLE IF NOT EXISTS deleted_customers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL DEFAULT '',
    email TEXT NULL,
    phone TEXT NULL,
    note TEXT NOT NULL DEFAULT '',
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fidelity_cards (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    solarium_purchases INTEGER NOT NULL DEFAULT 0,
    vouchers JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fidelity_cards_customer_id ON fidelity_cards (customer_id);
CREATE INDEX IF NOT EXISTS idx_fidelity_cards_vouchers ON fidelity_cards USING GIN (vouchers);

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    available_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    spent DOUBLE PRECISION NOT NULL DEFAULT 0,
    operations JSONB NOT NULL DEFAULT '[]'::jsonb,
    gift_cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_owner ON wallets (owner);