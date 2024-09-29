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
- `GetParamDuration(key string) (time.Duration, error)`: Get a URL query parameter as a time.Duration
- `GetParamTime(key, layout string) (time.Time, error)`: Get a URL query parameter as a time.Time
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

## Example

Here's a more comprehensive example demonstrating various features of xhttp:

```go
package main

import (
    "github.com/seefs001/xox/xhttp"
    "log"
    "net/http"
    "time"
)

type User struct {
    Name  string `json:"name" form:"name"`
    Email string `json:"email" form:"email"`
}

func main() {
    router := xhttp.NewRouter()

    router.HandleFunc("/users", xhttp.Wrap(func(c *xhttp.Context) {
        var user User
        if err := c.Bind(&user); err != nil {
            c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }

        // Process the user...

        c.JSON(http.StatusOK, user)
    }))

    router.HandleFunc("/time", xhttp.Wrap(func(c *xhttp.Context) {
        layout := c.GetParam("layout")
        if layout == "" {
            layout = time.RFC3339
        }

        c.String(http.StatusOK, time.Now().Format(layout))
    }))

    router.HandleFunc("/download", xhttp.Wrap(func(c *xhttp.Context) {
        err := c.StreamFile("path/to/file.pdf")
        if err != nil {
            c.String(http.StatusInternalServerError, "Error streaming file: %v", err)
        }
    }))

    log.Fatal(xhttp.ListenAndServe(":8080", router))
}
```

This example demonstrates request binding, parameter parsing, JSON responses, string responses, and file streaming.
