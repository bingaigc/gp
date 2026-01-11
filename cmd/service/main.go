package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bingaigc/gp/internal/config"
	"github.com/bingaigc/gp/internal/interfaces/server"
	"github.com/bingaigc/gp/internal/interfaces/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Version   = "2.0.0"
	BuildTime = "unknown"
)

// RunMode defines the service run mode
type RunMode string

const (
	ModeAll       RunMode = "all"       // Run both API and scheduler
	ModeAPI       RunMode = "api"       // Run only API server
	ModeScheduler RunMode = "scheduler" // Run only background scheduler
	ModeWorker    RunMode = "worker"    // Run only worker tasks
)

func main() {
	// Initialize logger
	initLogger()

	log.Info().
		Str("version", Version).
		Str("build_time", BuildTime).
		Msg("Starting Alpha Detector Service")

	// Get run mode from environment
	runMode := RunMode(getEnv("RUN_MODE", string(ModeAll)))
	log.Info().Str("run_mode", string(runMode)).Msg("Service run mode")

	// Load configuration
	cfg, err := config.Load(getEnv("CONFIG_PATH", "configs/config.yaml"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait group for goroutines
	var wg sync.WaitGroup

	// Start services based on run mode
	switch runMode {
	case ModeAll:
		startAPIServer(ctx, &wg, cfg)
		startScheduler(ctx, &wg, cfg)
	case ModeAPI:
		startAPIServer(ctx, &wg, cfg)
	case ModeScheduler:
		startScheduler(ctx, &wg, cfg)
	case ModeWorker:
		startWorker(ctx, &wg, cfg)
	default:
		log.Fatal().Str("mode", string(runMode)).Msg("Invalid run mode")
	}

	// Wait for shutdown signal
	<-sigChan
	log.Info().Msg("Shutdown signal received, gracefully stopping...")

	// Cancel context to stop all services
	cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("All services stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Warn().Msg("Shutdown timeout, forcing exit")
	}
}

func startAPIServer(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		port := getEnv("API_PORT", "8080")
		apiServer := server.NewAPIServer(cfg, port)

		log.Info().Str("port", port).Msg("Starting API server")

		if err := apiServer.Start(ctx); err != nil {
			log.Error().Err(err).Msg("API server error")
		}

		log.Info().Msg("API server stopped")
	}()
}

func startScheduler(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		scheduler := service.NewScheduler(cfg)

		log.Info().Msg("Starting background scheduler")

		if err := scheduler.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Scheduler error")
		}

		log.Info().Msg("Scheduler stopped")
	}()
}

func startWorker(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		worker := service.NewWorker(cfg)

		log.Info().Msg("Starting worker")

		if err := worker.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Worker error")
		}

		log.Info().Msg("Worker stopped")
	}()
}

func initLogger() {
	// Configure zerolog
	logLevel := getEnv("LOG_LEVEL", "info")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Use JSON format in production, pretty format in development
	if getEnv("LOG_FORMAT", "json") == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	log.Info().Str("level", logLevel).Msg("Logger initialized")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
