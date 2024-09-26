package main

import (
	"github.com/seefs001/xox/xlog"
)

func main() {
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

	// Now logs will be output to both console and file without duplication
	xlog.Info("This message will be logged to both console and file")

	// Use the Catch function to wrap operations that may produce errors
	xlog.Catch(func() error {
		// Simulate an operation that may produce an error
		return nil // or return an error
	})
}
