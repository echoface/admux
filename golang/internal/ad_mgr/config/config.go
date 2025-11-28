package config

import (
	"fmt"
	"github.com/echoface/admux/pkg/config"
)

// ============================================================================
// Ad Manager 专用配置示例
// ============================================================================

// AdMgrConfig 广告管理器配置
type AdMgrConfig struct {
	// 嵌入基础配置
	config.BaseConfig `yaml:",inline"`

	// 广告管理器特定配置
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	Adapters []AdapterConfig `yaml:"adapters"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Engine  string `yaml:"engine"`  // redis, memory
	TTL     int    `yaml:"ttl"`     // 过期时间（秒）
	MaxSize int    `yaml:"max_size"` // 最大缓存条目
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // ssp, dsp
	Endpoint string `yaml:"endpoint"`
	Enabled  bool   `yaml:"enabled"`
}

// LoadAdMgrConfig 加载广告管理器配置
func LoadAdMgrConfig() (*AdMgrConfig, error) {
	loader := config.NewLoader("admgr")
	var cfg AdMgrConfig

	if err := loader.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load AdMgr config: %w", err)
	}

	return &cfg, nil
}
