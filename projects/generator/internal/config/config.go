package config

import "context"

// GenerateRequest represents a configuration generation request
type GenerateRequest struct {
	ExperimentID string
	Template     string
	Parameters   map[string]string
}

// GeneratedConfig represents a generated configuration
type GeneratedConfig struct {
	ID           string
	ExperimentID string
	Content      string
	Version      string
}

// Template represents a configuration template
type Template struct {
	Name        string
	Description string
	Content     string
	Version     string
}

// ConfigManager interface for configuration management
type ConfigManager interface {
	GenerateConfig(ctx context.Context, req GenerateRequest) (*GeneratedConfig, error)
	ValidateConfig(ctx context.Context, cfg string) error
	GetTemplate(ctx context.Context, name string) (*Template, error)
	ListTemplates(ctx context.Context) ([]*Template, error)
	CreateTemplate(ctx context.Context, tmpl *Template) error
	UpdateTemplate(ctx context.Context, name string, tmpl *Template) error
	DeleteTemplate(ctx context.Context, name string) error
}