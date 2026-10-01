package events

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
)

type Repository struct{}

const eventColumns = `
	e.id, e.owner_id, e.type, e.title, e.slug, e.description, e.host_names, e.event_date, e.location,
	e.cover_image_url, e.avatar_url, e.theme, e.visibility, e.status, e.surprise_mode,
	e.show_reserver_names, e.published_at, l.id, e.created_at, e.updated_at`

const eventFrom = ` FROM events e JOIN gift_lists l ON l.event_id = e.id AND l.is_primary `

func scanEvent(row interface{ Scan(...any) error }) (Event, error) {
	var e Event
	err := row.Scan(&e.ID, &e.OwnerID, &e.Type, &e.Title, &e.Slug, &e.Description, &e.HostNames, &e.EventDate,
		&e.Location, &e.CoverImageURL, &e.AvatarURL, &e.Theme, &e.Visibility, &e.Status, &e.SurpriseMode,
		&e.ShowReserverNames, &e.PublishedAt, &e.ListID, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (Repository) Insert(ctx context.Context, db database.DBTX, e Event) error {
	_, err := db.Exec(ctx, `
		INSERT INTO events (id, owner_id, type, title, slug, description, host_names, event_date, location,
			cover_image_url, theme, visibility, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		e.ID, e.OwnerID, e.Type, e.Title, e.Slug, e.Description, e.HostNames, e.EventDate, e.Location,
		e.CoverImageURL, e.Theme, e.Visibility, e.Status)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (Repository) Get(ctx context.Context, db database.DBTX, id uuid.UUID) (Event, error) {
	return scanEvent(db.QueryRow(ctx, `SELECT `+eventColumns+eventFrom+`WHERE e.id = $1 AND e.deleted_at IS NULL`, id))
}

// ListForUser returns events the user owns or co-owns, newest first.
func (Repository) ListForUser(ctx context.Context, db database.DBTX, userID uuid.UUID) ([]Event, error) {
	rows, err := db.Query(ctx, `SELECT `+eventColumns+eventFrom+`
		WHERE e.deleted_at IS NULL AND (e.owner_id = $1 OR EXISTS (
			SELECT 1 FROM event_members m WHERE m.event_id = e.id AND m.user_id = $1 AND m.role = 'CO_OWNER'))
		ORDER BY e.created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (Repository) Update(ctx context.Context, db database.DBTX, e Event) error {
	_, err := db.Exec(ctx, `
		UPDATE events SET title = $2, slug = $3, description = $4, host_names = $5, event_date = $6,
			location = $7, cover_image_url = $8, theme = $9, visibility = $10, status = $11,
			surprise_mode = $12, show_reserver_names = $13, published_at = $14, updated_at = now()
		WHERE id = $1`,
		e.ID, e.Title, e.Slug, e.Description, e.HostNames, e.EventDate, e.Location, e.CoverImageURL,
		e.Theme, e.Visibility, e.Status, e.SurpriseMode, e.ShowReserverNames, e.PublishedAt)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (Repository) SoftDelete(ctx context.Context, db database.DBTX, id uuid.UUID, at time.Time) error {
	_, err := db.Exec(ctx, `UPDATE events SET deleted_at = $2, status = 'ARCHIVED', updated_at = now() WHERE id = $1`, id, at)
	return err
}

func (Repository) SlugExists(ctx context.Context, db database.DBTX, slug string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM events WHERE slug = $1)`, slug).Scan(&exists)
	return exists, err
}
