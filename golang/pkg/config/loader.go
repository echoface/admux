package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Loader 通用配置加载器
type Loader struct {
	ServiceName string
}

// NewLoader 创建配置加载器
func NewLoader(serviceName string) *Loader {
	return &Loader{
		ServiceName: serviceName,
	}
}

// Load 加载配置文件并解析到目标结构体
func (l *Loader) Load(configStruct interface{}) error {
	// 获取运行类型环境变量
	runType := os.Getenv("RUN_TYPE")
	if runType == "" {
		runType = "test"
	}

	// 验证运行类型
	if runType != "test" && runType != "prod" && runType != "dev" {
		return fmt.Errorf("invalid RUN_TYPE: %s, must be 'test', 'prod', or 'dev'", runType)
	}

	// 构建配置文件路径
	configDir := l.getConfigDir()
	configFile := filepath.Join(configDir, fmt.Sprintf("%s.yaml", runType))

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configFile)
	}

	// 设置viper配置
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// 设置环境变量前缀
	v.SetEnvPrefix(l.ServiceName)
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// 解析到配置结构体
	if err := v.Unmarshal(configStruct); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	fmt.Printf("Loaded %s config from: %s (run_type: %s)\n", l.ServiceName, configFile, runType)

	return nil
}

// getConfigDir 获取配置文件目录
func (l *Loader) getConfigDir() string {
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
