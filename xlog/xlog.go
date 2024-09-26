package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var defaultLogger = slog.New(NewColorConsoleHandler(os.Stdout, nil))

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	defaultLogger.Debug(fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	defaultLogger.Warn(fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	defaultLogger.Error(fmt.Sprintf(format, args...))
}

// ColorConsoleHandler implements a color console handler.
type ColorConsoleHandler struct {
	slog.Handler
	w io.Writer
}

// NewColorConsoleHandler creates a new ColorConsoleHandler.
func NewColorConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *ColorConsoleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColorConsoleHandler{
		Handler: slog.NewTextHandler(w, opts),
		w:       w,
	}
}

// Handle handles the log record with color output.
func (h *ColorConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()

	switch r.Level {
	case slog.LevelDebug:
		level = "\033[36m" + level + "\033[0m" // Cyan
	case slog.LevelInfo:
		level = "\033[32m" + level + "\033[0m" // Green
	case slog.LevelWarn:
		level = "\033[33m" + level + "\033[0m" // Yellow
	case slog.LevelError:
		level = "\033[31m" + level + "\033[0m" // Red
	}

	// Format output
	_, err := fmt.Fprintf(h.w, "%s [%s] %s\n", r.Time.Format(time.RFC3339), level, r.Message)
	if err != nil {
		return err
	}

	return h.Handler.Handle(ctx, r)
}

// Add adds a new handler to the logger.
func Add(handler slog.Handler) {
	defaultLogger = slog.New(handler)
}

// FileConfig represents the configuration for file logging.
type FileConfig struct {
	Filename   string
	MaxSize    int64 // in bytes
	MaxBackups int
	MaxAge     int // in days
}

// RotatingFileHandler implements a rotating file handler.
type RotatingFileHandler struct {
	slog.Handler
	config     FileConfig
	mu         sync.Mutex
	file       *os.File
	size       int64
	lastRotate time.Time
}

// NewRotatingFileHandler creates a new RotatingFileHandler.
func NewRotatingFileHandler(config FileConfig) (*RotatingFileHandler, error) {
	h := &RotatingFileHandler{
		config:     config,
		lastRotate: time.Now(),
	}
	if err := h.rotate(); err != nil {
		return nil, err
	}
	h.Handler = slog.NewJSONHandler(h.file, nil)
	return h, nil
}

// Handle handles the log record and rotates the file if necessary.
func (h *RotatingFileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.size >= h.config.MaxSize || time.Since(h.lastRotate) >= 24*time.Hour {
		if err := h.rotate(); err != nil {
			return err
		}
	}

	err := h.Handler.Handle(ctx, r)
	if err != nil {
		return err
	}

	// Estimate the size increase
	h.size += int64(len(r.Message) + 100) // 100 is a rough estimate for metadata

	return nil
}

// rotate rotates the log file.
func (h *RotatingFileHandler) rotate() error {
	if h.file != nil {
		h.file.Close()
	}

	// Rotate existing files
	for i := h.config.MaxBackups - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", h.config.Filename, i)
		newName := fmt.Sprintf("%s.%d", h.config.Filename, i+1)
		os.Rename(oldName, newName)
	}
	os.Rename(h.config.Filename, h.config.Filename+".1")

	// Open new file
	file, err := os.OpenFile(h.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	h.file = file
	h.size = 0
	h.lastRotate = time.Now()

	// Remove old files
	if h.config.MaxAge > 0 {
		cutoff := time.Now().Add(-time.Duration(h.config.MaxAge) * 24 * time.Hour)
		filepath.Walk(filepath.Dir(h.config.Filename), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.ModTime().Before(cutoff) {
				os.Remove(path)
			}
			return nil
		})
	}

	return nil
}

// AddRotatingFile adds a rotating file handler to the logger.
func AddRotatingFile(config FileConfig) error {
	handler, err := NewRotatingFileHandler(config)
	if err != nil {
		return err
	}
	Add(handler)
	return nil
}

// Catch wraps a function with error logging.
func Catch(f func() error) {
	if err := f(); err != nil {
		Error("Caught error", "error", err)
	}
}
