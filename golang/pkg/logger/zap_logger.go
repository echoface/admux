package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger implements Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zap-based logger
func NewZapLogger(config Config) (*ZapLogger, error) {
	var cores []zapcore.Core

	// Set log level
	level := getZapLevel(config.LogLevel)

	// Configure encoder based on environment
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if config.Environment == Dev {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configure output based on environment
	if config.Environment == Prod || config.Environment == Test {
		// In production/test, log to both file and stdout
		if config.LogFile != "" {
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   config.LogFile,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			})
			fileCore := zapcore.NewCore(encoder, fileWriter, level)
			cores = append(cores, fileCore)
		}

		stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, stdoutCore)
	} else {
		// In development, log to console only
		stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, stdoutCore)
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &ZapLogger{logger: logger}, nil
}

func (z *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Debug(msg, convertToZapFields(keysAndValues...)...)
}

func (z *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Info(msg, convertToZapFields(keysAndValues...)...)
}

func (z *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Warn(msg, convertToZapFields(keysAndValues...)...)
}

func (z *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Error(msg, convertToZapFields(keysAndValues...)...)
}

func (z *ZapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	z.logger.Fatal(msg, convertToZapFields(keysAndValues...)...)
}

// Sync flushes any buffered log entries
func (z *ZapLogger) Sync() error {
	return z.logger.Sync()
}

// getZapLevel converts string log level to zapcore.Level
func getZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertToZapFields converts variadic key-value pairs to zap fields
func convertToZapFields(keysAndValues ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keysAndValues[i+1]))
			}
		}
	}
	return fields
}

