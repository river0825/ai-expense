package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/riverlin/aiexpense/internal/usecase"
)

type AdminAuthHandler struct {
	loginUC  *usecase.AdminLoginUseCase
	verifyUC *usecase.AdminVerifyTokenUseCase
	logoutUC *usecase.AdminLogoutUseCase
}

func NewAdminAuthHandler(
	loginUC *usecase.AdminLoginUseCase,
	verifyUC *usecase.AdminVerifyTokenUseCase,
	logoutUC *usecase.AdminLogoutUseCase,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		loginUC:  loginUC,
		verifyUC: verifyUC,
		logoutUC: logoutUC,
	}
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid request"})
		return
	}

	result, err := h.loginUC.Execute(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAdminCredentials) {
			writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid credentials"})
			return
		}
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeAdminJSON(w, http.StatusOK, &Response{Status: "success", Data: map[string]interface{}{
		"token":      result.Token,
		"token_type": "Bearer",
		"expires_at": result.ExpiresAt,
	}})
}

func (h *AdminAuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Missing bearer token"})
		return
	}

	claims, err := h.verifyUC.Execute(r.Context(), token)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAdminToken) {
			writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid token"})
			return
		}
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeAdminJSON(w, http.StatusOK, &Response{Status: "success", Data: map[string]interface{}{
		"username":   claims.Username,
		"expires_at": claims.ExpiresAt,
	}})
}

func (h *AdminAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Missing bearer token"})
		return
	}

	if err := h.logoutUC.Execute(r.Context(), token); err != nil {
		writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	writeAdminJSON(w, http.StatusOK, &Response{Status: "success", Message: "Logged out"})
}

func (h *AdminAuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Missing bearer token"})
			return
		}

		if _, err := h.verifyUC.Execute(r.Context(), token); err != nil {
			if errors.Is(err, usecase.ErrInvalidAdminToken) {
				writeAdminJSON(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid token"})
				return
			}
			writeAdminJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
			return
		}

		next(w, r)
	}
}

func extractBearerToken(authorizationHeader string) string {
	parts := strings.SplitN(authorizationHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeAdminJSON(w http.ResponseWriter, statusCode int, payload *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
