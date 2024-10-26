package xhttpc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xcast"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xlog"
)

const (
	defaultTimeout               = 300 * time.Second
	defaultDialTimeout           = 5 * time.Second
	defaultKeepAlive             = 30 * time.Second
	defaultMaxIdleConns          = 100
	defaultIdleConnTimeout       = 90 * time.Second
	defaultTLSHandshakeTimeout   = 5 * time.Second
	defaultExpectContinueTimeout = 1 * time.Second
	defaultRetryCount            = 3
	defaultMaxBackoff            = 30 * time.Second
	defaultMaxBodyLogSize        = 1024 // 1KB
	defaultUserAgent             = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

// Client is a high-performance HTTP client with sensible defaults and advanced features
type Client struct {
	client           *http.Client
	retryConfig      RetryConfig
	userAgent        string
	debug            bool
	logOptions       LogOptions
	baseURL          string
	headers          http.Header
	cookies          []*http.Cookie
	queryParams      url.Values
	formData         url.Values
	authToken        string
	responseCallback func(*http.Response) error
	requestCallback  func(*http.Request) error
	forceContentType string
	mu               sync.RWMutex
}

// RetryConfig contains retry-related configuration
type RetryConfig struct {
	Enabled    bool
	Count      int
	MaxBackoff time.Duration
}

// LogOptions contains configuration for debug logging
type LogOptions struct {
	LogHeaders      bool
	LogBody         bool
	LogResponse     bool
	MaxBodyLogSize  int
	HeaderKeysToLog []string // New field to specify which header keys to log
}

// ClientOption allows customizing the Client
type ClientOption func(*Client) error

// StreamOption represents an option for streaming requests
type StreamOption func(*streamConfig)

// streamConfig holds the configuration for streaming requests
type streamConfig struct {
	ContentType string
	BufferSize  int
	Delimiter   byte
}

// WithStreamContentType sets the Content-Type header for streaming requests
func WithStreamContentType(contentType string) StreamOption {
	return func(sc *streamConfig) {
		sc.ContentType = contentType
	}
}

// WithStreamBufferSize sets the buffer size for streaming requests
func WithStreamBufferSize(size int) StreamOption {
	return func(sc *streamConfig) {
		sc.BufferSize = size
	}
}

// WithStreamDelimiter sets the delimiter for streaming requests
func WithStreamDelimiter(delimiter byte) StreamOption {
	return func(sc *streamConfig) {
		sc.Delimiter = delimiter
	}
}

// StreamConfig contains configuration for streaming requests
type StreamConfig struct {
	BufferSize int
	Delimiter  byte
}

// DefaultStreamConfig provides default settings for streaming
var DefaultStreamConfig = StreamConfig{
	BufferSize: 4096,
	Delimiter:  '\n',
}

// NewClient creates a new Client with default settings
func NewClient(options ...ClientOption) (*Client, error) {
	c := &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   defaultDialTimeout,
					KeepAlive: defaultKeepAlive,
				}).DialContext,
				MaxIdleConns:          defaultMaxIdleConns,
				IdleConnTimeout:       defaultIdleConnTimeout,
				TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
				ExpectContinueTimeout: defaultExpectContinueTimeout,
			},
		},
		retryConfig: RetryConfig{
			Enabled:    false, // Disable retry by default
			Count:      defaultRetryCount,
			MaxBackoff: defaultMaxBackoff,
		},
		userAgent: defaultUserAgent,
		debug:     false,
		logOptions: LogOptions{
			LogHeaders:      false,
			LogBody:         true,
			LogResponse:     true,
			MaxBodyLogSize:  defaultMaxBodyLogSize,
			HeaderKeysToLog: []string{}, // Don't log any headers by default
		},
		headers:     make(http.Header),
		queryParams: make(url.Values),
		formData:    make(url.Values),
	}

	// Set default headers
	c.headers.Set("User-Agent", c.userAgent)
	c.headers.Set("Accept", "application/json")
	c.headers.Set("Accept-Language", "en-US,en;q=0.9")

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, xerror.Wrap(err, "failed to apply client option")
		}
	}

	return c, nil
}

// WithTimeout sets the client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.client.Timeout = timeout
		return nil
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) error {
		c.retryConfig = config
		return nil
	}
}

// WithUserAgent sets the User-Agent header for all requests
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) error {
		if userAgent == "" {
			userAgent = defaultUserAgent
		}
		c.userAgent = userAgent
		c.headers.Set("User-Agent", userAgent)
		return nil
	}
}

