package config

import (
	"fmt"
	"time"

	"github.com/echoface/admux/pkg/config"
)

// ============================================================================
// ADX Engine 专用配置
// ============================================================================

// AdxServerConfig ADX服务器配置
type AdxServerConfig struct {
	// 基础配置（嵌入或复制通用配置）
	config.BaseConfig `yaml:",inline"`

	// ADX Engine 特定配置
	Redis    RedisConfig    `yaml:"redis"`
	SSPs     []SSPConfig    `yaml:"ssps"`
	Bidders  []BidderConfig `yaml:"bidders"`
	S3       S3Config       `yaml:"s3"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	MaxRetries   int           `yaml:"max_retries"`
}

// SSPConfig SSP配置
type SSPConfig struct {
	ID       string        `yaml:"id"`
	Name     string        `yaml:"name"`
	Endpoint string        `yaml:"endpoint"`
	Protocol string        `yaml:"protocol"`
	QPSLimit int           `yaml:"qps_limit"`
	Timeout  time.Duration `yaml:"timeout"`
	Enabled  bool          `yaml:"enabled"`
}

// BidderConfig Bidder配置
type BidderConfig struct {
	ID         string        `yaml:"id"`
	Name       string        `yaml:"name"`
	Endpoint   string        `yaml:"endpoint"`
	Timeout    time.Duration `yaml:"timeout"`
	QPSLimit   int           `yaml:"qps_limit"`
	AuthToken  string        `yaml:"auth_token"`
	Enabled    bool          `yaml:"enabled"`
	RetryCount int           `yaml:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// S3Config S3存储配置
type S3Config struct {
	Endpoint        string        `yaml:"endpoint"`
	AccessKeyID     string        `yaml:"access_key_id"`
	SecretAccessKey string        `yaml:"secret_access_key"`
	BucketName      string        `yaml:"bucket_name"`
	Prefix          string        `yaml:"prefix"`
	UseSSL          bool          `yaml:"use_ssl"`
	ScanInterval    time.Duration `yaml:"scan_interval"`
	Region          string        `yaml:"region"`
}

// ============================================================================
// 配置加载器
// ============================================================================

// LoadAdxConfig 加载ADX服务器配置
func LoadAdxConfig() (*AdxServerConfig, error) {
	loader := config.NewLoader("adxserver")
	var cfg AdxServerConfig

	if err := loader.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load ADX config: %w", err)
	}

	return &cfg, nil
}

// DefaultAdxConfig 获取默认ADX配置
func DefaultAdxConfig() *AdxServerConfig {
	return &AdxServerConfig{
		BaseConfig: *config.DefaultBaseConfig(),
	}
}
