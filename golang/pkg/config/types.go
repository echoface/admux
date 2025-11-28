package config

import (
	"fmt"
	"time"
)

// BaseConfig 基础配置（所有服务通用）
type BaseConfig struct {
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

	// 监控配置
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    string `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	EnableJSON bool   `yaml:"enable_json"`
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

// DefaultBaseConfig 获取默认基础配置
func DefaultBaseConfig() *BaseConfig {
	return &BaseConfig{
		Host:            "localhost",
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		MaxConnections:  1000,
		EnablePprof:     false,
		ShutdownTimeout: 30 * time.Second,
	}
}

// GetAddress 获取服务器地址
func (c *BaseConfig) GetAddress() string {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
