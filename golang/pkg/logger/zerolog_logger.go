package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZerologLogger implements Logger interface using zerolog
type ZerologLogger struct {
	logger zerolog.Logger
}

// NewZerologLogger creates a new zerolog-based logger
func NewZerologLogger(config Config) (Logger, error) {
	var writers []io.Writer

	// Set log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure output based on environment
	if config.Environment == Prod || config.Environment == Test {
		// In production/test, log to both file and stdout
		if config.LogFile != "" {
			fileWriter := &lumberjack.Logger{
				Filename:   config.LogFile,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			}
			writers = append(writers, fileWriter)
		}
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// In development, log to console with pretty formatting
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		})
	}

	multiWriter := zerolog.MultiLevelWriter(writers...)

	logger := zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return &ZerologLogger{logger: logger}, nil
}

func (z *ZerologLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Debug().Fields(parseKeyValues(keysAndValues...)).Msg(msg)
}

func (z *ZerologLogger) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Info().Fields(parseKeyValues(keysAndValues...)).Msg(msg)
}

func (z *ZerologLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Warn().Fields(parseKeyValues(keysAndValues...)).Msg(msg)
}

func (z *ZerologLogger) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Error().Fields(parseKeyValues(keysAndValues...)).Msg(msg)
}

func (z *ZerologLogger) Fatal(msg string, keysAndValues ...interface{}) {
	z.logger.Fatal().Fields(parseKeyValues(keysAndValues...)).Msg(msg)
}

// parseKeyValues converts variadic key-value pairs to a map
func parseKeyValues(keysAndValues ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if ok {
				fields[key] = keysAndValues[i+1]
			}
		}
	}
	return fields
}

