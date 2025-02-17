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
	"sync/atomic"
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
	mu             sync.RWMutex // Added mutex for thread-safe access
	reqIDCounter   uint64
	bufferPool     = sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
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
	mu.Lock()
	defer mu.Unlock()

	logConfig = config
	defaultLevel = config.Level
	defaultHandler = x.Must1(NewColorConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	defaultLogger = slog.New(defaultHandler)
}

// SetDefaultLogLevel sets the default logging level.
func SetDefaultLogLevel(level slog.Level) {
	mu.Lock()
	defer mu.Unlock()

	defaultLevel = level
	defaultHandler = x.Must1(NewColorConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	defaultLogger = slog.New(defaultHandler)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	processedArgs := processArgs(args...)
	log(slog.LevelDebug, msg, processedArgs...)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	log(slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string, args ...any) {
	processedArgs := processArgs(args...)
	log(slog.LevelInfo, msg, processedArgs...)
}

// Infof logs a formatted info message.
func Infof(format string, args ...any) {
	log(slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	processedArgs := processArgs(args...)
	log(slog.LevelWarn, msg, processedArgs...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	log(slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msgOrErr interface{}, args ...any) {
	var msg string
	if err, ok := msgOrErr.(error); ok {
		msg = err.Error()
	} else if str, ok := msgOrErr.(string); ok {
		msg = str
	} else {
		msg = fmt.Sprint(msgOrErr)
	}
	processedArgs := processArgs(args...)
	log(slog.LevelError, msg, processedArgs...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	log(slog.LevelError, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal error message and then calls os.Exit(1).
func Fatal(msg string, args ...any) {
	processedArgs := processArgs(args...)
	log(slog.LevelError, msg, processedArgs...)
	os.Exit(1)
}

// Fatalf logs a formatted fatal error message and then calls os.Exit(1).
func Fatalf(format string, args ...any) {
	log(slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// processArgs converts mixed args (slog.Attr and key-value pairs) to a slice of any
func processArgs(args ...any) []any {
	if len(args) == 0 {
		return nil
	}

	var processed []any

	for i := 0; i < len(args); {
		// Check if the current argument is a slog.Attr
		if attr, ok := args[i].(slog.Attr); ok {
			processed = append(processed, attr.Key, attr.Value)
			i++
			continue
		}

		// Handle key-value pairs
		if i+1 < len(args) {
			// If next item is not a slog.Attr, treat current and next as key-value pair
			if _, ok := args[i+1].(slog.Attr); !ok {
				processed = append(processed, args[i], args[i+1])
				i += 2
				continue
			}
		}

		// Handle single remaining item
		if i < len(args) {
			processed = append(processed, args[i])
			i++
		}
	}

	return processed
}

func logWithAttrs(level slog.Level, msg string, attrs ...slog.Attr) {
	mu.RLock()
	defer mu.RUnlock()

	// Convert slog.Attr slice to any slice
	args := make([]any, len(attrs)*2)
	for i, attr := range attrs {
		args[i*2] = attr.Key
		args[i*2+1] = attr.Value
	}
	defaultLogger.Log(context.Background(), level, msg, args...)
}

// log is a helper function to add file and line information if configured
func log(level slog.Level, msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	ctx := context.Background()
	reqID := GenReqID()
	ctx = WithReqID(ctx, reqID)

	if logConfig.IncludeFileAndLine {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// Use relative path for file
			if rel, err := filepath.Rel(filepath.Dir(file), file); err == nil {
				file = rel
			}
			// Get buffer from pool
			buf := bufferPool.Get().(*strings.Builder)
			buf.Reset()
			defer bufferPool.Put(buf)

			buf.WriteString(fmt.Sprintf("%s:%d", file, line))
			args = append(args, "source", buf.String())
		}
	}

	args = append(args, "req_id", reqID)
	defaultLogger.Log(ctx, level, msg, args...)
}

// ColorConsoleHandler implements a color console handler.
type ColorConsoleHandler struct {
	w             io.Writer
	opts          *slog.HandlerOptions
	attrs         []slog.Attr
	groups        []string
	format        string
	maxMessageLen int // 0 means no limit
	maxAttrLen    int // 0 means no limit
}

// NewColorConsoleHandler creates a new ColorConsoleHandler.
func NewColorConsoleHandler(w io.Writer, opts *slog.HandlerOptions) (*ColorConsoleHandler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: defaultLevel,
		}
	}
	return &ColorConsoleHandler{
		w:             w,
		opts:          opts,
		format:        "[%l][%s][%t] %m%a", // Default format: [level][source][time] message attributes
		maxMessageLen: 0,                   // Default to no limit
		maxAttrLen:    0,                   // Default to no limit
	}, nil
}

// SetMaxLengths sets the maximum lengths for message and attributes.
func (h *ColorConsoleHandler) SetMaxLengths(maxMessageLen, maxAttrLen int) {
	h.maxMessageLen = maxMessageLen
	h.maxAttrLen = maxAttrLen
}

// SetFormat sets the log format for ColorConsoleHandler.
// Placeholders: %t - time, %l - level, %s - source, %m - message, %a - attributes
func (h *ColorConsoleHandler) SetFormat(format string) {
	h.format = format
}

// Handle handles the log record with color output.
func (h *ColorConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.w == nil {
		return xerror.New("writer is nil")
	}

	level := r.Level.String()
	levelColor := xcolor.White

	switch r.Level {
	case slog.LevelDebug:
		levelColor = xcolor.Cyan
	case slog.LevelInfo:
		levelColor = xcolor.Green
	case slog.LevelWarn:
		levelColor = xcolor.Yellow
	case slog.LevelError:
		levelColor = xcolor.Red
	}

	level = xcolor.Colorize(levelColor, level)

	timeStr := r.Time.Format(time.RFC3339)
	msg := r.Message
	if h.maxMessageLen > 0 && len(msg) > h.maxMessageLen {
		msg = msg[:h.maxMessageLen] + "..."
	}

	var attrs []string
	var source string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			source = a.Value.String()
		} else {
			attrStr := fmt.Sprintf("%s=%v", a.Key, a.Value.Any())
			if h.maxAttrLen > 0 && len(attrStr) > h.maxAttrLen {
				attrStr = attrStr[:h.maxAttrLen] + "..."
			}
			attrs = append(attrs, attrStr)
		}
		return true
	})

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	prefix := strings.Join(h.groups, ".")
	if prefix != "" {
		prefix += "."
	}

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

// SetConsoleMaxLengths sets the maximum lengths for message and attributes in the console handler
func SetConsoleMaxLengths(maxMessageLen, maxAttrLen int) {
	if h, ok := defaultHandler.(*ColorConsoleHandler); ok {
		h.SetMaxLengths(maxMessageLen, maxAttrLen)
	} else if mh, ok := defaultHandler.(*MultiHandler); ok {
		for _, handler := range mh.handlers {
			if h, ok := handler.(*ColorConsoleHandler); ok {
				h.SetMaxLengths(maxMessageLen, maxAttrLen)
				break
			}
		}
	}
}

// Add adds a new handler to the existing handlers
func Add(handler slog.Handler) {
	mu.Lock()
	defer mu.Unlock()

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

// LogEntry represents a structured log entry
type LogEntry struct {
	Time     time.Time
	Level    slog.Level
	Message  string
	ReqID    string
	Source   string
	Fields   map[string]interface{}
	Error    error
}

// NewLogEntry creates a new log entry
func NewLogEntry(level slog.Level, msg string) *LogEntry {
	return &LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  make(map[string]interface{}),
	}
}

// WithField adds a field to the log entry
func (e *LogEntry) WithField(key string, value interface{}) *LogEntry {
	e.Fields[key] = value
	return e
}

// WithError adds an error to the log entry
func (e *LogEntry) WithError(err error) *LogEntry {
	e.Error = err
	if err != nil {
		e.Fields["error"] = err.Error()
	}
	return e
}

// WithReqID adds a request ID to the log entry
func (e *LogEntry) WithReqID(reqID string) *LogEntry {
	e.ReqID = reqID
	return e
}

// WithSource adds source information to the log entry
func (e *LogEntry) WithSource(source string) *LogEntry {
	e.Source = source
	return e
}

// Log logs the entry
func (e *LogEntry) Log() {
	args := make([]interface{}, 0, len(e.Fields)*2+4)
	
	if e.ReqID != "" {
		args = append(args, "req_id", e.ReqID)
	}
	if e.Source != "" {
		args = append(args, "source", e.Source)
	}
	
	for k, v := range e.Fields {
		args = append(args, k, v)
	}
	
	log(e.Level, e.Message, args...)
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
	writeCount  int
	closed      bool
	closeMu     sync.Mutex
	rotateSize  int64
	currentNum  int
}

// NewRotatingFileHandler creates a new RotatingFileHandler.
func NewRotatingFileHandler(config FileConfig) (*RotatingFileHandler, error) {
	dir := filepath.Dir(config.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, xerror.Wrap(err, "failed to create log directory")
	}

	h := &RotatingFileHandler{
		config:     config,
		lastRotate: time.Now(),
		logChan:    make(chan slog.Record, 1000),
		done:       make(chan struct{}),
		rotateSize: config.MaxSize, // Changed initial rotate size to MaxSize
		currentNum: 1,
	}

	if err := h.rotate(); err != nil {
		return nil, xerror.Wrap(err, "failed to rotate log file")
	}

	h.Handler = slog.NewJSONHandler(h.file, &slog.HandlerOptions{
		Level: config.Level,
	})

	h.buffer = bufio.NewWriter(h.file)
	h.flushTicker = time.NewTicker(time.Second)

	go h.run()

	return h, nil
}

// Handle handles the log record.
func (h *RotatingFileHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return xerror.New("handler is closed")
	}

	// Format the log entry first to get its size
	buf := bufferPool.Get().(*strings.Builder)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Format log entry
	fmt.Fprintf(buf, "[%s] %s %s\n", 
		r.Time.Format("2006-01-02 15:04:05.000"),
		r.Level.String(),
		r.Message)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(buf, "\t%s=%v\n", a.Key, a.Value)
		return true
	})

	logEntry := buf.String()
	entrySize := int64(len(logEntry))

	// Check if we need to rotate before writing
	if h.size+entrySize > h.config.MaxSize {
		if err := h.rotate(); err != nil {
			return xerror.Wrap(err, "failed to rotate log file")
		}
	}

	if _, err := h.buffer.WriteString(logEntry); err != nil {
		return xerror.Wrap(err, "failed to write to buffer")
	}

	h.size += entrySize
	h.writeCount++

	// Flush buffer if we've written enough entries
	if h.writeCount >= 100 || h.size >= h.config.MaxSize {
		if err := h.buffer.Flush(); err != nil {
			return xerror.Wrap(err, "failed to flush buffer")
		}
		h.writeCount = 0
	}

	return nil
}

// run handles periodic flushing and listens for shutdown signal.
func (h *RotatingFileHandler) run() {
	for {
		select {
		case <-h.flushTicker.C:
			h.mu.Lock()
			if h.buffer != nil {
				h.buffer.Flush()
			}
			h.mu.Unlock()
		case <-h.done:
			h.mu.Lock()
			if h.buffer != nil {
				h.buffer.Flush()
			}
			h.mu.Unlock()
			return
		}
	}
}

// shouldRotate checks if the file should be rotated
func (h *RotatingFileHandler) shouldRotate() bool {
	if h.file == nil {
		return true
	}

	// Check size-based rotation
	if h.config.MaxSize > 0 && h.size >= h.config.MaxSize {
		return true
	}

	// Check time-based rotation (daily)
	if !isSameDay(time.Now(), h.lastRotate) {
		return true
	}

	// Check if file has been deleted or moved
	if _, err := os.Stat(h.config.Filename); os.IsNotExist(err) {
		return true
	}

	return false
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
		if h.buffer != nil {
			if err := h.buffer.Flush(); err != nil {
				return xerror.Wrap(err, "failed to flush buffer before rotation")
			}
		}
		if err := h.file.Close(); err != nil {
			return xerror.Wrap(err, "failed to close file before rotation")
		}
	}

	if err := h.rotateFile(); err != nil {
		return xerror.Wrap(err, "failed to rotate file")
	}

	file, err := os.OpenFile(h.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return xerror.Wrap(err, "failed to open new log file")
	}

	h.file = file
	h.buffer = bufio.NewWriterSize(file, 32*1024) // 32KB buffer
	h.size = 0
	h.lastRotate = time.Now()
	h.writeCount = 0

	return nil
}

// rotateFile rotates the log file.
func (h *RotatingFileHandler) rotateFile() error {
	if _, err := os.Stat(h.config.Filename); os.IsNotExist(err) {
		return nil
	}

	now := time.Now()
	ext := filepath.Ext(h.config.Filename)
	base := strings.TrimSuffix(filepath.Base(h.config.Filename), ext)
	dir := filepath.Dir(h.config.Filename)
	
	// Create a unique name with timestamp and counter
	newName := filepath.Join(dir, fmt.Sprintf("%s.%s.%03d%s", 
		base,
		now.Format("20060102150405"),
		h.currentNum,
		ext))

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return xerror.Wrap(err, "failed to create directory")
	}

	// Rename the current file
	if err := os.Rename(h.config.Filename, newName); err != nil {
		return xerror.Wrap(err, "failed to rename log file")
	}

	h.currentNum++
	if h.currentNum > h.config.MaxBackups {
		h.currentNum = 0
	}

	// Cleanup old files asynchronously
	go h.removeOldFiles()

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
	h.closeMu.Lock()
	defer h.closeMu.Unlock()

	if h.closed {
		return nil
	}

	close(h.done)
	h.flushTicker.Stop()

	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error
	if h.buffer != nil {
		if err := h.buffer.Flush(); err != nil {
			errs = append(errs, xerror.Wrap(err, "failed to flush buffer"))
		}
	}

	if h.file != nil {
		if err := h.file.Sync(); err != nil {
			errs = append(errs, xerror.Wrap(err, "failed to sync file"))
		}
		if err := h.file.Close(); err != nil {
			errs = append(errs, xerror.Wrap(err, "failed to close file"))
		}
	}

	h.closed = true

	if len(errs) > 0 {
		return xerror.Errorf("multiple errors during close: %v", errs)
	}
	return nil
}

