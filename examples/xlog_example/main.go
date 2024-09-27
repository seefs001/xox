package main

import (
	"os"
	"time"

	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xlog_handlers"
)

func main() {
	xenv.Load()

	// Use the default console logger
	xlog.Info("This is an info message")
	xlog.Warn("This is a warning message")
	xlog.Error("This is an error message")

	// Add a file logger
	err := xlog.AddRotatingFile(xlog.FileConfig{
		Filename:   "app.log",
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxBackups: 3,
		MaxAge:     7, // 7 days
	})
	if err != nil {
		xlog.Error("Failed to add file logger", "error", err)
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

		xlog.Add(axiomHandler)
		xlog.Info("Axiom handler added")
	}

	// Now logs will be output to console, file, and Axiom (if configured)
	xlog.Info("This message will be logged to all configured handlers")

	// Use the Catch function to wrap operations that may produce errors
	xlog.Catch(func() error {
		// Simulate an operation that may produce an error
		return nil // or return an error
	})

	// Add a small delay to allow logs to be sent
	time.Sleep(2 * time.Second)

	// Gracefully shutdown all handlers
	xlog.Shutdown()
}
