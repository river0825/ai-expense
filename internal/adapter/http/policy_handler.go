package http

import (
	"net/http"
)

// GetPolicy godoc
// @Summary Get a legal policy document
// @Description Retrieve a policy by its key (e.g., privacy_policy, terms_of_use)
// @Tags legal
// @Accept json
// @Produce json
// @Param key path string true "Policy Key"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /api/policies/{key} [get]
func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.PathValue("key") // Go 1.22+ path value

	if key == "" {
		h.WriteJSON(w, http.StatusBadRequest, &Response{Status: "error", Error: "Policy key is required"})
		return
	}

	policy, err := h.getPolicyUC.Execute(ctx, key)
	if err != nil {
		h.WriteJSON(w, http.StatusInternalServerError, &Response{Status: "error", Error: err.Error()})
		return
	}

	if policy == nil {
		h.WriteJSON(w, http.StatusNotFound, &Response{Status: "error", Error: "Policy not found"})
		return
	}

	h.WriteJSON(w, http.StatusOK, &Response{Status: "success", Data: policy})
}
