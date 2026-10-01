package recommendations

import (
	"net/http"

	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type Handler struct{ provider Provider }

func NewHandler(p Provider) *Handler { return &Handler{provider: p} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/recommendations/categories", h.categories)
}

func (h *Handler) categories(w http.ResponseWriter, r *http.Request) {
	out, err := h.provider.SuggestCategories(r.Context(), Request{EventType: r.URL.Query().Get("eventType")})
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"categories": out, "source": "RULES"})
}
