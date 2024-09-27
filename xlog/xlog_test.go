package xlog_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColorConsoleHandler(t *testing.T) {
	assert := assert.New(t)
	buf := &bytes.Buffer{}
	handler, err := xlog.NewColorConsoleHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set the level to Debug
	})
	require.NoError(t, err, "NewColorConsoleHandler should not return an error")

	// Create a new logger with the test handler
	logger := slog.New(handler)
	xlog.SetLogger(logger) // Assuming you have a SetLogger function in xlog package

	xlog.Debug("This is a debug message")
	xlog.Info("This is an info message")
	xlog.Warn("This is a warning message")
	xlog.Error("This is an error message")

	output := buf.String()

	assert.Greater(len(output), 0, "Output should not be empty")
	assert.Contains(output, "DEBUG", "Output should contain DEBUG")
	assert.Contains(output, "INFO", "Output should contain INFO")
	assert.Contains(output, "WARN", "Output should contain WARN")
	assert.Contains(output, "ERROR", "Output should contain ERROR")
}

func TestRotatingFileHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	config := xlog.FileConfig{
		Filename:   "test.log",
		MaxSize:    1024, // 1KB for testing
		MaxBackups: 3,
		MaxAge:     1,
	}

	err := xlog.AddRotatingFile(config)
	require.NoError(err, "AddRotatingFile should not return an error")

	for i := 0; i < 100; i++ {
		xlog.Info("This is a test log message", "index", i)
	}

	// Verify that the log file was created
	_, err = os.Stat("test.log")
	assert.NoError(err, "Log file should exist")

	// Clean up
	os.Remove("test.log")
	for i := 1; i <= config.MaxBackups; i++ {
		os.Remove(fmt.Sprintf("test.log.%d", i))
	}
}

func TestCatch(t *testing.T) {
	assert := assert.New(t)
	buf := &bytes.Buffer{}
	handler, err := xlog.NewColorConsoleHandler(buf, nil)
	require.NoError(t, err, "NewColorConsoleHandler should not return an error")

	// Create a new logger with the test handler
	logger := slog.New(handler)
	xlog.SetLogger(logger) // Assuming you have a SetLogger function in xlog package

	xlog.Catch(func() error {
		return xerror.New("test error")
	})

	output := buf.String()
	assert.Contains(output, "Caught error", "Output should contain 'Caught error'")
	assert.Contains(output, "test error", "Output should contain 'test error'")
}
