package auth

import "github.com/karthikbalasubramani/netpilot-device-management/internal/user"

// AuthenticatedUser represents the verified identity associated with
// the current authenticated request.
//
// It contains only trusted information extracted from a successfully
// verified NetPilot access token.
//
// The raw JWT must never be stored in this structure.
type AuthenticatedUser struct {
	UserID  string
	Role    user.Role
	TokenID string
}
