// Package access resolves whether a user may manage an event and the
// resources under it. IDs coming from clients are never trusted: every
// lookup re-derives the owning event from the database.
package access

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type Role string

const (
	Owner   Role = "OWNER"
	CoOwner Role = "CO_OWNER"
)

// ErrEventNotFound is returned for missing events AND events the user cannot
// manage, so existence is not leaked.
var ErrEventNotFound = httpx.NotFound("EVENT_NOT_FOUND", "Evento não encontrado.")

const roleSQL = `
	CASE WHEN e.owner_id = $2 THEN 'OWNER'
	     ELSE (SELECT m.role FROM event_members m WHERE m.event_id = e.id AND m.user_id = $2 AND m.role = 'CO_OWNER')
	END`

func scanRole(ctx context.Context, db database.DBTX, query string, id, userID uuid.UUID, notFound *httpx.Error, dest ...any) (Role, error) {
	var role *string
	args := append([]any{&role}, dest...)
	if err := db.QueryRow(ctx, query, id, userID).Scan(args...); err != nil {
		if database.IsNoRows(err) {
			return "", notFound
		}
		return "", fmt.Errorf("access: %w", err)
	}
	if role == nil {
		return "", notFound
	}
	return Role(*role), nil
}

// Event checks the user's role on an event.
func Event(ctx context.Context, db database.DBTX, eventID, userID uuid.UUID) (Role, error) {
	return scanRole(ctx, db,
		`SELECT `+roleSQL+` FROM events e WHERE e.id = $1 AND e.deleted_at IS NULL`,
		eventID, userID, ErrEventNotFound)
}

var ErrListNotFound = httpx.NotFound("LIST_NOT_FOUND", "Lista não encontrada.")

// List resolves the event of a gift list the user can manage.
func List(ctx context.Context, db database.DBTX, listID, userID uuid.UUID) (eventID uuid.UUID, err error) {
	_, err = scanRole(ctx, db, `
		SELECT `+roleSQL+`, e.id FROM gift_lists l JOIN events e ON e.id = l.event_id
		WHERE l.id = $1 AND e.deleted_at IS NULL`, listID, userID, ErrListNotFound, &eventID)
	return eventID, err
}

var ErrItemNotFound = httpx.NotFound("ITEM_NOT_FOUND", "Item não encontrado.")

// Item resolves the list and event of an item the user can manage.
func Item(ctx context.Context, db database.DBTX, itemID, userID uuid.UUID) (eventID, listID uuid.UUID, err error) {
	_, err = scanRole(ctx, db, `
		SELECT `+roleSQL+`, e.id, l.id FROM list_items i
		JOIN gift_lists l ON l.id = i.gift_list_id
		JOIN events e ON e.id = l.event_id
		WHERE i.id = $1 AND e.deleted_at IS NULL`, itemID, userID, ErrItemNotFound, &eventID, &listID)
	return eventID, listID, err
}

var ErrCategoryNotFound = httpx.NotFound("CATEGORY_NOT_FOUND", "Categoria não encontrada.")

func Category(ctx context.Context, db database.DBTX, categoryID, userID uuid.UUID) (listID uuid.UUID, err error) {
	_, err = scanRole(ctx, db, `
		SELECT `+roleSQL+`, l.id FROM categories c
		JOIN gift_lists l ON l.id = c.gift_list_id
		JOIN events e ON e.id = l.event_id
		WHERE c.id = $1 AND e.deleted_at IS NULL`, categoryID, userID, ErrCategoryNotFound, &listID)
	return listID, err
}
