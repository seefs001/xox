package xlog

import (
	"bufio"
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
		Level:              slog.LevelDebug,
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
	format string
}

// NewColorConsoleHandler creates a new ColorConsoleHandler.
func NewColorConsoleHandler(w io.Writer, opts *slog.HandlerOptions) (*ColorConsoleHandler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: defaultLevel,
		}
	}
	return &ColorConsoleHandler{
		w:      w,
		opts:   opts,
		format: "%s[%l] [%t] %m%a", // Default format: source [level] [time] message attributes
	}, nil
}

// SetFormat sets the log format for ColorConsoleHandler.
// Placeholders: %t - time, %l - level, %s - source, %m - message, %a - attributes
func (h *ColorConsoleHandler) SetFormat(format string) {
	h.format = format
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
	var source string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			source = fmt.Sprintf("[%s] ", a.Value.String())
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

	// Format the log message
	logMsg := strings.ReplaceAll(h.format, "%s", source)
	logMsg = strings.ReplaceAll(logMsg, "%l", level)
	logMsg = strings.ReplaceAll(logMsg, "%t", timeStr)
	logMsg = strings.ReplaceAll(logMsg, "%m", prefix+msg)
	logMsg = strings.ReplaceAll(logMsg, "%a", attrStr)

	_, err := fmt.Fprintln(h.w, logMsg)
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

// SetConsoleFormat sets the format for the console handler
func SetConsoleFormat(format string) {
	if h, ok := defaultHandler.(*ColorConsoleHandler); ok {
		h.SetFormat(format)
	} else if mh, ok := defaultHandler.(*MultiHandler); ok {
		for _, handler := range mh.handlers {
			if h, ok := handler.(*ColorConsoleHandler); ok {
				h.SetFormat(format)
				break
			}
		}
	}
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
	config      FileConfig
	mu          sync.Mutex
	file        *os.File
	size        int64
	lastRotate  time.Time
	buffer      *bufio.Writer
	flushTicker *time.Ticker
	logChan     chan slog.Record
	done        chan struct{}
}

// NewRotatingFileHandler creates a new RotatingFileHandler.
func NewRotatingFileHandler(config FileConfig) (*RotatingFileHandler, error) {
	// Set default values if not provided
	if config.Filename == "" {
		return nil, xerror.New("filename must be specified")
	}
	if config.MaxSize == 0 {
		config.MaxSize = 100 * 1024 * 1024 // Default to 100MB
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 3 // Default to 3 backups
	}
	if config.MaxAge == 0 {
		config.MaxAge = 28 // Default to 28 days
	}
	if config.Level == 0 {
		config.Level = slog.LevelDebug // Default to Info level
	}

	h := &RotatingFileHandler{
		config:     config,
		lastRotate: time.Now(),
		logChan:    make(chan slog.Record, 1000),
		done:       make(chan struct{}),
	}

	if err := h.rotate(); err != nil {
		return nil, xerror.Wrap(err, "failed to rotate log file")
	}

	h.buffer = bufio.NewWriter(h.file)
	h.Handler = slog.NewJSONHandler(h.buffer, &slog.HandlerOptions{Level: config.Level})

	// Start background worker
	go h.processLogs()

	// Start periodic flushing
	h.flushTicker = time.NewTicker(1 * time.Second)
	go h.periodicFlush()

	return h, nil
}

// Handle queues the log record for processing.
func (h *RotatingFileHandler) Handle(ctx context.Context, r slog.Record) error {
	select {
	case h.logChan <- r:
		return nil
	default:
		return xerror.New("log channel full")
	}
}

// processLogs handles log records in the background.
func (h *RotatingFileHandler) processLogs() {
	for {
		select {
		case r, ok := <-h.logChan:
			if !ok {
				return
			}
			h.mu.Lock()
			if h.shouldRotate() {
				if err := h.rotate(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
					h.mu.Unlock()
					continue
				}
			}

			err := h.Handler.Handle(context.Background(), r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to handle log record: %v\n", err)
			}

			// Update size after successful write
			h.size += int64(len(r.Message) + 100) // 100 is a rough estimate for metadata
			h.mu.Unlock()
		case <-h.done:
			return
		}
	}
}

// shouldRotate checks if the file should be rotated
func (h *RotatingFileHandler) shouldRotate() bool {
	return !isSameDay(h.lastRotate, time.Now())
}

// isSameDay checks if two times are in the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// periodicFlush flushes the buffer periodically.
func (h *RotatingFileHandler) periodicFlush() {
	for {
		select {
		case <-h.flushTicker.C:
			h.mu.Lock()
			h.buffer.Flush()
			h.updateFileSize()
			h.mu.Unlock()
		case <-h.done:
			return
		}
	}
}

// updateFileSize updates the current file size
func (h *RotatingFileHandler) updateFileSize() {
	if info, err := h.file.Stat(); err == nil {
		h.size = info.Size()
	}
}

// rotate rotates the log file.
func (h *RotatingFileHandler) rotate() error {
	if h.file != nil {
		h.buffer.Flush()
		h.file.Close()
	}

	now := time.Now()
	newFilename := fmt.Sprintf("%s-%s", now.Format("2006-01-02"), filepath.Base(h.config.Filename))
	newFilePath := filepath.Join(filepath.Dir(h.config.Filename), newFilename)

	// Open new file
	file, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return xerror.Wrap(err, "failed to open new log file")
	}

	h.file = file
	h.size = 0
	h.lastRotate = now
	h.buffer = bufio.NewWriter(h.file)
	h.Handler = slog.NewJSONHandler(h.buffer, &slog.HandlerOptions{Level: h.config.Level})

	// Remove old files
	h.removeOldFiles()

	return nil
}

