package config

import (
	"fmt"
	"github.com/echoface/admux/pkg/config"
)

// ============================================================================
// DSP Engine 专用配置示例
// ============================================================================

// DSPEngineConfig DSP引擎配置
type DSPEngineConfig struct {
	// 嵌入基础配置
	config.BaseConfig `yaml:",inline"`

	// DSP引擎特定配置
	Auction    AuctionConfig    `yaml:"auction"`
	 bidders   []BidderConfig   `yaml:"bidders"`
	Analytics  AnalyticsConfig  `yaml:"analytics"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
}

// AuctionConfig 竞价配置
type AuctionConfig struct {
	Timeout      int `yaml:"timeout"`      // 竞价超时（毫秒）
	MinBidFloor  int `yaml:"min_bid_floor"` // 最低竞价
	MaxBidFloor  int `yaml:"max_bid_floor"` // 最高竞价
	WaitForFill  bool `yaml:"wait_for_fill"` // 是否等待填充
}

// DSP Bidder配置
type BidderConfig struct {
	ID         string `yaml:"id"`
	Name       string `yaml:"name"`
	Endpoint   string `yaml:"endpoint"`
	APIKey     string `yaml:"api_key"`
	QPSLimit   int    `yaml:"qps_limit"`
	Timeout    int    `yaml:"timeout"` // 毫秒
	Enabled    bool   `yaml:"enabled"`
	Weight     int    `yaml:"weight"` // 权重
}

// AnalyticsConfig 分析配置
type AnalyticsConfig struct {
	Enabled       bool   `yaml:"enabled"`
	DataRetention int    `yaml:"data_retention"` // 天
	BatchSize     int    `yaml:"batch_size"`
	FlushInterval int    `yaml:"flush_interval"` // 秒
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled   bool `yaml:"enabled"`
	Rate      int  `yaml:"rate"`      // 请求/秒
	Burst     int  `yaml:"burst"`     // 突发请求
	QueueSize int  `yaml:"queue_size"` // 队列大小
}

// LoadDSPEngineConfig 加载DSP引擎配置
func LoadDSPEngineConfig() (*DSPEngineConfig, error) {
	loader := config.NewLoader("dspengine")
	var cfg DSPEngineConfig

	if err := loader.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load DSP Engine config: %w", err)
	}

	return &cfg, nil
}