// WithProxy sets a proxy for the client
func WithProxy(proxyURL string) ClientOption {
	return func(c *Client) error {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return xerror.Wrap(err, "failed to parse proxy URL")
		}
		transport, ok := c.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.Proxy = http.ProxyURL(proxy)
		c.client.Transport = transport
		return nil
	}
}

// WithDebug enables or disables debug mode
func WithDebug(debug bool) ClientOption {
	return func(c *Client) error {
		c.debug = debug
		return nil
	}
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// WithLogOptions sets the logging options for debug mode
func WithLogOptions(options LogOptions) ClientOption {
	return func(c *Client) error {
		c.logOptions = options
		return nil
	}
}

// SetLogOptions sets the logging options for debug mode
func (c *Client) SetLogOptions(options LogOptions) {
	c.logOptions = options
}

// SetBaseURL sets the base URL for all requests
func (c *Client) SetBaseURL(url string) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set base URL", "error", err, "url", url)
		return c
	}
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

// SetHeader sets a header for all requests
func (c *Client) SetHeader(key, value string) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set header", "error", err, "key", key)
		return c
	}
	c.headers.Set(key, value)
	return c
}

// SetHeaders sets multiple headers for all requests
func (c *Client) SetHeaders(headers map[string]string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set headers", "error", err)
		return c
	}
	for k, v := range headers {
		c.headers.Set(k, v)
	}
	return c
}

// AddCookie adds a cookie for all requests
func (c *Client) AddCookie(cookie *http.Cookie) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to add cookie", "error", err)
		return c
	}
	c.cookies = append(c.cookies, cookie)
	return c
}

// SetQueryParam sets a query parameter for all requests
func (c *Client) SetQueryParam(key, value string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set query parameter", "error", err)
		return c
	}
	if c.queryParams == nil {
		c.queryParams = make(url.Values)
	}
	c.queryParams.Set(key, value)
	return c
}

// SetQueryParams sets multiple query parameters for all requests
func (c *Client) SetQueryParams(params map[string]string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set query parameters", "error", err)
		return c
	}
	for k, v := range params {
		c.queryParams.Set(k, v)
	}
	return c
}

// SetFormData sets form data for all requests
func (c *Client) SetFormData(data map[string]string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set form data", "error", err)
		return c
	}
	if c.formData == nil {
		c.formData = make(url.Values)
	}
	for k, v := range data {
		c.formData.Set(k, v)
	}
	return c
}

// SetBasicAuth sets basic auth for all requests
func (c *Client) SetBasicAuth(username, password string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set basic auth", "error", err)
		return c
	}
	c.SetHeader("Authorization", "Basic "+basicAuth(username, password))
	return c
}

// SetBearerToken sets bearer auth token for all requests
func (c *Client) SetBearerToken(token string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set bearer token", "error", err)
		return c
	}
	c.authToken = token
	return c
}

// AddQueryParam adds a query parameter for all requests
func (c *Client) AddQueryParam(key string, value interface{}) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to add query parameter", "error", err)
		return c
	}
	strValue, err := xcast.ToString(value)
	if err != nil {
		xlog.Warn("Failed to convert value to string", "error", err)
		return c
	}
	c.queryParams.Add(key, strValue)
	return c
}

// AddFormDataField adds a form data field for all requests
func (c *Client) AddFormDataField(key string, value interface{}) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to add form data field", "error", err)
		return c
	}
	strValue, err := xcast.ToString(value)
	if err != nil {
		xlog.Warn("Failed to convert value to string", "error", err)
		return c
	}
	c.formData.Add(key, strValue)
	return c
}

// Request performs an HTTP request
func (c *Client) Request(ctx context.Context, method, req_url string, body interface{}) (*http.Response, error) {
	if err := c.validateClient(); err != nil {
		return nil, err
	}
	fullURL := req_url
	if !isAbsoluteURL(req_url) && c.baseURL != "" {
		fullURL = c.baseURL + req_url
	}

	// Validate URL
	if _, err := url.Parse(fullURL); err != nil {
		return nil, xerror.Wrap(err, "invalid URL")
	}

	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyReader = v
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, xerror.Wrap(err, "failed to marshal request body")
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}
	req, err := c.createRequest(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to create request")
	}
	return c.doRequest(req)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	if err := c.validateClient(); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, url, nil)
}

// Post sends a POST request to the specified URL
func (c *Client) Post(ctx context.Context, url string, body interface{}, options ...RequestOption) (*http.Response, error) {
	return c.doRequestWithBody(ctx, http.MethodPost, url, body, options...)
}