// AddRotatingFile adds a rotating file handler to the logger
func AddRotatingFile(config FileConfig) error {
	handler, err := NewRotatingFileHandler(config)
	if err != nil {
		return xerror.Wrap(err, "failed to create rotating file handler")
	}

	mu.Lock()
	defer mu.Unlock()

	handlers = append(handlers, handler)

	var newHandler slog.Handler
	if mh, ok := defaultHandler.(*MultiHandler); ok {
		// If defaultHandler is already a MultiHandler, add the new handler to it
		mh.handlers = append(mh.handlers, handler)
		newHandler = mh
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
// If ctx is nil, it will use context.Background().
func Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var errs []error
	var wg sync.WaitGroup

	mu.RLock()
	for _, handler := range handlers {
		if sh, ok := handler.(ShutdownHandler); ok {
			wg.Add(1)
			go func(h ShutdownHandler) {
				defer wg.Done()
				if err := h.Shutdown(); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}(sh)
		}
	}
	mu.RUnlock()

	// Wait with context timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return xerror.Wrap(ctx.Err(), "shutdown timeout")
	case <-done:
		if len(errs) > 0 {
			return xerror.Wrap(errs[0], "shutdown error")
		}
		return nil
	}
}

// SetLogger sets the default logger
func SetLogger(logger *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()

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
	h.Handler = slog.NewJSONHandler(h.buffer, &slog.HandlerOptions{Level: level, AddSource: true})

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

	mu.Lock()
	defer mu.Unlock()

	handlers = append(handlers, handler)

	var newHandler slog.Handler
	if mh, ok := defaultHandler.(*MultiHandler); ok {
		// If defaultHandler is already a MultiHandler, add the new handler to it
		mh.handlers = append(mh.handlers, handler)
		newHandler = mh
	} else {
		// If not, create a new MultiHandler with both handlers
		newHandler = NewMultiHandler(defaultHandler, handler)
	}
	defaultHandler = newHandler
	defaultLogger = slog.New(defaultHandler)
	return nil
}

