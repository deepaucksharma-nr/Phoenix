// Package logging provides structured logging for Phoenix Platform
package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a logging field
type Field = zap.Field

// Common field constructors
var (
	String = zap.String
	Int    = zap.Int
	Int64  = zap.Int64
	Float64 = zap.Float64
	Bool   = zap.Bool
	Error  = zap.Error
	Any    = zap.Any
	Time   = zap.Time
	Duration = zap.Duration
)

// logger wraps zap.Logger
type logger struct {
	zap *zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config Config) (Logger, error) {
	zapConfig := zap.NewProductionConfig()
	
	// Set log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	
	// Set output format
	if config.Format == "json" {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set output paths
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}
	
	// Set initial fields
	zapConfig.InitialFields = map[string]interface{}{
		"service": config.ServiceName,
		"version": config.Version,
		"env":     config.Environment,
	}
	
	// Build logger
	zapLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // Skip wrapper function
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}
	
	return &logger{zap: zapLogger}, nil
}

// Config represents logger configuration
type Config struct {
	Level        string   `json:"level" yaml:"level"`
	Format       string   `json:"format" yaml:"format"`
	ServiceName  string   `json:"service_name" yaml:"service_name"`
	Version      string   `json:"version" yaml:"version"`
	Environment  string   `json:"environment" yaml:"environment"`
	OutputPaths  []string `json:"output_paths" yaml:"output_paths"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "phoenix",
		Version:     "unknown",
		Environment: "development",
		OutputPaths: []string{"stdout"},
	}
}

// NewDevelopmentLogger creates a logger suitable for development
func NewDevelopmentLogger() Logger {
	config := DefaultConfig()
	config.Level = "debug"
	config.Format = "console"
	
	l, _ := NewLogger(config)
	return l
}

// NewProductionLogger creates a logger suitable for production
func NewProductionLogger(serviceName string) Logger {
	config := DefaultConfig()
	config.ServiceName = serviceName
	config.Environment = os.Getenv("ENVIRONMENT")
	if config.Environment == "" {
		config.Environment = "production"
	}
	
	l, _ := NewLogger(config)
	return l
}

// Logger methods implementation

func (l *logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *logger) With(fields ...Field) Logger {
	return &logger{zap: l.zap.With(fields...)}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	// Extract common context values
	fields := []Field{}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, String("trace_id", traceID.(string)))
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, String("user_id", userID.(string)))
	}
	
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, String("request_id", requestID.(string)))
	}
	
	return l.With(fields...)
}

// Global logger instance
var globalLogger Logger

func init() {
	globalLogger = NewDevelopmentLogger()
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(l Logger) {
	globalLogger = l
}

// Global logging functions

// Debug logs a debug message
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func With(fields ...Field) Logger {
	return globalLogger.With(fields...)
}

// WithContext creates a child logger with context values
func WithContext(ctx context.Context) Logger {
	return globalLogger.WithContext(ctx)
}