package terminal

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error)
}

// Handler handles Terminal Chat requests for local testing
type Handler struct {
	useCase MessageProcessor
}

// NewHandler creates a new Terminal Chat handler
func NewHandler(useCase MessageProcessor) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// TerminalRequest represents a Terminal Chat message request
type TerminalRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// TerminalResponse represents a Terminal Chat response
type TerminalResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleMessage processes a terminal chat message (HTTP POST)
// Endpoint: POST /api/chat/terminal
func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	// Parse request
	var req TerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "Missing required fields: user_id, message",
		})
		return
	}

	// Map to UserMessage
	userMsg := &domain.UserMessage{
		UserID:    req.UserID,
		Content:   req.Message,
		Source:    "terminal",
		Timestamp: time.Now(),
	}

	// Process message
	resp, err := h.useCase.Execute(r.Context(), userMsg)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TerminalResponse{
		Status:  "success",
		Message: resp.Text,
		Data: map[string]interface{}{
			"user_id":          req.UserID,
			"original_message": req.Message,
			"result":           resp.Data,
		},
	})
}

// GetUserInfo retrieves user information
// Endpoint: GET /api/chat/terminal/user
func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
