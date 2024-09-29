package xlog_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

	tempDir := t.TempDir()
	config := xlog.FileConfig{
		Filename:   filepath.Join(tempDir, "test.log"),
		MaxSize:    100, // 100 bytes for testing
		MaxBackups: 3,
		MaxAge:     1,
		Level:      slog.LevelDebug,
	}

	err := xlog.AddRotatingFile(config)
	require.NoError(err, "AddRotatingFile should not return an error")

	// Test logging at different levels
	for i := 0; i < 20; i++ { // Increased to 20 to ensure rotation
		xlog.Debug("Debug message")
		xlog.Info("Info message")
		xlog.Warn("Warn message")
		xlog.Error("Error message")
	}

	// Wait for rotation to occur
	time.Sleep(2 * time.Second)

	// Verify that log files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log*"))
	require.NoError(err, "Error listing log files")

	t.Logf("Found %d log files", len(files)) // Add this line for debugging
	for _, file := range files {
		t.Logf("Log file: %s", file) // Add this line for debugging
	}

	assert.GreaterOrEqual(len(files), 2, "Should have at least 2 log files (current and rotated)")

	// Verify content of the log files
	for _, file := range files {
		content, err := os.ReadFile(file)
		require.NoError(err, "Error reading log file")
		logContent := string(content)

		assert.Contains(logContent, "Debug message", "Log should contain debug message")
		assert.Contains(logContent, "Info message", "Log should contain info message")
		assert.Contains(logContent, "Warn message", "Log should contain warn message")
		assert.Contains(logContent, "Error message", "Log should contain error message")
	}
}

func TestFixedFileHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "fixed.log")

	err := xlog.AddFixedFile(filename, slog.LevelDebug)
	require.NoError(err, "AddFixedFile should not return an error")

	xlog.Debug("Debug message")
	xlog.Info("Info message")
	xlog.Warn("Warn message")
	xlog.Error("Error message")

	// Wait for flush to occur
	time.Sleep(time.Second)

	// Verify content of the log file
	content, err := os.ReadFile(filename)
	require.NoError(err, "Error reading log file")
	logContent := string(content)

	assert.Contains(logContent, "Debug message", "Log should contain debug message")
	assert.Contains(logContent, "Info message", "Log should contain info message")
	assert.Contains(logContent, "Warn message", "Log should contain warn message")
	assert.Contains(logContent, "Error message", "Log should contain error message")
}

func TestShutdown(t *testing.T) {
	assert := assert.New(t)

	tempDir := t.TempDir()
	rotatingConfig := xlog.FileConfig{
		Filename:   filepath.Join(tempDir, "rotating.log"),
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     1,
		Level:      slog.LevelDebug,
	}

	fixedFilename := filepath.Join(tempDir, "fixed.log")

	err := xlog.AddRotatingFile(rotatingConfig)
	assert.NoError(err, "AddRotatingFile should not return an error")

	err = xlog.AddFixedFile(fixedFilename, slog.LevelDebug)
	assert.NoError(err, "AddFixedFile should not return an error")

	xlog.Info("Test message before shutdown")

	// Wait for logs to be written
	time.Sleep(time.Second)

	err = xlog.Shutdown()
	assert.NoError(err, "Shutdown should not return an error")

	// Verify that logs were written
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log*"))
	assert.NoError(err, "Error listing log files")

	logFound := false
	for _, file := range files {
		content, err := os.ReadFile(file)
		assert.NoError(err, "Error reading log file")
		if strings.Contains(string(content), "Test message before shutdown") {
			logFound = true
			break
		}
	}
	assert.True(logFound, "Log should contain test message in one of the files")

	fixedContent, err := os.ReadFile(fixedFilename)
	assert.NoError(err, "Error reading fixed log file")
	assert.Contains(string(fixedContent), "Test message before shutdown", "Fixed log should contain test message")
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
