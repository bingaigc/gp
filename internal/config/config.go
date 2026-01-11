package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	System        SystemConfig        `yaml:"system"`
	DataSource    DataSourceConfig    `yaml:"data_source"`
	AI            AIConfig            `yaml:"ai"`
	Risk          RiskConfig          `yaml:"risk"`
	Pipeline      PipelineConfig      `yaml:"pipeline"`
	Observability ObservabilityConfig `yaml:"observability"`
	Sinks         SinksConfig         `yaml:"sinks"`
	Filters       FiltersConfig       `yaml:"filters"`
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
	Provider string        `yaml:"provider"`
	Model    string        `yaml:"model"`
	BaseURL  string        `yaml:"base_url"`
	APIKey   string        `yaml:"api_key"`
	Timeout  time.Duration `yaml:"timeout"`
	MaxRetry int           `yaml:"max_retry"`
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

// SinksConfig 信号输出配置
type SinksConfig struct {
	Console  ConsoleSinkConfig  `yaml:"console"`
	JSONFile JSONFileSinkConfig `yaml:"json_file"`
	Webhook  WebhookSinkConfig  `yaml:"webhook"`
}

// ConsoleSinkConfig 控制台输出配置
type ConsoleSinkConfig struct {
	Enabled bool `yaml:"enabled"`
	Pretty  bool `yaml:"pretty"`
}

// JSONFileSinkConfig JSON文件输出配置
type JSONFileSinkConfig struct {
	Enabled   bool   `yaml:"enabled"`
	OutputDir string `yaml:"output_dir"`
	Pretty    bool   `yaml:"pretty"`
}

// WebhookSinkConfig Webhook输出配置
type WebhookSinkConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	Secret  string `yaml:"secret"`
	Format  string `yaml:"format"` // "dingtalk", "feishu", "generic"
}

// FiltersConfig 过滤器配置
type FiltersConfig struct {
	EnableBasic     bool     `yaml:"enable_basic"`
	EnableLiquidity bool     `yaml:"enable_liquidity"`
	MinVolume       int64    `yaml:"min_volume"`
	MinAmount       float64  `yaml:"min_amount"`
	Whitelist       []string `yaml:"whitelist"`
	Blacklist       []string `yaml:"blacklist"`
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
	// AI key is optional for demo mode
	if c.Pipeline.Workers <= 0 {
		return fmt.Errorf("pipeline workers must be > 0")
	}
	return nil
}

// ValidateForRealMode 验证实时模式配置
func (c *Config) ValidateForRealMode() error {
	if c.AI.APIKey == "" {
		return fmt.Errorf("AI API key is required for real-time mode")
	}
	if c.Pipeline.Workers <= 0 {
		return fmt.Errorf("pipeline workers must be > 0")
	}
	return nil
}
