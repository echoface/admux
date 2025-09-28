package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadConfig 加载配置文件，支持RUN_TYPE环境变量
func LoadConfig(serviceName string) (*ServerConfig, error) {
	// 获取运行类型环境变量
	runType := os.Getenv("RUN_TYPE")
	if runType == "" {
		runType = "test" // 默认为test环境
	}

	// 验证运行类型
	if runType != "test" && runType != "prod" {
		return nil, fmt.Errorf("invalid RUN_TYPE: %s, must be 'test' or 'prod'", runType)
	}

	// 构建配置文件路径
	configDir := getConfigDir(serviceName)
	configFile := filepath.Join(configDir, fmt.Sprintf("%s.yaml", runType))

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configFile)
	}

	// 设置viper配置
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// 设置环境变量前缀
	v.SetEnvPrefix(serviceName)
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// 解析到配置结构体
	var config ServerConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置运行时信息
	config.RunType = runType
	config.ConfigFile = configFile

	fmt.Printf("Loaded %s config from: %s (run_type: %s)\n", serviceName, configFile, runType)

	return &config, nil
}

// getConfigDir 获取配置文件目录
func getConfigDir(serviceName string) string {
	// 优先级：
	// 1. CONFIG_PATH环境变量
	// 2. 相对于可执行文件的conf目录
	// 3. 相对于当前工作目录的conf目录

	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return filepath.Join(configPath, "conf")
	}

	// 获取可执行文件路径
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		confPath := filepath.Join(exeDir, "conf")
		if _, err := os.Stat(confPath); err == nil {
			return confPath
		}
	}

	// 默认使用当前工作目录
	return "conf"
}

// GetRunType 获取当前运行类型
func GetRunType() string {
	runType := os.Getenv("RUN_TYPE")
	if runType == "" {
		return "test"
	}
	return runType
}

// IsProduction 判断是否为生产环境
func IsProduction() bool {
	return GetRunType() == "prod"
}

// IsTest 判断是否为测试环境
func IsTest() bool {
	return GetRunType() == "test"
}