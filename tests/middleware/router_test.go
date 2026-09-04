package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/TechTeam-ZUS/zus-go-common/middleware"
)

// mockAuthenticator rejects requests missing the "Authorization" header.
type mockAuthenticator struct{}

func (mockAuthenticator) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TestNewRouter(t *testing.T) {
	auth := mockAuthenticator{}
	router := middleware.NewRouter(
		func(r chi.Router) {
			r.Get("/public", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Group(func(r chi.Router) {
				r.Use(auth.Verify)
				r.Get("/private", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			})
		},
	)

	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{name: "public route without auth", path: "/public", expectedStatus: http.StatusOK},
		{name: "private route without auth", path: "/private", expectedStatus: http.StatusUnauthorized},
		{name: "private route with auth", path: "/private", authHeader: "Bearer token", expectedStatus: http.StatusOK},
		{name: "unknown route", path: "/missing", expectedStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestNewRouter_AppliesJSONContentType(t *testing.T) {
	router := middleware.NewRouter(
		func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestNewRouter_CallsAllRegistrars(t *testing.T) {
	called := []string{}
	router := middleware.NewRouter(
		func(r chi.Router) { called = append(called, "first") },
		func(r chi.Router) { called = append(called, "second") },
	)

	assert.NotNil(t, router)
	assert.Equal(t, []string{"first", "second"}, called)
}
