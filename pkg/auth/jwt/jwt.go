// Package jwt provides JWT token generation and validation for Phoenix Platform
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken indicates the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken indicates the token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidClaims indicates the token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims represents the JWT claims for Phoenix Platform
type Claims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	Namespace string   `json:"namespace,omitempty"`
}

// TokenGenerator generates JWT tokens
type TokenGenerator struct {
	secretKey      []byte
	issuer         string
	tokenDuration  time.Duration
	refreshDuration time.Duration
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(secretKey []byte, issuer string, tokenDuration, refreshDuration time.Duration) *TokenGenerator {
	return &TokenGenerator{
		secretKey:       secretKey,
		issuer:          issuer,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateToken generates a new JWT token
func (g *TokenGenerator) GenerateToken(userID, email string, roles []string, namespace string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    g.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(g.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		Namespace: namespace,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secretKey)
}

// GenerateRefreshToken generates a new refresh token
func (g *TokenGenerator) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    g.issuer,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(now.Add(g.refreshDuration)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secretKey)
}

// TokenValidator validates JWT tokens
type TokenValidator struct {
	secretKey []byte
	issuer    string
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(secretKey []byte, issuer string) *TokenValidator {
	return &TokenValidator{
		secretKey: secretKey,
		issuer:    issuer,
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *TokenValidator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Verify issuer
	if claims.Issuer != v.issuer {
		return nil, fmt.Errorf("%w: invalid issuer", ErrInvalidClaims)
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (v *TokenValidator) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidClaims
	}

	// Verify issuer
	if claims.Issuer != v.issuer {
		return "", fmt.Errorf("%w: invalid issuer", ErrInvalidClaims)
	}

	return claims.Subject, nil
}

// HasRole checks if the claims contain a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims contain any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the claims contain all of the specified roles
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}