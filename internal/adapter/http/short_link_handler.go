package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

type ShortLinkHandler struct {
	repo         domain.ShortLinkRepository
	dashboardURL string
}

func NewShortLinkHandler(repo domain.ShortLinkRepository, dashboardURL string) *ShortLinkHandler {
	return &ShortLinkHandler{
		repo:         repo,
		dashboardURL: dashboardURL,
	}
}

func (h *ShortLinkHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path (assuming mux handles pattern matching like /r/{id})
	// With Go 1.22+ mux, we can use r.PathValue("id")
	id := r.PathValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	link, err := h.repo.Get(ctx, id)
	if err != nil {
		// Generic 404 for security/simplicity
		http.NotFound(w, r)
		return
	}

	// Set cookie with the token
	http.SetCookie(w, &http.Cookie{
		Name:     "report_token",
		Value:    link.TargetToken,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),                           // Token validity (7 days)
		HttpOnly: false,                                                        // Must be false for JS to read if using localStorage backup or client-side auth logic
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https", // Auto-detect https
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to the actual report page (user dashboard)
	// The report page will check for the cookie or URL param
	redirectURL := fmt.Sprintf("%s/user/reports?token=%s", h.dashboardURL, link.TargetToken)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
