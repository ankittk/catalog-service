package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	tests := []struct {
		name         string
		userID       string
		email        string
		organization string
		role         string
		expectError  bool
	}{
		{
			name:         "valid token generation",
			userID:       "user-123",
			email:        "test@example.com",
			organization: "org-1",
			role:         "admin",
			expectError:  false,
		},
		{
			name:         "empty user ID",
			userID:       "",
			email:        "test@example.com",
			organization: "org-1",
			role:         "admin",
			expectError:  false, // JWT doesn't require non-empty user ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tt.userID, tt.email, tt.organization, tt.role)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate the generated token
			claims, err := jwtManager.ValidateToken(token)
			assert.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.email, claims.Email)
			assert.Equal(t, tt.organization, claims.Organization)
			assert.Equal(t, tt.role, claims.Role)
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	jwtManager := NewJWTManager(secretKey, tokenDuration)

	// Generate a valid token
	validToken, err := jwtManager.GenerateToken("user-123", "test@example.com", "org-1", "admin")
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "invalid token format",
			token:       "invalid.token.format",
			expectError: true,
		},
		{
			name:        "token with wrong signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
