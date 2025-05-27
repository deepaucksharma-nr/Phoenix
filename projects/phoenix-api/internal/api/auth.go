package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	internalModels "github.com/phoenix/platform/projects/phoenix-api/internal/models"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// RefreshRequest represents the token refresh request
type RefreshRequest struct {
	Token string `json:"token"`
}

// handleLogin handles user authentication
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user from store
	user, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Failed to get user")
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Warn().Str("username", req.Username).Msg("Invalid password attempt")
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	claims := services.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	token, _, expiresAt, err := s.jwtService.GenerateToken(claims)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Update last login
	if err := s.store.UpdateUserLastLogin(r.Context(), user.ID); err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to update last login")
		// Don't fail the request for this
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// handleRefreshToken handles token refresh
func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the existing token
	claims, err := s.jwtService.ValidateToken(req.Token)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Generate new token with same claims
	newToken, _, expiresAt, err := s.jwtService.GenerateToken(*claims)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		respondError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	response := LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     claims.Role,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get token from header
	token := extractToken(r)
	if token == "" {
		respondError(w, http.StatusBadRequest, "No token provided")
		return
	}

	// Revoke the token
	err := s.jwtService.RevokeToken(r.Context(), token, "User logout")
	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke token")
		// Still return success even if revocation fails
		// This prevents exposing internal errors to the user
	}

	// Get user info for logging (optional)
	claims, _ := s.jwtService.ValidateToken(token)
	if claims != nil {
		log.Info().Str("user_id", claims.UserID).Str("username", claims.Username).Msg("User logged out")
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// handleRegister handles user registration (optional, for development)
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Check if user already exists
	if _, err := s.store.GetUserByUsername(r.Context(), req.Username); err == nil {
		respondError(w, http.StatusConflict, "Username already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Create user
	user := &internalModels.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Default role
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.store.CreateUser(r.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token for immediate login
	claims := services.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	token, _, expiresAt, err := s.jwtService.GenerateToken(claims)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token for new user")
		// User was created, but token generation failed
		respondJSON(w, http.StatusCreated, map[string]string{
			"message": "User created successfully. Please login.",
			"user_id": user.ID,
		})
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	respondJSON(w, http.StatusCreated, response)
}

// handleGetProfile returns the current user's profile
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := s.store.GetUser(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user profile")
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	profile := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	respondJSON(w, http.StatusOK, profile)
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}
