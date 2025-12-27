package service

import (
	"cert-manager-backend/internal/model"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type PostgresAuthService struct {
	DB        *sql.DB
	JWTSecret []byte
	EmailSvc  EmailService // Dependency Injection
}

func NewAuthService(db *sql.DB, jwtSecret string, emailSvc EmailService) *PostgresAuthService {
	return &PostgresAuthService{
		DB:        db,
		JWTSecret: []byte(jwtSecret),
		EmailSvc:  emailSvc,
	}
}

// --- Helper: Generate & Hash Token ---
func generateToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	rawToken := hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])
	return rawToken, hashedToken, nil
}

// Register (Updated for V2)
func (s *PostgresAuthService) Register(ctx context.Context, req model.SignupRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 1. Generate Verification Token
	rawToken, hashedToken, err := generateToken()
	if err != nil {
		return err
	}
	expiry := time.Now().Add(24 * time.Hour) // 24 hour expiry

	// 2. Insert User (is_verified = FALSE)
	query := `
        INSERT INTO users (email, password_hash, organization_name, is_verified, verification_token_hash, verification_token_expiry)
        VALUES ($1, $2, $3, FALSE, $4, $5)
        RETURNING id
    `
	var userID string
	err = s.DB.QueryRowContext(ctx, query, req.Email, string(hashedPassword), req.OrgName, hashedToken, expiry).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 3. Send Verification Email (Async to not block HTTP)
	go func() {
		if err := s.EmailSvc.SendVerificationEmail(req.Email, rawToken); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}()

	return nil
}

// Login (Updated for V2)
func (s *PostgresAuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	var user model.User
	var hash string

	query := `
        SELECT id, email, organization_name, password_hash, (api_key_hash IS NOT NULL), email_enabled, is_verified
        FROM users 
        WHERE email = $1
    `

	err := s.DB.QueryRowContext(ctx, query, req.Email).Scan(
		&user.ID, &user.Email, &user.OrgName, &hash, &user.HasAPIKey, &user.EmailEnabled, &user.IsVerified,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid credentials")
	} else if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// CHECK: Is Verified?
	if !user.IsVerified {
		return nil, errors.New("email_not_verified")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: user}, nil
}

// VerifyEmail (New)
func (s *PostgresAuthService) VerifyEmail(ctx context.Context, token string) error {
	// Hash the input token to match DB
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	query := `
		UPDATE users 
		SET is_verified = TRUE, verification_token_hash = NULL, verification_token_expiry = NULL
		WHERE verification_token_hash = $1 
		  AND verification_token_expiry > NOW()
	`
	result, err := s.DB.ExecContext(ctx, query, hashedToken)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("invalid or expired token")
	}

	return nil
}

// RequestPasswordReset (New)
func (s *PostgresAuthService) RequestPasswordReset(ctx context.Context, email string) error {
	// 1. Check if user exists (silently return nil if not to prevent enumeration)
	var userID string
	err := s.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil // Silent success
	} else if err != nil {
		return err
	}

	// 2. Generate Token
	rawToken, hashedToken, err := generateToken()
	if err != nil {
		return err
	}
	expiry := time.Now().Add(1 * time.Hour)

	// 3. Update User
	query := `UPDATE users SET reset_token_hash = $1, reset_token_expiry = $2 WHERE id = $3`
	_, err = s.DB.ExecContext(ctx, query, hashedToken, expiry, userID)
	if err != nil {
		return err
	}

	// 4. Send Email
	go func() {
		if err := s.EmailSvc.SendPasswordResetEmail(email, rawToken); err != nil {
			fmt.Printf("Failed to send reset email: %v\n", err)
		}
	}()

	return nil
}

// ResetPassword (New)
func (s *PostgresAuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	// Hash new password
	newPassHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		UPDATE users 
		SET password_hash = $1, reset_token_hash = NULL, reset_token_expiry = NULL
		WHERE reset_token_hash = $2 
		  AND reset_token_expiry > NOW()
	`
	result, err := s.DB.ExecContext(ctx, query, string(newPassHash), hashedToken)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("invalid or expired token")
	}

	return nil
}

// RegenerateAPIKey (Existing)
func (s *PostgresAuthService) RegenerateAPIKey(ctx context.Context, userID string) (string, error) {
	apiKeyBytes := make([]byte, 16)
	rand.Read(apiKeyBytes)
	plainApiKey := "crt_live_" + hex.EncodeToString(apiKeyBytes)

	hash := sha256.Sum256([]byte(plainApiKey))
	apiKeyHash := hex.EncodeToString(hash[:])

	query := `UPDATE users SET api_key_hash = $1 WHERE id = $2`
	_, err := s.DB.ExecContext(ctx, query, apiKeyHash, userID)
	if err != nil {
		return "", fmt.Errorf("failed to update api key: %w", err)
	}

	return plainApiKey, nil
}

// GetUsersByIDs (Existing)
func (s *PostgresAuthService) GetUsersByIDs(ctx context.Context, userIDs []string) (map[string]model.User, error) {
	result := make(map[string]model.User)
	if len(userIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT id, email, organization_name, (api_key_hash IS NOT NULL), email_enabled, is_verified
		FROM users 
		WHERE id = ANY($1)
	`

	rows, err := s.DB.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to bulk fetch users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.OrgName, &u.HasAPIKey, &u.EmailEnabled, &u.IsVerified); err != nil {
			return nil, err
		}
		result[u.ID] = u
	}

	return result, nil
}

// GetProfile (Existing)
func (s *PostgresAuthService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	query := `
        SELECT id, email, organization_name, (api_key_hash IS NOT NULL), email_enabled, is_verified 
        FROM users 
        WHERE id = $1
    `
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.OrgName, &user.HasAPIKey, &user.EmailEnabled, &user.IsVerified,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile (Existing)
func (s *PostgresAuthService) UpdateProfile(ctx context.Context, userID string, req model.UpdateProfileRequest) error {
	query := `
        UPDATE users 
        SET organization_name = $1, email_enabled = $2 
        WHERE id = $3
    `
	_, err := s.DB.ExecContext(ctx, query, req.OrgName, req.EmailEnabled, userID)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	return nil
}

func (s *PostgresAuthService) generateJWT(user model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}