// PostForm sends a POST request with form data
func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return c.Post(ctx, url, strings.NewReader(data.Encode()), WithContentType("application/x-www-form-urlencoded"))
}

// Put sends a PUT request to the specified URL
func (c *Client) Put(ctx context.Context, url string, body interface{}, options ...RequestOption) (*http.Response, error) {
	return c.doRequestWithBody(ctx, http.MethodPut, url, body, options...)
}

// Patch sends a PATCH request to the specified URL
func (c *Client) Patch(ctx context.Context, url string, body interface{}, options ...RequestOption) (*http.Response, error) {
	return c.doRequestWithBody(ctx, http.MethodPatch, url, body, options...)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string) (*http.Response, error) {
	if err := c.validateClient(); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodDelete, url, nil)
}

// Do sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.doRequest(req)
}

// Head issues a HEAD request to the specified URL.
func (c *Client) Head(url string) (resp *http.Response, err error) {
	return c.Request(context.Background(), http.MethodHead, url, nil)
}

// JSONBody represents a JSON request body
type JSONBody map[string]interface{}

// FormData represents form data
type FormData map[string]string

// URLEncodedForm represents URL-encoded form data
type URLEncodedForm url.Values

// BinaryData represents binary data
type BinaryData []byte

// PostJSON sends a JSON-encoded POST request
func (c *Client) PostJSON(ctx context.Context, url string, body JSONBody) (*http.Response, error) {
	return c.Post(ctx, url, body, WithContentType("application/json"))
}

// PostFormData sends a multipart/form-data POST request
func (c *Client) PostFormData(ctx context.Context, url string, data FormData) (*http.Response, error) {
	return c.requestWithFormData(ctx, http.MethodPost, url, data)
}

// PostURLEncoded sends a x-www-form-urlencoded POST request
func (c *Client) PostURLEncoded(ctx context.Context, url string, data URLEncodedForm) (*http.Response, error) {
	return c.requestWithURLEncoded(ctx, http.MethodPost, url, data)
}

// PostBinary sends a binary POST request
func (c *Client) PostBinary(ctx context.Context, url string, data BinaryData) (*http.Response, error) {
	return c.requestWithBinary(ctx, http.MethodPost, url, data)
}

// PutJSON sends a JSON-encoded PUT request
func (c *Client) PutJSON(ctx context.Context, url string, body JSONBody) (*http.Response, error) {
	return c.Put(ctx, url, body, WithContentType("application/json"))
}

// PatchJSON sends a JSON-encoded PATCH request
func (c *Client) PatchJSON(ctx context.Context, url string, body JSONBody) (*http.Response, error) {
	return c.Patch(ctx, url, body, WithContentType("application/json"))
}

// PutFormData sends a multipart/form-data PUT request
func (c *Client) PutFormData(ctx context.Context, url string, data FormData) (*http.Response, error) {
	return c.requestWithFormData(ctx, http.MethodPut, url, data)
}

// PutURLEncoded sends a x-www-form-urlencoded PUT request
func (c *Client) PutURLEncoded(ctx context.Context, url string, data URLEncodedForm) (*http.Response, error) {
	return c.requestWithURLEncoded(ctx, http.MethodPut, url, data)
}

// PutBinary sends a binary PUT request
func (c *Client) PutBinary(ctx context.Context, url string, data BinaryData) (*http.Response, error) {
	return c.requestWithBinary(ctx, http.MethodPut, url, data)
}

// PatchFormData sends a multipart/form-data PATCH request
func (c *Client) PatchFormData(ctx context.Context, url string, data FormData) (*http.Response, error) {
	return c.requestWithFormData(ctx, http.MethodPatch, url, data)
}

// PatchURLEncoded sends a x-www-form-urlencoded PATCH request
func (c *Client) PatchURLEncoded(ctx context.Context, url string, data URLEncodedForm) (*http.Response, error) {
	return c.requestWithURLEncoded(ctx, http.MethodPatch, url, data)
}

// PatchBinary sends a binary PATCH request
func (c *Client) PatchBinary(ctx context.Context, url string, data BinaryData) (*http.Response, error) {
	return c.requestWithBinary(ctx, http.MethodPatch, url, data)
}

func (c *Client) requestWithJSON(ctx context.Context, method, url string, body JSONBody) (*http.Response, error) {
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, xerror.Wrap(err, "failed to marshal JSON body")
		}
	}
	fullURL := url
	if !isAbsoluteURL(url) && c.baseURL != "" {
		fullURL = c.baseURL + url
	}
	req, err := c.createRequest(ctx, method, fullURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req)
}

