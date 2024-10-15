# xlog

`xlog` is a flexible and feature-rich logging package for Go applications. It provides colored console output, rotating file logs, fixed file logs, and multi-handler support with easy integration for third-party logging services.

## Features

- Colored console output
- Rotating file logs
- Fixed file logs
- Multi-handler support
- Configurable log levels
- Source file and line information
- Easy integration with third-party logging services (e.g., Axiom)
- Custom color logging
- Graceful shutdown support

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

    // Set console format
    xlog.SetConsoleFormat("%s[%l] [%t] %m%a")

    // Log messages
    xlog.Debug("This is a debug message")
    xlog.Info("This is an info message")
    xlog.Warn("This is a warning message")
    xlog.Error("This is an error message")

    // Use formatting
    xlog.Infof("Hello, %s!", "world")

    // Add attributes
    xlog.Info("User logged in", "userId", 123, "username", "johndoe")

    // Use color logging
    xlog.RedLog(slog.LevelError, "This is a red error message")
    xlog.GreenLogf(slog.LevelInfo, "This is a green info message: %s", "Success!")

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

## Advanced Usage

### Using Multiple Handlers

```go
// Add rotating file handler
err := xlog.AddRotatingFile(xlog.FileConfig{
    Filename:   "/path/to/rotating.log",
    MaxSize:    100 * 1024 * 1024, // 100 MB
    MaxBackups: 3,
    MaxAge:     7, // days
    Level:      slog.LevelInfo,
})
if err != nil {
    // Handle error
}

// Add fixed file handler
err = xlog.AddFixedFile("/path/to/fixed.log", slog.LevelDebug)
if err != nil {
    // Handle error
}

// Add third-party handler (e.g., Axiom)
axiomHandler := xlog_handlers.NewAxiomHandler(apiToken, dataset)
xlog.Add(axiomHandler)
```

### Error Handling with Catch

```go
xlog.Catch(func() error {
    // Your code that might return an error
    return someFunction()
})
```

### Color Logging

```go
xlog.RedLog(slog.LevelError, "This is a red error message")
xlog.GreenLogf(slog.LevelInfo, "This is a green info message: %s", "Success!")
xlog.YellowLog(slog.LevelWarn, "This is a yellow warning")
xlog.BlueLogf(slog.LevelDebug, "Debug info: %v", someData)
```

## Best Practices

1. Always call `xlog.Shutdown()` before your application exits to ensure all logs are flushed and file handlers are closed properly.
2. Use structured logging with key-value pairs for better log analysis.
3. Configure appropriate log levels for different environments (e.g., Debug for development, Info for production).
4. Regularly rotate and archive log files to manage disk space and improve performance.
5. Use color logging to highlight important messages in console output.
6. Implement graceful shutdown to ensure all logs are properly written before the application exits.

## Example

Here's a comprehensive example demonstrating various features of the `xlog` package:

```go
package main

import (
    "os"
    "os/signal"
    "syscall"
    "time"

    "log/slog"

    "github.com/seefs001/xox/xenv"
    "github.com/seefs001/xox/xlog"
    "github.com/seefs001/xox/xlog_handlers"
)

func main() {
    xenv.Load()

    // Set console format
    xlog.SetConsoleFormat("%s[%l] [%t] %m%a")

    // Add rotating file logger
    err := xlog.AddRotatingFile(xlog.FileConfig{
        Filename:   "logs/rotating_app.log",
        MaxSize:    10 * 1024 * 1024, // 10MB
        MaxBackups: 3,
        MaxAge:     7, // 7 days
    })
    if err != nil {
        xlog.Error("Failed to add rotating file logger", "error", err)
    }

    // Add fixed file logger
    err = xlog.AddFixedFile("logs/fixed_app.log", slog.LevelDebug)
    if err != nil {
        xlog.Error("Failed to add fixed file logger", "error", err)
    }

    // Add Axiom handler if environment variables are set
    axiomApiToken := os.Getenv("AXIOM_API_TOKEN")
    axiomDataset := os.Getenv("AXIOM_DATASET")
    if axiomApiToken != "" && axiomDataset != "" {
        axiomHandler := xlog_handlers.NewAxiomHandler(axiomApiToken, axiomDataset)
        xlog.Add(axiomHandler)
    }

    // Log messages
    xlog.Info("Application started")
    xlog.Warn("This is a warning message")
    xlog.Error("This is an error message")

    // Use color logging
    xlog.GreenLog(slog.LevelInfo, "This is a green info message")
    xlog.YellowLog(slog.LevelWarn, "This is a yellow warning")

    // Use Catch for error handling
    xlog.Catch(func() error {
        // Simulate an operation that may produce an error
        return nil
    })

    // Set up graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        xlog.Info("Shutting down application")
        if err := xlog.Shutdown(); err != nil {
            xlog.Error("Failed to shutdown logger", "error", err)
        }
        os.Exit(0)
    }()

    // Main application loop
    for {
        xlog.Info("Application is running")
        time.Sleep(5 * time.Second)
    }
}
```

This example demonstrates how to set up multiple log handlers, use color logging, implement graceful shutdown, and integrate with third-party services like Axiom.

## API Reference

For a complete list of functions and their descriptions, please refer to the [GoDoc documentation](https://pkg.go.dev/github.com/seefs001/xox/xlog).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
