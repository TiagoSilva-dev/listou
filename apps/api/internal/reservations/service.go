package reservations

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/platform/audit"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

var (
	ErrItemNotFound       = httpx.NotFound("ITEM_NOT_FOUND", "Item não encontrado.")
	ErrReservationMissing = httpx.NotFound("RESERVATION_NOT_FOUND", "Reserva não encontrada.")
	ErrNotAvailable       = httpx.NewError(http.StatusConflict, "ITEM_NOT_AVAILABLE", "Este presente acabou de ser reservado por outra pessoa.")
	ErrReservationsOff    = httpx.NewError(http.StatusForbidden, "RESERVATIONS_DISABLED", "Esta lista não aceita reservas.")
	ErrNotActive          = httpx.NewError(http.StatusConflict, "RESERVATION_NOT_ACTIVE", "Esta reserva não está mais ativa.")
)

type Service struct {
	pool    *pgxpool.Pool
	ttl     time.Duration
	tracker *analytics.Recorder
	now     func() time.Time
}

func NewService(pool *pgxpool.Pool, ttl time.Duration, tracker *analytics.Recorder) *Service {
	return &Service{pool: pool, ttl: ttl, tracker: tracker, now: time.Now}
}

type itemLock struct {
	eventID      uuid.UUID
	slug         string
	allowed      bool
	published    bool
	private      bool
	desired      int
	purchased    int
	reserved     int
	archivedNull bool
}

