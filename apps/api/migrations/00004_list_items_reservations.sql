-- +goose Up
CREATE TABLE list_items (
    id                   UUID PRIMARY KEY,
    gift_list_id         UUID NOT NULL REFERENCES gift_lists (id) ON DELETE CASCADE,
    category_id          UUID REFERENCES categories (id) ON DELETE SET NULL,
    product_id           UUID REFERENCES products (id) ON DELETE SET NULL,
    title                TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 160),
    description          TEXT,
    notes                TEXT,
    image_url            TEXT,
    external_url         TEXT,
    price_reference_cents BIGINT CHECK (price_reference_cents >= 0),
    currency             CHAR(3) NOT NULL DEFAULT 'BRL',
    priority             TEXT NOT NULL DEFAULT 'MEDIUM' CHECK (priority IN ('HIGH', 'MEDIUM', 'LOW')),
    desired_quantity     INT NOT NULL DEFAULT 1 CHECK (desired_quantity BETWEEN 1 AND 999),
    purchased_quantity   INT NOT NULL DEFAULT 0 CHECK (purchased_quantity >= 0),
    -- Denormalized sum of ACTIVE reservations, maintained transactionally by
    -- the reservations module. The check below is the last line of defence
    -- against double-booking under concurrency.
    reserved_quantity    INT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    position             INT NOT NULL DEFAULT 0,
    archived_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT list_items_quantity_capacity CHECK (purchased_quantity + reserved_quantity <= desired_quantity)
);
CREATE INDEX list_items_list_idx ON list_items (gift_list_id, position, created_at) WHERE archived_at IS NULL;
CREATE INDEX list_items_category_idx ON list_items (category_id);

CREATE TABLE reservations (
    id            UUID PRIMARY KEY,
    list_item_id  UUID NOT NULL REFERENCES list_items (id) ON DELETE CASCADE,
    kind          TEXT NOT NULL CHECK (kind IN ('RESERVATION', 'PURCHASE')),
    quantity      INT NOT NULL DEFAULT 1 CHECK (quantity BETWEEN 1 AND 99),
    guest_name    TEXT NOT NULL CHECK (char_length(guest_name) BETWEEN 1 AND 80),
    guest_contact TEXT,
    message       TEXT,
    status        TEXT NOT NULL CHECK (status IN ('ACTIVE', 'CANCELLED', 'EXPIRED', 'CONFIRMED')),
    -- SHA-256 of the guest's management token; lets a guest cancel without an account.
    token_hash    BYTEA NOT NULL,
    expires_at    TIMESTAMPTZ,
    cancelled_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT reservations_token_hash_key UNIQUE (token_hash)
);
CREATE INDEX reservations_item_idx ON reservations (list_item_id, created_at DESC);
CREATE INDEX reservations_expiry_idx ON reservations (expires_at) WHERE status = 'ACTIVE';

-- +goose Down
DROP TABLE reservations;
DROP TABLE list_items;
