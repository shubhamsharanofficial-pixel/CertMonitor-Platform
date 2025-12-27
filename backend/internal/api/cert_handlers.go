package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cert-manager-backend/internal/model"
	"cert-manager-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type CertHandler struct {
	Service service.CertificateService
}

func NewCertHandler(svc service.CertificateService) *CertHandler {
	return &CertHandler{Service: svc}
}

// HandleIngest accepts the JSON payload from Agents.
func (h *CertHandler) HandleIngest(w http.ResponseWriter, r *http.Request) {
	var report model.AgentReport
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	if report.AgentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}
	if err := h.Service.ProcessReport(r.Context(), report); err != nil {
		http.Error(w, "Failed to process report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"success"}`))
}

// helper to try parsing ISO then Date
func parseDateOrTime(val string) (*time.Time, error) {
	// Try Full RFC3339 (ISO) first: "2025-12-20T15:00:00Z"
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t, nil
	}
	// Fallback to Date Only: "2025-12-20"
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("invalid date format")
}

// HandleListCerts fetches certificates with filtering options.
func (h *CertHandler) HandleListCerts(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()

	// 1. Pagination Defaults
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var opts []service.FilterOption
	opts = append(opts, service.WithPagination(limit, offset))

	// 2. Standard Filters
	if agentID := query.Get("agent_id"); agentID != "" {
		opts = append(opts, service.WithAgent(agentID))
	}
	if search := query.Get("search"); search != "" {
		opts = append(opts, service.WithSearch(search))
	}

	// 3. Date Filters
	var afterTime, beforeTime *time.Time
	if val := query.Get("valid_after"); val != "" {
		if t, err := parseDateOrTime(val); err == nil {
			afterTime = t
		}
	}
	if val := query.Get("valid_before"); val != "" {
		if t, err := parseDateOrTime(val); err == nil {
			beforeTime = t
		}
	}
	if afterTime != nil || beforeTime != nil {
		opts = append(opts, service.WithExpiryRange(afterTime, beforeTime))
	}

	// 4. Trust Filter (trusted=true/false)
	if trustStr := query.Get("trusted"); trustStr != "" {
		if isTrusted, err := strconv.ParseBool(trustStr); err == nil {
			opts = append(opts, service.WithTrust(&isTrusted))
		}
	}

	// 5. Status Filter (status=MISSING/ACTIVE)
	if status := query.Get("status"); status != "" {
		opts = append(opts, service.WithStatus(status))
	}

	// 6. Execute
	resp, err := h.Service.ListCertificates(r.Context(), userID, opts...)
	if err != nil {
		http.Error(w, "Failed to fetch certificates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// NEW: HandleDeleteInstance deletes a specific certificate instance
func (h *CertHandler) HandleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	instanceID := chi.URLParam(r, "id")
	if instanceID == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteInstance(r.Context(), userID, instanceID); err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

// NEW: HandlePruneMissing deletes all MISSING instances for the user
func (h *CertHandler) HandlePruneMissing(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.Service.DeleteAllMissingInstances(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to prune: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status":"pruned", "count":%d}`, count)))
}

// HandleGetStats returns the dashboard summary
func (h *CertHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized: User identity missing", http.StatusUnauthorized)
		return
	}

	stats, err := h.Service.GetDashboardStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
