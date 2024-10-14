package xmw

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xcolor"
	"github.com/seefs001/xox/xerror"
)

// Middleware defines the signature for middleware functions
type Middleware func(next http.Handler) http.Handler

// Use applies a list of middlewares to a http.Handler
func Use(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

var DefaultMiddlewareSet = []Middleware{
	Logger(),
	Recover(),
	Timeout(),
	CORS(),
	Compress(),
	Session(),
}

// LoggerConfig defines the config for Logger middleware
type LoggerConfig struct {
	Next         func(c *http.Request) bool
	Format       string
	TimeFormat   string
	TimeZone     string
	TimeInterval time.Duration
	Output       io.Writer
	ExcludePaths []string
	UseColor     bool
	LogHandler   func(msg string, attrs map[string]interface{})
}

// Logger returns a middleware that logs HTTP requests/responses.
func Logger(config ...LoggerConfig) Middleware {
	cfg := LoggerConfig{
		Next:         nil,
		Format:       "[${time}] ${status} - ${latency} ${method} ${path} ${query} ${ip} ${user_agent} ${body_size}\n",
		TimeFormat:   "2006-01-02 15:04:05",
		TimeZone:     "Local",
		TimeInterval: 500 * time.Millisecond,
		Output:       os.Stdout,
		UseColor:     true,
		LogHandler:   nil,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			if x.Contains(cfg.ExcludePaths, r.URL.Path) {
				return
			}

			attrs := map[string]interface{}{
				"time":       time.Now().Format(cfg.TimeFormat),
				"ip":         r.RemoteAddr,
				"user_agent": r.UserAgent(),
				"query":      r.URL.RawQuery,
				"body_size":  r.ContentLength,
				"status":     ww.statusCode,
				"latency":    duration,
				"method":     r.Method,
				"path":       r.URL.Path,
			}

			log := cfg.Format
			for key, value := range attrs {
				placeholder := "${" + key + "}"
				stringValue := fmt.Sprintf("%v", value)
				if cfg.UseColor && xcolor.IsColorEnabled() {
					stringValue = colorizeLogValue(key, ww.statusCode, stringValue)
				}
				log = strings.Replace(log, placeholder, stringValue, 1)
			}

			log = strings.Replace(log, "${query}", x.Ternary(attrs["query"] != "", "?"+attrs["query"].(string), ""), 1)

			if cfg.LogHandler != nil {
				cfg.LogHandler(log, attrs)
			} else if cfg.UseColor {
				xcolor.Print(xcolor.Reset, log)
			} else {
				_, _ = cfg.Output.Write([]byte(log))
			}
		})
	}
}

func colorizeLogValue(key string, statusCode int, value string) string {
	switch key {
	case "status":
		statusColor := xcolor.Green
		if statusCode >= 300 && statusCode < 400 {
			statusColor = xcolor.Yellow
		} else if statusCode >= 400 {
			statusColor = xcolor.Red
		}
		return xcolor.Sprint(statusColor, value)
	case "latency":
		return xcolor.Sprint(xcolor.Cyan, value)
	case "method":
		return xcolor.Sprint(xcolor.Blue, value)
	case "path":
		return xcolor.Sprint(xcolor.Purple, value)
	default:
		return value
	}
}

// RecoverConfig defines the config for Recover middleware.
type RecoverConfig struct {
	Next              func(c *http.Request) bool
	EnableStackTrace  bool
	StackTraceHandler func(c *http.Request, e interface{})
	ErrorLogger       func(msg string, keyvals ...interface{})
}

