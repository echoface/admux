package config

import (
	"fmt"
	"time"
)

// ServerConfig 服务器配置结构
type ServerConfig struct {
	// 运行时信息
	RunType    string `yaml:"-"`    // test/prod
	ConfigFile string `yaml:"-"`    // 配置文件路径

	// 服务器配置
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	MaxConnections  int           `yaml:"max_connections"`
	EnablePprof     bool          `yaml:"enable_pprof"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`

	// 日志配置
	Logging LoggingConfig `yaml:"logging"`

	// Redis配置
	Redis RedisConfig `yaml:"redis"`

	// 监控配置
	Monitoring MonitoringConfig `yaml:"monitoring"`

	// SSP配置
	SSPs []SSPConfig `yaml:"ssps"`

	// Bidder配置
	Bidders []BidderConfig `yaml:"bidders"`

	// 其他配置项...
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level        string `yaml:"level"`
	Output       string `yaml:"output"`
	FilePath     string `yaml:"file_path"`
	MaxSize      string `yaml:"max_size"`
	MaxBackups   int    `yaml:"max_backups"`
	MaxAge       int    `yaml:"max_age"`
	Compress     bool   `yaml:"compress"`
	EnableJSON   bool   `yaml:"enable_json"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Prometheus  PrometheusConfig  `yaml:"prometheus"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Jaeger      JaegerConfig      `yaml:"jaeger"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	Namespace string `yaml:"namespace"`
	Subsystem string `yaml:"subsystem"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Endpoint string        `yaml:"endpoint"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
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

// NewDefaultConfig 创建默认配置（保持向后兼容）
func NewDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

// GetAddress 获取服务器地址
func (c *ServerConfig) GetAddress() string {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}