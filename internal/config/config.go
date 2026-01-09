package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	System       SystemConfig       `yaml:"system"`
	DataSource   DataSourceConfig   `yaml:"data_source"`
	AI           AIConfig           `yaml:"ai"`
	Risk         RiskConfig         `yaml:"risk"`
	Pipeline     PipelineConfig     `yaml:"pipeline"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	LogLevel    string `yaml:"log_level"`
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	Primary  string        `yaml:"primary"`
	Timeout  time.Duration `yaml:"timeout"`
	MaxRetry int           `yaml:"max_retry"`
}

// AIConfig AI配置
type AIConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Timeout  time.Duration `yaml:"timeout"`
	MaxRetry int `yaml:"max_retry"`
}

// RiskConfig 风控配置
type RiskConfig struct {
	DailyMaxLoss         float64 `yaml:"daily_max_loss"`
	MaxSignalsPerDay     int     `yaml:"max_signals_per_day"`
	MaxConsecutiveLosses int     `yaml:"max_consecutive_losses"`
	MinConfidence        float64 `yaml:"min_confidence"`
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	Workers   int `yaml:"workers"`
	BatchSize int `yaml:"batch_size"`
	Limit     int `yaml:"limit"`
}

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
	EnableTracing bool `yaml:"enable_tracing"`
	EnableMetrics bool `yaml:"enable_metrics"`
}

// Load 加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// 从环境变量覆盖
	if apiKey := os.Getenv("ALPHA_KIMI_API_KEY"); apiKey != "" {
		cfg.AI.APIKey = apiKey
	}

	return &cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.AI.APIKey == "" {
		return fmt.Errorf("AI API key is required")
	}
	if c.Pipeline.Workers <= 0 {
		return fmt.Errorf("pipeline workers must be > 0")
	}
	return nil
}
