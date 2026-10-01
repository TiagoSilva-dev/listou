package curated

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

// Repository stores the curated catalog in Postgres.
type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const productColumns = `id, external_id, title, COALESCE(brand, ''), COALESCE(category, ''), COALESCE(emoji, ''),
	keywords, COALESCE(image_url, ''), active`

func scanProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(&p.ID, &p.ExternalID, &p.Title, &p.Brand, &p.Category, &p.Emoji, &p.Keywords, &p.ImageURL, &p.Active)
	return p, err
}

// List returns products newest first; onlyActive hides disabled ones (search path).
func (r *Repository) List(ctx context.Context, onlyActive bool) ([]Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+productColumns+` FROM curated_products
		WHERE ($1 = false OR active) ORDER BY created_at DESC, id DESC`, onlyActive)
	if err != nil {
		return nil, fmt.Errorf("curated: list products: %w", err)
	}
	defer rows.Close()
	var out []Product
	byID := map[uuid.UUID]int{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("curated: scan product: %w", err)
		}
		p.Offers = []Offer{}
		byID[p.ID] = len(out)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	orows, err := r.pool.Query(ctx, `SELECT product_id, merchant_code, url FROM curated_offers ORDER BY merchant_code`)
	if err != nil {
		return nil, fmt.Errorf("curated: list offers: %w", err)
	}
	defer orows.Close()
	for orows.Next() {
		var pid uuid.UUID
		var o Offer
		if err := orows.Scan(&pid, &o.Merchant, &o.URL); err != nil {
			return nil, fmt.Errorf("curated: scan offer: %w", err)
		}
		if i, ok := byID[pid]; ok {
			out[i].Offers = append(out[i].Offers, o)
		}
	}
	return out, orows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Product, error) {
	return get(ctx, r.pool, id)
}

func get(ctx context.Context, db database.DBTX, id uuid.UUID) (Product, error) {
	p, err := scanProduct(db.QueryRow(ctx, `SELECT `+productColumns+` FROM curated_products WHERE id = $1`, id))
	if err != nil {
		return Product{}, err
	}
	rows, err := db.Query(ctx, `SELECT merchant_code, url FROM curated_offers WHERE product_id = $1 ORDER BY merchant_code`, id)
	if err != nil {
		return Product{}, fmt.Errorf("curated: get offers: %w", err)
	}
	defer rows.Close()
	p.Offers = []Offer{}
	for rows.Next() {
		var o Offer
		if err := rows.Scan(&o.Merchant, &o.URL); err != nil {
			return Product{}, err
		}
		p.Offers = append(p.Offers, o)
	}
	return p, rows.Err()
}

// Merchants lists active stores whose enabled provider is CURATED.
func (r *Repository) Merchants(ctx context.Context) ([]Merchant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.code, m.name FROM merchants m
		JOIN affiliate_providers p ON p.merchant_id = m.id AND p.enabled AND p.adapter = $1
		WHERE m.active ORDER BY m.name`, Adapter)
	if err != nil {
		return nil, fmt.Errorf("curated: merchants: %w", err)
	}
	defer rows.Close()
	out := []Merchant{}
	for rows.Next() {
		var m Merchant
		if err := rows.Scan(&m.Code, &m.Name); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) Create(ctx context.Context, in Input) (uuid.UUID, error) {
	id := ids.New()
	err := database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO curated_products (id, external_id, title, brand, category, emoji, keywords, image_url, active)
			VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, NULLIF($8, ''), $9)`,
			id, "CUR-"+id.String(), in.Title, in.Brand, in.Category, EmojiFor(in.Category), in.Keywords, in.ImageURL, in.Active); err != nil {
			return fmt.Errorf("curated: insert product: %w", err)
		}
		return replaceOffers(ctx, tx, id, in.Offers)
	})
	return id, err
}

// Update replaces the product's fields and offers; false means it does not exist.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, in Input) (bool, error) {
	found := true
	err := database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE curated_products SET title = $2, brand = NULLIF($3, ''), category = NULLIF($4, ''), emoji = $5,
				keywords = $6, image_url = NULLIF($7, ''), active = $8, updated_at = now()
			WHERE id = $1`,
			id, in.Title, in.Brand, in.Category, EmojiFor(in.Category), in.Keywords, in.ImageURL, in.Active)
		if err != nil {
			return fmt.Errorf("curated: update product: %w", err)
		}
		if tag.RowsAffected() == 0 {
			found = false
			return nil
		}
		return replaceOffers(ctx, tx, id, in.Offers)
	})
	return found, err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM curated_products WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("curated: delete: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func replaceOffers(ctx context.Context, tx pgx.Tx, productID uuid.UUID, offers []Offer) error {
	if _, err := tx.Exec(ctx, `DELETE FROM curated_offers WHERE product_id = $1`, productID); err != nil {
		return fmt.Errorf("curated: clear offers: %w", err)
	}
	for _, o := range offers {
		if _, err := tx.Exec(ctx, `INSERT INTO curated_offers (id, product_id, merchant_code, url) VALUES ($1, $2, $3, $4)`,
			ids.New(), productID, o.Merchant, o.URL); err != nil {
			return fmt.Errorf("curated: insert offer: %w", err)
		}
	}
	return nil
}
