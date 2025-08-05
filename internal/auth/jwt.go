package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ankittk/catalog-service/internal/logger"
)

// Error definitions
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Claims represents the JWT claims
type Claims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
	Role         string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// TokenDuration returns the token duration
func (j *JWTManager) TokenDuration() time.Duration {
	return j.tokenDuration
}

// GenerateToken creates a new JWT token
func (j *JWTManager) GenerateToken(userID, email, organization, role string) (string, error) {
	claims := &Claims{
		UserID:       userID,
		Email:        email,
		Organization: organization,
		Role:         role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "catalog-service",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateSecretKey generates a random secret key
func GenerateSecretKey(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive, got %d", length)
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secret key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// HTTPMiddleware creates JWT authentication middleware for HTTP
func (j *JWTManager) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health check and OPTIONS requests
		if r.URL.Path == "/health" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			logger.Get().Warnw("Invalid authorization header", "error", err, "path", r.URL.Path)
			http.Error(w, "Unauthorized: Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := j.ValidateToken(tokenString)
		if err != nil {
			logger.Get().Warnw("Invalid JWT token", "error", err, "path", r.URL.Path)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GRPCUnaryInterceptor creates JWT authentication interceptor for gRPC
func (j *JWTManager) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for health check
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		tokenString, err := ExtractTokenFromHeader(authHeaders[0])
		if err != nil {
			logger.Get().Warnw("Invalid authorization header in gRPC", "error", err, "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header: %v", err)
		}

		// Validate token
		claims, err := j.ValidateToken(tokenString)
		if err != nil {
			logger.Get().Warnw("Invalid JWT token in gRPC", "error", err, "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "user", claims)
		return handler(ctx, req)
	}
}
