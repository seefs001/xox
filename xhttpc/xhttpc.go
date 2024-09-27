package xhttpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/seefs001/xox/xlog"
)

const (
	defaultTimeout               = 10 * time.Second
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
	client      *http.Client
	retryConfig RetryConfig
	userAgent   string
	debug       bool
	logOptions  LogOptions
	baseURL     string
	headers     http.Header
	cookies     []*http.Cookie
	queryParams url.Values
	formData    url.Values
	authToken   string
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
type ClientOption func(*Client)

// NewClient creates a new Client with default settings
func NewClient(options ...ClientOption) *Client {
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
			Enabled:    false,
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
		option(c)
	}

	return c
}

// WithTimeout sets the client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithUserAgent sets the User-Agent header for all requests
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
		c.headers.Set("User-Agent", userAgent)
	}
}

// WithProxy sets a proxy for the client
func WithProxy(proxyURL string) ClientOption {
	return func(c *Client) {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			xlog.Error("Failed to parse proxy URL", "error", err)
			return
		}
		transport, ok := c.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.Proxy = http.ProxyURL(proxy)
		c.client.Transport = transport
	}
}

// WithDebug enables or disables debug mode
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// WithLogOptions sets the logging options for debug mode
func WithLogOptions(options LogOptions) ClientOption {
	return func(c *Client) {
		c.logOptions = options
	}
}

// SetLogOptions sets the logging options for debug mode
func (c *Client) SetLogOptions(options LogOptions) {
	c.logOptions = options
}

// SetBaseURL sets the base URL for all requests
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// SetHeader sets a header for all requests
func (c *Client) SetHeader(key, value string) *Client {
	c.headers.Set(key, value)
	return c
}

// SetHeaders sets multiple headers for all requests
func (c *Client) SetHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.headers.Set(k, v)
	}
	return c
}

// AddCookie adds a cookie for all requests
func (c *Client) AddCookie(cookie *http.Cookie) *Client {
	c.cookies = append(c.cookies, cookie)
	return c
}

// SetQueryParam sets a query parameter for all requests
func (c *Client) SetQueryParam(key, value string) *Client {
	c.queryParams.Set(key, value)
	return c
}

// SetQueryParams sets multiple query parameters for all requests
func (c *Client) SetQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.queryParams.Set(k, v)
	}
	return c
}

// SetFormData sets form data for all requests
func (c *Client) SetFormData(data map[string]string) *Client {
	for k, v := range data {
		c.formData.Set(k, v)
	}
	return c
}

// SetBasicAuth sets basic auth for all requests
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.SetHeader("Authorization", "Basic "+basicAuth(username, password))
	return c
}

// SetBearerToken sets bearer auth token for all requests
func (c *Client) SetBearerToken(token string) *Client {
	c.authToken = token
	return c
}

// AddQueryParam adds a query parameter for all requests
func (c *Client) AddQueryParam(key string, value interface{}) *Client {
	c.queryParams.Add(key, fmt.Sprintf("%v", value))
	return c
}

// AddFormDataField adds a form data field for all requests
func (c *Client) AddFormDataField(key string, value interface{}) *Client {
	c.formData.Add(key, fmt.Sprintf("%v", value))
	return c
}

// Request performs an HTTP request
func (c *Client) Request(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, method, url, body)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, url, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, url, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, url, body)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, url, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, url, nil)
}

// Do sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.doRequest(req.Context(), req.Method, req.URL.String(), req.Body)
}

// Head issues a HEAD request to the specified URL.
func (c *Client) Head(url string) (resp *http.Response, err error) {
	return c.Request(context.Background(), http.MethodHead, url, nil)
}

// PostForm issues a POST request to the specified URL, with data's keys and
// values URL-encoded as the request body.
func (c *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(context.Background(), url, data)
}

// PostJSON performs a POST request with a JSON body
func (c *Client) PostJSON(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
	}
	return c.Post(ctx, url, bytes.NewReader(jsonBody))
}

// PutJSON performs a PUT request with a JSON body
func (c *Client) PutJSON(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
	}
	return c.Put(ctx, url, bytes.NewReader(jsonBody))
}

// PatchJSON performs a PATCH request with a JSON body
func (c *Client) PatchJSON(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
	}
	return c.Patch(ctx, url, bytes.NewReader(jsonBody))
}

// StreamResponse performs a streaming request and returns a channel of response chunks
func (c *Client) StreamResponse(ctx context.Context, method, url string, body interface{}) (<-chan []byte, <-chan error) {
	responseChan := make(chan []byte)
	errChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errChan)

		req, err := c.createRequest(ctx, method, url, body)
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		if c.debug {
			c.logRequest(req)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if c.debug && c.logOptions.LogResponse {
			c.logResponse(resp)
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				errChan <- fmt.Errorf("error reading stream: %w", err)
				return
			}

			select {
			case <-ctx.Done():
				return
			case responseChan <- line:
			}
		}
	}()

	return responseChan, errChan
}

// startTimeKey is the key used to store the start time in the request context
var startTimeKey = struct{}{}

