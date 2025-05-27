package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/phoenix/platform/pkg/auth/jwt"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
)

// JWTService provides JWT authentication services
type JWTService struct {
	generator  *jwt.TokenGenerator
	validator  *jwt.TokenValidator
	store      store.Store
}

// JWTClaims represents simplified claims for the API
type JWTClaims struct {
	UserID   string
	Username string
	Email    string
	Role     string
	JTI      string // JWT ID for blacklist tracking
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey []byte, issuer string, store store.Store) *JWTService {
	tokenDuration := 24 * time.Hour       // Access token valid for 24 hours
	refreshDuration := 7 * 24 * time.Hour // Refresh token valid for 7 days

	return &JWTService{
		generator: jwt.NewTokenGenerator(secretKey, issuer, tokenDuration, refreshDuration),
		validator: jwt.NewTokenValidator(secretKey, issuer),
		store:     store,
	}
}

// GenerateToken generates a new JWT token with the given claims
func (s *JWTService) GenerateToken(claims JWTClaims) (token string, jti string, expiresAt time.Time, err error) {
	// Generate a unique JTI if not provided
	if claims.JTI == "" {
		claims.JTI = uuid.New().String()
	}
	
	// Convert single role to roles array
	roles := []string{claims.Role}
	
	// The JWT package automatically generates a JTI in the ID field
	token, err = s.generator.GenerateToken(
		claims.UserID,
		claims.Email,
		roles,
		"", // No namespace for now
	)
	
	if err != nil {
		return "", "", time.Time{}, err
	}
	
	// Parse the token to get the JTI
	parsedClaims, _ := s.validator.ValidateToken(token)
	if parsedClaims != nil {
		jti = parsedClaims.ID // JWT ID is stored in the ID field
	}
	
	expiresAt = time.Now().Add(24 * time.Hour)
	return token, jti, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	claims, err := s.validator.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is blacklisted
	if s.store != nil {
		blacklisted, err := s.store.IsTokenBlacklisted(context.Background(), claims.ID)
		if err != nil {
			// Log error but don't fail validation - fail open for availability
			// In production, you might want to fail closed for security
			// log.Error().Err(err).Str("jti", claims.ID).Msg("Failed to check token blacklist")
		} else if blacklisted {
			return nil, jwt.ErrInvalidToken
		}
	}

	// Convert from jwt.Claims to our simplified JWTClaims
	role := "user"
	if len(claims.Roles) > 0 {
		role = claims.Roles[0]
	}

	return &JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Subject, // Subject contains user ID, we'll need to adjust this
		Email:    claims.Email,
		Role:     role,
		JTI:      claims.ID,
	}, nil
}

// GenerateRefreshToken generates a new refresh token
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	return s.generator.GenerateRefreshToken(userID)
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (s *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	return s.validator.ValidateRefreshToken(tokenString)
}

// RevokeToken adds a token to the blacklist
func (s *JWTService) RevokeToken(ctx context.Context, tokenString string, reason string) error {
	// Parse the token to get claims
	claims, err := s.validator.ValidateToken(tokenString)
	if err != nil {
		// Even if token is expired, we might want to blacklist it
		// to prevent any edge cases
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}
	
	// Add to blacklist
	if s.store != nil {
		err = s.store.BlacklistToken(
			ctx,
			claims.ID,                   // JTI
			claims.UserID,               // User ID
			claims.ExpiresAt.Time,       // Expiration time
			reason,                      // Reason for revocation
		)
		if err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}
	}
	
	return nil
}