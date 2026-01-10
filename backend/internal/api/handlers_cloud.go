package api

import (
	"cert-manager-backend/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CloudHandler manages the Agentless/Cloud Monitoring endpoints
type CloudHandler struct {
	Service service.AgentLessTargetService
}

// NewCloudHandler is the constructor
func NewCloudHandler(svc service.AgentLessTargetService) *CloudHandler {
	return &CloudHandler{
		Service: svc,
	}
}

// Request DTO
type AddTargetRequest struct {
	URL            string `json:"url"`
	FrequencyHours int    `json:"frequency_hours"`
}

type UpdateTargetRequest struct {
	FrequencyHours int `json:"frequency_hours"`
}

// POST /api/cloud/targets
func (h *CloudHandler) HandleAddTarget(w http.ResponseWriter, r *http.Request) {
	// 1. Parse
	var req AddTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Service Call
	target, err := h.Service.AddTarget(r.Context(), userID, req.URL, req.FrequencyHours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}

// PUT /api/cloud/targets/{id}
func (h *CloudHandler) HandleUpdateTarget(w http.ResponseWriter, r *http.Request) {
	var req UpdateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FrequencyHours < 1 {
		http.Error(w, "Frequency must be at least 1 hour", http.StatusBadRequest)
		return
	}

	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetID := chi.URLParam(r, "id")

	if err := h.Service.UpdateTarget(r.Context(), userID, targetID, req.FrequencyHours); err != nil {
		// Determine if it's a 404 or 500 based on error message, or generic 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated config or just 200 OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}

// GET /api/cloud/targets
func (h *CloudHandler) HandleListTargets(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targets, err := h.Service.ListTargets(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch targets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targets)
}

// DELETE /api/cloud/targets/{id}
func (h *CloudHandler) HandleDeleteTarget(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetID := chi.URLParam(r, "id")

	if err := h.Service.DeleteTarget(r.Context(), userID, targetID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
