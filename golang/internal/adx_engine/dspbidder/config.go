package dspbidder

import "time"

// BidderIndexConfig DSP索引管理器配置
type BidderIndexConfig struct {
	S3    S3StorageConfig `yaml:"s3"`
	Index IndexConfig     `yaml:"index"`
	Cache CacheConfig     `yaml:"cache"`
	Scan  ScanConfig      `yaml:"scan"`
}

// S3StorageConfig S3存储配置
type S3StorageConfig struct {
	Endpoint        string        `yaml:"endpoint"`
	AccessKeyID     string        `yaml:"access_key_id"`
	SecretAccessKey string        `yaml:"secret_access_key"`
	BucketName      string        `yaml:"bucket_name"`
	Prefix          string        `yaml:"prefix"`
	UseSSL          bool          `yaml:"use_ssl"`
	ScanInterval    time.Duration `yaml:"scan_interval"`
	Region          string        `yaml:"region"`
	PathStyle       bool          `yaml:"path_style"`
}

// IndexConfig 索引配置
type IndexConfig struct {
	BuildBatchSize    int           `yaml:"build_batch_size"`
	BuildTimeout      time.Duration `yaml:"build_timeout"`
	ReindexInterval   time.Duration `yaml:"reindex_interval"`
	AutoSwitch        bool          `yaml:"auto_switch"`
	EnablePersistence bool          `yaml:"enable_persistence"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Capacity        int           `yaml:"capacity"`
	TTL             time.Duration `yaml:"ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	EnableMetrics   bool          `yaml:"enable_metrics"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Interval    time.Duration `yaml:"interval"`
	RetryCount  int           `yaml:"retry_count"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
	Concurrency int           `yaml:"concurrency"`
	BatchSize   int           `yaml:"batch_size"`
}
