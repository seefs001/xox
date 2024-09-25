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
	"time"
)

// Client is a high-performance HTTP client with sensible defaults and advanced features
type Client struct {
	client     *http.Client
	retryCount int
	maxBackoff time.Duration
	userAgent  string
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

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

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
