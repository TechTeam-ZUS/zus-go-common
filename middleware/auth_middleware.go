package middleware

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Authenticator verifies an incoming request and attaches identity to context.
// AuthMiddleware is the built-in JWT implementation; consumers needing a
// different auth scheme can implement their own.
type Authenticator interface {
	Verify(next http.Handler) http.Handler
}

type AuthMiddleware struct {
	publicKey *rsa.PublicKey
}

func NewAuthMiddleware(secret string) (*AuthMiddleware, error) {
	rawStr := secret
	rawStr = strings.ReplaceAll(rawStr, "\\n", "\n")
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(rawStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %v", err)
	}
	return &AuthMiddleware{publicKey: key}, nil
}

func (m *AuthMiddleware) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.publicKey, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("token validation failed", "reason", sanitizeTokenError(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := token.Claims.GetSubject()
		if err != nil || userID == "" {
			logger.FromContext(r.Context()).Warn("user not found", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sanitizeTokenError(err error) string {
	if err == nil {
		return "token invalid"
	}
	// Return only the error type/category, not the raw message
	// which may contain token bytes or claims
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return "token expired"
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return "token not yet valid"
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "token malformed"
	case errors.Is(err, jwt.ErrSignatureInvalid):
		return "invalid signature"
	default:
		return "token validation error"
	}
}
