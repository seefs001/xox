package xhttp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
)

// ResponseWriter wraps http.ResponseWriter to provide additional functionality
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Status returns the HTTP status of the request
func (w *ResponseWriter) Status() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

// WriteJSON writes JSON response
func (w *ResponseWriter) WriteJSON(v interface{}) error {
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// WriteXML writes XML response
func (w *ResponseWriter) WriteXML(v interface{}) error {
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.Header().Set("Content-Type", "application/xml")
	return xml.NewEncoder(w).Encode(v)
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
	c.Writer.WriteHeader(code)
	return c.Writer.WriteJSON(v)
}

// XML sends an XML response
func (c *Context) XML(code int, v interface{}) error {
	c.Writer.WriteHeader(code)
	return c.Writer.WriteXML(v)
}

// String sends a string response
func (c *Context) String(code int, format string, values ...interface{}) error {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// GetParam retrieves a URL parameter
func (c *Context) GetParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetParamInt retrieves a URL parameter as an integer
func (c *Context) GetParamInt(key string) (int, error) {
	return x.StringToInt(c.GetParam(key))
}

// GetParamInt64 retrieves a URL parameter as an int64
func (c *Context) GetParamInt64(key string) (int64, error) {
	return x.StringToInt64(c.GetParam(key))
}

// GetParamFloat retrieves a URL parameter as a float64
func (c *Context) GetParamFloat(key string) (float64, error) {
	return x.StringToFloat64(c.GetParam(key))
}

// GetParamBool retrieves a URL parameter as a boolean
func (c *Context) GetParamBool(key string) bool {
	return x.StringToBool(c.GetParam(key))
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
	err := json.NewDecoder(c.Request.Body).Decode(v)
	if err != nil {
		return xerror.Wrap(err, "error decoding request body")
	}
	return nil
}

// GetBodyRaw returns the raw request body as a byte slice
func (c *Context) GetBodyRaw() ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	defer c.Request.Body.Close()
	return io.ReadAll(c.Request.Body)
}

// GetBodyXML decodes the request body as XML into the provided interface
func (c *Context) GetBodyXML(v interface{}) error {
	return xerror.Wrap(xml.NewDecoder(c.Request.Body).Decode(v), "error decoding XML request body")
}

// GetFormValue retrieves a form value
func (c *Context) GetFormValue(key string) string {
	return c.Request.FormValue(key)
}

// GetFormFile retrieves a file from a multipart form
func (c *Context) GetFormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, nil, xerror.Wrap(err, "failed to parse multipart form")
	}
	return c.Request.FormFile(key)
}

// ParseMultipartForm parses a multipart form
func (c *Context) ParseMultipartForm(maxMemory int64) error {
	return xerror.Wrap(c.Request.ParseMultipartForm(maxMemory), "error parsing multipart form")
}

// GetQueryParams retrieves all query parameters
func (c *Context) GetQueryParams() map[string][]string {
	return c.Request.URL.Query()
}

// GetBodyForm parses the request body as a form and returns the values
func (c *Context) GetBodyForm() (map[string][]string, error) {
	if err := c.Request.ParseForm(); err != nil {
		return nil, xerror.Wrap(err, "error parsing form")
	}
	return c.Request.PostForm, nil
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
	cookie, err := c.Request.Cookie(name)
	return cookie, xerror.Wrap(err, "error getting cookie")
}

// Handler defines the handler function signature
type Handler func(*Context)

// Wrap converts a Handler to http.HandlerFunc
func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(NewContext(w, r))
	}
}

// WrapHTTP wraps a standard http.HandlerFunc to use our Context
func WrapHTTP(h http.HandlerFunc) Handler {
	return func(c *Context) {
		h(c.Writer, c.Request)
	}
}

// Bind binds data from multiple sources (query, form, JSON body) to a struct based on tags
func (c *Context) Bind(v interface{}) error {
	if err := c.bindQuery(v); err != nil {
		return xerror.Wrap(err, "error binding query parameters")
	}
	if err := c.bindForm(v); err != nil {
		return xerror.Wrap(err, "error binding form data")
	}
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.GetBody(v); err != nil {
			return xerror.Wrap(err, "error binding JSON body")
		}
	}
	return nil
}

func (c *Context) bindQuery(v interface{}) error {
	return c.bindData(v, c.Request.URL.Query())
}

func (c *Context) bindForm(v interface{}) error {
	if err := c.Request.ParseForm(); err != nil {
		return xerror.Wrap(err, "error parsing form")
	}
	return c.bindData(v, c.Request.Form)
}

func (c *Context) bindData(v interface{}, data map[string][]string) error {
	return x.BindData(v, data)
}

// Router is a custom router that wraps http.ServeMux and provides additional functionality
type Router struct {
	*http.ServeMux
	routes map[string][]string
}

// NewRouter creates a new Router
func NewRouter() *Router {
	return &Router{
		ServeMux: http.NewServeMux(),
		routes:   make(map[string][]string),
	}
}

