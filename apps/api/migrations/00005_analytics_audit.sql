-- +goose Up
-- Outbound marketplace clicks recorded by GET /go/{offerId}. Privacy-first:
-- a random first-party visitor id, referrer host only, coarse device class.
CREATE TABLE click_events (
    id           UUID PRIMARY KEY,
    offer_id     UUID NOT NULL REFERENCES product_offers (id) ON DELETE CASCADE,
    merchant_id  UUID NOT NULL REFERENCES merchants (id),
    event_id     UUID REFERENCES events (id) ON DELETE CASCADE,
    list_item_id UUID REFERENCES list_items (id) ON DELETE SET NULL,
    visitor_id   UUID,
    referrer_host TEXT,
    utm_source   TEXT,
    utm_medium   TEXT,
    utm_campaign TEXT,
    device_class TEXT CHECK (device_class IN ('MOBILE', 'TABLET', 'DESKTOP', 'UNKNOWN')),
    price_cents  BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX click_events_event_idx ON click_events (event_id, created_at DESC);
CREATE INDEX click_events_offer_idx ON click_events (offer_id, created_at DESC);

-- Internal product analytics (LIST_VIEWED, ITEM_RESERVED, SHARE_CREATED...).
CREATE TABLE analytics_events (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL CHECK (name ~ '^[A-Z_]+$'),
    event_id     UUID REFERENCES events (id) ON DELETE CASCADE,
    list_item_id UUID REFERENCES list_items (id) ON DELETE SET NULL,
    user_id      UUID REFERENCES users (id) ON DELETE SET NULL,
    visitor_id   UUID,
    properties   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX analytics_events_event_idx ON analytics_events (event_id, name, created_at DESC);
CREATE INDEX analytics_events_name_idx ON analytics_events (name, created_at DESC);

CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY,
    actor_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    action        TEXT NOT NULL,
    entity_type   TEXT NOT NULL,
    entity_id     UUID,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id, created_at DESC);

-- +goose Down
DROP TABLE audit_logs;
DROP TABLE analytics_events;
DROP TABLE click_events;