func (c *Client) requestWithFormData(ctx context.Context, method, url string, data FormData) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if data != nil {
		for key, value := range data {
			_ = writer.WriteField(key, value)
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, xerror.Wrap(err, "failed to create form-data")
	}
	fullURL := url
	if !isAbsoluteURL(url) && c.baseURL != "" {
		fullURL = c.baseURL + url
	}
	req, err := c.createRequest(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.doRequest(req)
}

func (c *Client) requestWithURLEncoded(ctx context.Context, method, req_url string, data URLEncodedForm) (*http.Response, error) {
	var encodedData string
	if data != nil {
		encodedData = url.Values(data).Encode()
	}
	fullURL := req_url
	if !isAbsoluteURL(req_url) && c.baseURL != "" {
		fullURL = c.baseURL + req_url
	}
	req, err := c.createRequest(ctx, method, fullURL, strings.NewReader(encodedData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doRequest(req)
}

func (c *Client) requestWithBinary(ctx context.Context, method, url string, data BinaryData) (*http.Response, error) {
	fullURL := url
	if !isAbsoluteURL(url) && c.baseURL != "" {
		fullURL = c.baseURL + url
	}
	var bodyReader io.Reader
	if data != nil {
		bodyReader = bytes.NewReader(data)
	}
	req, err := c.createRequest(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	return c.doRequest(req)
}

// StreamResponse performs a streaming request and returns a channel of response chunks
func (c *Client) StreamResponse(ctx context.Context, method, url string, body interface{}, options ...StreamOption) (<-chan []byte, <-chan error) {
	config := &streamConfig{
		BufferSize: DefaultStreamConfig.BufferSize,
		Delimiter:  DefaultStreamConfig.Delimiter,
	}

	for _, option := range options {
		option(config)
	}

	responseChan := make(chan []byte)
	errChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errChan)

		req, err := c.createRequestWithBody(ctx, method, url, body)
		if err != nil {
			errChan <- xerror.Wrap(err, "failed to create request")
			return
		}

		if config.ContentType != "" {
			req.Header.Set("Content-Type", config.ContentType)
		}

		if c.debug {
			c.logRequest(req)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			errChan <- xerror.Wrap(err, "failed to send request")
			return
		}
		defer resp.Body.Close()

		if c.debug && c.logOptions.LogResponse {
			c.logResponse(resp)
		}

		reader := bufio.NewReaderSize(resp.Body, config.BufferSize)
		for {
			line, err := reader.ReadBytes(config.Delimiter)
			if err != nil {
				if err == io.EOF {
					if len(line) > 0 {
						responseChan <- line
					}
					return
				}
				errChan <- xerror.Wrap(err, "error reading stream")
				return
			}

			select {
			case <-ctx.Done():
				errChan <- xerror.Wrap(ctx.Err(), "context cancelled during streaming")
				return
			case responseChan <- line:
			}
		}
	}()

	return responseChan, errChan
}

// StreamJSON performs a streaming request and returns a channel of JSON-decoded objects
func (c *Client) StreamJSON(ctx context.Context, method, url string, body interface{}, options ...StreamOption) (<-chan interface{}, <-chan error) {
	responseChan := make(chan interface{})
	errChan := make(chan error, 1)

	bytesChan, bytesErrChan := c.StreamResponse(ctx, method, url, body, options...)

	go func() {
		defer close(responseChan)
		defer close(errChan)

		for {
			select {
			case bytes, ok := <-bytesChan:
				if !ok {
					return
				}
				var data interface{}
				err := json.Unmarshal(bytes, &data)
				if err != nil {
					errChan <- xerror.Wrap(err, "failed to unmarshal JSON")
					return
				}
				responseChan <- data
			case err, ok := <-bytesErrChan:
				if !ok {
					return
				}
				errChan <- err
			}
		}
	}()

	return responseChan, errChan
}

// StreamSSE performs a streaming request for Server-Sent Events
func (c *Client) StreamSSE(ctx context.Context, url string) (<-chan *SSEEvent, <-chan error) {
	eventChan := make(chan *SSEEvent)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		req, err := c.createRequestWithBody(ctx, http.MethodGet, url, nil)
		if err != nil {
			errChan <- xerror.Wrap(err, "failed to create request")
			return
		}
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.client.Do(req)
		if err != nil {
			errChan <- xerror.Wrap(err, "failed to send request")
			return
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			event, err := parseSSEEvent(reader)
			if err != nil {
				if err == io.EOF {
					return
				}
				errChan <- xerror.Wrap(err, "error parsing SSE event")
				return
			}

			select {
			case <-ctx.Done():
				errChan <- xerror.Wrap(ctx.Err(), "context cancelled during SSE streaming")
				return
			case eventChan <- event:
			}
		}
	}()

	return eventChan, errChan
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID    string
	Event string
	Data  string
}

func parseSSEEvent(reader *bufio.Reader) (*SSEEvent, error) {
	event := &SSEEvent{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			// End of event
			return event, nil
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue // Ignore lines without colon
		}

		field := line[:colonIndex]
		value := strings.TrimPrefix(line[colonIndex+1:], " ")

		switch field {
		case "id":
			event.ID = value
		case "event":
			event.Event = value
		case "data":
			event.Data += value + "\n"
		}
	}
}

