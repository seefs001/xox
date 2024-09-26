package xhttp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
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

// GetBodyRaw returns the raw request body as a byte slice
func (c *Context) GetBodyRaw() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// GetBodyXML decodes the request body as XML into the provided interface
func (c *Context) GetBodyXML(v interface{}) error {
	return xml.NewDecoder(c.Request.Body).Decode(v)
}

// GetFormValue retrieves a form value
func (c *Context) GetFormValue(key string) string {
	return c.Request.FormValue(key)
}

// GetFormFile retrieves a file from a multipart form
func (c *Context) GetFormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(key)
}

// ParseMultipartForm parses a multipart form
func (c *Context) ParseMultipartForm(maxMemory int64) error {
	return c.Request.ParseMultipartForm(maxMemory)
}

// GetQueryParams retrieves all query parameters
func (c *Context) GetQueryParams() map[string][]string {
	return c.Request.URL.Query()
}

// GetBodyForm parses the request body as a form and returns the values
func (c *Context) GetBodyForm() (map[string][]string, error) {
	if err := c.Request.ParseForm(); err != nil {
		return nil, err
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

// WrapHTTP wraps a standard http.HandlerFunc to use our Context
func WrapHTTP(h http.HandlerFunc) Handler {
	return func(c *Context) {
		h(c.Writer, c.Request)
	}
}

// Bind binds data from multiple sources (query, form, JSON body) to a struct based on tags
func (c *Context) Bind(v interface{}) error {
	if err := c.bindQuery(v); err != nil {
		return err
	}
	if err := c.bindForm(v); err != nil {
		return err
	}
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.GetBody(v); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) bindQuery(v interface{}) error {
	return c.bindData(v, c.Request.URL.Query())
}

func (c *Context) bindForm(v interface{}) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	return c.bindData(v, c.Request.Form)
}

func (c *Context) bindData(v interface{}, data map[string][]string) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		tag := field.Tag.Get("form")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		if tag == "" {
			tag = field.Tag.Get("query")
		}
		if tag == "" {
			continue
		}

		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}

		if values, ok := data[name]; ok && len(values) > 0 {
			if err := setField(fieldValue, values); err != nil {
				return err
			}
		}
	}

	return nil
}

func setField(field reflect.Value, values []string) error {
	if len(values) == 0 {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(values[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(values[0])
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intValue, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(values[0], 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		slice := reflect.MakeSlice(field.Type(), len(values), len(values))
		for i, value := range values {
			if err := setField(slice.Index(i), []string{value}); err != nil {
				return err
			}
		}
		field.Set(slice)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			timeValue, err := time.Parse(time.RFC3339, values[0])
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(timeValue))
		}
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), values)
	case reflect.Interface:
		if field.IsNil() {
			field.Set(reflect.ValueOf(values[0]))
		} else {
			return setField(field.Elem(), values)
		}
	default:
		return nil // Unsupported type
	}
	return nil
}
