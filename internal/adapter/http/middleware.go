package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// Middleware defines a function that wraps a http.Handler
type Middleware func(http.Handler) http.Handler

// ChainMiddleware chains multiple middlewares
func ChainMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// AuthMiddleware validates JWT tokens and sets user_id in context
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Check query param for backward compatibility or specific use cases (like easy testing)
			// But strictly speaking, we should prefer header.
			// Let's support "token" query param as seen in existing code, but maybe deprecate it?
			// The existing code mostly looked at "token" query param.
			// Ideally we move to Header.
			// Let's check query param "token" as fallback.
			token := r.URL.Query().Get("token")
			if token != "" {
				authHeader = "Bearer " + token
			}
		}

		if authHeader == "" {
			h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// Dev mode bypass
		if h.isDev && tokenString == "test-user" {
			ctx := context.WithValue(r.Context(), "user_id", "test-user")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		claims, err := h.tokenManager.ValidateToken(tokenString)
		if err != nil {
			h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid token"})
			return
		}

		slog.Info("JWT Claims", "claims", claims)

		userID, err := h.tokenManager.GetUserIDFromClaims(claims)
		if err != nil {
			h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid user ID in token"})
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware validates Admin API Key
func (h *Handler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.adminAPIKey == "" {
			// If no key configured, allow access (or maybe deny? existing code allowed if empty)
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")
		if key != h.adminAPIKey {
			h.WriteJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
// NewAdminSessionMiddleware creates a middleware that validates Admin JWT token
func NewAdminSessionMiddleware(uc *usecase.AdminVerifyTokenUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"status":"error","error":"Missing authorization header"}`))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"status":"error","error":"Invalid authorization header format"}`))
				return
			}

			tokenString := parts[1]

			if _, err := uc.Execute(r.Context(), tokenString); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"status":"error","error":"Invalid token or session expired"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminSessionMiddleware validates Admin JWT token and session
func (h *Handler) AdminSessionMiddleware(next http.Handler) http.Handler {
	return NewAdminSessionMiddleware(h.adminVerifyTokenUC)(next)
}
