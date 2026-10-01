package lists

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/catalog"
	"github.com/listou/listou/apps/api/internal/decision"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
	"github.com/listou/listou/apps/api/internal/recommendations"
)

var (
	ErrQuantityConflict = httpx.NewError(http.StatusConflict, "ITEM_QUANTITY_CONFLICT",
		"A quantidade não pode ser menor que o total já reservado ou comprado.")
	ErrCategoryTaken = httpx.NewError(http.StatusConflict, "CATEGORY_EXISTS", "Já existe uma categoria com este nome.")
)

type Service struct {
	pool    *pgxpool.Pool
	repo    Repository
	catalog *catalog.Service
	decider decision.Engine
	tracker *analytics.Recorder
}

func NewService(pool *pgxpool.Pool, cat *catalog.Service, decider decision.Engine, tracker *analytics.Recorder) *Service {
	return &Service{pool: pool, catalog: cat, decider: decider, tracker: tracker}
}

// CreatePrimaryList implements events.ListSeeder.
func (s *Service) CreatePrimaryList(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, title string, suggestions []recommendations.CategorySuggestion) (uuid.UUID, error) {
	l := GiftList{ID: ids.New(), EventID: eventID, Title: title}
	if err := s.repo.InsertList(ctx, tx, l); err != nil {
		return uuid.Nil, err
	}
	pos := 0
	for ci, sug := range suggestions {
		emoji := sug.Emoji
		c := Category{ID: ids.New(), Name: sug.Name, Emoji: &emoji, Position: ci}
		if err := s.repo.InsertCategory(ctx, tx, l.ID, c); err != nil {
			return uuid.Nil, fmt.Errorf("seed category: %w", err)
		}
		for _, d := range sug.Desires {
			e := d.Emoji
			priority := "MEDIUM"
			if d.Importance == recommendations.Essential {
				priority = "HIGH"
			} else if d.Importance == recommendations.Optional {
				priority = "LOW"
			}
			it := Item{ID: ids.New(), ListID: l.ID, CategoryID: &c.ID, Title: d.Title, Emoji: &e, Currency: "BRL",
				Priority: priority, DesiredQuantity: d.Quantity, Position: pos}
			pos++
			if err := s.repo.InsertItem(ctx, tx, it); err != nil {
				return uuid.Nil, err
			}
		}
	}
	return l.ID, nil
}

type ListView struct {
	List       GiftList   `json:"list"`
	Categories []Category `json:"categories"`
	Items      []Item     `json:"items"`
}

func (s *Service) ForEvent(ctx context.Context, userID, eventID uuid.UUID) (ListView, error) {
	if _, err := access.Event(ctx, s.pool, eventID, userID); err != nil {
		return ListView{}, err
	}
	l, err := s.repo.PrimaryList(ctx, s.pool, eventID)
	if err != nil {
		return ListView{}, fmt.Errorf("primary list: %w", err)
	}
	v, err := s.View(ctx, s.pool, l)
	if err != nil {
		return ListView{}, err
	}
	if err := s.redactIfSurprise(ctx, eventID, v.Items); err != nil {
		return ListView{}, err
	}
	return v, nil
}

// redactIfSurprise hides per-item reserved/purchased state from the owner when
// the event is in surprise mode ("I don't want to know who bought what").
// Aggregate totals remain available on the dashboard.
func (s *Service) redactIfSurprise(ctx context.Context, eventID uuid.UUID, items []Item) error {
	var surprise bool
	if err := s.pool.QueryRow(ctx, `SELECT surprise_mode FROM events WHERE id = $1`, eventID).Scan(&surprise); err != nil {
		return fmt.Errorf("surprise flag: %w", err)
	}
	if !surprise {
		return nil
	}
	for i := range items {
		items[i].PurchasedQuantity, items[i].ReservedQuantity = 0, 0
		items[i].AvailableQuantity, items[i].Status = items[i].DesiredQuantity, Available
	}
	return nil
}

// View loads categories and hydrated items for a list (also used by the public page).
func (s *Service) View(ctx context.Context, db database.DBTX, l GiftList) (ListView, error) {
	cats, err := s.repo.Categories(ctx, db, l.ID)
	if err != nil {
		return ListView{}, err
	}
	items, err := s.repo.Items(ctx, db, l.ID)
	if err != nil {
		return ListView{}, err
	}
	if err := s.hydrate(ctx, db, items); err != nil {
		return ListView{}, err
	}
	return ListView{List: l, Categories: cats, Items: items}, nil
}

func (s *Service) PrimaryList(ctx context.Context, db database.DBTX, eventID uuid.UUID) (GiftList, error) {
	return s.repo.PrimaryList(ctx, db, eventID)
}

