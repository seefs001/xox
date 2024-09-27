package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xcolor"
	"github.com/seefs001/xox/xerror"
)

var (
	defaultLogger  *slog.Logger
	defaultHandler slog.Handler
	logConfig      LogConfig
	defaultLevel   slog.Level
	handlers       []slog.Handler
)

// LogConfig represents the configuration for logging.
type LogConfig struct {
	IncludeFileAndLine bool
	Level              slog.Level
}

func init() {
	logConfig = LogConfig{
		IncludeFileAndLine: true,
		Level:              slog.LevelInfo,
	}
	defaultLevel = logConfig.Level
	defaultHandler = x.Must1(NewColorConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	defaultLogger = slog.New(defaultHandler)
}

// SetLogConfig sets the logging configuration.
func SetLogConfig(config LogConfig) {
	logConfig = config
	defaultLevel = config.Level
	defaultHandler = x.Must1(NewColorConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	defaultLogger = slog.New(defaultHandler)
}

// SetDefaultLogLevel sets the default logging level.
func SetDefaultLogLevel(level slog.Level) {
	defaultLevel = level
	defaultHandler = x.Must1(NewColorConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	defaultLogger = slog.New(defaultHandler)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	log(slog.LevelDebug, msg, args...)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	log(slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string, args ...any) {
	log(slog.LevelInfo, msg, args...)
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	log(slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	log(slog.LevelWarn, msg, args...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	log(slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msg string, args ...any) {
	log(slog.LevelError, msg, args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	log(slog.LevelError, fmt.Sprintf(format, args...))
}

// log is a helper function to add file and line information if configured
func log(level slog.Level, msg string, args ...any) {
	if logConfig.IncludeFileAndLine {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// Use relative path for file
			rel, err := filepath.Rel(filepath.Dir(file), file)
			if err == nil {
				file = rel
			}
			// Format file:line to be clickable in most IDEs
			fileInfo := fmt.Sprintf("%s:%d", file, line)
			args = append(args, "source", fileInfo)
		}
	}
	defaultLogger.Log(context.Background(), level, msg, args...)
}

// ColorConsoleHandler implements a color console handler.
type ColorConsoleHandler struct {
	w      io.Writer
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

// NewColorConsoleHandler creates a new ColorConsoleHandler.
func NewColorConsoleHandler(w io.Writer, opts *slog.HandlerOptions) (*ColorConsoleHandler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: defaultLevel,
		}
	}
	return &ColorConsoleHandler{
		w:    w,
		opts: opts,
	}, nil
}

// Handle handles the log record with color output.
func (h *ColorConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()

	switch r.Level {
	case slog.LevelDebug:
		level = xcolor.Colorize(xcolor.Cyan, level)
	case slog.LevelInfo:
		level = xcolor.Colorize(xcolor.Green, level)
	case slog.LevelWarn:
		level = xcolor.Colorize(xcolor.Yellow, level)
	case slog.LevelError:
		level = xcolor.Colorize(xcolor.Red, level)
	}

	// Format output
	timeStr := r.Time.Format(time.RFC3339)
	msg := r.Message

	// Apply attributes
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, a.Value.String()))
		} else {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		}
		return true
	})
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	// Apply groups
	prefix := strings.Join(h.groups, ".")
	if prefix != "" {
		prefix += "."
	}

	_, err := fmt.Fprintf(h.w, "%s [%s] %s%s%s\n", timeStr, level, prefix, msg, attrStr)
	return xerror.Wrap(err, "failed to write log message")
}

// Enabled implements the slog.Handler interface.
func (h *ColorConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// WithAttrs implements the slog.Handler interface.
func (h *ColorConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return &newHandler
}

// WithGroup implements the slog.Handler interface.
func (h *ColorConsoleHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.groups = append(newHandler.groups, name)
	return &newHandler
}

// Add adds a new handler to the existing handlers
func Add(handler slog.Handler) {
	handlers = append(handlers, handler)
	if mh, ok := defaultHandler.(*MultiHandler); ok {
		// If defaultHandler is already a MultiHandler, add the new handler to it
		mh.handlers = append(mh.handlers, handler)
	} else {
		// If not, create a new MultiHandler with both handlers
		defaultHandler = NewMultiHandler(defaultHandler, handler)
	}
	defaultLogger = slog.New(defaultHandler)
}

// NewMultiHandler creates a new MultiHandler
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// MultiHandler implements a handler that writes to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

// Enabled implements the Handler interface
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle implements the Handler interface
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return xerror.Wrap(errs[0], "failed to handle log record")
	}
	return nil
}

