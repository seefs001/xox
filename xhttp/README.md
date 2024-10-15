# xhttp

xhttp is a powerful and flexible HTTP library for Go that provides a high-level abstraction over the standard net/http package. It offers a convenient Context-based API, middleware support, and various utility functions for handling HTTP requests and responses.

## Features

- Context-based request handling
- Easy-to-use router with grouping support
- JSON and XML response helpers
- File serving and streaming
- Request binding and parameter parsing
- Middleware support
- IP address handling
- Basic authentication helper
- Cookie management
- Redirection support
- Custom response writer

## Installation

```bash
go get github.com/seefs001/xox/xhttp
```

## Quick Start

```go
package main

import (
    "github.com/seefs001/xox/xhttp"
    "log"
    "net/http"
)

func main() {
    router := xhttp.NewRouter()

    router.HandleFunc("/", xhttp.Wrap(func(c *xhttp.Context) {
        c.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
    }))

    log.Fatal(xhttp.ListenAndServe(":8080", router))
}
```

## API Reference

### Context

`Context` is the core type in xhttp, providing access to the request and response writer, as well as utility methods for handling HTTP operations.

#### Request Parsing

- `GetParam(key string) string`: Get a URL query parameter
- `GetParamInt(key string) (int, error)`: Get a URL query parameter as an integer
- `GetParamInt64(key string) (int64, error)`: Get a URL query parameter as an int64
- `GetParamFloat(key string) (float64, error)`: Get a URL query parameter as a float64
- `GetParamBool(key string) bool`: Get a URL query parameter as a boolean
- `GetParamUint(key string) (uint, error)`: Get a URL query parameter as an unsigned integer
- `GetParamUint64(key string) (uint64, error)`: Get a URL query parameter as a uint64
- `GetParamDuration(key string) (time.Duration, error)`: Get a URL parameter as a time.Duration
- `GetParamTime(key, layout string) (time.Time, error)`: Get a URL parameter as a time.Time
- `GetHeader(key string) string`: Get a request header
- `GetBody(v interface{}) error`: Decode the request body into a struct
- `GetBodyRaw() ([]byte, error)`: Get the raw request body
- `GetBodyXML(v interface{}) error`: Decode the request body as XML
- `GetFormValue(key string) string`: Get a form value
- `GetFormFile(key string) (multipart.File, *multipart.FileHeader, error)`: Get a file from a multipart form
- `ParseMultipartForm(maxMemory int64) error`: Parse a multipart form
- `GetQueryParams() map[string][]string`: Get all query parameters
- `GetBodyForm() (map[string][]string, error)`: Parse the request body as a form and return the values
- `BasicAuth() (username, password string, ok bool)`: Get the username and password from Basic Auth
- `IsAjax() bool`: Check if the request is an AJAX request
- `ClientIP() string`: Get the client's IP address

#### Response Writing

- `JSON(code int, v interface{}) error`: Send a JSON response
- `XML(code int, v interface{}) error`: Send an XML response
- `String(code int, format string, values ...interface{}) error`: Send a string response
- `SetHeader(key, value string)`: Set a response header
- `SetStatus(code int)`: Set the HTTP status code
- `NoContent(code int)`: Send a response with no content
- `File(filepath string)`: Send a file as the response
- `StreamFile(filepath string) error`: Stream a file as the response
- `Redirect(code int, url string)`: Send a redirect response
- `SetCookie(cookie *http.Cookie)`: Set a cookie
- `GetCookie(name string) (*http.Cookie, error)`: Get a cookie

#### Data Binding

- `Bind(v interface{}) error`: Bind data from multiple sources (query, form, JSON body) to a struct
- `MustBind(v interface{})`: Bind data to a struct and panic if there's an error

### Router

`Router` is a custom router that wraps http.ServeMux and provides additional functionality.

```go
router := xhttp.NewRouter()

// Handle a route
router.HandleFunc("/api/users", handleUsers)

// Create a route group
api := router.Group("/api")
api.HandleFunc("/products", handleProducts)

// Print the routing tree
router.PrintRoutes()
```

### Middleware

xhttp supports middleware through the `NewRouterWithMiddleware` function:

```go
loggingMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

router := xhttp.NewRouterWithMiddleware(loggingMiddleware)
```

### Server

- `ListenAndServe(addr string, handler http.Handler) error`: Start an HTTP server
- `ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error`: Start an HTTPS server

## Comprehensive Example

Here's a more comprehensive example demonstrating various features of xhttp:

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/seefs001/xox/xhttp"
    "github.com/seefs001/xox/xlog"
    "github.com/seefs001/xox/xmw"
)

func UserHandler(c *xhttp.Context) {
    userID := c.GetParam("id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
        return
    }

    user := map[string]string{
        "id":   userID,
        "name": fmt.Sprintf("User %s", userID),
    }

    c.JSON(http.StatusOK, user)
}

func HeaderEchoHandler(c *xhttp.Context) {
    headers := make(map[string]string)
    for key, values := range c.Request.Header {
        headers[key] = values[0]
    }

    c.JSON(http.StatusOK, headers)
}

func ParamHandler(c *xhttp.Context) {
    id, err := c.GetParamInt("id")
    if err != nil {
        c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
        return
    }

    score, err := c.GetParamFloat("score")
    if err != nil {
        c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid score"})
        return
    }

    c.JSON(http.StatusOK, map[string]interface{}{
        "id":    id,
        "score": score,
    })
}

func BodyHandler(c *xhttp.Context) {
    var data map[string]interface{}
    if err := c.GetBody(&data); err != nil {
        c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
        return
    }

    c.JSON(http.StatusOK, data)
}

func CookieHandler(c *xhttp.Context) {
    cookie := &http.Cookie{
        Name:    "example",
        Value:   "test",
        Expires: time.Now().Add(24 * time.Hour),
    }
    c.SetCookie(cookie)

    retrievedCookie, err := c.GetCookie("example")
    if err != nil {
        c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cookie"})
        return
    }

    c.JSON(http.StatusOK, map[string]string{"cookie_value": retrievedCookie.Value})
}

func main() {
    xlog.Info("Starting xhttp example server")

    mux := http.NewServeMux()

    mux.HandleFunc("/user", xhttp.Wrap(UserHandler))
    mux.HandleFunc("/echo-headers", xhttp.Wrap(HeaderEchoHandler))
    mux.HandleFunc("/params", xhttp.Wrap(ParamHandler))
    mux.HandleFunc("/body", xhttp.Wrap(BodyHandler))
    mux.HandleFunc("/cookie", xhttp.Wrap(CookieHandler))

    middlewareStack := []xmw.Middleware{
        xmw.Logger(xmw.LoggerConfig{
            Output:   log.Writer(),
            UseColor: true,
        }),
        xmw.Recover(xmw.RecoverConfig{
            EnableStackTrace: true,
        }),
        xmw.Timeout(xmw.TimeoutConfig{
            Timeout: 5 * time.Second,
        }),
        xmw.CORS(xmw.CORSConfig{
            AllowOrigins: []string{"*"},
            AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        }),
        xmw.Compress(xmw.CompressConfig{
            Level: 5,
        }),
    }

    handler := xmw.Use(mux, middlewareStack...)

    xlog.Info("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

This example demonstrates how to use various features of xhttp, including parameter parsing, JSON responses, cookie handling, and integration with middleware from the xmw package.

## Contributing

Contributions to xhttp are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.