// Rotate forces a rotation of the current log file.
func (h *RotatingFileHandler) Rotate() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rotate()
}

// ColorLog logs a message with a custom color.
func ColorLog(level slog.Level, color xcolor.ColorCode, msg string, args ...any) {
	processedArgs := processArgs(args...)
	coloredMsg := xcolor.Colorize(color, msg)
	log(level, coloredMsg, processedArgs...)
}

// ColorLogf logs a formatted message with a custom color.
func ColorLogf(level slog.Level, color xcolor.ColorCode, format string, args ...any) {
	coloredMsg := xcolor.Colorize(color, fmt.Sprintf(format, args...))
	log(level, coloredMsg)
}

// RedLog logs a message in red.
func RedLog(level slog.Level, msg string, args ...any) {
	processedArgs := processArgs(args...)
	ColorLog(level, xcolor.Red, msg, processedArgs...)
}

// RedLogf logs a formatted message in red.
func RedLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Red, format, args...)
}

// GreenLog logs a message in green.
func GreenLog(level slog.Level, msg string, args ...any) {
	processedArgs := processArgs(args...)
	ColorLog(level, xcolor.Green, msg, processedArgs...)
}

// GreenLogf logs a formatted message in green.
func GreenLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Green, format, args...)
}

// YellowLog logs a message in yellow.
func YellowLog(level slog.Level, msg string, args ...any) {
	processedArgs := processArgs(args...)
	ColorLog(level, xcolor.Yellow, msg, processedArgs...)
}

