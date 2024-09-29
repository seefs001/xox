# xlog

`xlog` is a flexible and feature-rich logging package for Go applications. It provides colored console output, rotating file logs, and multi-handler support.

## Features

- Colored console output
- Rotating file logs
- Fixed file logs
- Multi-handler support
- Configurable log levels
- Source file and line information

## Installation

```bash
go get github.com/seefs001/xox/xlog
```

## Quick Start

```go
package main

import (
    "github.com/seefs001/xox/xlog"
    "log/slog"
)

func main() {
    // Set log level (optional, default is Debug)
    xlog.SetDefaultLogLevel(slog.LevelInfo)

    // Log messages
    xlog.Debug("This is a debug message")
    xlog.Info("This is an info message")
    xlog.Warn("This is a warning message")
    xlog.Error("This is an error message")

    // Use formatting
    xlog.Infof("Hello, %s!", "world")

    // Add attributes
    xlog.Info("User logged in", "userId", 123, "username", "johndoe")

    // Shutdown logging (important for file handlers)
    defer xlog.Shutdown()
}
```

## Configuration

### Setting Log Config

```go
xlog.SetLogConfig(xlog.LogConfig{
    IncludeFileAndLine: true,
    Level:              slog.LevelDebug,
})
```

### Adding a Rotating File Handler

```go
err := xlog.AddRotatingFile(xlog.FileConfig{
    Filename:   "/path/to/log/file.log",
    MaxSize:    10 * 1024 * 1024, // 10 MB
    MaxBackups: 5,
    MaxAge:     7, // days
    Level:      slog.LevelInfo,
})
if err != nil {
    // Handle error
}
```

### Adding a Fixed File Handler

```go
err := xlog.AddFixedFile("/path/to/fixed/log/file.log", slog.LevelDebug)
if err != nil {
    // Handle error
}
```

### Setting Console Format

```go
xlog.SetConsoleFormat("%s[%l] [%t] %m%a")
```

Placeholders:
- `%s`: Source (file and line)
- `%l`: Log level
- `%t`: Timestamp
- `%m`: Message
- `%a`: Attributes

## API Reference

### Logging Functions

- `Debug(msg string, args ...any)`
- `Debugf(format string, args ...any)`
- `Info(msg string, args ...any)`
- `Infof(format string, args ...any)`
- `Warn(msg string, args ...any)`
- `Warnf(format string, args ...any)`
- `Error(msg string, args ...any)`
- `Errorf(format string, args ...any)`

### Configuration Functions

- `SetLogConfig(config LogConfig)`
- `SetDefaultLogLevel(level slog.Level)`
- `AddRotatingFile(config FileConfig) error`
- `AddFixedFile(filename string, level slog.Level) error`
- `SetConsoleFormat(format string)`
- `SetLogger(logger *slog.Logger)`

### Utility Functions

- `Catch(f func() error)`: Wraps a function with error logging
- `Shutdown() error`: Shuts down all handlers

## Advanced Usage

### Using Multiple Handlers

```go
// Add console handler (default)
consoleHandler, _ := xlog.NewColorConsoleHandler(os.Stdout, nil)

// Add rotating file handler
rotatingHandler, _ := xlog.NewRotatingFileHandler(xlog.FileConfig{
    Filename:   "/path/to/rotating.log",
    MaxSize:    100 * 1024 * 1024, // 100 MB
    MaxBackups: 3,
    MaxAge:     7, // days
    Level:      slog.LevelInfo,
})

// Add fixed file handler
fixedHandler, _ := xlog.NewFixedFileHandler("/path/to/fixed.log", slog.LevelDebug)

// Create multi-handler
multiHandler := xlog.NewMultiHandler(consoleHandler, rotatingHandler, fixedHandler)

// Set logger with multi-handler
logger := slog.New(multiHandler)
xlog.SetLogger(logger)
```

### Error Handling with Catch

```go
xlog.Catch(func() error {
    // Your code that might return an error
    return someFunction()
})
```

## Best Practices

1. Always call `xlog.Shutdown()` before your application exits to ensure all logs are flushed and file handlers are closed properly.
2. Use structured logging with key-value pairs for better log analysis.
3. Configure appropriate log levels for different environments (e.g., Debug for development, Info for production).
4. Regularly rotate and archive log files to manage disk space and improve performance.