// Handle registers the handler for the given pattern
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.ServeMux.Handle(pattern, handler)
	r.addRoute(pattern, "")
}

// HandleFunc registers the handler function for the given pattern
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.ServeMux.HandleFunc(pattern, handler)
	r.addRoute(pattern, "")
}

// Group creates a new route group
func (r *Router) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		router: r,
		prefix: prefix,
	}
}

// addRoute adds a route to the routes map
func (r *Router) addRoute(pattern, method string) {
	r.routes[pattern] = append(r.routes[pattern], method)
}

// PrintRoutes prints the routing tree
func (r *Router) PrintRoutes() {
	fmt.Println("Routing Tree:")
	for pattern, methods := range r.routes {
		fmt.Printf("├── %s\n", pattern)
		for _, method := range methods {
			if method != "" {
				fmt.Printf("│   └── %s\n", method)
			}
		}
	}
}

// RouterGroup represents a group of routes with a common prefix
type RouterGroup struct {
	router *Router
	prefix string
}

// Handle registers the handler for the given pattern within the group
func (g *RouterGroup) Handle(pattern string, handler http.Handler) {
	fullPattern := g.prefix + pattern
	g.router.ServeMux.Handle(fullPattern, handler)
	g.router.addRoute(fullPattern, "")
}

// HandleFunc registers the handler function for the given pattern within the group
func (g *RouterGroup) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	fullPattern := g.prefix + pattern
	g.router.ServeMux.HandleFunc(fullPattern, handler)
	g.router.addRoute(fullPattern, "")
}

// Group creates a new sub-group
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		router: g.router,
		prefix: g.prefix + prefix,
	}
}

// ListenAndServe starts the HTTP server
func ListenAndServe(addr string, handler http.Handler) error {
	return xerror.Wrap(http.ListenAndServe(addr, handler), "error starting HTTP server")
}

// ListenAndServeTLS starts the HTTPS server
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	return xerror.Wrap(http.ListenAndServeTLS(addr, certFile, keyFile, handler), "error starting HTTPS server")
}

// GetParamUint retrieves a URL parameter as an unsigned integer
func (c *Context) GetParamUint(key string) (uint, error) {
	return x.StringToUint(c.GetParam(key))
}

// GetParamUint64 retrieves a URL parameter as an uint64
func (c *Context) GetParamUint64(key string) (uint64, error) {
	return x.StringToUint64(c.GetParam(key))
}

// GetParamDuration retrieves a URL parameter as a time.Duration
func (c *Context) GetParamDuration(key string) (time.Duration, error) {
	return x.StringToDuration(c.GetParam(key))
}

// GetParamTime retrieves a URL parameter as a time.Time
func (c *Context) GetParamTime(key, layout string) (time.Time, error) {
	return time.Parse(layout, c.GetParam(key))
}

// SetStatus sets the HTTP status code
func (c *Context) SetStatus(code int) {
	c.Writer.WriteHeader(code)
}

// NoContent sends a response with no content
func (c *Context) NoContent(code int) {
	c.SetStatus(code)
}

// File sends a file as the response
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// StreamFile streams a file as the response
func (c *Context) StreamFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return xerror.Wrap(err, "error opening file")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return xerror.Wrap(err, "error getting file info")
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", stat.Name()))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	_, err = io.Copy(c.Writer, file)
	return xerror.Wrap(err, "error streaming file")
}

// BasicAuth returns the username and password provided in the request's
// Authorization header, if the request uses HTTP Basic Authentication.
func (c *Context) BasicAuth() (username, password string, ok bool) {
	return c.Request.BasicAuth()
}

// IsAjax checks if the request is an AJAX request
func (c *Context) IsAjax() bool {
	return c.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// ClientIP returns the client's IP address
func (c *Context) ClientIP() string {
	// Add X-Original-Forwarded-For check
	for _, h := range []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-Ip"} {
		if ip := c.Request.Header.Get(h); ip != "" {
			return strings.Split(ip, ",")[0]
		}
	}

	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}

	return c.Request.RemoteAddr
}

// MustBind binds data to a struct and panics if there's an error
func (c *Context) MustBind(v interface{}) {
	if err := c.Bind(v); err != nil {
		panic(xerror.Wrap(err, "error binding data"))
	}
}

// NewRouterWithMiddleware creates a new http.ServeMux with provided middlewares
func NewRouterWithMiddleware(middlewares ...func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	handler := http.Handler(mux)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return mux
}

// New utility methods
func (c *Context) GetBodyString() (string, error) {
	data, err := c.GetBodyRaw()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Context) AbortWithError(code int, err error) {
	c.Writer.WriteHeader(code)
	c.Writer.WriteJSON(map[string]string{"error": err.Error()})
}

func (c *Context) GetRequestID() string {
	return c.GetHeader("X-Request-ID")
}

func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
		return true
	}
	return false
}
