package xmw

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

			done := make(chan bool)
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if cfg.TimeoutHandler != nil {
					cfg.TimeoutHandler(w, r)
				} else {
					if cfg.ErrorLogger != nil {
						cfg.ErrorLogger("Request timed out", "uri", r.RequestURI)
					}
					http.Error(w, "Request timed out", http.StatusGatewayTimeout)
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
		return nil, xerror.New("session not found")
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
	if session != nil {
		session[key] = value
	}
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
	return nil
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

			sessionID := getSessionID(r, cfg.CookieName)
			if sessionID == "" {
				sessionID = generateSessionID()
				http.SetCookie(w, &http.Cookie{
					Name:   cfg.CookieName,
					Value:  sessionID,
					MaxAge: cfg.MaxAge,
				})
			}

			session, err := cfg.Store.Get(sessionID)
			if xerror.Is(err, xerror.ErrNotFound) {
				session = make(map[string]interface{})
			} else if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), cfg.SessionName, session)
			ctx = context.WithValue(ctx, "sessionManager", sessionManager)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)

			if err := cfg.Store.Set(sessionID, session); err != nil {
				// Log the error, but don't interrupt the response
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
