// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/phoenix-vnext/platform/cmd/api-gateway/internal/handlers"
	"github.com/phoenix-vnext/platform/cmd/api-gateway/internal/middleware"
	"github.com/phoenix-vnext/platform/pkg/auth"
)

// TestAPIGatewayIntegration tests the API Gateway integration
func TestAPIGatewayIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	// Setup
	logger := zap.NewNop()
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	router := setupTestRouter(logger, jwtManager)

	// Test cases
	t.Run("Health Check", testHealthCheck(router))
	t.Run("Authentication", testAuthentication(router, jwtManager))
	t.Run("Protected Endpoints", testProtectedEndpoints(router, jwtManager))
	t.Run("Role-Based Access", testRoleBasedAccess(router, jwtManager))
}

func setupTestRouter(logger *zap.Logger, jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogging(logger))

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth handler
	authHandler := handlers.NewAuthHandler(jwtManager, logger)
	
	// API routes
	v1 := router.Group("/api/v1")
	
	// Auth routes
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/refresh", authHandler.Refresh)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(middleware.AuthConfig{
		JWTManager: jwtManager,
		SkipPaths:  []string{"/api/v1/auth/"},
		Logger:     logger,
	}))

	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)

	// Test endpoints
	protected.GET("/test/user", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "user endpoint"})
	})

	protected.GET("/test/admin", 
		middleware.RequireRoles("admin"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin endpoint"})
		},
	)

	return router
}

func testHealthCheck(router *gin.Engine) func(*testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	}
}

func testAuthentication(router *gin.Engine, jwtManager *auth.JWTManager) func(*testing.T) {
	return func(t *testing.T) {
		// Test login with valid credentials
		loginReq := map[string]string{
			"email":    "admin@phoenix.io",
			"password": "admin123",
		}
		body, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var loginResp handlers.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.Token)
		assert.Equal(t, "admin@phoenix.io", loginResp.User.Email)
		assert.Contains(t, loginResp.User.Roles, "admin")

		// Test login with invalid credentials
		invalidReq := map[string]string{
			"email":    "admin@phoenix.io",
			"password": "wrongpassword",
		}
		body, _ = json.Marshal(invalidReq)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func testProtectedEndpoints(router *gin.Engine, jwtManager *auth.JWTManager) func(*testing.T) {
	return func(t *testing.T) {
		// Get a valid token
		token, err := jwtManager.GenerateToken("user-001", "admin@phoenix.io", []string{"admin", "user"}, "tenant-001")
		require.NoError(t, err)

		// Test accessing protected endpoint with token
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var userInfo handlers.UserInfo
		err = json.Unmarshal(w.Body.Bytes(), &userInfo)
		require.NoError(t, err)
		assert.Equal(t, "admin@phoenix.io", userInfo.Email)

		// Test accessing protected endpoint without token
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Test with invalid token
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func testRoleBasedAccess(router *gin.Engine, jwtManager *auth.JWTManager) func(*testing.T) {
	return func(t *testing.T) {
		// Admin token
		adminToken, err := jwtManager.GenerateToken("user-001", "admin@phoenix.io", []string{"admin", "user"}, "tenant-001")
		require.NoError(t, err)

		// User token
		userToken, err := jwtManager.GenerateToken("user-002", "user@phoenix.io", []string{"user"}, "tenant-001")
		require.NoError(t, err)

		// Test admin accessing admin endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/test/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test user accessing admin endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/test/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		// Test user accessing user endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/test/user", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}