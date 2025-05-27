package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthStatusCommand(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".phoenix", "config.yaml")
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	require.NoError(t, err)

	// Reset viper for clean test
	viper.Reset()
	viper.SetConfigFile(configPath)

	tests := []struct {
		name           string
		setupConfig    func()
		expectedOutput string
		expectedError  bool
	}{
		{
			name: "authenticated with valid token",
			setupConfig: func() {
				viper.Set("auth_token", "valid-token")
				viper.Set("api_url", "http://localhost:8080")
				viper.Set("username", "test@example.com")
				err := viper.WriteConfig()
				require.NoError(t, err)
			},
			expectedOutput: "Authenticated as: test@example.com",
			expectedError:  false,
		},
		{
			name: "not authenticated",
			setupConfig: func() {
				viper.Set("auth_token", "")
				viper.Set("api_url", "http://localhost:8080")
				err := viper.WriteConfig()
				require.NoError(t, err)
			},
			expectedOutput: "Not authenticated",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			tt.setupConfig()

			// Capture output
			var buf bytes.Buffer
			cmd := statusCmd
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := buf.String()
			assert.Contains(t, output, tt.expectedOutput)
		})
	}
}

func TestAuthLogoutCommand(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".phoenix", "config.yaml")
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	require.NoError(t, err)

	// Reset viper and set config
	viper.Reset()
	viper.SetConfigFile(configPath)
	viper.Set("auth_token", "test-token")
	viper.Set("username", "test@example.com")
	err = viper.WriteConfig()
	require.NoError(t, err)

	// Capture output
	var buf bytes.Buffer
	cmd := logoutCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command
	err = cmd.Execute()
	assert.NoError(t, err)

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Successfully logged out")

	// Verify token was cleared
	viper.ReadInConfig()
	assert.Empty(t, viper.GetString("auth_token"))
	assert.Empty(t, viper.GetString("username"))
}

// MockLoginServer creates a mock HTTP server for login testing
func mockLoginServer(t *testing.T, username, password string, success bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		if success {
			response := fmt.Sprintf(`{
				"token": "mock-jwt-token",
				"expires_at": "%s",
				"user": {
					"id": "user-123",
					"email": "%s",
					"name": "Test User"
				}
			}`, time.Now().Add(24*time.Hour).Format(time.RFC3339), username)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid credentials"}`))
		}
	}))
}