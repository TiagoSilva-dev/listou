package lists

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /api/v1/events/{id}/list", require(h.forEvent))
	mux.Handle("PATCH /api/v1/lists/{id}", require(h.updateList))
	mux.Handle("POST /api/v1/lists/{id}/categories", require(h.createCategory))
	mux.Handle("PATCH /api/v1/categories/{id}", require(h.updateCategory))
	mux.Handle("DELETE /api/v1/categories/{id}", require(h.deleteCategory))
	mux.Handle("POST /api/v1/lists/{id}/items", require(h.createItem))
	mux.Handle("GET /api/v1/lists/{id}/items", require(h.items))
	mux.Handle("PATCH /api/v1/items/{id}", require(h.updateItem))
	mux.Handle("DELETE /api/v1/items/{id}", require(h.deleteItem))
}

func pathID(r *http.Request, notFound error) (uuid.UUID, error) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		return uuid.Nil, notFound
	}
	return id, nil
}

// run is a tiny helper that removes handler boilerplate: parse the path id,
// optionally decode a body, call the service, encode the result.
func run[In any, Out any](w http.ResponseWriter, r *http.Request, notFound error, status int, key string,
	fn func(userID, id uuid.UUID, in In) (Out, error)) {
	u, _ := auth.UserFrom(r.Context())
	id, err := pathID(r, notFound)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in In
	if r.Method == http.MethodPost || r.Method == http.MethodPatch {
		if err := httpx.Decode(r, &in); err != nil {
			httpx.Fail(w, r, err)
			return
		}
	}
	out, err := fn(u.ID, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if key == "" {
		w.WriteHeader(status)
		return
	}
	httpx.JSON(w, status, map[string]any{key: out})
}

type none struct{}

func (h *Handler) forEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, err := pathID(r, access.ErrEventNotFound)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	v, err := h.svc.ForEvent(r.Context(), u.ID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) updateList(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrListNotFound, http.StatusOK, "list", func(u, id uuid.UUID, in ListInput) (GiftList, error) {
		return h.svc.UpdateList(r.Context(), u, id, in)
	})
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrListNotFound, http.StatusCreated, "category", func(u, id uuid.UUID, in CategoryInput) (Category, error) {
		return h.svc.CreateCategory(r.Context(), u, id, in)
	})
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrCategoryNotFound, http.StatusOK, "category", func(u, id uuid.UUID, in CategoryInput) (Category, error) {
		return h.svc.UpdateCategory(r.Context(), u, id, in)
	})
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrCategoryNotFound, http.StatusNoContent, "", func(u, id uuid.UUID, _ none) (none, error) {
		return none{}, h.svc.DeleteCategory(r.Context(), u, id)
	})
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrListNotFound, http.StatusCreated, "item", func(u, id uuid.UUID, in ItemInput) (Item, error) {
		return h.svc.CreateItem(r.Context(), u, id, in)
	})
}

func (h *Handler) items(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrListNotFound, http.StatusOK, "items", func(u, id uuid.UUID, _ none) ([]Item, error) {
		return h.svc.Items(r.Context(), u, id)
	})
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrItemNotFound, http.StatusOK, "item", func(u, id uuid.UUID, in ItemInput) (Item, error) {
		return h.svc.UpdateItem(r.Context(), u, id, in)
	})
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	run(w, r, access.ErrItemNotFound, http.StatusNoContent, "", func(u, id uuid.UUID, _ none) (none, error) {
		return none{}, h.svc.DeleteItem(r.Context(), u, id)
	})
}