// WithAttrs implements the Handler interface
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(handlers...)
}

// WithGroup implements the Handler interface
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(handlers...)
}

// FileConfig represents the configuration for file logging.
type FileConfig struct {
	Filename   string
	MaxSize    int64 // in bytes
	MaxBackups int
	MaxAge     int // in days
	Level      slog.Level
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
		return nil, xerror.Wrap(err, "failed to rotate log file")
	}
	h.Handler = slog.NewJSONHandler(h.file, &slog.HandlerOptions{Level: config.Level})
	return h, nil
}

// Handle handles the log record and rotates the file if necessary.
func (h *RotatingFileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.size >= h.config.MaxSize || time.Since(h.lastRotate) >= 24*time.Hour {
		if err := h.rotate(); err != nil {
			return xerror.Wrap(err, "failed to rotate log file")
		}
	}

	err := h.Handler.Handle(ctx, r)
	if err != nil {
		return xerror.Wrap(err, "failed to handle log record")
	}

	// Estimate the size increase
	h.size += int64(len(r.Message) + 100) // 100 is a rough estimate for metadata

	return nil
}

// rotate rotates the log file.
func (h *RotatingFileHandler) rotate() error {
	if h.file != nil {
		if err := h.file.Close(); err != nil {
			return xerror.Wrap(err, "failed to close log file")
		}
	}

	// Rotate existing files
	for i := h.config.MaxBackups - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", h.config.Filename, i)
		newName := fmt.Sprintf("%s.%d", h.config.Filename, i+1)
		if err := os.Rename(oldName, newName); err != nil && !os.IsNotExist(err) {
			return xerror.Wrap(err, "failed to rename log file")
		}
	}
	if err := os.Rename(h.config.Filename, h.config.Filename+".1"); err != nil && !os.IsNotExist(err) {
		return xerror.Wrap(err, "failed to rename current log file")
	}

	// Open new file
	file, err := os.OpenFile(h.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return xerror.Wrap(err, "failed to open new log file")
	}

	h.file = file
	h.size = 0
	h.lastRotate = time.Now()

	// Remove old files
	if h.config.MaxAge > 0 {
		cutoff := time.Now().Add(-time.Duration(h.config.MaxAge) * 24 * time.Hour)
		err := filepath.Walk(filepath.Dir(h.config.Filename), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.ModTime().Before(cutoff) {
				if err := os.Remove(path); err != nil {
					return xerror.Wrap(err, "failed to remove old log file")
				}
			}
			return nil
		})
		if err != nil {
			return xerror.Wrap(err, "failed to remove old log files")
		}
	}

	return nil
}

// AddRotatingFile adds a rotating file handler to the logger
func AddRotatingFile(config FileConfig) error {
	handler, err := NewRotatingFileHandler(config)
	if err != nil {
		return xerror.Wrap(err, "failed to create rotating file handler")
	}

	var newHandler slog.Handler
	if mh, ok := defaultHandler.(*MultiHandler); ok {
		// If defaultHandler is already a MultiHandler, add the new handler to it
		newHandlers := append(mh.handlers, handler)
		newHandler = NewMultiHandler(newHandlers...)
	} else {
		// If not, create a new MultiHandler with both handlers
		newHandler = NewMultiHandler(defaultHandler, handler)
	}

	defaultHandler = newHandler
	defaultLogger = slog.New(defaultHandler)
	return nil
}

// Catch wraps a function with error logging.
func Catch(f func() error) {
	if err := f(); err != nil {
		Error("Caught error", "error", err)
	}
}

// ShutdownHandler is an interface for handlers that need to be shut down.
type ShutdownHandler interface {
	Shutdown() error
}

// Shutdown shuts down all handlers that implement the ShutdownHandler interface.
func Shutdown() error {
	var errs []error
	for _, handler := range handlers {
		if sh, ok := handler.(ShutdownHandler); ok {
			if err := sh.Shutdown(); err != nil {
				errs = append(errs, xerror.Wrap(err, "failed to shutdown handler"))
			}
		}
	}
	if len(errs) > 0 {
		return xerror.Wrap(errs[0], "failed to shutdown one or more handlers")
	}
	return nil
}

// SetLogger sets the default logger
func SetLogger(logger *slog.Logger) {
	defaultLogger = logger
	defaultHandler = logger.Handler()
}
