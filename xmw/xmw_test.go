package xmw

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := Use(handler, middleware1, middleware2)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Test-1") != "true" {
		t.Error("Middleware 1 was not applied")
	}
	if rec.Header().Get("X-Test-2") != "true" {
		t.Error("Middleware 2 was not applied")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	buf := new(bytes.Buffer)
	loggerMiddleware := Logger(LoggerConfig{
		Format: "${method} ${path}\n",
		Output: buf,
	})

	finalHandler := loggerMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "GET /test") {
		t.Error("Logger did not log the request correctly")
	}
}

func TestRecover(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	recoverMiddleware := Recover()
	finalHandler := recoverMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestTimeout(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	timeoutMiddleware := Timeout(TimeoutConfig{
		Timeout: 100 * time.Millisecond,
	})

	finalHandler := timeoutMiddleware(slowHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("Expected status code %d, got %d", http.StatusGatewayTimeout, rec.Code)
	}
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS(CORSConfig{
		AllowOrigins: []string{"http://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"X-Custom-Header"},
	})

	finalHandler := corsMiddleware(handler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Error("CORS middleware did not set the correct Allow-Origin header")
	}
	if rec.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Error("CORS middleware did not set the correct Allow-Methods header")
	}
	if rec.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header" {
		t.Error("CORS middleware did not set the correct Allow-Headers header")
	}
}

func TestCompress(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	compressMiddleware := Compress()
	finalHandler := compressMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Compress middleware did not set the correct Content-Encoding header")
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello, World!" {
		t.Error("Compress middleware did not correctly compress the response")
	}
}

func TestRateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimitMiddleware := RateLimit(RateLimitConfig{
		Max:      2,
		Duration: time.Minute,
	})

	finalHandler := rateLimitMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	rec3 := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec1, req)
	finalHandler.ServeHTTP(rec2, req)
	finalHandler.ServeHTTP(rec3, req)

	if rec1.Code != http.StatusOK || rec2.Code != http.StatusOK {
		t.Error("First two requests should be allowed")
	}
	if rec3.Code != http.StatusTooManyRequests {
		t.Error("Third request should be rate limited")
	}
}

func TestBasicAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	basicAuthMiddleware := BasicAuth(BasicAuthConfig{
		Users: map[string]string{
			"testuser": "testpass",
		},
		Realm: "Test Realm",
	})

	finalHandler := basicAuthMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("testuser:testpass")))
	rec := httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Error("Basic Auth middleware did not allow valid credentials")
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("wronguser:wrongpass")))
	rec = httptest.NewRecorder()

	finalHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Error("Basic Auth middleware did not block invalid credentials")
	}
}

func TestResponseWriter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	rw := &responseWriter{ResponseWriter: httptest.NewRecorder(), statusCode: http.StatusOK}
	handler.ServeHTTP(rw, httptest.NewRequest("GET", "/", nil))

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.statusCode)
	}
}

func TestGzipResponseWriter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	buf := new(bytes.Buffer)
	gw := gzip.NewWriter(buf)
	grw := gzipResponseWriter{ResponseWriter: httptest.NewRecorder(), Writer: gw}

	handler.ServeHTTP(grw, httptest.NewRequest("GET", "/", nil))
	gw.Close()

	reader, err := gzip.NewReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello, World!" {
		t.Error("GzipResponseWriter did not correctly write the response")
	}
}
