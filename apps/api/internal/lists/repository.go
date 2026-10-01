package lists

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
)

type Repository struct{}

func (Repository) InsertList(ctx context.Context, db database.DBTX, l GiftList) error {
	_, err := db.Exec(ctx, `
		INSERT INTO gift_lists (id, event_id, title, is_primary, allow_reservations) VALUES ($1, $2, $3, true, true)`,
		l.ID, l.EventID, l.Title)
	if err != nil {
		return fmt.Errorf("insert list: %w", err)
	}
	return nil
}

const listColumns = `id, event_id, title, description, allow_reservations, allow_group_contributions`

func scanList(row interface{ Scan(...any) error }) (GiftList, error) {
	var l GiftList
	err := row.Scan(&l.ID, &l.EventID, &l.Title, &l.Description, &l.AllowReservations, &l.AllowGroupContributions)
	return l, err
}

func (Repository) PrimaryList(ctx context.Context, db database.DBTX, eventID uuid.UUID) (GiftList, error) {
	return scanList(db.QueryRow(ctx, `SELECT `+listColumns+` FROM gift_lists WHERE event_id = $1 AND is_primary`, eventID))
}

func (Repository) List(ctx context.Context, db database.DBTX, id uuid.UUID) (GiftList, error) {
	return scanList(db.QueryRow(ctx, `SELECT `+listColumns+` FROM gift_lists WHERE id = $1`, id))
}

func (Repository) UpdateList(ctx context.Context, db database.DBTX, l GiftList) error {
	_, err := db.Exec(ctx, `UPDATE gift_lists SET title = $2, description = $3, allow_reservations = $4, updated_at = now() WHERE id = $1`,
		l.ID, l.Title, l.Description, l.AllowReservations)
	return err
}

func (Repository) InsertCategory(ctx context.Context, db database.DBTX, listID uuid.UUID, c Category) error {
	_, err := db.Exec(ctx, `INSERT INTO categories (id, gift_list_id, name, emoji, position) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, listID, c.Name, c.Emoji, c.Position)
	return err
}

func (Repository) Categories(ctx context.Context, db database.DBTX, listID uuid.UUID) ([]Category, error) {
	rows, err := db.Query(ctx, `SELECT id, name, emoji, position FROM categories WHERE gift_list_id = $1 ORDER BY position, created_at`, listID)
	if err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}
	defer rows.Close()
	out := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Emoji, &c.Position); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (Repository) Category(ctx context.Context, db database.DBTX, id uuid.UUID) (Category, error) {
	var c Category
	err := db.QueryRow(ctx, `SELECT id, name, emoji, position FROM categories WHERE id = $1`, id).Scan(&c.ID, &c.Name, &c.Emoji, &c.Position)
	return c, err
}

func (Repository) UpdateCategory(ctx context.Context, db database.DBTX, c Category) error {
	_, err := db.Exec(ctx, `UPDATE categories SET name = $2, emoji = $3, position = $4, updated_at = now() WHERE id = $1`,
		c.ID, c.Name, c.Emoji, c.Position)
	return err
}

func (Repository) DeleteCategory(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	_, err := db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (Repository) CategoryInList(ctx context.Context, db database.DBTX, categoryID, listID uuid.UUID) (bool, error) {
	var ok bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1 AND gift_list_id = $2)`, categoryID, listID).Scan(&ok)
	return ok, err
}

func (Repository) NextCategoryPosition(ctx context.Context, db database.DBTX, listID uuid.UUID) (int, error) {
	var n int
	err := db.QueryRow(ctx, `SELECT COALESCE(MAX(position) + 1, 0) FROM categories WHERE gift_list_id = $1`, listID).Scan(&n)
	return n, err
}

const itemColumns = `id, gift_list_id, category_id, product_id, title, description, notes, image_url, emoji,
	external_url, price_reference_cents, currency, priority, desired_quantity, purchased_quantity,
	reserved_quantity, position, archived_at, created_at`

func scanItem(row interface{ Scan(...any) error }) (Item, error) {
	var it Item
	err := row.Scan(&it.ID, &it.ListID, &it.CategoryID, &it.ProductID, &it.Title, &it.Description, &it.Notes,
		&it.ImageURL, &it.Emoji, &it.ExternalURL, &it.PriceReferenceCents, &it.Currency, &it.Priority,
		&it.DesiredQuantity, &it.PurchasedQuantity, &it.ReservedQuantity, &it.Position, &it.ArchivedAt, &it.CreatedAt)
	return it, err
}

func (Repository) InsertItem(ctx context.Context, db database.DBTX, it Item) error {
	_, err := db.Exec(ctx, `
		INSERT INTO list_items (id, gift_list_id, category_id, product_id, title, description, notes, image_url, emoji,
			external_url, price_reference_cents, currency, priority, desired_quantity, position)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		it.ID, it.ListID, it.CategoryID, it.ProductID, it.Title, it.Description, it.Notes, it.ImageURL, it.Emoji,
		it.ExternalURL, it.PriceReferenceCents, it.Currency, it.Priority, it.DesiredQuantity, it.Position)
	if err != nil {
		return fmt.Errorf("insert item: %w", err)
	}
	return nil
}

func (Repository) Item(ctx context.Context, db database.DBTX, id uuid.UUID, forUpdate bool) (Item, error) {
	q := `SELECT ` + itemColumns + ` FROM list_items WHERE id = $1`
	if forUpdate {
		q += ` FOR UPDATE`
	}
	return scanItem(db.QueryRow(ctx, q, id))
}

func (Repository) Items(ctx context.Context, db database.DBTX, listID uuid.UUID) ([]Item, error) {
	rows, err := db.Query(ctx, `SELECT `+itemColumns+` FROM list_items
		WHERE gift_list_id = $1 AND archived_at IS NULL ORDER BY position, created_at`, listID)
	if err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (Repository) UpdateItem(ctx context.Context, db database.DBTX, it Item) error {
	_, err := db.Exec(ctx, `
		UPDATE list_items SET category_id = $2, product_id = $3, title = $4, description = $5, notes = $6,
			image_url = $7, emoji = $8, external_url = $9, price_reference_cents = $10, priority = $11,
			desired_quantity = $12, purchased_quantity = $13, position = $14, updated_at = now()
		WHERE id = $1`,
		it.ID, it.CategoryID, it.ProductID, it.Title, it.Description, it.Notes, it.ImageURL, it.Emoji,
		it.ExternalURL, it.PriceReferenceCents, it.Priority, it.DesiredQuantity, it.PurchasedQuantity, it.Position)
	return err
}

func (Repository) NextItemPosition(ctx context.Context, db database.DBTX, listID uuid.UUID) (int, error) {
	var n int
	err := db.QueryRow(ctx, `SELECT COALESCE(MAX(position) + 1, 0) FROM list_items WHERE gift_list_id = $1`, listID).Scan(&n)
	return n, err
}

func (Repository) HasReservations(ctx context.Context, db database.DBTX, itemID uuid.UUID) (bool, error) {
	var ok bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM reservations WHERE list_item_id = $1)`, itemID).Scan(&ok)
	return ok, err
}

func (Repository) ArchiveItem(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE list_items SET archived_at = now(), updated_at = now() WHERE id = $1`, id)
	return err
}

func (Repository) DeleteItem(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	_, err := db.Exec(ctx, `DELETE FROM list_items WHERE id = $1`, id)
	return err
}
