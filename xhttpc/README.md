# xhttpc

`xhttpc` is a high-performance HTTP client for Go with sensible defaults and advanced features.

## Features

- Retry mechanism with exponential backoff
- Customizable timeout and transport options
- Support for various request types (JSON, form data, URL-encoded, binary)
- Streaming support for responses and Server-Sent Events (SSE)
- Debug logging with customizable options
- Easy-to-use fluent interface for request building
- Context support for cancellation and timeouts
- Automatic handling of base URLs
- Custom header and cookie management
- Proxy support

## Installation

```bash
go get github.com/seefs001/xox/xhttpc
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/seefs001/xox/xhttpc"
)

func main() {
    client, err := xhttpc.NewClient()
    if err != nil {
        panic(err)
    }

    resp, err := client.Get(context.Background(), "https://api.example.com/users")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // Process the response...
    fmt.Println("Status:", resp.Status)
}
```

## Usage

### Creating a Client

```go
client, err := xhttpc.NewClient(
    xhttpc.WithTimeout(10*time.Second),
    xhttpc.WithRetryConfig(xhttpc.RetryConfig{
        Enabled:    true,
        Count:      3,
        MaxBackoff: 30 * time.Second,
    }),
    xhttpc.WithUserAgent("MyApp/1.0"),
    xhttpc.WithDebug(true),
)
```

### Making Requests

#### GET Request

```go
resp, err := client.Get(context.Background(), "https://api.example.com/users")
```

#### POST Request (JSON)

```go
data := map[string]interface{}{
    "name": "John Doe",
    "age":  30,
}
resp, err := client.PostJSON(context.Background(), "https://api.example.com/users", data)
```

#### PUT Request (Form Data)

```go
formData := xhttpc.FormData{
    "key1": "value1",
    "key2": "value2",
}
resp, err := client.PutFormData(context.Background(), "https://api.example.com/update", formData)
```

#### DELETE Request

```go
resp, err := client.Delete(context.Background(), "https://api.example.com/users/123")
```

### Streaming Responses

```go
ch, errCh := client.StreamResponse(context.Background(), "GET", "https://api.example.com/stream",
    nil,
    xhttpc.WithStreamBufferSize(4096),
    xhttpc.WithStreamDelimiter('\n'),
)

for {
    select {
    case chunk, ok := <-ch:
        if !ok {
            return
        }
        // Process chunk...
    case err := <-errCh:
        // Handle error...
        return
    }
}
```

### Server-Sent Events (SSE)

```go
eventCh, errCh := client.StreamSSE(context.Background(), "https://api.example.com/events")

for {
    select {
    case event, ok := <-eventCh:
        if !ok {
            return
        }
        fmt.Printf("Received event: %+v\n", event)
    case err := <-errCh:
        // Handle error...
        return
    }
}
```

## Advanced Configuration

### Custom Transport

```go
client, err := xhttpc.NewClient(
    xhttpc.WithCustomTransport(&http.Transport{
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    }),
)
```

### TLS Configuration

```go
tlsConfig := &tls.Config{
    // Custom TLS settings...
}
client, err := xhttpc.NewClient(
    xhttpc.WithTLSConfig(tlsConfig),
)
```

### Proxy Configuration

```go
client, err := xhttpc.NewClient(
    xhttpc.WithProxy("http://proxy.example.com:8080"),
)
```

### Request and Response Callbacks

```go
client, err := xhttpc.NewClient(
    xhttpc.WithRequestCallback(func(req *http.Request) error {
        // Modify request before sending
        return nil
    }),
    xhttpc.WithResponseCallback(func(resp *http.Response) error {
        // Process response after receiving
        return nil
    }),
)
```

## Debug Logging

Enable debug logging for detailed request and response information:

```go
client.SetDebug(true)
client.SetLogOptions(xhttpc.LogOptions{
    LogHeaders:      true,
    LogBody:         true,
    LogResponse:     true,
    MaxBodyLogSize:  1024,
    HeaderKeysToLog: []string{"Content-Type", "Authorization"},
})
```

## Default Client

For quick usage without creating a new client instance:

```go
defaultClient := xhttpc.GetDefaultClient()
resp, err := defaultClient.Get(context.Background(), "https://api.example.com/users")
```

## Additional Features

### Setting Base URL

```go
client.SetBaseURL("https://api.example.com")
// Now you can use relative paths in requests
resp, err := client.Get(context.Background(), "/users")
```

### Adding Headers

```go
client.SetHeader("X-API-Key", "your-api-key")
```

### Adding Cookies

```go
client.AddCookie(&http.Cookie{Name: "session", Value: "123456"})
```

### Setting Query Parameters

```go
client.SetQueryParam("page", "1")
client.SetQueryParam("limit", "10")
```

### Fluent Interface

```go
resp, err := client.
    SetBaseURL("https://api.example.com").
    SetHeader("X-API-Key", "your-api-key").
    SetQueryParam("limit", "10").
    Get(context.Background(), "/users")
```

### Automatic JSON Decoding

```go
var users []User
err := client.GetJSONAndDecode(context.Background(), "/users", &users)
```

### Custom Content Type

```go
resp, err := client.Post(context.Background(), "/data", data, xhttpc.WithContentType("application/xml"))
```

This README provides a comprehensive overview of the `xhttpc` package, including its main features, installation instructions, usage examples, and advanced configuration options. It covers the core functionality and demonstrates how to use the package effectively in various scenarios.
