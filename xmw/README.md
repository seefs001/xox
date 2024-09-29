# XMW - Extensible Middleware for Web Applications

XMW is a powerful and flexible middleware package for Go web applications. It provides a set of common middleware functions that can be easily integrated into your web server to handle various cross-cutting concerns such as logging, recovery, timeout handling, CORS, compression, rate limiting, basic authentication, and session management.

## Features

- Modular and composable middleware architecture
- Easy integration with standard `http.Handler` interface
- Customizable options for each middleware
- Built-in support for common web application needs

## Installation

```bash
go get github.com/seefs001/xox/xmw
```

## Usage

Here's a basic example of how to use XMW middleware:

```go
package main

import (
    "net/http"
    "github.com/seefs001/xox/xmw"
)

func main() {
    // Create your main handler
    mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Apply middleware
    handler := xmw.Use(mainHandler,
        xmw.Logger(),
        xmw.Recover(),
        xmw.Timeout(),
        xmw.CORS(),
    )

    // Start the server
    http.ListenAndServe(":8080", handler)
}
```

## Available Middleware

### Logger

Logs HTTP requests and responses.

```go
loggerMiddleware := xmw.Logger(xmw.LoggerConfig{
    Format:     "[${time}] ${status} - ${latency} ${method} ${path}",
    TimeFormat: "2006-01-02 15:04:05",
    Output:     os.Stdout,
    UseColor:   true,
})
```

### Recover

Recovers from panics and sends a 500 Internal Server Error response.

```go
recoverMiddleware := xmw.Recover(xmw.RecoverConfig{
    EnableStackTrace: true,
})
```

### Timeout

Sets a timeout for request handling.

```go
timeoutMiddleware := xmw.Timeout(xmw.TimeoutConfig{
    Timeout: 30 * time.Second,
})
```

### CORS

Handles Cross-Origin Resource Sharing (CORS) headers.

```go
corsMiddleware := xmw.CORS(xmw.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
})
```

### Compress

Compresses the response using gzip compression.

```go
compressMiddleware := xmw.Compress(xmw.CompressConfig{
    Level:   gzip.DefaultCompression,
    MinSize: 1024,
})
```

### RateLimit

Limits the number of requests from a single client.

```go
rateLimitMiddleware := xmw.RateLimit(xmw.RateLimitConfig{
    Max:      100,
    Duration: time.Minute,
})
```

### BasicAuth

Implements HTTP Basic Authentication.

```go
basicAuthMiddleware := xmw.BasicAuth(xmw.BasicAuthConfig{
    Users: map[string]string{
        "user": "password",
    },
    Realm: "Restricted",
})
```

### Session

Manages user sessions.

```go
sessionMiddleware := xmw.Session(xmw.SessionConfig{
    Store:      xmw.NewMemoryStore(),
    CookieName: "session_id",
    MaxAge:     3600,
})
```

## Customization

Most middleware functions accept a configuration struct that allows you to customize their behavior. Refer to the individual middleware documentation for specific configuration options.

## Combining Middleware

You can use the `xmw.Use` function to combine multiple middleware:

```go
handler := xmw.Use(mainHandler,
    xmw.Logger(),
    xmw.Recover(),
    xmw.Timeout(xmw.TimeoutConfig{Timeout: 10 * time.Second}),
    xmw.CORS(),
)
```

## Performance Considerations

While XMW is designed to be efficient, be mindful of the number and order of middleware you apply, as each additional layer can impact performance. Profile your application to ensure optimal performance.

## Best Practices

1. Order your middleware carefully. For example, place the Logger middleware first to capture all requests, including those that might be terminated by subsequent middleware.
2. Use the Recover middleware to prevent panics from crashing your server.
3. Configure timeouts to prevent long-running requests from consuming resources.
4. Use CORS middleware when your API needs to be accessed from different domains.
5. Apply compression for larger responses to reduce bandwidth usage.
6. Implement rate limiting to protect your server from abuse.

## Support

For issues, feature requests, or questions, please open an issue in the GitHub repository.
