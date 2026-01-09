package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Test loading the default config file
	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if cfg == nil {
		t.Fatal("Config is nil")
	}
	
	// Check some default values
	if cfg.System.Name == "" {
		t.Error("System name should not be empty")
	}
	
	if cfg.Pipeline.Workers <= 0 {
		t.Error("Pipeline workers should be positive")
	}
	
	if cfg.Risk.DailyMaxLoss <= 0 {
		t.Error("Daily max loss should be positive")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	// Set environment variable
	testKey := "test-api-key-12345"
	os.Setenv("ALPHA_KIMI_API_KEY", testKey)
	defer os.Unsetenv("ALPHA_KIMI_API_KEY")
	
	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if cfg.AI.APIKey != testKey {
		t.Errorf("AI API Key = %v, want %v", cfg.AI.APIKey, testKey)
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Loading nonexistent config should fail")
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Name:        "Test",
			Version:     "1.0.0",
			Environment: "test",
			LogLevel:    "info",
		},
		DataSource: DataSourceConfig{
			Primary:  "eastmoney",
			Timeout:  10 * time.Second,
			MaxRetry: 3,
		},
		AI: AIConfig{
			Provider: "kimi",
			Model:    "moonshot-v1-8k",
			BaseURL:  "https://api.moonshot.cn/v1",
			APIKey:   "test-key",
			Timeout:  30 * time.Second,
			MaxRetry: 2,
		},
		Risk: RiskConfig{
			DailyMaxLoss:         50000,
			MaxSignalsPerDay:     20,
			MaxConsecutiveLosses: 3,
			MinConfidence:        0.6,
		},
		Pipeline: PipelineConfig{
			Workers:   3,
			BatchSize: 20,
			Limit:     100,
		},
		Sinks: SinksConfig{
			Console: ConsoleSinkConfig{
				Enabled: true,
				Pretty:  true,
			},
		},
		Filters: FiltersConfig{
			EnableBasic:     true,
			EnableLiquidity: false,
			MinVolume:       100000,
			MinAmount:       10000000,
		},
		Observability: ObservabilityConfig{
			EnableTracing: true,
			EnableMetrics: true,
		},
	}
	
	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}
}

func TestConfig_ValidateForRealMode(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Name:        "Test",
			Environment: "test",
		},
		AI: AIConfig{
			APIKey: "", // Empty API key
		},
	}
	
	err := cfg.ValidateForRealMode()
	if err == nil {
		t.Error("ValidateForRealMode should fail with empty API key")
	}
}