// removeOldFiles removes files older than MaxAge
func (h *RotatingFileHandler) removeOldFiles() {
	if h.config.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -h.config.MaxAge)
		dir := filepath.Dir(h.config.Filename)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				if strings.HasPrefix(filepath.Base(path), filepath.Base(h.config.Filename)) {
					if info.ModTime().Before(cutoff) {
						os.Remove(path) // Ignore errors
					}
				}
			}
			return nil
		})
	}
}

// Close closes the RotatingFileHandler.
func (h *RotatingFileHandler) Close() error {
	close(h.done)
	close(h.logChan)
	h.flushTicker.Stop()

	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.buffer.Flush(); err != nil {
		return xerror.Wrap(err, "failed to flush buffer")
	}
	return h.file.Close()
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
		if sh, ok := handler.(interface{ Close() error }); ok {
			if err := sh.Close(); err != nil {
				errs = append(errs, xerror.Wrap(err, "failed to close handler"))
			}
		}
	}

	// Close the default handler if it's a RotatingFileHandler or FixedFileHandler
	if h, ok := defaultHandler.(interface{ Close() error }); ok {
		if err := h.Close(); err != nil {
			errs = append(errs, xerror.Wrap(err, "failed to close default handler"))
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

// FixedFileHandler implements a handler that writes to a single file without rotation.
type FixedFileHandler struct {
	slog.Handler
	mu     sync.Mutex
	file   *os.File
	buffer *bufio.Writer
}

// NewFixedFileHandler creates a new FixedFileHandler.
func NewFixedFileHandler(filename string, level slog.Level) (*FixedFileHandler, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to open log file")
	}

	h := &FixedFileHandler{
		file: file,
	}

	h.buffer = bufio.NewWriter(h.file)
	h.Handler = slog.NewJSONHandler(h.buffer, &slog.HandlerOptions{Level: level})

	// Start periodic flushing
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			h.mu.Lock()
			h.buffer.Flush()
			h.mu.Unlock()
		}
	}()

	return h, nil
}

// Handle handles the log record.
func (h *FixedFileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.Handler.Handle(ctx, r)
}

// Close closes the FixedFileHandler.
func (h *FixedFileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.buffer.Flush(); err != nil {
		return xerror.Wrap(err, "failed to flush buffer")
	}
	return h.file.Close()
}

// AddFixedFile adds a fixed file handler to the logger
func AddFixedFile(filename string, level slog.Level) error {
	handler, err := NewFixedFileHandler(filename, level)
	if err != nil {
		return xerror.Wrap(err, "failed to create fixed file handler")
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
