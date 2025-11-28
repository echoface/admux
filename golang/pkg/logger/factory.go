package logger

// LoggerType represents the type of logger implementation
type LoggerType string

const (
	Zap     LoggerType = "zap"
	Zerolog LoggerType = "zerolog"
)

var Default Logger = MustNew(Zap, DefaultConfig())

func MustNew(loggerType LoggerType, config Config) Logger {
	logger, err := New(loggerType, config)
	panicIfErr(err)
	return logger
}

// New creates a new logger based on the specified type and configuration
func New(loggerType LoggerType, config Config) (Logger, error) {
	switch loggerType {
	case Zerolog:
		return NewZerologLogger(config)
	case Zap:
		return NewZapLogger(config)
	default:
		// Default to zerolog
		return NewZerologLogger(config)
	}
}

// DefaultConfig returns a default configuration for the logger
func DefaultConfig() Config {
	return Config{
		Environment: Dev,
		LogLevel:    "info",
		LogFile:     "",
		MaxSize:     100,  // 100MB
		MaxBackups:  3,    // keep 3 backups
		MaxAge:      30,   // 30 days
		Compress:    true, // compress rotated files
	}
}

// NewProduction creates a logger configured for production environment
func NewProduction(loggerType LoggerType, logFile string) (Logger, error) {
	config := DefaultConfig()
	config.Environment = Prod
	config.LogLevel = "info"
	config.LogFile = logFile
	return New(loggerType, config)
}

// NewTest creates a logger configured for test environment
func NewTest(loggerType LoggerType, logFile string) (Logger, error) {
	config := DefaultConfig()
	config.Environment = Test
	config.LogLevel = "debug"
	config.LogFile = logFile
	return New(loggerType, config)
}

// NewDevelopment creates a logger configured for development environment
func NewDevelopment(loggerType LoggerType) (Logger, error) {
	config := DefaultConfig()
	config.Environment = Dev
	config.LogLevel = "debug"
	return New(loggerType, config)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
