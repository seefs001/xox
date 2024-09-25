package xmw

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seefs001/xox/xlog"
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

// LoggerConfig defines the config for Logger middleware
type LoggerConfig struct {
	Next         func(c *http.Request) bool
	Format       string
	TimeFormat   string
	TimeZone     string
	TimeInterval time.Duration
	Output       io.Writer
	ExcludePaths []string
	UseColor     bool // Add UseColor field to support optional coloring
}

// Logger returns a middleware that logs HTTP requests/responses.
func Logger(config ...LoggerConfig) Middleware {
	cfg := LoggerConfig{
		Next:         nil,
		Format:       "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat:   "15:04:05",
		TimeZone:     "Local",
		TimeInterval: 500 * time.Millisecond,
		Output:       os.Stdout,
		UseColor:     true, // Default to no color
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

			for _, path := range cfg.ExcludePaths {
				if r.URL.Path == path {
					return
				}
			}

			log := strings.Replace(cfg.Format, "${time}", time.Now().Format(cfg.TimeFormat), 1)
			if cfg.UseColor {
				statusColor := "\x1b[32m" // Green for 2xx
				if ww.statusCode >= 300 && ww.statusCode < 400 {
					statusColor = "\x1b[33m" // Yellow for 3xx
				} else if ww.statusCode >= 400 {
					statusColor = "\x1b[31m" // Red for 4xx and 5xx
				}
				log = strings.Replace(log, "${status}", statusColor+strconv.Itoa(ww.statusCode)+"\x1b[0m", 1)
				log = strings.Replace(log, "${latency}", "\x1b[36m"+duration.String()+"\x1b[0m", 1)
				log = strings.Replace(log, "${method}", "\x1b[34m"+r.Method+"\x1b[0m", 1)
			} else {
				log = strings.Replace(log, "${status}", strconv.Itoa(ww.statusCode), 1)
				log = strings.Replace(log, "${latency}", duration.String(), 1)
				log = strings.Replace(log, "${method}", r.Method, 1)
			}
			log = strings.Replace(log, "${path}", r.URL.Path, 1)

			_, _ = cfg.Output.Write([]byte(log))
		})
	}
}

// RecoverConfig defines the config for Recover middleware.
type RecoverConfig struct {
	Next              func(c *http.Request) bool
	EnableStackTrace  bool
	StackTraceHandler func(c *http.Request, e interface{})
}

// Recover returns a middleware which recovers from panics anywhere in the chain
func Recover(config ...RecoverConfig) Middleware {
	cfg := RecoverConfig{
		Next:              nil,
		EnableStackTrace:  false,
		StackTraceHandler: nil,
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
						} else {
							xlog.Error("Panic recovered", "error", err, "stack", string(stack))
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
}

// Timeout wraps a handler in a timeout.
func Timeout(config TimeoutConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Next != nil && config.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), config.Timeout)
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
				if config.TimeoutHandler != nil {
					config.TimeoutHandler(w, r)
				} else {
					xlog.Warn("Request timed out", "uri", r.RequestURI)
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

// RateLimitConfig defines the config for RateLimit middleware.
type RateLimitConfig struct {
	Next     func(c *http.Request) bool
	Max      int
	Duration time.Duration
	KeyFunc  func(*http.Request) string
}

// RateLimit returns a middleware that limits the number of requests
// TODO: Fix this
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
		cfg = config[0]
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
func BasicAuth(config BasicAuthConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Next != nil && config.Next(r) {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if config.AuthFunc != nil {
				if !config.AuthFunc(user, pass) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else if storedPass, ok := config.Users[user]; !ok || storedPass != pass {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
