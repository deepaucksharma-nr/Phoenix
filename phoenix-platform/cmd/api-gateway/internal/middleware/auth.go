package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phoenix/platform/pkg/auth"
	"go.uber.org/zap"
)

// AuthConfig contains configuration for auth middleware
type AuthConfig struct {
	JWTManager      *auth.JWTManager
	SkipPaths       []string
	RequiredRoles   map[string][]string // path -> required roles
	Logger          *zap.Logger
}

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should skip authentication
		if shouldSkipAuth(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Extract token from header
		authHeader := c.GetHeader("Authorization")
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			config.Logger.Debug("failed to extract token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := config.JWTManager.ValidateToken(token)
		if err != nil {
			config.Logger.Debug("failed to validate token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check required roles for the path
		if requiredRoles, ok := config.RequiredRoles[c.Request.URL.Path]; ok {
			if !claims.HasAnyRole(requiredRoles...) {
				config.Logger.Debug("insufficient permissions",
					zap.String("user_id", claims.UserID),
					zap.Strings("user_roles", claims.Roles),
					zap.Strings("required_roles", requiredRoles),
				)
				c.JSON(http.StatusForbidden, gin.H{
					"error": "insufficient permissions",
				})
				c.Abort()
				return
			}
		}

		// Set claims in context
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)

		// Log authenticated request
		config.Logger.Debug("authenticated request",
			zap.String("user_id", claims.UserID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.Next()
	}
}

// shouldSkipAuth checks if the path should skip authentication
func shouldSkipAuth(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// RequireRoles creates a middleware that requires specific roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims",
			})
			c.Abort()
			return
		}

		if !userClaims.HasAnyRole(roles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClaims retrieves claims from gin context
func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	return claims.(*auth.Claims), true
}

// GetUserID retrieves user ID from gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetTenantID retrieves tenant ID from gin context
func GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	return tenantID.(string), true
}