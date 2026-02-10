package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beaesthetic/consent-service/internal/application"
	"github.com/beaesthetic/consent-service/internal/config"
	mongoinfra "github.com/beaesthetic/consent-service/internal/infra/mongo"
	httpport "github.com/beaesthetic/consent-service/internal/port/http"
	"github.com/beaesthetic/consent-service/pkg/health"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Int("port", cfg.Server.Port).
		Str("database", cfg.MongoDB.Database).
		Msg("Starting consent-service")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.ConnectionString))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to disconnect from MongoDB")
		}
	}()

	// Ping MongoDB to verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	log.Info().Msg("Connected to MongoDB")

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Initialize repositories
	policyRepo := mongoinfra.NewPolicyRepository(db)
	consentRepo := mongoinfra.NewConsentRepository(db)
	linkRepo := mongoinfra.NewConsentLinkRepository(db)

	// Create indexes
	if err := consentRepo.EnsureIndexes(); err != nil {
		log.Warn().Err(err).Msg("Failed to create consent indexes")
	}
	if err := linkRepo.EnsureIndexes(); err != nil {
		log.Warn().Err(err).Msg("Failed to create consent link indexes")
	}

	// Initialize services
	policyService := application.NewPolicyService(policyRepo)
	consentService := application.NewConsentService(consentRepo, policyRepo)
	linkService := application.NewLinkService(linkRepo, policyRepo, cfg.Server.BaseURL)

	// Initialize HTTP router
	router := httpport.NewRouter(policyService, consentService, linkService)

	// Register health check endpoint
	health.RegisterGinHealthCheck(router.Engine(), mongoClient)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router.Engine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("HTTP server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
