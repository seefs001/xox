package xhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

// ResponseWriter wraps http.ResponseWriter to provide additional functionality
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteJSON writes JSON response
func (w *ResponseWriter) WriteJSON(v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// Context is a custom context type that includes request and response writer
type Context struct {
	Request *http.Request
	Writer  *ResponseWriter
}

// NewContext creates a new Context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		Writer:  &ResponseWriter{ResponseWriter: w},
	}
}

// JSON sends a JSON response
func (c *Context) JSON(code int, v interface{}) error {
	c.Writer.StatusCode = code
	return c.Writer.WriteJSON(v)
}

// GetParam retrieves a URL parameter
func (c *Context) GetParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetParamInt retrieves a URL parameter as an integer
func (c *Context) GetParamInt(key string) (int, error) {
	return strconv.Atoi(c.GetParam(key))
}

// GetParamFloat retrieves a URL parameter as a float64
func (c *Context) GetParamFloat(key string) (float64, error) {
	return strconv.ParseFloat(c.GetParam(key), 64)
}

// GetHeader retrieves a header value
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// GetContext retrieves the request's context
func (c *Context) GetContext() context.Context {
	return c.Request.Context()
}

// WithValue adds a value to the request's context
func (c *Context) WithValue(key, val interface{}) {
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), key, val))
}

// GetBody decodes the request body into the provided interface
func (c *Context) GetBody(v interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}

// Redirect sends a redirect response
func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.Writer, c.Request, url, code)
}

// SetCookie sets a cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// GetCookie retrieves a cookie
func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// Handler defines the handler function signature
type Handler func(*Context)

// Wrap converts a Handler to http.HandlerFunc
func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(NewContext(w, r))
	}
}
