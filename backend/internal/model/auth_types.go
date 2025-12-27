package model

// User represents a tenant/user in the system
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	OrgName      string `json:"organization_name"`
	HasAPIKey    bool   `json:"has_api_key"`
	EmailEnabled bool   `json:"email_enabled"`

	// New Field for Phase 2
	IsVerified bool `json:"is_verified"`
}

// SignupRequest is the payload for POST /api/signup
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"orgName"`
}

// LoginRequest is the payload for POST /api/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful login/signup
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateProfileRequest is the payload for PUT /api/profile
type UpdateProfileRequest struct {
	OrgName      string `json:"organization_name"`
	EmailEnabled bool   `json:"email_enabled"`
}

// NEW: Request for Password Reset (Step 1)
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// NEW: Request for Password Reset Completion (Step 2)
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// NEW: Request for Email Verification
type VerifyEmailRequest struct {
	Token string `json:"token"`
}
