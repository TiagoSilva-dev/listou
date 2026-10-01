-- +goose Up
-- Reference data required in every environment. Merchants are real
-- marketplaces; their providers point at the MOCK adapter until a real
-- adapter is implemented (see docs/affiliate-providers.md).
INSERT INTO merchants (id, code, name, website_url) VALUES
    ('0192f000-0000-7000-8000-000000000001', 'AMAZON', 'Amazon', NULL),
    ('0192f000-0000-7000-8000-000000000002', 'MERCADO_LIVRE', 'Mercado Livre', NULL),
    ('0192f000-0000-7000-8000-000000000003', 'SHOPEE', 'Shopee', NULL);

INSERT INTO affiliate_providers (id, code, merchant_id, adapter, name, enabled) VALUES
    ('0192f000-0000-7000-8000-000000000101', 'AMAZON_MOCK', '0192f000-0000-7000-8000-000000000001', 'MOCK', 'Amazon (demonstração)', true),
    ('0192f000-0000-7000-8000-000000000102', 'MERCADO_LIVRE_MOCK', '0192f000-0000-7000-8000-000000000002', 'MOCK', 'Mercado Livre (demonstração)', true),
    ('0192f000-0000-7000-8000-000000000103', 'SHOPEE_MOCK', '0192f000-0000-7000-8000-000000000003', 'MOCK', 'Shopee (demonstração)', true);

-- +goose Down
DELETE FROM affiliate_providers WHERE code IN ('AMAZON_MOCK', 'MERCADO_LIVRE_MOCK', 'SHOPEE_MOCK');
DELETE FROM merchants WHERE code IN ('AMAZON', 'MERCADO_LIVRE', 'SHOPEE');
