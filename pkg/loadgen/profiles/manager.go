package profiles

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ProfileConfig represents a load simulation profile configuration
type ProfileConfig struct {
	Description string                 `yaml:"description"`
	Aliases     []string               `yaml:"aliases"`
	Parameters  map[string]interface{} `yaml:"parameters"`
	ResourceImpact struct {
		CPU     string `yaml:"cpu"`
		Memory  string `yaml:"memory"`
		Network string `yaml:"network"`
	} `yaml:"resource_impact"`
}

// LoadProfilesConfig represents the complete load profiles configuration
type LoadProfilesConfig struct {
	Profiles map[string]ProfileConfig `yaml:"profiles"`
	Global   struct {
		OTLPEndpoint   string `yaml:"otlp_endpoint"`
		PushgatewayURL string `yaml:"pushgateway_url"`
		RateLimitMax   int    `yaml:"rate_limit_max"`
		CleanupTimeout string `yaml:"cleanup_timeout"`
	} `yaml:"global"`
	ResourceLimits struct {
		MaxCPUPercent int `yaml:"max_cpu_percent"`
		MaxMemoryMB   int `yaml:"max_memory_mb"`
		MaxOpenFiles  int `yaml:"max_open_files"`
	} `yaml:"resource_limits"`
	Telemetry struct {
		Enabled  bool     `yaml:"enabled"`
		Interval string   `yaml:"interval"`
		Metrics  []string `yaml:"metrics"`
	} `yaml:"telemetry"`
}

// ProfileManager manages load simulation profiles
type ProfileManager struct {
	config        *LoadProfilesConfig
	profileLookup map[string]string // Maps aliases to profile names
}

// NewProfileManager creates a new profile manager
func NewProfileManager(configPath string) (*ProfileManager, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config LoadProfilesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand environment variables
	config.Global.OTLPEndpoint = os.ExpandEnv(config.Global.OTLPEndpoint)
	config.Global.PushgatewayURL = os.ExpandEnv(config.Global.PushgatewayURL)

	// Build alias lookup map
	profileLookup := make(map[string]string)
	for name, profile := range config.Profiles {
		profileLookup[name] = name
		for _, alias := range profile.Aliases {
			profileLookup[alias] = name
		}
	}

	return &ProfileManager{
		config:        &config,
		profileLookup: profileLookup,
	}, nil
}

// GetProfile returns a profile configuration by name or alias
func (pm *ProfileManager) GetProfile(nameOrAlias string) (*ProfileConfig, error) {
	// Look up the canonical name
	canonicalName, exists := pm.profileLookup[nameOrAlias]
	if !exists {
		return nil, fmt.Errorf("unknown profile: %s", nameOrAlias)
	}

	profile, exists := pm.config.Profiles[canonicalName]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", canonicalName)
	}

	return &profile, nil
}

// ListProfiles returns all available profile names
func (pm *ProfileManager) ListProfiles() []string {
	names := make([]string, 0, len(pm.config.Profiles))
	for name := range pm.config.Profiles {
		names = append(names, name)
	}
	return names
}

// GetDefaultDuration returns the default duration for a profile
func (pm *ProfileManager) GetDefaultDuration(profileName string) (time.Duration, error) {
	profile, err := pm.GetProfile(profileName)
	if err != nil {
		return 0, err
	}

	if params, ok := profile.Parameters["duration_default"].(string); ok {
		return time.ParseDuration(params)
	}

	// Default to 5 minutes if not specified
	return 5 * time.Minute, nil
}

// GetMaxDuration returns the maximum allowed duration for a profile
func (pm *ProfileManager) GetMaxDuration(profileName string) (time.Duration, error) {
	profile, err := pm.GetProfile(profileName)
	if err != nil {
		return 0, err
	}

	if params, ok := profile.Parameters["duration_max"].(string); ok {
		return time.ParseDuration(params)
	}

	// Default to 1 hour if not specified
	return time.Hour, nil
}

// ValidateDuration checks if a duration is valid for a profile
func (pm *ProfileManager) ValidateDuration(profileName string, duration time.Duration) error {
	maxDuration, err := pm.GetMaxDuration(profileName)
	if err != nil {
		return err
	}

	if duration > maxDuration {
		return fmt.Errorf("duration %v exceeds maximum %v for profile %s", duration, maxDuration, profileName)
	}

	if duration < time.Second {
		return fmt.Errorf("duration must be at least 1 second")
	}

	return nil
}

// GetResourceLimits returns the configured resource limits
func (pm *ProfileManager) GetResourceLimits() (cpuPercent, memoryMB, openFiles int) {
	return pm.config.ResourceLimits.MaxCPUPercent,
		pm.config.ResourceLimits.MaxMemoryMB,
		pm.config.ResourceLimits.MaxOpenFiles
}

// GetOTLPEndpoint returns the configured OTLP endpoint
func (pm *ProfileManager) GetOTLPEndpoint() string {
	return pm.config.Global.OTLPEndpoint
}

// GetCleanupTimeout returns the cleanup timeout duration
func (pm *ProfileManager) GetCleanupTimeout() (time.Duration, error) {
	return time.ParseDuration(pm.config.Global.CleanupTimeout)
}

// GetProfileDescription returns a human-readable description of a profile
func (pm *ProfileManager) GetProfileDescription(profileName string) (string, error) {
	profile, err := pm.GetProfile(profileName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s (Impact: CPU=%s, Memory=%s, Network=%s)",
		profile.Description,
		profile.ResourceImpact.CPU,
		profile.ResourceImpact.Memory,
		profile.ResourceImpact.Network,
	), nil
}

// GetTelemetryConfig returns telemetry configuration
func (pm *ProfileManager) GetTelemetryConfig() (enabled bool, interval time.Duration, metrics []string) {
	if !pm.config.Telemetry.Enabled {
		return false, 0, nil
	}

	interval, _ = time.ParseDuration(pm.config.Telemetry.Interval)
	if interval == 0 {
		interval = 10 * time.Second
	}

	return true, interval, pm.config.Telemetry.Metrics
}