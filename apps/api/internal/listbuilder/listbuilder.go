// Package listbuilder is the AI List Builder (milestone 10): it suggests
// desires that complement an existing list and lets the owner add the chosen
// ones in one step. Suggestions come from a recommendations.Provider, so an
// LLM or Jev backed provider can replace the rules without touching this
// package. It suggests desires only, never products, prices or links, and it
// is gated by the AI_LIST_BUILDER feature flag.
package listbuilder

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/catalog/textnorm"
	"github.com/listou/listou/apps/api/internal/events"
	"github.com/listou/listou/apps/api/internal/lists"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
	"github.com/listou/listou/apps/api/internal/platform/validate"
	"github.com/listou/listou/apps/api/internal/recommendations"
)

const (
	maxPrompt   = 300
	maxApplying = 40
)

var ErrDisabled = httpx.NewError(http.StatusNotFound, "FEATURE_DISABLED", "Este recurso ainda não está disponível.")

type Service struct {
	events  *events.Service
	lists   *lists.Service
	builder recommendations.Builder
	tracker *analytics.Recorder
	enabled bool
}

func NewService(ev *events.Service, ls *lists.Service, b recommendations.Builder, t *analytics.Recorder, enabled bool) *Service {
	return &Service{events: ev, lists: ls, builder: b, tracker: t, enabled: enabled}
}

type SuggestInput struct {
	Prompt string `json:"prompt"`
}

type Suggestions struct {
	Categories []recommendations.CategorySuggestion `json:"categories"`
	Source     string                               `json:"source"`
}

func (s *Service) Suggest(ctx context.Context, userID, eventID uuid.UUID, in SuggestInput) (Suggestions, error) {
	if !s.enabled {
		return Suggestions{}, ErrDisabled
	}
	in.Prompt = strings.TrimSpace(in.Prompt)
	if len([]rune(in.Prompt)) > maxPrompt {
		return Suggestions{}, httpx.Validation(map[string]string{"prompt": "Use até 300 caracteres."})
	}
	ev, err := s.events.Get(ctx, userID, eventID) // also authorizes the caller
	if err != nil {
		return Suggestions{}, err
	}
	view, err := s.lists.ForEvent(ctx, userID, eventID)
	if err != nil {
		return Suggestions{}, err
	}
	existing := make([]string, 0, len(view.Items))
	for _, it := range view.Items {
		existing = append(existing, it.Title)
	}
	cats, err := s.builder.Suggest(ctx, recommendations.BuildRequest{EventType: ev.Type, Prompt: in.Prompt, Existing: existing})
	if err != nil {
		return Suggestions{}, err
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.AISuggested, EventID: &eventID, UserID: &userID,
		Props: map[string]any{"hasPrompt": in.Prompt != "", "categories": len(cats)}})
	return Suggestions{Categories: cats, Source: "RULES"}, nil
}

type ApplyDesire struct {
	Title    string `json:"title"`
	Emoji    string `json:"emoji"`
	Quantity int    `json:"quantity"`
	Priority string `json:"priority"`
}

type ApplyCategory struct {
	Name    string        `json:"name"`
	Emoji   string        `json:"emoji"`
	Desires []ApplyDesire `json:"desires"`
}

type ApplyInput struct {
	Categories []ApplyCategory `json:"categories"`
}

type ApplyResult struct {
	AddedItems        int `json:"addedItems"`
	CreatedCategories int `json:"createdCategories"`
}

// Apply adds the chosen desires, reusing categories that already exist by name.
func (s *Service) Apply(ctx context.Context, userID, eventID uuid.UUID, in ApplyInput) (ApplyResult, error) {
	if !s.enabled {
		return ApplyResult{}, ErrDisabled
	}
	total := 0
	for _, c := range in.Categories {
		total += len(c.Desires)
	}
	if total == 0 || total > maxApplying {
		return ApplyResult{}, httpx.Validation(map[string]string{"categories": "Escolha de 1 a 40 itens."})
	}
	view, err := s.lists.ForEvent(ctx, userID, eventID)
	if err != nil {
		return ApplyResult{}, err
	}
	byName := map[string]uuid.UUID{}
	for _, c := range view.Categories {
		byName[textnorm.Normalize(c.Name)] = c.ID
	}
	var res ApplyResult
	for _, c := range in.Categories {
		name := strings.TrimSpace(c.Name)
		if name == "" || len(c.Desires) == 0 {
			continue
		}
		key := textnorm.Normalize(name)
		catID, ok := byName[key]
		if !ok {
			ci := lists.CategoryInput{Name: &name, Emoji: validate.Trim(&c.Emoji)}
			cat, err := s.lists.CreateCategory(ctx, userID, view.List.ID, ci)
			if err != nil {
				return res, err
			}
			catID = cat.ID
			byName[key] = catID
			res.CreatedCategories++
		}
		cid := catID.String()
		for _, d := range c.Desires {
			title := strings.TrimSpace(d.Title)
			qty := d.Quantity
			if qty < 1 {
				qty = 1
			}
			prio := d.Priority
			if prio == "" {
				prio = "MEDIUM"
			}
			it := lists.ItemInput{Title: &title, Emoji: validate.Trim(&d.Emoji), DesiredQuantity: &qty, Priority: &prio, CategoryID: &cid}
			if _, err := s.lists.CreateItem(ctx, userID, view.List.ID, it); err != nil {
				return res, err
			}
			res.AddedItems++
		}
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.AIApplied, EventID: &eventID, UserID: &userID,
		Props: map[string]any{"items": res.AddedItems}})
	return res, nil
}

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("POST /api/v1/events/{id}/suggestions", require(h.suggest))
	mux.Handle("POST /api/v1/events/{id}/suggestions/apply", require(h.apply))
}

func (h *Handler) suggest(w http.ResponseWriter, r *http.Request) {
	u, id, ok := h.parse(w, r)
	if !ok {
		return
	}
	var in SuggestInput
	if r.ContentLength != 0 {
		if err := httpx.Decode(r, &in); err != nil {
			httpx.Fail(w, r, err)
			return
		}
	}
	out, err := h.svc.Suggest(r.Context(), u, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (h *Handler) apply(w http.ResponseWriter, r *http.Request) {
	u, id, ok := h.parse(w, r)
	if !ok {
		return
	}
	var in ApplyInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out, err := h.svc.Apply(r.Context(), u, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, out)
}

func (h *Handler) parse(w http.ResponseWriter, r *http.Request) (user, event uuid.UUID, ok bool) {
	u, _ := auth.UserFrom(r.Context())
	id, valid := ids.Parse(r.PathValue("id"))
	if !valid {
		httpx.Fail(w, r, access.ErrEventNotFound)
		return uuid.Nil, uuid.Nil, false
	}
	return u.ID, id, true
}
