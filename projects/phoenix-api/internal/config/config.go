package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	PrometheusURL  string
	PushgatewayURL string
	JWTSecret      string
	Environment    string
	Features       Features
}

type Features struct {
	UsePushgateway bool
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"),
		PrometheusURL:  getEnv("PROMETHEUS_URL", "http://localhost:9090"),
		PushgatewayURL: getEnv("PUSHGATEWAY_URL", "http://localhost:9091"),
		JWTSecret:      getEnv("JWT_SECRET", "phoenix-secret-key"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		Features: Features{
			UsePushgateway: getEnvBool("USE_PUSHGATEWAY", false),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}