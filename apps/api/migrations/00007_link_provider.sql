-- +goose Up
-- Merchant and provider for products added by pasting a link. The merchant
-- name shown to users comes from each offer's store name (see product_offers.metadata).
INSERT INTO merchants (id, code, name, website_url) VALUES
    ('0192f000-0000-7000-8000-000000000004', 'LINK', 'Link da loja', NULL);

INSERT INTO affiliate_providers (id, code, merchant_id, adapter, name, enabled) VALUES
    ('0192f000-0000-7000-8000-000000000104', 'LINK_PASTED', '0192f000-0000-7000-8000-000000000004', 'LINK', 'Link colado pelo usuário', true);

-- +goose Down
DELETE FROM product_offers WHERE merchant_id = '0192f000-0000-7000-8000-000000000004';
DELETE FROM affiliate_providers WHERE code = 'LINK_PASTED';
DELETE FROM merchants WHERE code = 'LINK';
