package logger

// Logger defines the standard behavior for our loggers.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
}

// Environment represents the deployment environment
type Environment string

const (
	Dev  Environment = "dev"
	Test Environment = "test"
	Prod Environment = "prod"
)

// Config holds the configuration for logger initialization
type Config struct {
	Environment Environment
	LogLevel    string
	LogFile     string
	MaxSize     int  // maximum size in megabytes before rotation
	MaxBackups  int  // maximum number of old log files to retain
	MaxAge      int  // maximum number of days to retain old log files
	Compress    bool // whether to compress rotated log files
}