func (s *Service) hydrate(ctx context.Context, db database.DBTX, items []Item) error {
	var pids []uuid.UUID
	for _, it := range items {
		if it.ProductID != nil {
			pids = append(pids, *it.ProductID)
		}
	}
	products, offers, err := s.catalog.Hydrate(ctx, db, pids)
	if err != nil {
		return err
	}
	for i := range items {
		it := &items[i]
		if it.ProductID != nil {
			if p, ok := products[*it.ProductID]; ok {
				it.Product = &ProductSummary{ID: p.ID, CanonicalTitle: p.CanonicalTitle, Brand: p.Brand, ImageURL: p.ImageURL, Emoji: p.Emoji, Demo: p.Demo}
			}
			it.Offers = offers[*it.ProductID]
		}
		it.finalize()
	}
	return nil
}

func (s *Service) UpdateList(ctx context.Context, userID, listID uuid.UUID, in ListInput) (GiftList, error) {
	if _, err := access.List(ctx, s.pool, listID, userID); err != nil {
		return GiftList{}, err
	}
	l, err := s.repo.List(ctx, s.pool, listID)
	if err != nil {
		return GiftList{}, err
	}
	if in.Title != nil && *in.Title != "" && len(*in.Title) <= 120 {
		l.Title = *in.Title
	}
	if in.Description != nil {
		l.Description = in.Description
	}
	if in.AllowReservations != nil {
		l.AllowReservations = *in.AllowReservations
	}
	return l, s.repo.UpdateList(ctx, s.pool, l)
}

func (s *Service) CreateCategory(ctx context.Context, userID, listID uuid.UUID, in CategoryInput) (Category, error) {
	if err := in.Validate(true); err != nil {
		return Category{}, err
	}
	if _, err := access.List(ctx, s.pool, listID, userID); err != nil {
		return Category{}, err
	}
	pos, err := s.repo.NextCategoryPosition(ctx, s.pool, listID)
	if err != nil {
		return Category{}, err
	}
	c := Category{ID: ids.New(), Name: *in.Name, Emoji: in.Emoji, Position: pos}
	if err := s.repo.InsertCategory(ctx, s.pool, listID, c); err != nil {
		if database.IsUniqueViolation(err, "categories_list_name_key") {
			return Category{}, ErrCategoryTaken
		}
		return Category{}, err
	}
	return c, nil
}

func (s *Service) UpdateCategory(ctx context.Context, userID, categoryID uuid.UUID, in CategoryInput) (Category, error) {
	if err := in.Validate(false); err != nil {
		return Category{}, err
	}
	if _, err := access.Category(ctx, s.pool, categoryID, userID); err != nil {
		return Category{}, err
	}
	c, err := s.repo.Category(ctx, s.pool, categoryID)
	if err != nil {
		return Category{}, err
	}
	if in.Name != nil {
		c.Name = *in.Name
	}
	if in.Emoji != nil {
		c.Emoji = in.Emoji
	}
	if in.Position != nil {
		c.Position = *in.Position
	}
	if err := s.repo.UpdateCategory(ctx, s.pool, c); err != nil {
		if database.IsUniqueViolation(err, "categories_list_name_key") {
			return Category{}, ErrCategoryTaken
		}
		return Category{}, err
	}
	return c, nil
}

func (s *Service) DeleteCategory(ctx context.Context, userID, categoryID uuid.UUID) error {
	if _, err := access.Category(ctx, s.pool, categoryID, userID); err != nil {
		return err
	}
	return s.repo.DeleteCategory(ctx, s.pool, categoryID)
}

func (s *Service) Items(ctx context.Context, userID, listID uuid.UUID) ([]Item, error) {
	eventID, err := access.List(ctx, s.pool, listID, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.Items(ctx, s.pool, listID)
	if err != nil {
		return nil, err
	}
	if err := s.hydrate(ctx, s.pool, items); err != nil {
		return nil, err
	}
	return items, s.redactIfSurprise(ctx, eventID, items)
}

func (s *Service) CreateItem(ctx context.Context, userID, listID uuid.UUID, in ItemInput) (Item, error) {
	if err := in.Validate(true); err != nil {
		return Item{}, err
	}
	eventID, err := access.List(ctx, s.pool, listID, userID)
	if err != nil {
		return Item{}, err
	}
	it := Item{ID: ids.New(), ListID: listID, Currency: "BRL", Priority: "MEDIUM", DesiredQuantity: 1}
	if err := s.applyInput(ctx, &it, in); err != nil {
		return Item{}, err
	}
	if it.CategoryID == nil {
		it.CategoryID = s.classify(ctx, listID, it.Title)
	}
	if it.Position, err = s.repo.NextItemPosition(ctx, s.pool, listID); err != nil {
		return Item{}, err
	}
	if err := s.repo.InsertItem(ctx, s.pool, it); err != nil {
		return Item{}, err
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.ItemCreated, EventID: &eventID, ItemID: &it.ID, UserID: &userID,
		Props: map[string]any{"withProduct": it.ProductID != nil}})
	return s.loadItem(ctx, it.ID)
}

