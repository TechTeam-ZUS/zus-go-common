package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/middleware"
)

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
			auth, err := middleware.NewAuthMiddleware(tt.secret)
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
	auth, err := middleware.NewAuthMiddleware(pubPEM)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + signToken(t, key, "user-123", time.Now().Add(time.Hour)),
			expectedStatus: http.StatusOK,
			expectedUserID: "user-123",
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
			name:           "missing subject",
			authHeader:     "Bearer " + signToken(t, key, "", time.Now().Add(time.Hour)),
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
			expectedUserID: "user-123",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