// Create reserves or records a purchase. The item row is locked FOR UPDATE so
// concurrent guests are serialized; the CHECK constraint is the backstop.
func (s *Service) Create(ctx context.Context, slug string, itemID uuid.UUID, visitor *uuid.UUID, in CreateInput) (Reservation, error) {
	if err := in.Validate(); err != nil {
		return Reservation{}, err
	}
	var res Reservation
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var l itemLock
		err := tx.QueryRow(ctx, `
			SELECT e.id, e.slug, g.allow_reservations, e.status = 'PUBLISHED', e.visibility = 'PRIVATE',
				i.desired_quantity, i.purchased_quantity, i.reserved_quantity, i.archived_at IS NULL
			FROM list_items i
			JOIN gift_lists g ON g.id = i.gift_list_id
			JOIN events e ON e.id = g.event_id
			WHERE i.id = $1 AND e.slug = $2 AND e.deleted_at IS NULL
			FOR UPDATE OF i`, itemID, slug).Scan(&l.eventID, &l.slug, &l.allowed, &l.published, &l.private,
			&l.desired, &l.purchased, &l.reserved, &l.archivedNull)
		if err != nil {
			if database.IsNoRows(err) {
				return ErrItemNotFound
			}
			return fmt.Errorf("lock item: %w", err)
		}
		if !l.published || l.private || !l.archivedNull {
			return ErrItemNotFound
		}
		if !l.allowed {
			return ErrReservationsOff
		}
		// Expired reservations free capacity before we count it.
		freed, err := expireItem(ctx, tx, itemID, s.now())
		if err != nil {
			return err
		}
		l.reserved -= freed
		if l.desired-l.purchased-l.reserved < in.Quantity {
			return ErrNotAvailable
		}

		token, hash := ids.Token()
		res = Reservation{ID: ids.New(), ItemID: itemID, Kind: in.Kind, Quantity: in.Quantity, CreatedAt: s.now(), ManageToken: token}
		var expires *time.Time
		status := "ACTIVE"
		if in.Kind == "RESERVATION" {
			t := s.now().Add(s.ttl)
			expires = &t
		} else {
			status = "CONFIRMED"
		}
		res.Status, res.ExpiresAt = status, expires
		if _, err := tx.Exec(ctx, `
			INSERT INTO reservations (id, list_item_id, kind, quantity, guest_name, guest_contact, message, status, token_hash, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			res.ID, itemID, in.Kind, in.Quantity, in.GuestName, in.GuestContact, in.Message, status, hash, expires); err != nil {
			return fmt.Errorf("insert reservation: %w", err)
		}
		col := "reserved_quantity"
		if in.Kind == "PURCHASE" {
			col = "purchased_quantity"
		}
		if _, err := tx.Exec(ctx, `UPDATE list_items SET `+col+` = `+col+` + $2, updated_at = now() WHERE id = $1`, itemID, in.Quantity); err != nil {
			if database.IsCheckViolation(err, "list_items_quantity_capacity") {
				return ErrNotAvailable
			}
			return fmt.Errorf("update item: %w", err)
		}
		return nil
	})
	if err != nil {
		return Reservation{}, err
	}
	var eventID uuid.UUID
	if err := s.pool.QueryRow(ctx, `SELECT g.event_id FROM list_items i JOIN gift_lists g ON g.id = i.gift_list_id WHERE i.id = $1`, itemID).Scan(&eventID); err == nil {
		name := analytics.ItemReserved
		if in.Kind == "PURCHASE" {
			name = analytics.ItemPurchased
		}
		s.tracker.Track(ctx, analytics.Event{Name: name, EventID: &eventID, ItemID: &itemID, VisitorID: visitor,
			Props: map[string]any{"quantity": float64(in.Quantity)}})
	}
	return res, nil
}

// expireItem marks overdue ACTIVE reservations of an item as EXPIRED and
// releases their quantity. Returns the released quantity.
func expireItem(ctx context.Context, tx pgx.Tx, itemID uuid.UUID, now time.Time) (int, error) {
	var freed int
	err := tx.QueryRow(ctx, `
		WITH expired AS (
			UPDATE reservations SET status = 'EXPIRED', updated_at = now()
			WHERE list_item_id = $1 AND status = 'ACTIVE' AND expires_at IS NOT NULL AND expires_at <= $2
			RETURNING quantity)
		SELECT COALESCE(SUM(quantity), 0)::int FROM expired`, itemID, now).Scan(&freed)
	if err != nil {
		return 0, fmt.Errorf("expire: %w", err)
	}
	if freed > 0 {
		if _, err := tx.Exec(ctx, `UPDATE list_items SET reserved_quantity = reserved_quantity - $2, updated_at = now() WHERE id = $1`, itemID, freed); err != nil {
			return 0, fmt.Errorf("release: %w", err)
		}
	}
	return freed, nil
}

// ExpireDue releases all overdue reservations; run periodically.
func (s *Service) ExpireDue(ctx context.Context) (int, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT list_item_id FROM reservations WHERE status = 'ACTIVE' AND expires_at <= $1`, s.now())
	if err != nil {
		return 0, err
	}
	var itemIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		itemIDs = append(itemIDs, id)
	}
	rows.Close()
	total := 0
	for _, id := range itemIDs {
		err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `SELECT 1 FROM list_items WHERE id = $1 FOR UPDATE`, id); err != nil {
				return err
			}
			n, err := expireItem(ctx, tx, id, s.now())
			total += n
			return err
		})
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Cancel releases a reservation using the guest's manage token.
func (s *Service) Cancel(ctx context.Context, reservationID uuid.UUID, token string) error {
	if token == "" || len(token) > 128 {
		return ErrReservationMissing
	}
	var itemID, eventID uuid.UUID
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var status, kind string
		var qty int
		err := tx.QueryRow(ctx, `
			SELECT r.list_item_id, r.status, r.kind, r.quantity, g.event_id
			FROM reservations r JOIN list_items i ON i.id = r.list_item_id JOIN gift_lists g ON g.id = i.gift_list_id
			WHERE r.id = $1 AND r.token_hash = $2`, reservationID, ids.HashToken(token)).Scan(&itemID, &status, &kind, &qty, &eventID)
		if err != nil {
			if database.IsNoRows(err) {
				return ErrReservationMissing
			}
			return fmt.Errorf("find reservation: %w", err)
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM list_items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
			return err
		}
		// Re-read under the lock: expiry may have just released it.
		if err := tx.QueryRow(ctx, `SELECT status FROM reservations WHERE id = $1`, reservationID).Scan(&status); err != nil {
			return err
		}
		if status != "ACTIVE" && status != "CONFIRMED" {
			return ErrNotActive
		}
		if _, err := tx.Exec(ctx, `UPDATE reservations SET status = 'CANCELLED', cancelled_at = now(), updated_at = now() WHERE id = $1`, reservationID); err != nil {
			return err
		}
		col := "reserved_quantity"
		if kind == "PURCHASE" {
			col = "purchased_quantity"
		}
		_, err = tx.Exec(ctx, `UPDATE list_items SET `+col+` = GREATEST(`+col+` - $2, 0), updated_at = now() WHERE id = $1`, itemID, qty)
		return err
	})
	if err != nil {
		return err
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.ItemUnreserved, EventID: &eventID, ItemID: &itemID})
	return nil
}

