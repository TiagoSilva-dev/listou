package auth

import (
	"net/http"
	"time"

	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

const SessionCookie = "listou_session"

type CookieConfig struct {
	Secure bool
	Domain string
}

type Handler struct {
	svc    *Service
	cookie CookieConfig
}

func NewHandler(svc *Service, cookie CookieConfig) *Handler {
	return &Handler{svc: svc, cookie: cookie}
}

func (h *Handler) Register(mux *http.ServeMux, limit httpx.Middleware) {
	mux.Handle("POST /api/v1/auth/register", limit(http.HandlerFunc(h.register)))
	mux.Handle("POST /api/v1/auth/login", limit(http.HandlerFunc(h.login)))
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.Handle("GET /api/v1/me", h.Require(http.HandlerFunc(h.me)))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, sess, err := h.svc.Register(r.Context(), in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	h.setCookie(w, sess.Token, sess.ExpiresAt)
	httpx.JSON(w, http.StatusCreated, map[string]any{"user": u})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, sess, err := h.svc.Login(r.Context(), in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	h.setCookie(w, sess.Token, sess.ExpiresAt)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": u})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(SessionCookie); err == nil {
		if err := h.svc.Logout(r.Context(), c.Value); err != nil {
			httpx.Fail(w, r, err)
			return
		}
	}
	h.setCookie(w, "", time.Unix(0, 0))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	u, _ := UserFrom(r.Context())
	httpx.JSON(w, http.StatusOK, map[string]any{"user": u})
}

func (h *Handler) setCookie(w http.ResponseWriter, value string, expires time.Time) {
	c := &http.Cookie{
		Name:     SessionCookie,
		Value:    value,
		Path:     "/",
		Domain:   h.cookie.Domain,
		Expires:  expires,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	}
	if value == "" {
		c.MaxAge = -1
	}
	http.SetCookie(w, c)
}

// Require rejects requests without a valid session and stores the user in the context.
func (h *Handler) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(SessionCookie)
		if err != nil {
			httpx.Fail(w, r, httpx.ErrUnauthorized)
			return
		}
		u, err := h.svc.Authenticate(r.Context(), c.Value)
		if err != nil {
			httpx.Fail(w, r, err)
			return
		}
		httpx.SetLogUserID(r.Context(), u.ID.String())
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
	})
}

// RequireFunc is Require for plain handler functions.
func (h *Handler) RequireFunc(fn http.HandlerFunc) http.Handler { return h.Require(fn) }

// RequireAdminFunc is RequireFunc restricted to users with the ADMIN role.
func (h *Handler) RequireAdminFunc(fn http.HandlerFunc) http.Handler {
	return h.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := UserFrom(r.Context()); !u.IsAdmin() {
			httpx.Fail(w, r, httpx.ErrForbidden)
			return
		}
		fn(w, r)
	}))
}

// Optional returns the signed-in user for routes that also serve anonymous visitors.
func (h *Handler) Optional(r *http.Request) *User {
	c, err := r.Cookie(SessionCookie)
	if err != nil {
		return nil
	}
	u, err := h.svc.Authenticate(r.Context(), c.Value)
	if err != nil {
		return nil
	}
	return &u
}