// WithCustomTransport sets a custom transport for the client
func WithCustomTransport(transport http.RoundTripper) ClientOption {
	return func(c *Client) error {
		c.client.Transport = transport
		return nil
	}
}

// WithTLSConfig sets a custom TLS configuration for the client
func WithTLSConfig(tlsConfig *tls.Config) ClientOption {
	return func(c *Client) error {
		transport, ok := c.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.TLSClientConfig = tlsConfig
		c.client.Transport = transport
		return nil
	}
}

// WithDialContext sets a custom DialContext function for the client
func WithDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) ClientOption {
	return func(c *Client) error {
		transport, ok := c.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.DialContext = dialContext
		c.client.Transport = transport
		return nil
	}
}

// WithResponseCallback sets a callback function to be called after each response
func WithResponseCallback(callback func(*http.Response) error) ClientOption {
	return func(c *Client) error {
		c.responseCallback = callback
		return nil
	}
}

// WithRequestCallback sets a callback function to be called before each request
func WithRequestCallback(callback func(*http.Request) error) ClientOption {
	return func(c *Client) error {
		c.requestCallback = callback
		return nil
	}
}

// startTimeKey is the key used to store the start time in the request context
var startTimeKey = struct{}{}

// GetClient returns the underlying http.Client
func (c *Client) GetClient() *http.Client {
	return c.client
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.debug {
		c.logRequest(req)
		req = req.WithContext(context.WithValue(req.Context(), startTimeKey, time.Now()))
	}

	var resp *http.Response
	var err error

	if c.retryConfig.Enabled {
		backoff := time.Millisecond * 100

		err = c.retryWithBackoff(req.Context(), func() error {
			clonedReq := req.Clone(req.Context())
			if clonedReq.Body != nil {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return err
				}
				req.Body = io.NopCloser(bytes.NewBuffer(body))
				clonedReq.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			resp, err = c.client.Do(clonedReq)
			if err != nil {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > c.retryConfig.MaxBackoff {
					backoff = c.retryConfig.MaxBackoff
				}
			}
			return err
		})
	} else {
		resp, err = c.client.Do(req)
	}

	if err != nil {
		return nil, xerror.Wrap(err, "request failed")
	}

	if c.debug && c.logOptions.LogResponse {
		c.logResponse(resp)
	}

	if c.responseCallback != nil {
		if err := c.responseCallback(resp); err != nil {
			xlog.Error("Response callback failed", "error", err)
		}
	}

	return resp, nil
}

func (c *Client) createRequest(ctx context.Context, method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to create request")
	}

	// Ensure User-Agent is set
	if c.userAgent == "" {
		c.userAgent = defaultUserAgent
	}
	req.Header.Set("User-Agent", c.userAgent)

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Set custom headers
	for k, v := range c.headers {
		req.Header[k] = v
	}

	// Set cookies
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// Set query parameters
	q := req.URL.Query()
	for k, v := range c.queryParams {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Set auth token
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	if c.forceContentType != "" {
		req.Header.Set("Content-Type", c.forceContentType)
	}

	return req, nil
}

func (c *Client) retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error
	for i := 0; i < c.retryConfig.Count; i++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i == c.retryConfig.Count-1 {
			break
		}

		backoffDuration := time.Duration(float64(c.retryConfig.MaxBackoff) * (1 - math.Exp(float64(-i))))
		timer := time.NewTimer(backoffDuration)

		select {
		case <-ctx.Done():
			timer.Stop()
			return xerror.Wrap(ctx.Err(), "context cancelled during retry")
		case <-timer.C:
			// Continue to the next iteration
		}
	}

	return xerror.Wrap(lastErr, "max retries reached")
}