// classify suggests one of the list's categories for an uncategorized item.
func (s *Service) classify(ctx context.Context, listID uuid.UUID, title string) *uuid.UUID {
	cats, err := s.repo.Categories(ctx, s.pool, listID)
	if err != nil || len(cats) == 0 {
		return nil
	}
	labels := make([]string, len(cats))
	for i, c := range cats {
		labels[i] = c.Name
	}
	d, err := s.decider.Classify(ctx, title, labels)
	if err != nil || d.Label == "" {
		return nil
	}
	for _, c := range cats {
		if c.Name == d.Label {
			id := c.ID
			return &id
		}
	}
	return nil
}

// applyInput copies validated fields onto the item, resolving category and
// product references against the database (never trusting client ids).
func (s *Service) applyInput(ctx context.Context, it *Item, in ItemInput) error {
	if in.ProductID != nil {
		if *in.ProductID == "" {
			it.ProductID = nil
		} else {
			pid, ok := ids.Parse(*in.ProductID)
			if !ok {
				return catalog.ErrProductNotFound
			}
			p, offers, err := s.catalog.Get(ctx, pid)
			if err != nil {
				return err
			}
			it.ProductID = &p.ID
			if it.Title == "" && (in.Title == nil || *in.Title == "") {
				it.Title = p.CanonicalTitle
			}
			if it.ImageURL == nil {
				it.ImageURL = p.ImageURL
			}
			if it.Emoji == nil {
				it.Emoji = p.Emoji
			}
			if it.PriceReferenceCents == nil && len(offers) > 0 {
				it.PriceReferenceCents = offers[0].PriceCents
			}
		}
	}
	if in.Title != nil && *in.Title != "" {
		it.Title = *in.Title
	}
	if in.CategoryID != nil {
		if *in.CategoryID == "" {
			it.CategoryID = nil
		} else {
			cid, ok := ids.Parse(*in.CategoryID)
			if !ok {
				return access.ErrCategoryNotFound
			}
			okList, err := s.repo.CategoryInList(ctx, s.pool, cid, it.ListID)
			if err != nil {
				return err
			}
			if !okList {
				return access.ErrCategoryNotFound
			}
			it.CategoryID = &cid
		}
	}
	setStr := func(dst **string, v *string) {
		if v != nil {
			if *v == "" {
				*dst = nil
			} else {
				val := *v
				*dst = &val
			}
		}
	}
	setStr(&it.Description, in.Description)
	setStr(&it.Notes, in.Notes)
	setStr(&it.ImageURL, in.ImageURL)
	setStr(&it.Emoji, in.Emoji)
	setStr(&it.ExternalURL, in.ExternalURL)
	if in.PriceReferenceCents != nil {
		it.PriceReferenceCents = in.PriceReferenceCents
	}
	if in.Priority != nil {
		it.Priority = *in.Priority
	}
	if in.DesiredQuantity != nil {
		it.DesiredQuantity = *in.DesiredQuantity
	}
	if in.PurchasedQuantity != nil {
		it.PurchasedQuantity = *in.PurchasedQuantity
	}
	if in.Position != nil {
		it.Position = *in.Position
	}
	return nil
}

func (s *Service) loadItem(ctx context.Context, id uuid.UUID) (Item, error) {
	it, err := s.repo.Item(ctx, s.pool, id, false)
	if err != nil {
		return Item{}, err
	}
	items := []Item{it}
	if err := s.hydrate(ctx, s.pool, items); err != nil {
		return Item{}, err
	}
	var eventID uuid.UUID
	if err := s.pool.QueryRow(ctx, `SELECT event_id FROM gift_lists WHERE id = $1`, it.ListID).Scan(&eventID); err != nil {
		return Item{}, err
	}
	return items[0], s.redactIfSurprise(ctx, eventID, items)
}

func (s *Service) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, in ItemInput) (Item, error) {
	if err := in.Validate(false); err != nil {
		return Item{}, err
	}
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, _, err := access.Item(ctx, tx, itemID, userID); err != nil {
			return err
		}
		it, err := s.repo.Item(ctx, tx, itemID, true)
		if err != nil {
			return err
		}
		if err := s.applyInput(ctx, &it, in); err != nil {
			return err
		}
		if it.PurchasedQuantity+it.ReservedQuantity > it.DesiredQuantity {
			return ErrQuantityConflict
		}
		return s.repo.UpdateItem(ctx, tx, it)
	})
	if err != nil {
		if database.IsCheckViolation(err, "list_items_quantity_capacity") {
			return Item{}, ErrQuantityConflict
		}
		return Item{}, err
	}
	return s.loadItem(ctx, itemID)
}

// DeleteItem archives items with reservation history (keeps guests' records
// consistent) and hard-deletes the rest.
func (s *Service) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, _, err := access.Item(ctx, tx, itemID, userID); err != nil {
			return err
		}
		has, err := s.repo.HasReservations(ctx, tx, itemID)
		if err != nil {
			return err
		}
		if has {
			return s.repo.ArchiveItem(ctx, tx, itemID)
		}
		return s.repo.DeleteItem(ctx, tx, itemID)
	})
}
