package httpx

import (
	"net/http"
	"strings"
)

// SameOrigin rejects state-changing browser requests whose Origin is not one
// of the allowed web origins. Combined with SameSite=Lax session cookies this
// covers CSRF without a token round-trip. Requests without an Origin header
// (server-to-server calls from the BFF, curl) are allowed through; browsers
// always send Origin on cross-site POST/PATCH/DELETE.
func SameOrigin(allowed ...string) Middleware {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[strings.TrimRight(a, "/")] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			if origin := r.Header.Get("Origin"); origin != "" {
				if _, ok := set[origin]; !ok {
					Fail(w, r, NewError(http.StatusForbidden, "CSRF_REJECTED", "Origem da requisição não permitida."))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