func (c *Client) logRequest(req *http.Request) {
	if req == nil {
		xlog.Warn("Attempted to log nil request")
		return
	}

	reqID := xlog.GenReqID()
	xlog.Debug("HTTP Request",
		"req_id", reqID,
		"method", req.Method,
		"url", req.URL.String(),
		"proto", req.Proto,
		"content_length", req.ContentLength,
		"remote_addr", req.RemoteAddr,
	)

	if c.logOptions.LogHeaders && req.Header != nil {
		for key, values := range req.Header {
			if c.shouldLogHeader(key) {
				xlog.Debug("Request Header",
					"req_id", reqID,
					"key", key,
					"value", strings.Join(values, ", "))
			}
		}
	}

	if c.logOptions.LogBody && req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			xlog.Warn("Failed to read request body",
				"req_id", reqID,
				"error", err)
		} else {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) > c.logOptions.MaxBodyLogSize {
				xlog.Debug("Request Body (truncated)",
					"req_id", reqID,
					"body", string(body[:c.logOptions.MaxBodyLogSize]))
			} else {
				xlog.Debug("Request Body",
					"req_id", reqID,
					"body", string(body))
			}
		}
	}
}

func (c *Client) logResponse(resp *http.Response) {
	if resp == nil {
		xlog.Warn("Attempted to log nil response")
		return
	}

	reqID := xlog.GenReqID()
	xlog.Debug("HTTP Response",
		"req_id", reqID,
		"status", resp.Status,
		"status_code", resp.StatusCode,
		"proto", resp.Proto,
		"content_length", resp.ContentLength,
	)

	if c.logOptions.LogHeaders && resp.Header != nil {
		for key, values := range resp.Header {
			if c.shouldLogHeader(key) {
				xlog.Debug("Response Header",
					"req_id", reqID,
					"key", key,
					"value", strings.Join(values, ", "))
			}
		}
	}

	if c.logOptions.LogBody && resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			xlog.Warn("Failed to read response body",
				"req_id", reqID,
				"error", err)
		} else {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) > c.logOptions.MaxBodyLogSize {
				xlog.Debug("Response Body (truncated)",
					"req_id", reqID,
					"body", string(body[:c.logOptions.MaxBodyLogSize]))
			} else {
				xlog.Debug("Response Body",
					"req_id", reqID,
					"body", string(body))
			}
		}
	}

	if resp.Request != nil && resp.Request.Context() != nil {
		if startTimeValue := resp.Request.Context().Value(startTimeKey); startTimeValue != nil {
			if startTime, ok := startTimeValue.(time.Time); ok {
				duration := time.Since(startTime)
				xlog.Debug("Response Timing",
					"req_id", reqID,
					"duration", duration.String(),
					"start_time", startTime.Format(time.RFC3339),
					"end_time", time.Now().Format(time.RFC3339),
				)
			}
		}
	}
}

func (c *Client) shouldLogHeader(key string) bool {
	if len(c.logOptions.HeaderKeysToLog) == 0 {
		return true // Log all headers if no specific keys are set
	}
	for _, allowedKey := range c.logOptions.HeaderKeysToLog {
		if strings.EqualFold(key, allowedKey) {
			return true
		}
	}
	return false
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// WithBearerToken sets bearer auth token for all requests
func WithBearerToken(token string) ClientOption {
	return func(c *Client) error {
		c.SetBearerToken(token)
		return nil
	}
}

// WithBaseURL sets the base URL for all requests
func WithBaseURL(url string) ClientOption {
	return func(c *Client) error {
		c.SetBaseURL(url)
		return nil
	}
}

// isAbsoluteURL checks if the given URL is absolute
func isAbsoluteURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func (c *Client) createRequestWithBody(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyReader = v
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, xerror.Wrap(err, "failed to marshal request body")
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}
	req, err := c.createRequest(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if c.forceContentType == "" {
		switch body.(type) {
		case JSONBody:
			req.Header.Set("Content-Type", "application/json")
		case FormData:
		case URLEncodedForm:
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case BinaryData:
			req.Header.Set("Content-Type", "application/octet-stream")
		}
	}

	return req, nil
}

// SetForceContentType sets the Content-Type header for all requests
func (c *Client) SetForceContentType(contentType string) *Client {
	if err := c.validateClient(); err != nil {
		xlog.Error("Failed to set force content type", "error", err)
		return c
	}
	c.forceContentType = contentType
	return c
}

var defaultClient = x.Must1(NewClient())

// GetDefaultClient returns the default client
func GetDefaultClient() *Client {
	return defaultClient
}

// Add this new method to validate the client
func (c *Client) validateClient() error {
	if c == nil {
		return xerror.New("client is nil")
	}
	if c.client == nil {
		return xerror.New("http client is nil")
	}
	return nil
}

// GetJSONAndDecode sends a GET request and decodes the JSON response into the provided interface
func (c *Client) GetJSONAndDecode(ctx context.Context, url string, result interface{}) error {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return xerror.Wrap(err, "failed to send GET request")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return xerror.Wrap(err, "failed to decode response body")
	}

	return nil
}

