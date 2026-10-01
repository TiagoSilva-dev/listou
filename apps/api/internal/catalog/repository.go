package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Repository struct{}

type enabledMerchant struct {
	ID      uuid.UUID
	Code    string
	Name    string
	Adapter string
}

func (Repository) EnabledMerchants(ctx context.Context, db database.DBTX) ([]enabledMerchant, error) {
	rows, err := db.Query(ctx, `
		SELECT m.id, m.code, m.name, p.adapter FROM affiliate_providers p
		JOIN merchants m ON m.id = p.merchant_id
		WHERE p.enabled AND m.active ORDER BY m.name`)
	if err != nil {
		return nil, fmt.Errorf("enabled merchants: %w", err)
	}
	defer rows.Close()
	var out []enabledMerchant
	for rows.Next() {
		var m enabledMerchant
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Adapter); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpsertProduct stores a provider product and its offers for the enabled
// merchants, reusing an existing product when any offer is already known.
func (Repository) UpsertProduct(ctx context.Context, db database.DBTX, pd affiliate.ProductData, merchants map[string]uuid.UUID, now time.Time) (uuid.UUID, error) {
	var productID uuid.UUID
	for _, o := range pd.Offers {
		mID, ok := merchants[o.MerchantCode]
		if !ok {
			continue
		}
		err := db.QueryRow(ctx, `SELECT product_id FROM product_offers WHERE merchant_id = $1 AND external_product_id = $2`,
			mID, o.ExternalID).Scan(&productID)
		if err == nil {
			break
		}
		if !database.IsNoRows(err) {
			return uuid.Nil, fmt.Errorf("find offer: %w", err)
		}
	}
	meta, _ := json.Marshal(map[string]any{"emoji": pd.Emoji, "demo": pd.Demo, "externalId": pd.ExternalID})
	if productID == uuid.Nil {
		productID = ids.New()
		if _, err := db.Exec(ctx, `
			INSERT INTO products (id, canonical_title, brand, description, gtin, image_url, category, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			productID, pd.Title, pd.Brand, pd.Description, pd.GTIN, pd.ImageURL, pd.Category, meta); err != nil {
			return uuid.Nil, fmt.Errorf("insert product: %w", err)
		}
	} else if _, err := db.Exec(ctx, `
		UPDATE products SET canonical_title = $2, brand = $3, description = $4, image_url = $5, category = $6,
			metadata = $7, updated_at = now() WHERE id = $1`,
		productID, pd.Title, pd.Brand, pd.Description, pd.ImageURL, pd.Category, meta); err != nil {
		return uuid.Nil, fmt.Errorf("update product: %w", err)
	}
	for _, o := range pd.Offers {
		mID, ok := merchants[o.MerchantCode]
		if !ok {
			continue
		}
		offerMeta, _ := json.Marshal(map[string]any{"demo": pd.Demo, "storeName": o.StoreName})
		if _, err := db.Exec(ctx, `
			INSERT INTO product_offers (id, product_id, merchant_id, external_product_id, title, price_cents,
				original_price_cents, currency, availability, product_url, image_url, last_synced_at, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (merchant_id, external_product_id) DO UPDATE SET
				title = EXCLUDED.title, price_cents = EXCLUDED.price_cents,
				original_price_cents = EXCLUDED.original_price_cents, availability = EXCLUDED.availability,
				product_url = EXCLUDED.product_url, image_url = EXCLUDED.image_url,
				last_synced_at = EXCLUDED.last_synced_at, metadata = EXCLUDED.metadata, updated_at = now()`,
			ids.New(), productID, mID, o.ExternalID, o.Title, o.PriceCents, o.OriginalPriceCents, o.Currency,
			string(o.Availability), o.ProductURL, o.ImageURL, now, offerMeta); err != nil {
			return uuid.Nil, fmt.Errorf("upsert offer: %w", err)
		}
	}
	return productID, nil
}

const productColumns = `p.id, p.canonical_title, p.brand, p.description, p.image_url, p.category,
	p.metadata->>'emoji', COALESCE((p.metadata->>'demo')::boolean, false)`

func scanProduct(row interface{ Scan(...any) error }) (Product, error) {
	var p Product
	err := row.Scan(&p.ID, &p.CanonicalTitle, &p.Brand, &p.Description, &p.ImageURL, &p.Category, &p.Emoji, &p.Demo)
	return p, err
}

func (Repository) Product(ctx context.Context, db database.DBTX, id uuid.UUID) (Product, error) {
	return scanProduct(db.QueryRow(ctx, `SELECT `+productColumns+` FROM products p WHERE p.id = $1`, id))
}

func (Repository) Products(ctx context.Context, db database.DBTX, idList []uuid.UUID) (map[uuid.UUID]Product, error) {
	out := map[uuid.UUID]Product{}
	if len(idList) == 0 {
		return out, nil
	}
	rows, err := db.Query(ctx, `SELECT `+productColumns+` FROM products p WHERE p.id = ANY($1)`, idList)
	if err != nil {
		return nil, fmt.Errorf("products: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out[p.ID] = p
	}
	return out, rows.Err()
}

// OffersForProducts returns offers from active merchants, cheapest first.
func (Repository) OffersForProducts(ctx context.Context, db database.DBTX, productIDs []uuid.UUID) (map[uuid.UUID][]Offer, error) {
	out := map[uuid.UUID][]Offer{}
	if len(productIDs) == 0 {
		return out, nil
	}
	rows, err := db.Query(ctx, `
		SELECT o.id, o.product_id, m.code, COALESCE(o.metadata->>'storeName', m.name), o.title, o.price_cents, o.original_price_cents, o.currency,
			o.availability, o.image_url, o.last_synced_at, COALESCE((o.metadata->>'demo')::boolean, false)
		FROM product_offers o JOIN merchants m ON m.id = o.merchant_id
		WHERE o.product_id = ANY($1) AND m.active
		ORDER BY o.product_id, o.availability = 'OUT_OF_STOCK', o.price_cents NULLS LAST`, productIDs)
	if err != nil {
		return nil, fmt.Errorf("offers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var o Offer
		if err := rows.Scan(&o.ID, &o.ProductID, &o.Merchant.Code, &o.Merchant.Name, &o.Title, &o.PriceCents,
			&o.OriginalPriceCents, &o.Currency, &o.Availability, &o.ImageURL, &o.LastSyncedAt, &o.Demo); err != nil {
			return nil, err
		}
		o.GoURL = "/go/" + o.ID.String()
		out[o.ProductID] = append(out[o.ProductID], o)
	}
	return out, rows.Err()
}
