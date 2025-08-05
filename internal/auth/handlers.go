package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ankittk/catalog-service/internal/logger"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Organization string    `json:"organization"`
	Role         string    `json:"role"`
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	jwtManager *JWTManager
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtManager *JWTManager) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
	}
}

// Login handles user login and token generation
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Get().Warnw("Failed to decode login request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Organization == "" {
		http.Error(w, "Email, password, and organization are required", http.StatusBadRequest)
		return
	}

	// In a real application, you would validate credentials against a database
	// For demo purposes, we'll use a simple validation
	userID, role, err := h.validateCredentials(req.Email, req.Password, req.Organization)
	if err != nil {
		logger.Get().Warnw("Invalid credentials", "email", req.Email, "organization", req.Organization)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(userID, req.Email, req.Organization, role)
	if err != nil {
		logger.Get().Errorw("Failed to generate token", "error", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(h.jwtManager.TokenDuration())

	// Create response
	response := LoginResponse{
		Token:        token,
		ExpiresAt:    expiresAt,
		UserID:       userID,
		Email:        req.Email,
		Organization: req.Organization,
		Role:         role,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Get().Errorw("Failed to encode login response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Get().Infow("User logged in successfully",
		"user_id", userID,
		"email", req.Email,
		"organization", req.Organization,
		"role", role)
}

// validateCredentials validates user credentials
func (h *AuthHandler) validateCredentials(email, password, organization string) (string, string, error) {
	// Demo credentials: in production, use a proper authentication system
	demoUsers := map[string]struct {
		Password     string
		Organization string
		Role         string
	}{
		"admin@org1.com": {
			Password:     "admin123",
			Organization: "org-1",
			Role:         "admin",
		},
		"user@org1.com": {
			Password:     "user123",
			Organization: "org-1",
			Role:         "user",
		},
		"admin@org2.com": {
			Password:     "admin123",
			Organization: "org-2",
			Role:         "admin",
		},
		"user@org2.com": {
			Password:     "user123",
			Organization: "org-2",
			Role:         "user",
		},
		"admin@org3.com": {
			Password:     "admin123",
			Organization: "org-3",
			Role:         "admin",
		},
		"user@org3.com": {
			Password:     "user123",
			Organization: "org-3",
			Role:         "user",
		},
	}

	user, exists := demoUsers[email]
	if !exists {
		return "", "", ErrInvalidCredentials
	}

	if user.Password != password || user.Organization != organization {
		return "", "", ErrInvalidCredentials
	}

	// Generate a simple user ID
	userID := "user-" + email[:len(email)-4] // Remove @xxx.com

	return userID, user.Role, nil
}
