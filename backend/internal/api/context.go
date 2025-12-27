package api

import (
	"context"
	"net/http"
)

// Define the key type internally
type contextKey string

// The key itself is private to this package
const userIDKey contextKey = "userID"

// SetUserInContext puts the ID into the request context (Used by Middleware)
func SetUserInContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

// GetUserFromContext retrieves the ID safely (Used by Handlers)
// Returns empty string if not found
func GetUserFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
