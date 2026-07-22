CREATE TABLE customers (
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

CREATE TABLE deleted_customers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL DEFAULT '',
    email TEXT NULL,
    phone TEXT NULL,
    note TEXT NOT NULL DEFAULT '',
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE fidelity_cards (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    solarium_purchases INTEGER NOT NULL DEFAULT 0,
    vouchers JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    available_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    spent DOUBLE PRECISION NOT NULL DEFAULT 0,
    operations JSONB NOT NULL DEFAULT '[]'::jsonb,
    gift_cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
