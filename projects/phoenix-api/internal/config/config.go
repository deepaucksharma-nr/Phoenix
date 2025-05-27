package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	PrometheusURL  string
	PushgatewayURL string
	JWTSecret      string
	Environment    string
	Features       Features
	CostRates      CostRates
	Timeouts       Timeouts
}

type Features struct {
	UsePushgateway bool
}

type CostRates struct {
	MetricsIngestionPerMillion float64
	StorageRetentionPerGB      float64
	CPUCostPerCore             float64
	MemoryCostPerGB            float64
}

type Timeouts struct {
	AgentPollTimeout  time.Duration
	TaskAssignTimeout time.Duration
	HeartbeatInterval time.Duration
}

func Load() *Config {
	// Require critical secrets in production
	env := getEnv("ENVIRONMENT", "development")

	// Database URL construction from components for security
	dbURL := constructDatabaseURL()

	// JWT secret must be provided in production
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" && env == "production" {
		panic("JWT_SECRET must be set in production environment")
	}
	if jwtSecret == "" {
		// Only for development - generate a random secret
		jwtSecret = fmt.Sprintf("dev-secret-%d", time.Now().Unix())
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    dbURL,
		PrometheusURL:  getEnv("PROMETHEUS_URL", "http://localhost:9090"),
		PushgatewayURL: getEnv("PUSHGATEWAY_URL", "http://localhost:9091"),
		JWTSecret:      jwtSecret,
		Environment:    env,
		Features: Features{
			UsePushgateway: getEnvBool("USE_PUSHGATEWAY", false),
		},
		CostRates: CostRates{
			MetricsIngestionPerMillion: getEnvFloat("COST_METRICS_PER_MILLION", 50.0),
			StorageRetentionPerGB:      getEnvFloat("COST_STORAGE_PER_GB", 10.0),
			CPUCostPerCore:             getEnvFloat("COST_CPU_PER_CORE", 100.0),
			MemoryCostPerGB:            getEnvFloat("COST_MEMORY_PER_GB", 20.0),
		},
		Timeouts: Timeouts{
			AgentPollTimeout:  getEnvDuration("AGENT_POLL_TIMEOUT", 30*time.Second),
			TaskAssignTimeout: getEnvDuration("TASK_ASSIGN_TIMEOUT", 5*time.Minute),
			HeartbeatInterval: getEnvDuration("HEARTBEAT_INTERVAL", 1*time.Minute),
		},
	}
}

func constructDatabaseURL() string {
	// Get database connection components
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbName := getEnv("DATABASE_NAME", "phoenix")
	dbUser := getEnv("DATABASE_USER", "phoenix")
	dbPassword := getEnv("DATABASE_PASSWORD", "")
	dbSSLMode := getEnv("DATABASE_SSL_MODE", "disable")

	// In production, password must be provided
	if getEnv("ENVIRONMENT", "development") == "production" && dbPassword == "" {
		panic("DATABASE_PASSWORD must be set in production environment")
	}

	// Default password only for local development
	if dbPassword == "" {
		dbPassword = "phoenix"
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
