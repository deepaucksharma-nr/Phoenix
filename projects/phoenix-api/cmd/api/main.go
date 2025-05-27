package main

import (
	"context"
	"database/sql" // Only for type reference in migration function
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	commonstore "github.com/phoenix/platform/pkg/common/store"
	"github.com/phoenix/platform/projects/phoenix-api/internal/api"
	"github.com/phoenix/platform/projects/phoenix-api/internal/config"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/phoenix/platform/projects/phoenix-api/internal/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg("No .env file found")
	}

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Load configuration
	cfg := config.Load()

	// Initialize store
	postgresStore, err := commonstore.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create postgres store")
	}
	defer postgresStore.Close()

	// Run migrations using the store's DB connection
	if err := runMigrations(postgresStore, cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	pipelineStore := store.NewPostgresPipelineDeploymentStore(postgresStore)

	// Initialize WebSocket hub
	zapLogger, _ := zap.NewProduction()
	hub := websocket.NewHub(zapLogger)
	go hub.Run()

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Create composite store
	compositeStore := store.NewCompositeStore(postgresStore, pipelineStore)
	
	// Initialize API server
	apiServer, err := api.NewServer(compositeStore, hub, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create API server")
	}

	// Start task queue background worker
	go apiServer.GetTaskQueue().Run(context.Background())

	// Start token blacklist cleanup job (runs every hour)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := compositeStore.CleanupExpiredTokens(ctx); err != nil {
					log.Error().Err(err).Msg("Failed to cleanup expired tokens")
				}
				cancel()
			}
		}
	}()

	// Setup routes
	apiServer.SetupRoutes(r)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info().Msg("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Server shutdown failed")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Starting Phoenix API server")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func runMigrations(dbProvider interface{ DB() *sql.DB }, databaseURL string) error {
	driver, err := postgres.WithInstance(dbProvider.DB(), &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}