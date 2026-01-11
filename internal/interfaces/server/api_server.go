package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bingaigc/gp/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// APIServer represents the HTTP API server
type APIServer struct {
	cfg    *config.Config
	port   string
	router *gin.Engine
	server *http.Server
}

// NewAPIServer creates a new API server instance
func NewAPIServer(cfg *config.Config, port string) *APIServer {
	// Set Gin mode based on environment
	if cfg != nil {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())

	return &APIServer{
		cfg:    cfg,
		port:   port,
		router: router,
	}
}

// Start starts the API server
func (s *APIServer) Start(ctx context.Context) error {
	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.port),
		Handler: s.router,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Info().Msgf("API server listening on :%s", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return s.shutdown()
	case err := <-errChan:
		return err
	}
}

// shutdown gracefully shuts down the server
func (s *APIServer) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down API server...")
	return s.server.Shutdown(ctx)
}

// setupRoutes sets up all API routes
func (s *APIServer) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/ready", s.handleReady)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Status endpoints
		v1.GET("/status", s.handleStatus)
		v1.GET("/version", s.handleVersion)

		// Job trigger endpoints
		v1.POST("/jobs/scan", s.handleScanJob)
		v1.POST("/jobs/predict", s.handlePredictJob)
		v1.POST("/jobs/backtest", s.handleBacktestJob)

		// Query endpoints
		v1.GET("/signals", s.handleGetSignals)
		v1.GET("/metrics", s.handleGetMetrics)
	}
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("ip", clientIP).
			Msg("HTTP request")
	}
}
