# Logger Package

A production-grade logging package that provides unified logging interface with support for multiple logging backends (zerolog and zap) and automatic log file rotation.

## Features

- **Unified Interface**: Single `Logger` interface for both zerolog and zap implementations
- **Environment Support**: Configurable for development, test, and production environments
- **File Rotation**: Automatic log file rotation with configurable size, backup count, and retention
- **Structured Logging**: Support for key-value pairs in log messages
- **Multiple Outputs**: Console output for development, file + console for production

## Installation

The package requires the following dependencies:

```bash
go get github.com/rs/zerolog
go get go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
```

## Usage

### Basic Usage

```go
import "github.com/echoface/admux/pkg/logger"

// Development environment (console output only)
log, err := logger.NewDevelopment(logger.Zerolog)
if err != nil {
    panic(err)
}

log.Info("Application started", "version", "1.0.0", "environment", "development")
```

### Production Environment with File Rotation

```go
// Production environment with file rotation
log, err := logger.NewProduction(logger.Zap, "/var/log/myapp/app.log")
if err != nil {
    panic(err)
}

log.Info("Production logger initialized",
    "logger", "zap",
    "environment", "production"
)
```

### Custom Configuration

```go
config := logger.Config{
    Environment: logger.Production,
    LogLevel:    "warn",
    LogFile:     "/var/log/myapp/error.log",
    MaxSize:     500,  // 500MB
    MaxBackups:  10,   // keep 10 backups
    MaxAge:      90,   // 90 days
    Compress:    true, // compress rotated files
}

log, err := logger.New(logger.Zerolog, config)
if err != nil {
    panic(err)
}
```

### Structured Logging

```go
log.Info("User action completed",
    "user_id", 12345,
    "action", "purchase",
    "amount", 99.99,
    "currency", "USD",
)

log.Error("Failed to process request",
    "error", "database connection timeout",
    "retry_count", 3,
    "endpoint", "/api/v1/orders",
)
```

## Logger Types

### Zerolog
- **Pros**: Fast, zero-allocation JSON logger
- **Cons**: Less feature-rich than zap
- **Best for**: High-performance applications

### Zap
- **Pros**: Highly configurable, structured logging
- **Cons**: Slightly slower than zerolog
- **Best for**: Applications requiring rich logging features

## Configuration

### Environment Settings

- **Development**: Console output with colored formatting, debug level
- **Test**: Console + file output, debug level
- **Production**: Console + file output, info level

### Log Levels

- `debug` - Detailed debug information
- `info` - General operational information
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (causes program exit)

### File Rotation Settings

- `MaxSize`: Maximum size in megabytes before rotation (default: 100MB)
- `MaxBackups`: Maximum number of old log files to retain (default: 3)
- `MaxAge`: Maximum number of days to retain old log files (default: 30)
- `Compress`: Whether to compress rotated log files (default: true)

## Examples

See `example.go` for complete usage examples.

## Testing

Run the tests to verify both logger implementations:

```bash
go test ./pkg/logger/... -v
```

## License

This package is part of the admux project.
