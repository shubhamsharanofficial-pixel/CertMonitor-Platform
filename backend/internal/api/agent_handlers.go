package api

import (
	"cert-manager-backend/internal/assets"
	"cert-manager-backend/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type AgentHandler struct {
	Service service.AgentService
}

func NewAgentHandler(svc service.AgentService) *AgentHandler {
	return &AgentHandler{Service: svc}
}

// HandleListAgents returns a list of agents for the logged-in user
func (h *AgentHandler) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	agents, err := h.Service.ListAgents(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch agents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleDeleteAgent removes a specific agent
func (h *AgentHandler) HandleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		http.Error(w, "Agent ID required", http.StatusBadRequest)
		return
	}

	err := h.Service.DeleteAgent(r.Context(), userID, agentID)
	if err != nil {
		http.Error(w, "Failed to delete agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

// HandleGetInstallScript serves the dynamic install.sh
// This endpoint is PUBLIC (no auth required to download the template).
func (h *AgentHandler) HandleGetInstallScript(w http.ResponseWriter, r *http.Request) {
	// 1. Determine Base URL dynamically
	// Logic: Use "https" if TLS or proxy headers exist, otherwise "http".
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	// Host includes port if present (e.g., localhost:8080 or api.site.com)
	baseURL := fmt.Sprintf("%s://%s/api", scheme, r.Host)

	// 2. Load Script Template from embedded assets
	script := assets.GetInstallScript()

	// 3. Inject the detected Base URL
	// The script contains: BASE_URL="{{BASE_URL}}"
	finalScript := strings.ReplaceAll(script, "{{BASE_URL}}", baseURL)

	// 4. Serve as shell script
	w.Header().Set("Content-Type", "text/x-shellscript")
	w.Write([]byte(finalScript))
}
