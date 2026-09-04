package middleware_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/middleware"
)

// mockRevocationCheck returns a TokenValidation simulating a pure
// revocation/store check: passes ctx through unchanged, or fails with err.
func mockRevocationCheck(err error) middleware.TokenValidation {
	return func(ctx context.Context, token *jwt.Token) (context.Context, error) {
		return ctx, err
	}
}

// mockEnrichment returns a TokenValidation that replicates the old
// sub-claim-to-UserIDKey behavior, or always returns the given error.
func mockEnrichment(err error) middleware.TokenValidation {
	return func(ctx context.Context, token *jwt.Token) (context.Context, error) {
		if err != nil {
			return ctx, err
		}
		userID, gerr := token.Claims.GetSubject()
		if gerr != nil || userID == "" {
			return ctx, errors.New("missing subject")
		}
		return context.WithValue(ctx, middleware.UserIDKey, userID), nil
	}
}

func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return key, string(pubPEM)
}

func signToken(t *testing.T, key *rsa.PrivateKey, subject string, expiresAt time.Time) string {
	t.Helper()
	claims := jwt.RegisteredClaims{Subject: subject, ExpiresAt: jwt.NewNumericDate(expiresAt)}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func signTokenNotYetValid(t *testing.T, key *rsa.PrivateKey, subject string) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func signTokenHS256(t *testing.T, subject string) string {
	t.Helper()
	claims := jwt.RegisteredClaims{Subject: subject, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("attacker-controlled-secret"))
	require.NoError(t, err)
	return signed
}

func TestNewAuthMiddleware(t *testing.T) {
	_, pubPEM := generateRSAKeyPair(t)

	tests := []struct {
		name        string
		secret      string
		expectedErr bool
	}{
		{name: "valid PEM key", secret: pubPEM, expectedErr: false},
		{name: "invalid PEM key", secret: "not-a-key", expectedErr: true},
		{name: "empty secret", secret: "", expectedErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := middleware.NewAuthMiddleware(tt.secret, nil)
			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, auth)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, auth)
		})
	}
}

func TestAuthMiddleware_Verify(t *testing.T) {
	key, pubPEM := generateRSAKeyPair(t)
	otherKey, _ := generateRSAKeyPair(t)

	tests := []struct {
		name           string
		authHeader     string
		validate       middleware.TokenValidation
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "valid token, no validate",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed token",
			authHeader:     "Bearer not-a-jwt",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(-time.Hour)),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "signed with wrong key",
			authHeader:     "Bearer " + signToken(t, otherKey, "user-123", time.Now().Add(time.Hour)),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "not-yet-valid token",
			authHeader:     "Bearer " + signTokenNotYetValid(t, key, "user-123"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "algorithm confusion: HS256-signed token rejected",
			authHeader:     "Bearer " + signTokenHS256(t, "user-123"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "token without Bearer prefix is still accepted",
			authHeader:     signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "lowercase bearer prefix is not stripped, fails to parse",
			authHeader:     "bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Bearer with empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing subject rejected by validate",
			authHeader:     "Bearer " + signToken(t, key, "", time.Now().Add(time.Hour)),
			validate:       mockEnrichment(nil),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "validate sets userID in context",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			validate:       mockEnrichment(nil),
			expectedStatus: http.StatusOK,
			expectedUserID: "user-123",
		},
		{
			name:           "validate error rejects request",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			validate:       mockEnrichment(errors.New("enrich failed")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "revocation check accepts token",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			validate:       mockRevocationCheck(nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "revocation check rejects revoked token",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			validate:       mockRevocationCheck(errors.New("token revoked")),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := middleware.NewAuthMiddleware(pubPEM, tt.validate)
			require.NoError(t, err)

			var gotUserID any
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUserID = r.Context().Value(middleware.UserIDKey)
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			auth.Verify(next).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedUserID != "" {
				assert.Equal(t, tt.expectedUserID, gotUserID)
			}
		})
	}
}
