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

	"github.com/seefs001/xox/xcolor"
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].Format != "" {
			cfg.Format = config[0].Format
		}
		if config[0].TimeFormat != "" {
			cfg.TimeFormat = config[0].TimeFormat
		}
		if config[0].TimeZone != "" {
			cfg.TimeZone = config[0].TimeZone
		}
		if config[0].TimeInterval != 0 {
			cfg.TimeInterval = config[0].TimeInterval
		}
		if config[0].Output != nil {
			cfg.Output = config[0].Output
		}
		if len(config[0].ExcludePaths) > 0 {
			cfg.ExcludePaths = config[0].ExcludePaths
		}
		cfg.UseColor = config[0].UseColor
		if config[0].LogHandler != nil {
			cfg.LogHandler = config[0].LogHandler
		}
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

			for _, path := range cfg.ExcludePaths {
				if r.URL.Path == path {
					return
				}
			}

			attrs := make(map[string]interface{})
			attrs["time"] = time.Now().Format(cfg.TimeFormat)
			attrs["ip"] = r.RemoteAddr
			attrs["user_agent"] = r.UserAgent()
			attrs["query"] = r.URL.RawQuery
			attrs["body_size"] = r.ContentLength
			attrs["status"] = ww.statusCode
			attrs["latency"] = duration
			attrs["method"] = r.Method
			attrs["path"] = r.URL.Path

			log := cfg.Format
			for key, value := range attrs {
				placeholder := "${" + key + "}"
				stringValue := fmt.Sprintf("%v", value)
				if cfg.UseColor && xcolor.IsColorEnabled() {
					switch key {
					case "status":
						statusColor := xcolor.Green
						if ww.statusCode >= 300 && ww.statusCode < 400 {
							statusColor = xcolor.Yellow
						} else if ww.statusCode >= 400 {
							statusColor = xcolor.Red
						}
						stringValue = xcolor.Sprint(statusColor, stringValue)
					case "latency":
						stringValue = xcolor.Sprint(xcolor.Cyan, stringValue)
					case "method":
						stringValue = xcolor.Sprint(xcolor.Blue, stringValue)
					case "path":
						stringValue = xcolor.Sprint(xcolor.Purple, stringValue)
					}
				}
				log = strings.Replace(log, placeholder, stringValue, 1)
			}

			if attrs["query"] != "" {
				log = strings.Replace(log, "${query}", "?"+attrs["query"].(string), 1)
			} else {
				log = strings.Replace(log, "${query}", "", 1)
			}

			if cfg.LogHandler != nil {
				cfg.LogHandler(log, attrs)
			} else {
				if cfg.UseColor {
					xcolor.Print(xcolor.Reset, log)
				} else {
					_, _ = cfg.Output.Write([]byte(log))
				}
			}
		})
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].EnableStackTrace {
			cfg.EnableStackTrace = config[0].EnableStackTrace
		}
		if config[0].StackTraceHandler != nil {
			cfg.StackTraceHandler = config[0].StackTraceHandler
		}
		if config[0].ErrorLogger != nil {
			cfg.ErrorLogger = config[0].ErrorLogger
		}
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].Timeout != 0 {
			cfg.Timeout = config[0].Timeout
		}
		if config[0].TimeoutHandler != nil {
			cfg.TimeoutHandler = config[0].TimeoutHandler
		}
		if config[0].ErrorLogger != nil {
			cfg.ErrorLogger = config[0].ErrorLogger
		}
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if len(config[0].AllowOrigins) > 0 {
			cfg.AllowOrigins = config[0].AllowOrigins
		}
		if len(config[0].AllowMethods) > 0 {
			cfg.AllowMethods = config[0].AllowMethods
		}
		if len(config[0].AllowHeaders) > 0 {
			cfg.AllowHeaders = config[0].AllowHeaders
		}
		if config[0].AllowCredentials {
			cfg.AllowCredentials = config[0].AllowCredentials
		}
		if len(config[0].ExposeHeaders) > 0 {
			cfg.ExposeHeaders = config[0].ExposeHeaders
		}
		if config[0].MaxAge > 0 {
			cfg.MaxAge = config[0].MaxAge
		}
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].Level != 0 {
			cfg.Level = config[0].Level
		}
		if config[0].MinSize > 0 {
			cfg.MinSize = config[0].MinSize
		}
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

// RateLimitConfig defines the config for RateLimit middleware.
type RateLimitConfig struct {
	Next     func(c *http.Request) bool
	Max      int
	Duration time.Duration
	KeyFunc  func(*http.Request) string
}

// RateLimit returns a middleware that limits the number of requests
func RateLimit(config ...RateLimitConfig) Middleware {
	cfg := RateLimitConfig{
		Next:     nil,
		Max:      100,
		Duration: time.Minute,
		KeyFunc: func(r *http.Request) string {
			return r.RemoteAddr
		},
	}

	if len(config) > 0 {
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].Max > 0 {
			cfg.Max = config[0].Max
		}
		if config[0].Duration > 0 {
			cfg.Duration = config[0].Duration
		}
		if config[0].KeyFunc != nil {
			cfg.KeyFunc = config[0].KeyFunc
		}
	}

	type limiter struct {
		lastTime time.Time
		tokens   int
	}

	limiters := make(map[string]*limiter)
	mu := &sync.Mutex{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := cfg.KeyFunc(r)
			mu.Lock()
			l, exists := limiters[key]
			if !exists {
				l = &limiter{lastTime: time.Now(), tokens: cfg.Max}
				limiters[key] = l
			}

			now := time.Now()
			elapsed := now.Sub(l.lastTime)
			l.lastTime = now

			l.tokens += int(elapsed.Seconds() * float64(cfg.Max) / cfg.Duration.Seconds())
			if l.tokens > cfg.Max {
				l.tokens = cfg.Max
			}

			if l.tokens < 1 {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			l.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if len(config[0].Users) > 0 {
			cfg.Users = config[0].Users
		}
		if config[0].Realm != "" {
			cfg.Realm = config[0].Realm
		}
		if config[0].AuthFunc != nil {
			cfg.AuthFunc = config[0].AuthFunc
		}
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
	if session, ok := m.sessions[sessionID]; ok {
		return session, nil
	}
	return nil, fmt.Errorf("session not found")
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
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].Store != nil {
			cfg.Store = config[0].Store
		}
		if config[0].CookieName != "" {
			cfg.CookieName = config[0].CookieName
		}
		if config[0].MaxAge != 0 {
			cfg.MaxAge = config[0].MaxAge
		}
		if config[0].SessionName != "" {
			cfg.SessionName = config[0].SessionName
		}
	}

	sessionManager := NewSessionManager(cfg.Store, cfg.CookieName, cfg.MaxAge, cfg.SessionName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Next != nil && cfg.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			var sessionID string
			cookie, err := r.Cookie(cfg.CookieName)
			if err == nil {
				sessionID = cookie.Value
			}

			if sessionID == "" {
				sessionID = generateSessionID()
				http.SetCookie(w, &http.Cookie{
					Name:   cfg.CookieName,
					Value:  sessionID,
					MaxAge: cfg.MaxAge,
				})
			}

			session, err := cfg.Store.Get(sessionID)
			if err != nil {
				session = make(map[string]interface{})
			}

			ctx := context.WithValue(r.Context(), cfg.SessionName, session)
			ctx = context.WithValue(ctx, "sessionManager", sessionManager)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)

			cfg.Store.Set(sessionID, session)
		})
	}
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
