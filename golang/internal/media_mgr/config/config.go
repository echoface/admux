package config

import (
	"fmt"
	"github.com/echoface/admux/pkg/config"
)

// ============================================================================
// Media Manager 专用配置示例
// ============================================================================

// MediaMgrConfig 媒体管理器配置
type MediaMgrConfig struct {
	// 嵌入基础配置
	config.BaseConfig `yaml:",inline"`

	// 媒体管理器特定配置
	Storage   StorageConfig   `yaml:"storage"`
	Process   ProcessConfig   `yaml:"process"`
	CDN       CDNConfig       `yaml:"cdn"`
	Transcode TranscodeConfig `yaml:"transcode"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string `yaml:"type"`       // local, s3, oss
	Path       string `yaml:"path"`       // 本地路径或存储bucket
	MaxSize    int    `yaml:"max_size"`   // 最大文件大小（MB）
	AllowTypes []string `yaml:"allow_types"` // 允许的文件类型
}

// ProcessConfig 处理配置
type ProcessConfig struct {
	Workers      int `yaml:"workers"`       // 工作线程数
	QueueSize    int `yaml:"queue_size"`    // 队列大小
	RetryTimes   int `yaml:"retry_times"`   // 重试次数
	RetryDelay   int `yaml:"retry_delay"`   // 重试延迟（秒）
}

// CDNConfig CDN配置
type CDNConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Domain     string `yaml:"domain"`
	Provider   string `yaml:"provider"` // cloudflare, aliyun
	CacheTTL   int    `yaml:"cache_ttl"` // 缓存时间（秒）
}

// TranscodeConfig 转码配置
type TranscodeConfig struct {
	Enabled        bool              `yaml:"enabled"`
	OutputFormats  []string          `yaml:"output_formats"` // mp4, hls, dash
	Qualities      map[string]int    `yaml:"qualities"`      // quality -> bitrate
	FFmpegPath     string            `yaml:"ffmpeg_path"`
	DefaultQuality string            `yaml:"default_quality"`
}

// LoadMediaMgrConfig 加载媒体管理器配置
func LoadMediaMgrConfig() (*MediaMgrConfig, error) {
	loader := config.NewLoader("mediamgr")
	var cfg MediaMgrConfig

	if err := loader.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load Media Mgr config: %w", err)
	}

	return &cfg, nil
}
