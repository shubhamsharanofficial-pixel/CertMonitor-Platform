package api

import (
	"encoding/json"
	"net/http"

	"cert-manager-backend/internal/model"
	"cert-manager-backend/internal/service"
)

type AuthHandler struct {
	Service service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{Service: svc}
}

// HandleSignup (Updated) - No longer returns token
func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req model.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.OrgName == "" {
		http.Error(w, "Email, Password, and Organization Name are required", http.StatusBadRequest)
		return
	}

	err := h.Service.Register(r.Context(), req)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success", "message":"Please check your email to verify your account."}`))
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	resp, err := h.Service.Login(r.Context(), req)
	if err != nil {
		if err.Error() == "email_not_verified" {
			http.Error(w, "Please verify your email address before logging in.", http.StatusForbidden)
			return
		}
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// NEW: HandleVerifyEmail
func (h *AuthHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req model.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Service.VerifyEmail(r.Context(), req.Token); err != nil {
		http.Error(w, "Verification failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"verified"}`))
}

// NEW: HandleForgotPassword
func (h *AuthHandler) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Always return OK even if email not found (Security)
	h.Service.RequestPasswordReset(r.Context(), req.Email)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"sent"}`))
}

// NEW: HandleResetPassword
func (h *AuthHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		http.Error(w, "Reset failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"reset"}`))
}

// HandleRegenerateKey handles POST /api/key/regenerate
func (h *AuthHandler) HandleRegenerateKey(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized: User identity missing", http.StatusUnauthorized)
		return
	}

	newKey, err := h.Service.RegenerateAPIKey(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to regenerate key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"api_key": newKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetProfile returns the current user's details and settings
func (h *AuthHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized: User identity missing", http.StatusUnauthorized)
		return
	}

	user, err := h.Service.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleUpdateProfile updates OrgName and Email Preferences
func (h *AuthHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized: User identity missing", http.StatusUnauthorized)
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateProfile(r.Context(), userID, req); err != nil {
		http.Error(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated profile or simple success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}
