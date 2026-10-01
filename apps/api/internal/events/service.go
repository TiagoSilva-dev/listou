package events

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/platform/audit"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
	"github.com/listou/listou/apps/api/internal/recommendations"
)

var ErrSlugTaken = httpx.NewError(http.StatusConflict, "SLUG_TAKEN", "Este endereço já está em uso. Tente outro.")

// ListSeeder creates the primary gift list for a new event. Implemented by
// the lists module so events never writes list tables directly.
type ListSeeder interface {
	CreatePrimaryList(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, title string, suggestions []recommendations.CategorySuggestion) (uuid.UUID, error)
}

type Service struct {
	pool    *pgxpool.Pool
	repo    Repository
	lists   ListSeeder
	recs    recommendations.Provider
	tracker *analytics.Recorder
	now     func() time.Time
}

func NewService(pool *pgxpool.Pool, lists ListSeeder, recs recommendations.Provider, tracker *analytics.Recorder) *Service {
	return &Service{pool: pool, lists: lists, recs: recs, tracker: tracker, now: time.Now}
}

func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, in CreateInput) (Event, error) {
	date, err := in.Validate()
	if err != nil {
		return Event{}, err
	}
	e := Event{
		ID: ids.New(), OwnerID: ownerID, Type: in.Type, Title: in.Title, Description: in.Description,
		HostNames: in.HostNames, EventDate: date, Location: in.Location, CoverImageURL: in.CoverImageURL,
		Theme: defaultTheme[in.Type], Visibility: "UNLISTED", Status: "DRAFT",
	}
	if in.Theme != nil {
		e.Theme = *in.Theme
	}
	if in.Visibility != nil {
		e.Visibility = *in.Visibility
	}

	var suggestions []recommendations.CategorySuggestion
	if in.Template == "SUGGESTED" {
		if suggestions, err = s.recs.SuggestCategories(ctx, recommendations.Request{EventType: in.Type}); err != nil {
			return Event{}, fmt.Errorf("create event: suggestions: %w", err)
		}
	}

	base := in.Slug
	userChoseSlug := base != nil
	candidate := ""
	if userChoseSlug {
		candidate = *base
	} else {
		candidate = suggestSlug(in.Title, in.HostNames)
	}

	for attempt := 0; attempt < 6; attempt++ {
		e.Slug = candidate
		err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
			if err := s.repo.Insert(ctx, tx, e); err != nil {
				return err
			}
			listID, err := s.lists.CreatePrimaryList(ctx, tx, e.ID, e.Title, suggestions)
			e.ListID = listID
			return err
		})
		if err == nil {
			break
		}
		if !database.IsUniqueViolation(err, "events_slug_key") {
			return Event{}, err
		}
		if userChoseSlug {
			return Event{}, ErrSlugTaken
		}
		candidate = withSuffix(suggestSlug(in.Title, in.HostNames))
	}
	if err != nil {
		return Event{}, ErrSlugTaken
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.ListCreated, EventID: &e.ID, UserID: &ownerID,
		Props: map[string]any{"type": e.Type, "template": in.Template}})
	return s.repo.Get(ctx, s.pool, e.ID)
}

func suggestSlug(title string, hostNames *string) string {
	src := title
	if hostNames != nil && *hostNames != "" {
		src = *hostNames
	}
	slug := Slugify(src)
	if !ValidSlug(slug) {
		slug = withSuffix("lista")
	}
	return slug
}

func withSuffix(slug string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	if len(slug) > 54 {
		slug = slug[:54]
	}
	return fmt.Sprintf("%s-%d", slug, n.Int64()+1000)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Event, error) {
	return s.repo.ListForUser(ctx, s.pool, userID)
}

func (s *Service) Get(ctx context.Context, userID, eventID uuid.UUID) (Event, error) {
	if _, err := access.Event(ctx, s.pool, eventID, userID); err != nil {
		return Event{}, err
	}
	return s.repo.Get(ctx, s.pool, eventID)
}

func (s *Service) Update(ctx context.Context, userID, eventID uuid.UUID, in UpdateInput) (Event, error) {
	date, err := in.Validate()
	if err != nil {
		return Event{}, err
	}
	var out Event
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := access.Event(ctx, tx, eventID, userID); err != nil {
			return err
		}
		e, err := s.repo.Get(ctx, tx, eventID)
		if err != nil {
			return err
		}
		wasPublished := e.Status == "PUBLISHED"
		apply(&e, in, date)
		if e.Status == "PUBLISHED" && e.PublishedAt == nil {
			now := s.now()
			e.PublishedAt = &now
		}
		if err := s.repo.Update(ctx, tx, e); err != nil {
			if database.IsUniqueViolation(err, "events_slug_key") {
				return ErrSlugTaken
			}
			return err
		}
		if !wasPublished && e.Status == "PUBLISHED" {
			if err := audit.Log(ctx, tx, &userID, "EVENT_PUBLISHED", "event", e.ID, nil); err != nil {
				return err
			}
		}
		out = e
		return nil
	})
	return out, err
}

func apply(e *Event, in UpdateInput, date *time.Time) {
	if in.Title != nil {
		e.Title = *in.Title
	}
	if in.HostNames != nil {
		e.HostNames = nilIfEmpty(*in.HostNames)
	}
	if in.Description != nil {
		e.Description = nilIfEmpty(*in.Description)
	}
	if in.EventDate != nil {
		e.EventDate = date
	}
	if in.Location != nil {
		e.Location = nilIfEmpty(*in.Location)
	}
	if in.CoverImageURL != nil {
		e.CoverImageURL = nilIfEmpty(*in.CoverImageURL)
	}
	if in.Slug != nil {
		e.Slug = *in.Slug
	}
	if in.Visibility != nil {
		e.Visibility = *in.Visibility
	}
	if in.Status != nil {
		e.Status = *in.Status
	}
	if in.Theme != nil {
		e.Theme = *in.Theme
	}
	if in.SurpriseMode != nil {
		e.SurpriseMode = *in.SurpriseMode
	}
	if in.ShowReserverNames != nil {
		e.ShowReserverNames = *in.ShowReserverNames
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) Delete(ctx context.Context, userID, eventID uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		role, err := access.Event(ctx, tx, eventID, userID)
		if err != nil {
			return err
		}
		if role != access.Owner {
			return httpx.ErrForbidden
		}
		if err := s.repo.SoftDelete(ctx, tx, eventID, s.now()); err != nil {
			return err
		}
		return audit.Log(ctx, tx, &userID, "EVENT_DELETED", "event", eventID, nil)
	})
}

func (s *Service) SlugAvailable(ctx context.Context, slug string) (bool, error) {
	if !ValidSlug(slug) {
		return false, nil
	}
	exists, err := s.repo.SlugExists(ctx, s.pool, slug)
	return !exists, err
}
