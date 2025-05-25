package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config manages CLI configuration
type Config struct {
	configPath string
	viper      *viper.Viper
}

// New creates a new config instance
func New() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	configPath := filepath.Join(home, ".phoenix", "config.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Set defaults
	v.SetDefault("api.endpoint", "http://localhost:8080")
	v.SetDefault("output.format", "table")

	// Try to read existing config
	v.ReadInConfig()

	return &Config{
		configPath: configPath,
		viper:      v,
	}
}

// GetConfigPath returns the path to the config file
func (c *Config) GetConfigPath() string {
	return c.configPath
}

// GetToken returns the stored authentication token
func (c *Config) GetToken() string {
	return c.viper.GetString("auth.token")
}

// SetToken stores the authentication token
func (c *Config) SetToken(token string) error {
	c.viper.Set("auth.token", token)
	return c.save()
}

// ClearToken removes the authentication token
func (c *Config) ClearToken() error {
	c.viper.Set("auth.token", "")
	return c.save()
}

// GetAPIEndpoint returns the API endpoint
func (c *Config) GetAPIEndpoint() string {
	endpoint := c.viper.GetString("api.endpoint")
	if endpoint == "" {
		return "http://localhost:8080"
	}
	return endpoint
}

// SetAPIEndpoint stores the API endpoint
func (c *Config) SetAPIEndpoint(endpoint string) error {
	c.viper.Set("api.endpoint", endpoint)
	return c.save()
}

// GetOutputFormat returns the output format
func (c *Config) GetOutputFormat() string {
	format := c.viper.GetString("output.format")
	if format == "" {
		return "table"
	}
	return format
}

// SetOutputFormat stores the output format
func (c *Config) SetOutputFormat(format string) error {
	c.viper.Set("output.format", format)
	return c.save()
}

// save writes the configuration to disk
func (c *Config) save() error {
	// Ensure directory exists
	dir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	if err := c.viper.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := c.viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}