func (c *Client) doRequest(ctx context.Context, method, reqURL string, body interface{}) (*http.Response, error) {
	fullURL := c.baseURL + reqURL
	req, err := c.createRequest(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	if c.debug {
		c.logRequest(req)
		// Set the start time before sending the request
		ctx = context.WithValue(ctx, startTimeKey, time.Now())
		req = req.WithContext(ctx)
	}

	var resp *http.Response
	if c.retryConfig.Enabled {
		operation := func() error {
			var err error
			resp, err = c.client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to send request: %w", err)
			}
			if resp.StatusCode >= 500 {
				return fmt.Errorf("server error: %d", resp.StatusCode)
			}
			return nil
		}

		err = c.retryWithBackoff(ctx, operation)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err = c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	if c.debug && c.logOptions.LogResponse {
		c.logResponse(resp)
	}

	return resp, nil
}

func (c *Client) createRequest(ctx context.Context, method, reqURL string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case url.Values:
			bodyReader = strings.NewReader(v.Encode())
			contentType = "application/x-www-form-urlencoded"
		case io.Reader:
			bodyReader = v
			// Check if the reader is a *bytes.Reader containing JSON data
			if jsonReader, ok := v.(*bytes.Reader); ok {
				jsonData, _ := io.ReadAll(jsonReader)
				if json.Valid(jsonData) {
					contentType = "application/json"
					bodyReader = bytes.NewReader(jsonData)
				}
			}
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
			contentType = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set method-specific headers
	switch method {
	case http.MethodGet:
		req.Header.Set("Cache-Control", "no-cache")
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	// Set default headers
	req.Header.Set("User-Agent", c.userAgent)
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

	// Set form data
	if (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) && contentType == "application/x-www-form-urlencoded" {
		req.PostForm = c.formData
	}

	// Set auth token
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	return req, nil
}

func (c *Client) retryWithBackoff(ctx context.Context, operation func() error) error {
	var err error
	for i := 0; i < c.retryConfig.Count; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i == c.retryConfig.Count-1 {
			break
		}

		backoffDuration := c.calculateBackoff(i)
		timer := time.NewTimer(backoffDuration)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue to the next iteration
		}
	}

	return fmt.Errorf("max retries reached: %w", err)
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(time.Second)
	max := float64(c.retryConfig.MaxBackoff)
	temp := math.Min(max, math.Pow(2, float64(attempt))*backoff)
	backoff = temp/2 + rand.Float64()*(temp/2)
	return time.Duration(backoff)
}

func (c *Client) logRequest(req *http.Request) {
	xlog.Info("HTTP Request", "method", req.Method, "url", req.URL.String())

	if c.logOptions.LogHeaders {
		for key, values := range req.Header {
			if c.shouldLogHeader(key) {
				for _, value := range values {
					xlog.Info("Request Header", "key", key, "value", value)
				}
			}
		}
	}

	if c.logOptions.LogBody && req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the body
		if len(body) > c.logOptions.MaxBodyLogSize {
			xlog.Info("Request Body (truncated)", "body", string(body[:c.logOptions.MaxBodyLogSize]))
		} else {
			xlog.Info("Request Body", "body", string(body))
		}
	}

	xlog.Info("Request Details",
		"method", req.Method,
		"url", req.URL.String(),
		"protocol", req.Proto,
		"host", req.Host,
		"content_length", req.ContentLength,
		"transfer_encoding", req.TransferEncoding,
		"close", req.Close,
		"trailer", req.Trailer,
		"remote_addr", req.RemoteAddr,
		"request_uri", req.RequestURI,
	)

	if c.debug {
		xlog.Info("Custom Client Settings",
			"base_url", c.baseURL,
			"user_agent", c.userAgent,
			"retry_enabled", c.retryConfig.Enabled,
			"retry_count", c.retryConfig.Count,
			"max_backoff", c.retryConfig.MaxBackoff,
			"debug_mode", c.debug,
		)
	}
}

func (c *Client) logResponse(resp *http.Response) {
	xlog.Info("HTTP Response", "status", resp.Status, "status_code", resp.StatusCode)

	if c.logOptions.LogHeaders {
		for key, values := range resp.Header {
			if c.shouldLogHeader(key) {
				for _, value := range values {
					xlog.Info("Response Header", "key", key, "value", value)
				}
			}
		}
	}

	if c.logOptions.LogBody && resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the body
		if len(body) > c.logOptions.MaxBodyLogSize {
			xlog.Info("Response Body (truncated)", "body", string(body[:c.logOptions.MaxBodyLogSize]))
		} else {
			xlog.Info("Response Body", "body", string(body))
		}
	}

	xlog.Info("Response Details",
		"status", resp.Status,
		"status_code", resp.StatusCode,
		"protocol", resp.Proto,
		"content_length", resp.ContentLength,
		"transfer_encoding", resp.TransferEncoding,
		"uncompressed", resp.Uncompressed,
		"trailer", resp.Trailer,
	)

	if c.debug {
		// Safely get and use startTimeKey
		if startTimeValue := resp.Request.Context().Value(startTimeKey); startTimeValue != nil {
			if startTime, ok := startTimeValue.(time.Time); ok {
				duration := time.Since(startTime)
				xlog.Info("Response Timing",
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
		return false // Don't log any headers if no specific keys are set
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
	return func(c *Client) {
		c.SetBearerToken(token)
	}
}

// WithBaseURL sets the base URL for all requests
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.SetBaseURL(url)
	}
}
