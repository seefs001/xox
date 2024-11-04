package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xlog_handlers"
)

func main() {
	xenv.Load()

	// Set console format to put source at the beginning, followed by level, then time
	xlog.SetConsoleFormat("%s[%l] [%t] %m%a")

	xlog.Info("This is an info message", slog.String("key", "value"), slog.Int("int", 123), slog.Any("any", "any"))

	// Use the default console logger
	xlog.Info("This is an info message")
	xlog.Warn("This is a warning message")
	xlog.Error("This is an error message")

	// Add a rotating file logger
	err := xlog.AddRotatingFile(xlog.FileConfig{
		Filename:   "logs/rotating_app.log",
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxBackups: 3,
		MaxAge:     7, // 7 days
	})
	if err != nil {
		xlog.Error("Failed to add rotating file logger", "error", err)
	}

	// Add a fixed file logger
	err = xlog.AddFixedFile("logs/fixed_app.log", slog.LevelDebug)
	if err != nil {
		xlog.Error("Failed to add fixed file logger", "error", err)
	}

	// Add Axiom handler if environment variables are set
	axiomApiToken := os.Getenv("AXIOM_API_TOKEN")
	axiomDataset := os.Getenv("AXIOM_DATASET")
	if axiomApiToken != "" && axiomDataset != "" {
		xlog.Info("Axiom handler adding")
		axiomHandler := xlog_handlers.NewAxiomHandler(axiomApiToken, axiomDataset)

		// Enable debug mode for Axiom handler
		axiomHandler.SetDebug(true)
		axiomHandler.SetLogOptions(xhttpc.LogOptions{
			LogHeaders:      true,
			LogBody:         true,
			LogResponse:     true,
			MaxBodyLogSize:  1024,
			HeaderKeysToLog: []string{"Content-Type", "Authorization"},
		})

		// Set Axiom flush interval to 5 seconds (adjust as needed)
		axiomHandler.SetFlushInterval(5 * time.Second)

		xlog.Add(axiomHandler)
		xlog.Info("Axiom handler added")
	}

	// Now logs will be output to console, rotating file, fixed file, and Axiom (if configured)
	xlog.Info("This message will be logged to all configured handlers")

	// Use the Catch function to wrap operations that may produce errors
	xlog.Catch(func() error {
		// Simulate an operation that may produce an error
		return nil // or return an error
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a done channel to signal when shutdown is complete
	done := make(chan bool, 1)

	// Start a goroutine to handle shutdown
	go func() {
		// Wait for interrupt signal
		<-sigChan
		xlog.Info("Shutting down application")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Perform shutdown tasks
		if err := xlog.Shutdown(ctx); err != nil {
			xlog.Error("Failed to shutdown logger", "error", err)
		}

		// Signal that shutdown is complete
		done <- true
	}()

	// Main application loop
	for {
		select {
		case <-done:
			return
		default:
			// Your application logic here
			xlog.Info("Application is running")
			time.Sleep(1 * time.Second) // Changed from 5 seconds to 1 second for more frequent logging
		}
	}
}
