package logger

import (
	"os"
	"testing"
)

func TestZerologLogger(t *testing.T) {
	// Test development environment
	logger, err := NewDevelopment(Zerolog)
	if err != nil {
		t.Fatalf("Failed to create zerolog logger: %v", err)
	}

	logger.Info("Test info message", "key1", "value1", "key2", 123)
	logger.Debug("Test debug message", "debug_key", "debug_value")
	logger.Warn("Test warn message", "warn_key", "warn_value")
	logger.Error("Test error message", "error_key", "error_value")

	// Test production environment with file output
	config := Config{
		Environment: Prod,
		LogLevel:    "info",
		LogFile:     "test_production.log",
		MaxSize:     1, // 1MB for testing
		MaxBackups:  1,
		MaxAge:      1,
		Compress:    false,
	}

	prodLogger, err := New(Zerolog, config)
	if err != nil {
		t.Fatalf("Failed to create production zerolog logger: %v", err)
	}

	prodLogger.Info("Production test message", "env", "production")

	// Clean up test file
	defer os.Remove("test_production.log")
}

func TestZapLogger(t *testing.T) {
	// Test development environment
	logger, err := NewDevelopment(Zap)
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}

	logger.Info("Test info message", "key1", "value1", "key2", 123)
	logger.Debug("Test debug message", "debug_key", "debug_value")
	logger.Warn("Test warn message", "warn_key", "warn_value")
	logger.Error("Test error message", "error_key", "error_value")

	// Test production environment with file output
	config := Config{
		Environment: Prod,
		LogLevel:    "info",
		LogFile:     "test_production_zap.log",
		MaxSize:     1, // 1MB for testing
		MaxBackups:  1,
		MaxAge:      1,
		Compress:    false,
	}

	prodLogger, err := New(Zap, config)
	if err != nil {
		t.Fatalf("Failed to create production zap logger: %v", err)
	}

	prodLogger.Info("Production test message", "env", "production")

	// Clean up test file
	defer os.Remove("test_production_zap.log")
}

func TestLoggerInterface(t *testing.T) {
	// Test that both implementations satisfy the Logger interface
	var _ Logger = (*ZerologLogger)(nil)
	var _ Logger = (*ZapLogger)(nil)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Environment != Dev {
		t.Errorf("Expected default environment to be Development, got %v", config.Environment)
	}
	if config.LogLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got %v", config.LogLevel)
	}
	if config.MaxSize != 100 {
		t.Errorf("Expected default max size to be 100, got %v", config.MaxSize)
	}
}

func TestEnvironmentSpecificLoggers(t *testing.T) {
	// Test production logger
	prodLogger, err := NewProduction(Zerolog, "test_prod.log")
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}
	prodLogger.Info("Production logger test")
	defer os.Remove("test_prod.log")

	// Test test logger
	testLogger, err := NewTest(Zap, "test_test.log")
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	testLogger.Debug("Test logger debug message")
	defer os.Remove("test_test.log")

	// Test development logger
	devLogger, err := NewDevelopment(Zerolog)
	if err != nil {
		t.Fatalf("Failed to create development logger: %v", err)
	}
	devLogger.Debug("Development logger debug message")
}

