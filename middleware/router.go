package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouteRegistrar registers routes on r. auth is passed through so the
// registrar decides per-route/group whether to apply auth.Verify.
type RouteRegistrar func(r chi.Router, auth Authenticator)

// NewRouter builds the base router with standard middleware, then hands
// control to each registrar to register its own routes.
func NewRouter(auth Authenticator, routes ...RouteRegistrar) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Compress(5))
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Use(JSONResponseMiddleware)

	for _, register := range routes {
		register(r, auth)
	}

	return r
}

func JSONResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