// PostJSONAndDecode sends a POST request with JSON body and decodes the response into the provided interface
func (c *Client) PostJSONAndDecode(ctx context.Context, url string, body JSONBody, result interface{}) error {
	resp, err := c.PostJSON(ctx, url, body)
	if err != nil {
		return xerror.Wrap(err, "failed to send POST request")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return xerror.Wrap(err, "failed to decode response body")
	}

	return nil
}

// RequestOption represents an option for customizing requests
type RequestOption func(*requestOptions)

type requestOptions struct {
	contentType string
	headers     http.Header
}

// WithContentType sets the Content-Type header for the request
func WithContentType(contentType string) RequestOption {
	return func(opts *requestOptions) {
		opts.contentType = contentType
	}
}

// WithHeader adds a header to the request
func WithHeader(key, value string) RequestOption {
	return func(opts *requestOptions) {
		if opts.headers == nil {
			opts.headers = make(http.Header)
		}
		opts.headers.Add(key, value)
	}
}

func (c *Client) doRequestWithBody(ctx context.Context, method, url string, body interface{}, options ...RequestOption) (*http.Response, error) {
	if err := c.validateClient(); err != nil {
		return nil, err
	}

	opts := &requestOptions{}
	for _, option := range options {
		option(opts)
	}

	var bodyReader io.Reader
	var contentType string

	switch v := body.(type) {
	case nil:
		// No body
	case io.Reader:
		bodyReader = v
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = bytes.NewReader(v)
	case JSONBody:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, xerror.Wrap(err, "failed to marshal JSON body")
		}
		bodyReader = bytes.NewReader(jsonData)
		contentType = "application/json"
	default:
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, xerror.Wrap(err, "failed to marshal body")
		}
		bodyReader = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	if opts.contentType != "" {
		contentType = opts.contentType
	}

	fullURL := url
	if !isAbsoluteURL(url) && c.baseURL != "" {
		fullURL = c.baseURL + url
	}

	req, err := c.createRequest(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, values := range opts.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return c.doRequest(req)
}

// Request represents an HTTP request and provides methods for building and executing it
type Request struct {
	client      *Client
	httpRequest *http.Request
	response    *http.Response
	body        []byte
	err         error
	receivedAt  time.Time
}

// NewRequest creates a new Request object for building and executing a request
func (c *Client) NewRequest() *Request {
	return &Request{
		client: c,
		httpRequest: &http.Request{
			Header: make(http.Header),
			URL:    &url.URL{},
		},
	}
}

// SetBody sets the request body
func (r *Request) SetBody(body interface{}) *Request {
	switch v := body.(type) {
	case string:
		r.httpRequest.Body = io.NopCloser(strings.NewReader(v))
	case []byte:
		r.httpRequest.Body = io.NopCloser(bytes.NewReader(v))
	case io.Reader:
		r.httpRequest.Body = io.NopCloser(v)
	default:
		jsonBody, err := json.Marshal(body)
		if err != nil {
			r.err = xerror.Wrap(err, "failed to marshal request body")
			return r
		}
		r.httpRequest.Body = io.NopCloser(bytes.NewReader(jsonBody))
		r.httpRequest.Header.Set("Content-Type", "application/json")
	}
	return r
}

// SetHeader sets a header for the request
func (r *Request) SetHeader(key, value string) *Request {
	r.httpRequest.Header.Set(key, value)
	return r
}

// SetQueryParam sets a query parameter for the request
func (r *Request) SetQueryParam(key, value string) *Request {
	q := r.httpRequest.URL.Query()
	q.Set(key, value)
	r.httpRequest.URL.RawQuery = q.Encode()
	return r
}

