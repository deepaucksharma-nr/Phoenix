package config

import "time"

type Config struct {
	APIURL         string
	HostID         string
	PollInterval   time.Duration
	ConfigDir      string
	PushgatewayURL string

	// NRDOT Collector configuration
	UseNRDOT       bool
	NRLicenseKey   string
	NROTLPEndpoint string
	CollectorType  string // "otel" or "nrdot"
}

// GetAPIEndpoint returns the full URL for an API endpoint
func (c *Config) GetAPIEndpoint(path string) string {
	return c.APIURL + "/api/v1/agent" + path
}