// Recover returns a middleware which recovers from panics anywhere in the chain
func Recover(config ...RecoverConfig) Middleware {
	cfg := RecoverConfig{
		Next:              nil,
		EnableStackTrace:  false,
		StackTraceHandler: nil,
		ErrorLogger:       nil,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if cfg.EnableStackTrace {
						stack := debug.Stack()
						if cfg.StackTraceHandler != nil {
							cfg.StackTraceHandler(r, err)
						} else if cfg.ErrorLogger != nil {
							cfg.ErrorLogger("Panic recovered", "error", err, "stack", string(stack))
						}
					}
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutConfig defines the config for Timeout middleware.
type TimeoutConfig struct {
	Next           func(c *http.Request) bool
	Timeout        time.Duration
	TimeoutHandler func(w http.ResponseWriter, r *http.Request)
	ErrorLogger    func(msg string, keyvals ...interface{})
}

type timeoutResponseWriter struct {
	http.ResponseWriter
	mu         sync.Mutex
	written    bool
	statusCode int
}

func (w *timeoutResponseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.written {
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
		w.written = true
	}
}

func (w *timeoutResponseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.written {
		w.statusCode = http.StatusOK
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// Timeout wraps a handler in a timeout.
func Timeout(config ...TimeoutConfig) Middleware {
	cfg := TimeoutConfig{
		Next:           nil,
		Timeout:        30 * time.Second,
		TimeoutHandler: nil,
		ErrorLogger:    nil,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			tw := &timeoutResponseWriter{ResponseWriter: w}
			done := make(chan bool, 1)
			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				tw.mu.Lock()
				defer tw.mu.Unlock()
				if !tw.written {
					if cfg.TimeoutHandler != nil {
						cfg.TimeoutHandler(w, r)
					} else {
						if cfg.ErrorLogger != nil {
							cfg.ErrorLogger("Request timed out", "uri", r.RequestURI)
						}
						http.Error(w, "Request timed out", http.StatusGatewayTimeout)
					}
				}
			}
		})
	}
}

// CORSConfig defines the config for CORS middleware.
type CORSConfig struct {
	Next             func(c *http.Request) bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// CORS returns a middleware that adds CORS headers to the response
func CORS(config ...CORSConfig) Middleware {
	cfg := CORSConfig{
		Next:         nil,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH"},
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			if cfg.AllowOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowed := range cfg.AllowOrigins {
					if origin == allowed {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(cfg.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
			}

			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CompressConfig defines the config for Compress middleware.
type CompressConfig struct {
	Next    func(c *http.Request) bool
	Level   int
	MinSize int
}

// Compress returns a middleware that compresses the response using gzip compression
func Compress(config ...CompressConfig) Middleware {
	cfg := CompressConfig{
		Next:    nil,
		Level:   -1,
		MinSize: 1024,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			gw, _ := gzip.NewWriterLevel(w, cfg.Level)
			defer gw.Close()

			gzw := gzipResponseWriter{ResponseWriter: w, Writer: gw}
			next.ServeHTTP(gzw, r)
		})
	}
}

// RateLimitConfig defines the configuration for the RateLimit middleware
type RateLimitConfig struct {
	Next       func(c *http.Request) bool
	Max        int
	Duration   time.Duration
	Message    string
	StatusCode int
	KeyFunc    func(*http.Request) string
}

// RateLimit returns a middleware that limits the number of requests
func RateLimit(config ...RateLimitConfig) Middleware {
	cfg := RateLimitConfig{
		Next:       nil,
		Max:        100,
		Duration:   time.Minute,
		Message:    "Too many requests, please try again later.",
		StatusCode: http.StatusTooManyRequests,
		KeyFunc:    defaultKeyFunc,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	// Ensure we have a valid key function
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = defaultKeyFunc
	}

	// Ensure we have a valid status code
	if cfg.StatusCode == 0 {
		cfg.StatusCode = http.StatusTooManyRequests
	}

	type client struct {
		count    int
		lastSeen time.Time
	}
	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := cfg.KeyFunc(r)
			mu.Lock()
			if _, found := clients[key]; !found {
				clients[key] = &client{count: 0, lastSeen: time.Now()}
			}

			c := clients[key]
			if time.Since(c.lastSeen) > cfg.Duration {
				c.count = 0
				c.lastSeen = time.Now()
			}

			if c.count >= cfg.Max {
				mu.Unlock()
				http.Error(w, cfg.Message, cfg.StatusCode)
				return
			}

			c.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// defaultKeyFunc generates a default key for rate limiting based on the request's RemoteAddr
func defaultKeyFunc(r *http.Request) string {
	return r.RemoteAddr
}

// BasicAuthConfig defines the config for BasicAuth middleware.
type BasicAuthConfig struct {
	Next     func(c *http.Request) bool
	Users    map[string]string
	Realm    string
	AuthFunc func(string, string) bool
}

// BasicAuth returns a middleware that performs basic authentication
func BasicAuth(config ...BasicAuthConfig) Middleware {
	cfg := BasicAuthConfig{
		Next:     nil,
		Users:    make(map[string]string),
		Realm:    "Restricted",
		AuthFunc: nil,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if cfg.AuthFunc != nil {
				if !cfg.AuthFunc(user, pass) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else if storedPass, ok := cfg.Users[user]; !ok || storedPass != pass {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	Get(sessionID string) (map[string]interface{}, error)
	Set(sessionID string, data map[string]interface{}) error
	Delete(sessionID string) error
}

// MemoryStore implements SessionStore using in-memory storage
type MemoryStore struct {
	sessions map[string]map[string]interface{}
	mu       sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]map[string]interface{}),
	}
}

func (m *MemoryStore) Get(sessionID string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, xerror.ErrNotFound
	}
	return session, nil
}

func (m *MemoryStore) Set(sessionID string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = data
	return nil
}

func (m *MemoryStore) Delete(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

// DefaultSessionName is the default name for the session in the context
const DefaultSessionName = "ctx_session"

// SessionConfig defines the config for Session middleware
type SessionConfig struct {
	Next        func(c *http.Request) bool
	Store       SessionStore
	CookieName  string
	MaxAge      int
	SessionName string
}

// SessionManager handles session operations
type SessionManager struct {
	store       SessionStore
	cookieName  string
	maxAge      int
	sessionName string
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(store SessionStore, cookieName string, maxAge int, sessionName string) *SessionManager {
	if store == nil {
		store = NewMemoryStore()
	}
	if cookieName == "" {
		cookieName = "session_id"
	}
	if maxAge <= 0 {
		maxAge = 86400 // 1 day default
	}
	if sessionName == "" {
		sessionName = DefaultSessionName
	}
	return &SessionManager{
		store:       store,
		cookieName:  cookieName,
		maxAge:      maxAge,
		sessionName: sessionName,
	}
}

// Get retrieves a value from the session
func (sm *SessionManager) Get(r *http.Request, key string) (interface{}, bool) {
	session := sm.getSession(r)
	if session == nil {
		return nil, false
	}
	value, ok := session[key]
	return value, ok
}

// Set sets a value in the session
func (sm *SessionManager) Set(r *http.Request, key string, value interface{}) {
	session := sm.getSession(r)
	if session == nil {
		// If session doesn't exist, create a new one
		session = make(map[string]interface{})
		ctx := context.WithValue(r.Context(), sm.sessionName, session)
		*r = *r.WithContext(ctx)
	}
	session[key] = value
}

// Delete removes a value from the session
func (sm *SessionManager) Delete(r *http.Request, key string) {
	session := sm.getSession(r)
	if session != nil {
		delete(session, key)
	}
}

// Clear removes all values from the session
func (sm *SessionManager) Clear(r *http.Request) {
	session := sm.getSession(r)
	if session != nil {
		for key := range session {
			delete(session, key)
		}
	}
}

// getSession retrieves the session from the request context
func (sm *SessionManager) getSession(r *http.Request) map[string]interface{} {
	if session, ok := r.Context().Value(sm.sessionName).(map[string]interface{}); ok {
		return session
	}
	// If session doesn't exist, create a new one
	session := make(map[string]interface{})
	ctx := context.WithValue(r.Context(), sm.sessionName, session)
	*r = *r.WithContext(ctx)
	return session
}

// Session returns a middleware that handles session management
func Session(config ...SessionConfig) Middleware {
	cfg := SessionConfig{
		Next:        nil,
		Store:       NewMemoryStore(),
		CookieName:  "session_id",
		MaxAge:      86400, // 1 day
		SessionName: DefaultSessionName,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	sessionManager := NewSessionManager(cfg.Store, cfg.CookieName, cfg.MaxAge, cfg.SessionName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			var session map[string]interface{}
			var sessionID string

			cookie, err := r.Cookie(cfg.CookieName)
			if err == nil {
				sessionID = cookie.Value
				session, err = cfg.Store.Get(sessionID)
				if err != nil {
					// If session not found or any other error, create a new session
					session = make(map[string]interface{})
					sessionID = generateSessionID()
				}
			} else {
				// No cookie found, create a new session
				session = make(map[string]interface{})
				sessionID = generateSessionID()
			}

			// Always set the cookie, refreshing expiration
			http.SetCookie(w, &http.Cookie{
				Name:   cfg.CookieName,
				Value:  sessionID,
				MaxAge: cfg.MaxAge,
			})

			ctx := context.WithValue(r.Context(), cfg.SessionName, session)
			ctx = context.WithValue(ctx, "sessionManager", sessionManager)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)

			// Save session after the handler has been called
			if err := cfg.Store.Set(sessionID, session); err != nil {
				fmt.Printf("Error saving session: %v\n", err)
			}
		})
	}
}

func getSessionID(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		return cookie.Value
	}
	return ""
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetSessionManager retrieves the SessionManager from the request context
func GetSessionManager(r *http.Request) *SessionManager {
	if sm, ok := r.Context().Value("sessionManager").(*SessionManager); ok {
		return sm
	}
	return nil
}

// responseWriter is a wrapper for http.ResponseWriter that allows us to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// gzipResponseWriter is a wrapper for http.ResponseWriter that allows us to gzip the response
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}

// StaticConfig defines the config for Static middleware
type StaticConfig struct {
	Next      func(c *http.Request) bool
	Root      string
	Index     string
	Browse    bool
	MaxAge    int
	Prefix    string
	APIPrefix string
	Files     map[string]File
	EmbedFS   embed.FS
}

// File represents a file that can be served
type File struct {
	Content      io.Reader
	ContentType  string
	LastModified time.Time
	ETag         string
}

// Static returns a middleware that serves static files, API requests, and io.Reader content
func Static(config ...StaticConfig) Middleware {
	cfg := StaticConfig{
		Next:      nil,
		Root:      "public",
		Index:     "index.html",
		Browse:    false,
		MaxAge:    0,
		Prefix:    "/",
		APIPrefix: "/api",
		Files:     make(map[string]File),
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	var fsys http.FileSystem
	if cfg.EmbedFS != (embed.FS{}) {
		fsys = http.FS(cfg.EmbedFS)
	} else {
		fsys = http.Dir(cfg.Root)
	}
	fileServer := http.FileServer(fsys)
	filesMutex := &sync.RWMutex{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Check if it's an API request
			if strings.HasPrefix(r.URL.Path, cfg.APIPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			// Strip prefix if necessary
			path := r.URL.Path
			if cfg.Prefix != "/" {
				if strings.HasPrefix(path, cfg.Prefix) {
					path = strings.TrimPrefix(path, cfg.Prefix)
				} else {
					next.ServeHTTP(w, r)
					return
				}
			}

			// If path is empty or ends with '/', append index.html
			if path == "" || path[len(path)-1] == '/' {
				path += cfg.Index
			}

			// Check if we have a File for this path
			filesMutex.RLock()
			file, exists := cfg.Files[path]
			filesMutex.RUnlock()

			if exists {
				serveFile(w, r, file, cfg.MaxAge)
				return
			}

			// If not found in Files, check the filesystem or embed.FS
			f, err := fsys.Open(path)
			if err != nil {
				if os.IsNotExist(err) {
					// Try serving index.html for SPA
					indexPath := filepath.Join(cfg.Root, cfg.Index)
					if indexFile, err := fsys.Open(indexPath); err == nil {
						defer indexFile.Close()
						stat, _ := indexFile.Stat()
						file := File{
							Content:      indexFile,
							ContentType:  "text/html",
							LastModified: stat.ModTime(),
							ETag:         generateETag(stat),
						}
						serveFile(w, r, file, cfg.MaxAge)
						return
					}
				}
				http.NotFound(w, r)
				return
			}
			defer f.Close()

			stat, err := f.Stat()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if stat.IsDir() {
				indexPath := filepath.Join(path, cfg.Index)
				if indexFile, err := fsys.Open(indexPath); err == nil {
					defer indexFile.Close()
					indexStat, _ := indexFile.Stat()
					file := File{
						Content:      indexFile,
						ContentType:  "text/html",
						LastModified: indexStat.ModTime(),
						ETag:         generateETag(indexStat),
					}
					serveFile(w, r, file, cfg.MaxAge)
					return
				}
				if !cfg.Browse {
					http.NotFound(w, r)
					return
				}
				// Use the built-in file server to handle directory listing
				fileServer.ServeHTTP(w, r)
				return
			}

			// Serve the file
			file = File{
				Content:      f.(io.ReadSeeker),
				ContentType:  getContentType(stat.Name()),
				LastModified: stat.ModTime(),
				ETag:         generateETag(stat),
			}
			serveFile(w, r, file, cfg.MaxAge)
		})
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, file File, maxAge int) {
	// Set Content-Type
	if file.ContentType != "" {
		w.Header().Set("Content-Type", file.ContentType)
	}

	// Set Last-Modified
	w.Header().Set("Last-Modified", file.LastModified.UTC().Format(http.TimeFormat))

	// Set ETag
	w.Header().Set("ETag", file.ETag)

	// Set Cache-Control
	if maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	}

	// Check If-Modified-Since
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && file.LastModified.Before(t.Add(time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Check If-None-Match
	if r.Header.Get("If-None-Match") == file.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Serve the content
	http.ServeContent(w, r, "", file.LastModified, file.Content.(io.ReadSeeker))
}

func generateETag(info os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size())
}

func getContentType(name string) string {
	ext := filepath.Ext(name)
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// AddFile adds or updates a file in the Files map
func (cfg *StaticConfig) AddFile(path string, content io.Reader, contentType string, lastModified time.Time) {
	contentBytes, _ := io.ReadAll(content)
	etag := fmt.Sprintf(`"%x"`, md5.Sum(contentBytes))
	cfg.Files[path] = File{
		Content:      io.NopCloser(bytes.NewReader(contentBytes)),
		ContentType:  contentType,
		LastModified: lastModified,
		ETag:         etag,
	}
}