const (
	// Content type constants
	ContentTypeJSON          = "application/json"
	ContentTypeXML           = "application/xml"
	ContentTypeForm          = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm = "multipart/form-data"
	ContentTypeTextPlain     = "text/plain"
	ContentTypeOctetStream   = "application/octet-stream"
	ContentTypeHTML          = "text/html"

	// Content type constants with charset
	ContentTypeJSONUTF8      = "application/json; charset=utf-8"
	ContentTypeXMLUTF8       = "application/xml; charset=utf-8"
	ContentTypeFormUTF8      = "application/x-www-form-urlencoded; charset=utf-8"
	ContentTypeTextPlainUTF8 = "text/plain; charset=utf-8"
	ContentTypeHTMLUTF8      = "text/html; charset=utf-8"

	// Additional charsets
	ContentTypeJSONGBK      = "application/json; charset=gbk"
	ContentTypeXMLGBK       = "application/xml; charset=gbk"
	ContentTypeTextPlainGBK = "text/plain; charset=gbk"
	ContentTypeHTMLGBK      = "text/html; charset=gbk"

	ContentTypeJSONGB2312      = "application/json; charset=gb2312"
	ContentTypeXMLGB2312       = "application/xml; charset=gb2312"
	ContentTypeTextPlainGB2312 = "text/plain; charset=gb2312"
	ContentTypeHTMLGB2312      = "text/html; charset=gb2312"

	// Default content type
	DefaultContentType = ContentTypeJSONUTF8
)

// Get sends a GET request with specified content type
func (r *Request) Get(ctx context.Context, contentType, url string) *Request {
	if contentType == "" {
		contentType = DefaultContentType
	}
	r.SetHeader("Content-Type", contentType)
	return r.Execute(ctx, http.MethodGet, url)
}

// Post sends a POST request with specified content type
func (r *Request) Post(ctx context.Context, contentType, url string) *Request {
	if contentType == "" {
		contentType = DefaultContentType
	}
	r.SetHeader("Content-Type", contentType)
	return r.Execute(ctx, http.MethodPost, url)
}

// Put sends a PUT request with specified content type
func (r *Request) Put(ctx context.Context, contentType, url string) *Request {
	if contentType == "" {
		contentType = DefaultContentType
	}
	r.SetHeader("Content-Type", contentType)
	return r.Execute(ctx, http.MethodPut, url)
}

// Patch sends a PATCH request with specified content type
func (r *Request) Patch(ctx context.Context, contentType, url string) *Request {
	if contentType == "" {
		contentType = DefaultContentType
	}
	r.SetHeader("Content-Type", contentType)
	return r.Execute(ctx, http.MethodPatch, url)
}

// Delete sends a DELETE request with specified content type
func (r *Request) Delete(ctx context.Context, contentType, url string) *Request {
	if contentType == "" {
		contentType = DefaultContentType
	}
	r.SetHeader("Content-Type", contentType)
	return r.Execute(ctx, http.MethodDelete, url)
}

// Execute sends the request
func (r *Request) Execute(ctx context.Context, method, requestURL string) *Request {
	if r.err != nil {
		return r
	}

	fullURL := requestURL
	if !isAbsoluteURL(requestURL) && r.client.baseURL != "" {
		fullURL = r.client.baseURL + requestURL
	}

	// Parse the URL and merge query parameters
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		r.err = xerror.Wrap(err, "failed to parse URL")
		return r
	}

	// Merge existing query parameters from the Request with the URL's query parameters
	mergedQuery := parsedURL.Query()
	for k, v := range r.httpRequest.URL.Query() {
		for _, value := range v {
			mergedQuery.Add(k, value)
		}
	}
	parsedURL.RawQuery = mergedQuery.Encode()
	fullURL = parsedURL.String()

	req, err := r.client.createRequest(ctx, method, fullURL, r.httpRequest.Body)
	if err != nil {
		r.err = err
		return r
	}

	// Copy headers from r.httpRequest to req
	for key, values := range r.httpRequest.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	r.httpRequest = req
	r.response, r.err = r.client.doRequest(req)
	r.receivedAt = time.Now()

	if r.err == nil {
		r.body, r.err = io.ReadAll(r.response.Body)
		r.response.Body.Close()
		r.response.Body = io.NopCloser(bytes.NewBuffer(r.body))
	}

	return r
}

// Result stores the response body into the provided interface
func (r *Request) Result(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.Unmarshal(r.body, v)
}

// String returns the response body as a string
func (r *Request) String() string {
	if r.err != nil {
		return ""
	}
	return string(r.body)
}

// Bytes returns the response body as a byte slice
func (r *Request) Bytes() []byte {
	if r.err != nil {
		return nil
	}
	return r.body
}

// Error returns any error that occurred during the request
func (r *Request) Error() error {
	return r.err
}

// StatusCode returns the response status code
func (r *Request) StatusCode() int {
	if r.response == nil {
		return 0
	}
	return r.response.StatusCode
}

// Header returns the response headers
func (r *Request) Header() http.Header {
	if r.response == nil {
		return nil
	}
	return r.response.Header
}

// Time returns the time when the response was received
func (r *Request) Time() time.Time {
	return r.receivedAt
}
