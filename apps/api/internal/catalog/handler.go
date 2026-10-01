package catalog

import (
	"net/http"
	"strconv"

	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register wires the routes. linkLimit throttles the endpoints that make the
// server fetch a user-supplied URL.
func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler, linkLimit httpx.Middleware) {
	mux.Handle("GET /api/v1/products/search", require(h.search))
	mux.Handle("POST /api/v1/products/import", require(h.importProduct))
	mux.Handle("POST /api/v1/products/link-preview", linkLimit(require(h.linkPreview)))
	mux.Handle("POST /api/v1/products/import-link", linkLimit(require(h.importLink)))
	mux.Handle("GET /api/v1/products/{id}", require(h.get))
	mux.Handle("GET /api/v1/products/{id}/offers", require(h.offers))
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 40 {
		limit = 24
	}
	res, err := h.svc.Search(r.Context(), SearchParams{
		Query: q.Get("q"), Sort: q.Get("sort"), Category: q.Get("category"), Limit: limit, UserID: &u.ID,
	})
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if res == nil {
		res = []SearchResult{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"results": res})
}

type importInput struct {
	ProviderCode string `json:"providerCode"`
	ExternalID   string `json:"externalId"`
}

func (h *Handler) importProduct(w http.ResponseWriter, r *http.Request) {
	var in importInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, offers, err := h.svc.Import(r.Context(), in.ProviderCode, in.ExternalID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p, "offers": offers})
}

type linkInput struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (h *Handler) linkPreview(w http.ResponseWriter, r *http.Request) {
	var in linkInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, err := h.svc.PreviewLink(r.Context(), in.URL)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"preview": p})
}

func (h *Handler) importLink(w http.ResponseWriter, r *http.Request) {
	var in linkInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, offers, err := h.svc.ImportLink(r.Context(), in.URL, in.Title)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p, "offers": offers})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrProductNotFound)
		return
	}
	p, offers, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p, "offers": offers})
}

func (h *Handler) offers(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrProductNotFound)
		return
	}
	_, offers, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"offers": offers})
}
