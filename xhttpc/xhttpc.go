package xhttpc

import (
	"bytes"
	"context"
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

// Client is a high-performance HTTP client with sensible defaults and advanced features
type Client struct {
	client     *http.Client
	retryCount int
	maxBackoff time.Duration
	userAgent  string
	debug      bool
	logOptions LogOptions
}

// LogOptions contains configuration for debug logging
type LogOptions struct {
	LogHeaders     bool
	LogBody        bool
	LogResponse    bool
	MaxBodyLogSize int
}

// ClientOption allows customizing the Client
type ClientOption func(*Client)

// NewClient creates a new Client with default settings
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		retryCount: 3,
		maxBackoff: 30 * time.Second,
		userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		debug:      false,
		logOptions: LogOptions{
			LogHeaders:     false,
			LogBody:        true,
			LogResponse:    true,
			MaxBodyLogSize: 1024, // Default to 1KB
		},
	}

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

// WithRetryCount sets the number of retries for failed requests
func WithRetryCount(count int) ClientOption {
	return func(c *Client) {
		c.retryCount = count
	}
}

// WithMaxBackoff sets the maximum backoff duration
func WithMaxBackoff(maxBackoff time.Duration) ClientOption {
	return func(c *Client) {
		c.maxBackoff = maxBackoff
	}
}

// WithUserAgent sets the User-Agent header for all requests
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithHTTPProxy sets an HTTP proxy for the client
func WithHTTPProxy(proxyURL string) ClientOption {
	return func(c *Client) {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			// Handle error (e.g., log it)
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

// WithSOCKS5Proxy sets a SOCKS5 proxy for the client
func WithSOCKS5Proxy(proxyAddr string) ClientOption {
	return func(c *Client) {
		transport, ok := c.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, "tcp", proxyAddr)
		}
		c.client.Transport = transport
	}
}

// WithDebug enables or disables debug mode
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithLogOptions sets the logging options for debug mode
func WithLogOptions(options LogOptions) ClientOption {
	return func(c *Client) {
		c.logOptions = options
	}
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, url, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, url, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, url, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, url, nil)
}

func (c *Client) doRequest(ctx context.Context, method, reqURL string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case url.Values:
			bodyReader = strings.NewReader(v.Encode())
			contentType = "application/x-www-form-urlencoded"
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

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("User-Agent", c.userAgent)

	if c.debug {
		c.logRequest(req)
	}

	var resp *http.Response
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

	if c.debug && c.logOptions.LogResponse {
		c.logResponse(resp)
	}

	return resp, nil
}

func (c *Client) retryWithBackoff(ctx context.Context, operation func() error) error {
	var err error
	for i := 0; i < c.retryCount; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i == c.retryCount-1 {
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
	max := float64(c.maxBackoff)
	temp := math.Min(max, math.Pow(2, float64(attempt))*backoff)
	rand.Seed(time.Now().UnixNano())
	backoff = temp/2 + rand.Float64()*(temp/2)
	return time.Duration(backoff)
}

// SetRequestHeaders sets custom headers for all requests
func (c *Client) SetRequestHeaders(headers map[string]string) {
	c.client.Transport = &headerTransport{
		base:    c.client.Transport,
		headers: headers,
	}
}

type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

func (c *Client) logRequest(req *http.Request) {
	xlog.Info("HTTP Request", "method", req.Method, "url", req.URL.String())

	if c.logOptions.LogHeaders {
		for key, values := range req.Header {
			for _, value := range values {
				xlog.Info("Request Header", "key", key, "value", value)
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
}

func (c *Client) logResponse(resp *http.Response) {
	xlog.Info("HTTP Response", "status", resp.Status)

	if c.logOptions.LogHeaders {
		for key, values := range resp.Header {
			for _, value := range values {
				xlog.Info("Response Header", "key", key, "value", value)
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
}
