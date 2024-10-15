# xenv

`xenv` is a Go package for managing environment variables with extended functionality. It provides an easy way to load environment variables from `.env` files and access them with type-safe getters.

## Features

- Load environment variables from `.env` files
- Recursive search for `.env` files in parent directories
- Type-safe getters for various data types
- Default values for missing environment variables
- JSON parsing for complex data structures
- Set and unset environment variables

## Installation

```bash
go get github.com/seefs001/xox/xenv
```

## Usage

### Loading Environment Variables

```go
import "github.com/seefs001/xox/xenv"

func main() {
    err := xenv.Load()
    if err != nil {
        // Handle error
    }

    // Or with custom options
    err = xenv.Load(xenv.LoadOptions{Filename: "custom.env"})
    if err != nil {
        // Handle error
    }
}
```

### Accessing Environment Variables

```go
// Get a string value
value := xenv.Get("KEY")

// Get a string value with a default
value := xenv.GetDefault("KEY", "default_value")

// Get a boolean value
boolValue := xenv.GetBool("BOOL_KEY")

// Get an integer value
intValue, err := xenv.GetInt("INT_KEY")

// Get a float64 value
floatValue, err := xenv.GetFloat64("FLOAT_KEY")

// Get a slice of strings
slice := xenv.GetSlice("SLICE_KEY", ",")

// Get a map of strings
mapValue := xenv.GetMap("MAP_KEY", ",", ":")

// Get a duration value
duration, err := xenv.GetDuration("DURATION_KEY")

// Get and parse JSON
var jsonData struct {
    Field string `json:"field"`
}
err := xenv.GetJSON("JSON_KEY", &jsonData)
```

### Setting and Unsetting Environment Variables

```go
// Set an environment variable
err := xenv.Set("KEY", "value")

// Unset an environment variable
err := xenv.Unset("KEY")
```

### Advanced Usage

```go
// Must get (panics if not set)
value := xenv.MustGet("REQUIRED_KEY")

// Get int64 value
int64Value, err := xenv.GetInt64("INT64_KEY")

// Get uint value
uintValue, err := xenv.GetUint("UINT_KEY")

// Get uint64 value
uint64Value, err := xenv.GetUint64("UINT64_KEY")
```

## API Reference

### Load Functions

- `Load(options ...LoadOptions) error`
  Loads environment variables from a `.env` file.

  Options:
  ```go
  type LoadOptions struct {
      Filename string // Custom filename (default: ".env")
  }
  ```

### Getter Functions

- `Get(key string) string`
- `GetDefault(key, defaultValue string) string`
- `GetBool(key string) bool`
- `GetInt(key string) (int, error)`
- `GetFloat64(key string) (float64, error)`
- `GetSlice(key, sep string) []string`
- `GetMap(key, pairSep, kvSep string) map[string]string`
- `GetDuration(key string) (time.Duration, error)`
- `GetJSON(key string, v interface{}) error`
- `MustGet(key string) string`
- `GetInt64(key string) (int64, error)`
- `GetUint(key string) (uint, error)`
- `GetUint64(key string) (uint64, error)`

### Setter Functions

- `Set(key, value string) error`
- `Unset(key string) error`

## Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/seefs001/xox/xenv"
)

func main() {
    err := xenv.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Access environment variables
    apiKey := xenv.MustGet("API_KEY")
    debug := xenv.GetBool("DEBUG")
    port, err := xenv.GetInt("PORT")
    if err != nil {
        port = 8080 // Default port
    }

    fmt.Printf("API Key: %s\n", apiKey)
    fmt.Printf("Debug Mode: %v\n", debug)
    fmt.Printf("Port: %d\n", port)

    // Parse JSON configuration
    var config struct {
        DatabaseURL string `json:"db_url"`
        MaxConnections int `json:"max_connections"`
    }
    err = xenv.GetJSON("APP_CONFIG", &config)
    if err != nil {
        log.Printf("Failed to parse APP_CONFIG: %v", err)
    } else {
        fmt.Printf("Database URL: %s\n", config.DatabaseURL)
        fmt.Printf("Max Connections: %d\n", config.MaxConnections)
    }

    // Set and unset environment variables
    err = xenv.Set("TEMP_VAR", "temporary value")
    if err != nil {
        log.Printf("Failed to set TEMP_VAR: %v", err)
    }

    tempValue := xenv.Get("TEMP_VAR")
    fmt.Printf("TEMP_VAR: %s\n", tempValue)

    err = xenv.Unset("TEMP_VAR")
    if err != nil {
        log.Printf("Failed to unset TEMP_VAR: %v", err)
    }
}
```

This package is part of the `github.com/seefs001/xox` module.