// YellowLogf logs a formatted message in yellow.
func YellowLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Yellow, format, args...)
}

// BlueLog logs a message in blue.
func BlueLog(level slog.Level, msg string, args ...any) {
	processedArgs := processArgs(args...)
	ColorLog(level, xcolor.Blue, msg, processedArgs...)
}

// BlueLogf logs a formatted message in blue.
func BlueLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Blue, format, args...)
}

// PurpleLog logs a message in purple.
func PurpleLog(level slog.Level, msg string, args ...any) {
	processedArgs := processArgs(args...)
	ColorLog(level, xcolor.Purple, msg, processedArgs...)
}

// PurpleLogf logs a formatted message in purple.
func PurpleLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Purple, format, args...)
}

// CyanLog logs a message in cyan.
func CyanLog(level slog.Level, msg string, args ...any) {
	ColorLog(level, xcolor.Cyan, msg, args...)
}

// CyanLogf logs a formatted message in cyan.
func CyanLogf(level slog.Level, format string, args ...any) {
	ColorLogf(level, xcolor.Cyan, format, args...)
}

// DebugContext logs a debug message with context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	logContext(ctx, slog.LevelDebug, msg, args...)
}

// InfoContext logs an info message with context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	logContext(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext logs a warning message with context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	logContext(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext logs an error message with context.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	logContext(ctx, slog.LevelError, msg, args...)
}

// logContext is a helper function to add file and line information if configured, with context
func logContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	if logConfig.IncludeFileAndLine {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			rel, err := filepath.Rel(filepath.Dir(file), file)
			if err == nil {
				file = rel
			}
			fileInfo := fmt.Sprintf("%s:%d", file, line)
			args = append(args, "source", fileInfo)
		}
	}
	defaultLogger.LogAttrs(ctx, level, msg, slog.Any("args", args))
}

// GenReqID generates a unique request ID
func GenReqID() string {
	id := atomic.AddUint64(&reqIDCounter, 1)
	return fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), id)
}

// contextKey is a type for context keys
type contextKey int

const (
	reqIDKey contextKey = iota
)

// WithReqID adds a request ID to the context
func WithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDKey, reqID)
}

// GetReqID gets the request ID from context
func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(reqIDKey).(string); ok {
		return reqID
	}
	return ""
}

// isSameDay checks if two times are on the same day.
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
