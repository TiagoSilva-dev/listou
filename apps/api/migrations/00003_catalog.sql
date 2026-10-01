-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE merchants (
    id          UUID PRIMARY KEY,
    code        TEXT NOT NULL CHECK (code ~ '^[A-Z0-9_]+$'),
    name        TEXT NOT NULL,
    website_url TEXT,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT merchants_code_key UNIQUE (code)
);

-- Which adapter builds outbound links for a merchant. Only non-secret
-- settings live in config; credentials come from the environment.
CREATE TABLE affiliate_providers (
    id          UUID PRIMARY KEY,
    code        TEXT NOT NULL CHECK (code ~ '^[A-Z0-9_]+$'),
    merchant_id UUID NOT NULL REFERENCES merchants (id) ON DELETE CASCADE,
    -- Which Go adapter implements this provider (MOCK, AMAZON, ...).
    adapter     TEXT NOT NULL CHECK (adapter ~ '^[A-Z0-9_]+$'),
    name        TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    config      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT affiliate_providers_code_key UNIQUE (code)
);
CREATE UNIQUE INDEX affiliate_providers_one_enabled_per_merchant ON affiliate_providers (merchant_id) WHERE enabled;

CREATE TABLE products (
    id              UUID PRIMARY KEY,
    canonical_title TEXT NOT NULL,
    brand           TEXT,
    description     TEXT,
    gtin            TEXT,
    image_url       TEXT,
    category        TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX products_gtin_key ON products (gtin) WHERE gtin IS NOT NULL;
CREATE INDEX products_title_trgm_idx ON products USING gin (canonical_title gin_trgm_ops);

CREATE TABLE product_offers (
    id                   UUID PRIMARY KEY,
    product_id           UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    merchant_id          UUID NOT NULL REFERENCES merchants (id),
    external_product_id  TEXT NOT NULL,
    title                TEXT NOT NULL,
    price_cents          BIGINT CHECK (price_cents >= 0),
    original_price_cents BIGINT CHECK (original_price_cents >= 0),
    currency             CHAR(3) NOT NULL DEFAULT 'BRL',
    availability         TEXT NOT NULL DEFAULT 'UNKNOWN' CHECK (availability IN ('IN_STOCK', 'OUT_OF_STOCK', 'UNKNOWN')),
    product_url          TEXT NOT NULL,
    image_url            TEXT,
    last_synced_at       TIMESTAMPTZ,
    metadata             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT product_offers_merchant_external_key UNIQUE (merchant_id, external_product_id)
);
CREATE INDEX product_offers_product_idx ON product_offers (product_id);

-- +goose Down
DROP TABLE product_offers;
DROP TABLE products;
DROP TABLE affiliate_providers;
DROP TABLE merchants;
