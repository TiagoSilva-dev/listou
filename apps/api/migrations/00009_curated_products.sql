-- +goose Up
-- The hand-curated catalog moves from an embedded JSON file to tables so the
-- admin panel can maintain it. external_id is the stable id the catalog layer
-- stores on materialized products (never rename it).
CREATE TABLE curated_products (
    id          UUID PRIMARY KEY,
    external_id TEXT NOT NULL CHECK (external_id <> ''),
    title       TEXT NOT NULL CHECK (title <> ''),
    brand       TEXT,
    category    TEXT,
    emoji       TEXT,
    keywords    TEXT NOT NULL DEFAULT '',
    image_url   TEXT,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT curated_products_external_id_key UNIQUE (external_id)
);

CREATE TABLE curated_offers (
    id            UUID PRIMARY KEY,
    product_id    UUID NOT NULL REFERENCES curated_products (id) ON DELETE CASCADE,
    merchant_code TEXT NOT NULL CHECK (merchant_code ~ '^[A-Z0-9_]+$'),
    url           TEXT NOT NULL CHECK (url ~* '^https?://'),
    CONSTRAINT curated_offers_one_per_merchant UNIQUE (product_id, merchant_code)
);

INSERT INTO curated_products (id, external_id, title, brand, category, emoji, keywords, image_url) VALUES
    ('0192f000-0000-7000-8000-000000000201', 'ML-AIRFRYER-PHILCO-11L', 'Fritadeira Air Fryer Oven 11 Litros 1800w 8 Funções Philco PAF11B', 'Philco', 'Cozinha', '🍟', 'air fryer airfryer fritadeira sem oleo oven forno 11l', 'https://http2.mlstatic.com/D_NQ_NP_633098-MLA92276027266_092025-O.webp'),
    ('0192f000-0000-7000-8000-000000000202', 'ML-PANELAS-GAMMA-VIENNA-5', 'Jogo De Panelas Cerâmica Antiaderente Gamma Vienna 5 Peças Preto', 'Gamma', 'Cozinha', '🍲', 'jogo de panelas conjunto panela ceramica antiaderente 5 pecas', 'https://http2.mlstatic.com/D_NQ_NP_624797-MLA114048065297_072026-O.webp'),
    ('0192f000-0000-7000-8000-000000000203', 'ML-FAQUEIRO-BAMBU-24', 'Faqueiro 24 Peças Inox Bambu Mesa Elegante Talheres Hold On', 'Hold On', 'Cozinha', '🍴', 'faqueiro talheres inox jogo 24 pecas bambu mesa', 'https://http2.mlstatic.com/D_NQ_NP_710382-MLA114241185556_082026-O.webp');

INSERT INTO curated_offers (id, product_id, merchant_code, url) VALUES
    ('0192f000-0000-7000-8000-000000000211', '0192f000-0000-7000-8000-000000000201', 'MERCADO_LIVRE', 'https://meli.la/2Ne9wWJ'),
    ('0192f000-0000-7000-8000-000000000212', '0192f000-0000-7000-8000-000000000202', 'MERCADO_LIVRE', 'https://meli.la/1pZVy8u'),
    ('0192f000-0000-7000-8000-000000000213', '0192f000-0000-7000-8000-000000000203', 'MERCADO_LIVRE', 'https://meli.la/1NVeWQU');

-- +goose Down
DROP TABLE curated_offers;
DROP TABLE curated_products;
