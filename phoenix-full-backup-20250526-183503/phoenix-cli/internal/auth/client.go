package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles authentication with the Phoenix API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// NewClient creates a new authentication client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Login authenticates with the Phoenix API and returns a JWT token
func (c *Client) Login(username, password string) (string, error) {
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error != "" {
			return "", fmt.Errorf("authentication failed: %s", errorResp.Error)
		}
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return loginResp.Token, nil
}

// VerifyToken verifies a token and returns user information
func (c *Client) VerifyToken(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/auth/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &userInfo, nil
}