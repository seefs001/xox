package xhttpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client, "NewClient should not return nil")

	assert.Equal(t, 3, client.retryCount, "Default retryCount should be 3")
	assert.Equal(t, 30*time.Second, client.maxBackoff, "Default maxBackoff should be 30s")
	assert.NotEmpty(t, client.userAgent, "Default userAgent should be set")
}

func TestClientOptions(t *testing.T) {
	client := NewClient(
		WithTimeout(5*time.Second),
		WithRetryCount(5),
		WithMaxBackoff(10*time.Second),
		WithUserAgent("TestAgent"),
	)

	assert.Equal(t, 5*time.Second, client.client.Timeout, "Timeout should be 5s")
	assert.Equal(t, 5, client.retryCount, "RetryCount should be 5")
	assert.Equal(t, 10*time.Second, client.maxBackoff, "MaxBackoff should be 10s")
	assert.Equal(t, "TestAgent", client.userAgent, "UserAgent should be TestAgent")
}

func TestClientMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("GET response"))
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	client := NewClient()

	t.Run("GET", func(t *testing.T) {
		resp, err := client.Get(context.Background(), server.URL)
		require.NoError(t, err, "GET request should not fail")
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "GET response", string(body), "Unexpected GET response")
	})

	t.Run("POST", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		resp, err := client.Post(context.Background(), server.URL, data)
		require.NoError(t, err, "POST request should not fail")
		defer resp.Body.Close()

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err, "Should decode JSON response")
		assert.Equal(t, "value", result["key"], "Unexpected POST response")
	})

	t.Run("PUT", func(t *testing.T) {
		data := map[string]string{"key": "updated"}
		resp, err := client.Put(context.Background(), server.URL, data)
		require.NoError(t, err, "PUT request should not fail")
		defer resp.Body.Close()

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err, "Should decode JSON response")
		assert.Equal(t, "updated", result["key"], "Unexpected PUT response")
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.Delete(context.Background(), server.URL)
		require.NoError(t, err, "DELETE request should not fail")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status 204")
	})
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithRetryCount(3), WithMaxBackoff(100*time.Millisecond))

	resp, err := client.Get(context.Background(), server.URL)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	assert.Equal(t, 3, attempts, "Expected 3 attempts")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestSetRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"), "X-Custom-Header should be set")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	client.SetRequestHeaders(map[string]string{
		"X-Custom-Header": "test-value",
	})

	resp, err := client.Get(context.Background(), server.URL)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Get(ctx, server.URL)
	assert.Error(t, err, "Expected error due to context cancellation")
	assert.Equal(t, context.DeadlineExceeded, err, "Expected DeadlineExceeded error")
}
