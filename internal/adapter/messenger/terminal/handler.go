package terminal

import (
	"encoding/json"
	"net/http"
)

// Handler handles Terminal Chat requests for local testing
type Handler struct {
	useCase *TerminalUseCase
}

// NewHandler creates a new Terminal Chat handler
func NewHandler(useCase *TerminalUseCase) *Handler {
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

	// Process message
	response, err := h.useCase.HandleMessage(r.Context(), req.UserID, req.Message)
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
	json.NewEncoder(w).Encode(response)
}

// GetUserInfo retrieves user information
// Endpoint: GET /api/chat/terminal/user
func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "Method not allowed. Use GET.",
		})
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "Missing user_id query parameter",
		})
		return
	}

	user, err := h.useCase.GetUserInfo(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(TerminalResponse{
			Status:  "error",
			Message: "User not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TerminalResponse{
		Status:  "success",
		Message: "User information retrieved",
		Data:    user,
	})
}
