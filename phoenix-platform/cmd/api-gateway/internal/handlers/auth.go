package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phoenix/platform/pkg/auth"
	"go.uber.org/zap"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	jwtManager *auth.JWTManager
	logger     *zap.Logger
	// In production, this would use a user service
	// For demo, we'll use hardcoded users
	users map[string]User
}

// User represents a user for authentication
type User struct {
	ID       string
	Email    string
	Password string // In production, this would be hashed
	Roles    []string
	TenantID string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtManager *auth.JWTManager, logger *zap.Logger) *AuthHandler {
	// Demo users - in production, this would come from a database
	users := map[string]User{
		"admin@phoenix.io": {
			ID:       "user-001",
			Email:    "admin@phoenix.io",
			Password: "admin123", // Never store plaintext passwords in production!
			Roles:    []string{"admin", "user"},
			TenantID: "tenant-001",
		},
		"user@phoenix.io": {
			ID:       "user-002",
			Email:    "user@phoenix.io",
			Password: "user123",
			Roles:    []string{"user"},
			TenantID: "tenant-001",
		},
		"viewer@phoenix.io": {
			ID:       "user-003",
			Email:    "viewer@phoenix.io",
			Password: "viewer123",
			Roles:    []string{"viewer"},
			TenantID: "tenant-002",
		},
	}

	return &AuthHandler{
		jwtManager: jwtManager,
		logger:     logger,
		users:      users,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id,omitempty"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Find user
	user, exists := h.users[req.Email]
	if !exists || user.Password != req.Password {
		h.logger.Debug("login failed", zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, user.Roles, user.TenantID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	h.logger.Info("user logged in", 
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Roles:    user.Roles,
			TenantID: user.TenantID,
		},
	})
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Refresh token
	newToken, err := h.jwtManager.RefreshToken(req.Token)
	if err != nil {
		h.logger.Debug("token refresh failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid or expired token",
		})
		return
	}

	// Get claims from old token to return user info
	claims, _ := h.jwtManager.ValidateToken(req.Token)

	c.JSON(http.StatusOK, LoginResponse{
		Token:     newToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: UserInfo{
			ID:       claims.UserID,
			Email:    claims.Email,
			Roles:    claims.Roles,
			TenantID: claims.TenantID,
		},
	})
}

// Me returns current user information
func (h *AuthHandler) Me(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	userClaims := claims.(*auth.Claims)
	c.JSON(http.StatusOK, UserInfo{
		ID:       userClaims.UserID,
		Email:    userClaims.Email,
		Roles:    userClaims.Roles,
		TenantID: userClaims.TenantID,
	})
}

// Logout handles user logout
// In a stateful system, this would invalidate the token
// For JWT, logout is typically handled client-side
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	h.logger.Info("user logged out", zap.String("user_id", userID.(string)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}