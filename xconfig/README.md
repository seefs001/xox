# xconfig

xconfig is a flexible and powerful configuration management package for Go applications. It provides a simple interface to load, access, and manage configuration data from various sources, including JSON, environment variables, structs, and maps.

## Features

- Load configuration from JSON, environment variables, structs, and maps
- Support for nested configuration structures
- Type-safe access to configuration values
- Thread-safe operations
- Flatten nested maps into dot-notation keys
- Parse configuration to structs
- Default configuration instance for easy global usage
- Support for custom environment variable prefixes
- Ability to load configuration from files

## Installation

To install xconfig, use `go get`:

```bash
go get github.com/seefs001/xox/xconfig
```

## Usage

### Creating a Config Instance

```go
import "github.com/seefs001/xox/xconfig"

// Create a new Config instance
config := xconfig.NewConfig()

// Or use the default Config instance
xconfig.Put("key", "value")
```

### Loading Configuration

#### From JSON

```go
jsonStr := `{"database": {"host": "localhost", "port": 5432}}`
err := config.LoadFromJSON(jsonStr)
```

#### From Environment Variables

```go
// Load all environment variables
err := config.LoadFromEnv()

// Load environment variables with a specific prefix
err := config.LoadFromEnv(xconfig.WithEnvPrefix("APP_"))

// Load from an environment file
err := config.LoadFromEnv(xconfig.WithEnvFile("/path/to/env.json"))
```

#### From Struct

```go
type AppConfig struct {
    ServerPort int    `config:"server_port"`
    Debug      bool   `config:"debug_mode"`
}

cfg := AppConfig{ServerPort: 8080, Debug: true}
err := config.LoadFromStruct(cfg)
```

#### From Map

```go
data := map[string]any{
    "api_key": "secret",
    "max_connections": 100,
}
config.LoadFromMap(data)
```

### Accessing Configuration Values

```go
// Get string value
strValue, err := config.GetString("key")

// Get integer value
intValue, err := config.GetInt("key")

// Get int32 value
int32Value, err := config.GetInt32("key")

// Get boolean value
boolValue, err := config.GetBool("key")

// Get all configuration data
allData := config.GetAll()
```

### Parsing Configuration to Struct

```go
type DatabaseConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

var dbConfig DatabaseConfig
err := config.ParseToStruct(&dbConfig)
```

### Using the Default Config Instance

xconfig provides a default Config instance for convenience:

```go
xconfig.Put("app_name", "MyApp")
appName, err := xconfig.GetString("app_name")
```

### Working with Nested Structures

xconfig supports nested structures using dot notation:

```go
jsonStr := `{
    "database": {
        "master": {
            "host": "localhost",
            "port": 5432
        },
        "slave": [
            {"host": "slave1", "port": 5433},
            {"host": "slave2", "port": 5434}
        ]
    }
}`
config.LoadFromJSON(jsonStr)

masterHost, _ := config.GetString("database.master.host")
slavePort, _ := config.GetInt("database.slave.1.port")
```

### Flattening Nested Maps

xconfig can flatten nested maps into dot-notation keys:

```go
nestedMap := map[string]any{
    "app": map[string]any{
        "name": "MyApp",
        "version": "1.0.0",
    },
}
config.FlattenMap(nestedMap, "")

appName, _ := config.GetString("app.name")
```

## Advanced Usage

### Custom Environment Variable Prefix

```go
config := xconfig.NewConfig()
err := config.LoadFromEnv(xconfig.WithEnvPrefix("MYAPP_"))
```

### Loading from Multiple Sources

```go
config := xconfig.NewConfig()
config.LoadFromJSON(jsonStr)
config.LoadFromEnv()
config.LoadFromStruct(structConfig)
```

### Thread-safe Operations

All operations in xconfig are thread-safe, allowing for concurrent access from multiple goroutines.

### Error Handling

All getter methods return an error as the second return value. Always check for errors:

```go
value, err := config.GetString("key")
if err != nil {
    // Handle error
}
```

## Best Practices

1. Use a single configuration instance throughout your application.
2. Load configuration early in your application's lifecycle.
3. Use strongly typed getters (GetString, GetInt, etc.) to ensure type safety.
4. Utilize the dot notation for accessing nested configurations.
5. When working with environment variables, consider using a prefix to avoid conflicts.

## Example

Here's a comprehensive example showcasing various features of xconfig:

```go
package main

import (
    "fmt"
    "github.com/seefs001/xox/xconfig"
)

func main() {
    // Load configuration from JSON
    jsonConfig := `{
        "app": {
            "name": "MyApp",
            "version": "1.0.0"
        },
        "database": {
            "host": "localhost",
            "port": 5432
        }
    }`
    xconfig.LoadFromJSON(jsonConfig)

    // Load configuration from environment variables
    xconfig.LoadFromEnv(xconfig.WithEnvPrefix("APP_"))

    // Access configuration values
    appName, _ := xconfig.GetString("app.name")
    dbPort, _ := xconfig.GetInt("database.port")

    fmt.Printf("App Name: %s\n", appName)
    fmt.Printf("Database Port: %d\n", dbPort)

    // Parse configuration to struct
    type Config struct {
        App struct {
            Name    string `json:"name"`
            Version string `json:"version"`
        } `json:"app"`
        Database struct {
            Host string `json:"host"`
            Port int    `json:"port"`
        } `json:"database"`
    }

    var cfg Config
    err := xconfig.ParseToStruct(&cfg)
    if err != nil {
        fmt.Printf("Error parsing config: %v\n", err)
    } else {
        fmt.Printf("Parsed Config: %+v\n", cfg)
    }
}
```

This example demonstrates loading configuration from JSON and environment variables, accessing values, and parsing the configuration to a struct.

## Contributing

Contributions to xconfig are welcome! Please submit issues and pull requests on the GitHub repository.

## License

xconfig is released under the MIT License. See the LICENSE file for details.
