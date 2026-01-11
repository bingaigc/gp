package service

import (
	"context"
	"time"

	"github.com/bingaigc/gp/internal/config"
	"github.com/rs/zerolog/log"
)

// Scheduler handles periodic background tasks
type Scheduler struct {
	cfg *config.Config
}

// NewScheduler creates a new scheduler instance
func NewScheduler(cfg *config.Config) *Scheduler {
	return &Scheduler{
		cfg: cfg,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	log.Info().Msg("Scheduler started")

	// Get schedule interval from environment or use default
	scanInterval := getScheduleInterval("SCAN_INTERVAL", 1*time.Hour)
	predictInterval := getScheduleInterval("PREDICT_INTERVAL", 6*time.Hour)

	// Create tickers for periodic tasks
	scanTicker := time.NewTicker(scanInterval)
	predictTicker := time.NewTicker(predictInterval)
	defer scanTicker.Stop()
	defer predictTicker.Stop()

	// Run initial scan on startup if enabled
	if getBoolEnv("RUN_ON_STARTUP", true) {
		log.Info().Msg("Running initial scan on startup")
		s.runScanTask(ctx)
	}

	// Main scheduler loop
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Scheduler context cancelled, stopping")
			return nil

		case <-scanTicker.C:
			log.Info().Msg("Scheduled scan task triggered")
			s.runScanTask(ctx)

		case <-predictTicker.C:
			log.Info().Msg("Scheduled predict task triggered")
			s.runPredictTask(ctx)
		}
	}
}

// runScanTask executes the scan logic
func (s *Scheduler) runScanTask(ctx context.Context) {
	start := time.Now()
	log.Info().Msg("Starting scan task")

	// TODO: Implement actual scan logic
	// This should call the scan functionality from CLI commands
	// For now, we'll just simulate work
	select {
	case <-ctx.Done():
		log.Warn().Msg("Scan task cancelled")
		return
	case <-time.After(5 * time.Second):
		// Simulated work
	}

	duration := time.Since(start)
	log.Info().
		Dur("duration", duration).
		Msg("Scan task completed")
}

// runPredictTask executes the predict logic
func (s *Scheduler) runPredictTask(ctx context.Context) {
	start := time.Now()
	log.Info().Msg("Starting predict task")

	// TODO: Implement actual predict logic
	select {
	case <-ctx.Done():
		log.Warn().Msg("Predict task cancelled")
		return
	case <-time.After(3 * time.Second):
		// Simulated work
	}

	duration := time.Since(start)
	log.Info().
		Dur("duration", duration).
		Msg("Predict task completed")
}

// getScheduleInterval gets interval from environment or uses default
func getScheduleInterval(key string, defaultValue time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Warn().
			Str("key", key).
			Str("value", value).
			Err(err).
			Msg("Failed to parse interval, using default")
		return defaultValue
	}

	return duration
}