// CancelByOwner lets an event manager release any reservation of their event.
func (s *Service) CancelByOwner(ctx context.Context, userID, reservationID uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var itemID uuid.UUID
		var status, kind string
		var qty int
		err := tx.QueryRow(ctx, `
			SELECT r.list_item_id, r.status, r.kind, r.quantity
			FROM reservations r JOIN list_items i ON i.id = r.list_item_id
			JOIN gift_lists g ON g.id = i.gift_list_id JOIN events e ON e.id = g.event_id
			WHERE r.id = $1 AND e.deleted_at IS NULL AND (e.owner_id = $2 OR EXISTS (
				SELECT 1 FROM event_members m WHERE m.event_id = e.id AND m.user_id = $2 AND m.role = 'CO_OWNER'))`,
			reservationID, userID).Scan(&itemID, &status, &kind, &qty)
		if err != nil {
			if database.IsNoRows(err) {
				return ErrReservationMissing
			}
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM list_items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
			return err
		}
		if status != "ACTIVE" && status != "CONFIRMED" {
			return ErrNotActive
		}
		if _, err := tx.Exec(ctx, `UPDATE reservations SET status = 'CANCELLED', cancelled_at = now(), updated_at = now() WHERE id = $1`, reservationID); err != nil {
			return err
		}
		col := "reserved_quantity"
		if kind == "PURCHASE" {
			col = "purchased_quantity"
		}
		if _, err := tx.Exec(ctx, `UPDATE list_items SET `+col+` = GREATEST(`+col+` - $2, 0), updated_at = now() WHERE id = $1`, itemID, qty); err != nil {
			return err
		}
		return audit.Log(ctx, tx, &userID, "RESERVATION_CANCELLED_BY_OWNER", "reservation", reservationID, nil)
	})
}

// Confirm turns an active reservation into a confirmed purchase.
func (s *Service) Confirm(ctx context.Context, reservationID uuid.UUID, token string) error {
	if token == "" || len(token) > 128 {
		return ErrReservationMissing
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var itemID uuid.UUID
		var qty int
		err := tx.QueryRow(ctx, `SELECT list_item_id, quantity FROM reservations WHERE id = $1 AND token_hash = $2`,
			reservationID, ids.HashToken(token)).Scan(&itemID, &qty)
		if err != nil {
			if database.IsNoRows(err) {
				return ErrReservationMissing
			}
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM list_items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
			return err
		}
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM reservations WHERE id = $1`, reservationID).Scan(&status); err != nil {
			return err
		}
		if status != "ACTIVE" {
			return ErrNotActive
		}
		if _, err := tx.Exec(ctx, `UPDATE reservations SET status = 'CONFIRMED', kind = 'PURCHASE', expires_at = NULL, updated_at = now() WHERE id = $1`, reservationID); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE list_items SET reserved_quantity = reserved_quantity - $2, purchased_quantity = purchased_quantity + $2, updated_at = now() WHERE id = $1`, itemID, qty)
		return err
	})
}
