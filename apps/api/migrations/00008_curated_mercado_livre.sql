-- +goose Up
-- Mercado Livre now serves the hand-curated catalog (affiliate links we
-- generate in the partner's panel) instead of the fictitious MOCK data.
UPDATE affiliate_providers
SET adapter = 'CURATED', code = 'MERCADO_LIVRE_CURATED', name = 'Mercado Livre (catálogo curado)', updated_at = now()
WHERE code = 'MERCADO_LIVRE_MOCK';

-- +goose Down
UPDATE affiliate_providers
SET adapter = 'MOCK', code = 'MERCADO_LIVRE_MOCK', name = 'Mercado Livre (demonstração)', updated_at = now()
WHERE code = 'MERCADO_LIVRE_CURATED';
