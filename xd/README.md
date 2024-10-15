# XD - Extensible Dependency Injection Container for Go

XD is a powerful and flexible dependency injection container for Go applications. It provides a simple and intuitive API for managing dependencies, supporting named services, struct injection, and more.

## Features

- Simple service registration and retrieval
- Support for named services
- Struct injection with tag-based configuration
- Thread-safe operations
- Customizable logging
- Easy cloning and clearing of containers
- Recursive struct injection
- Type-safe service retrieval using generics

## Installation

To install XD, use `go get`:

```bash
go get github.com/seefs001/xox/xd
```

## Usage

### Creating a Container

```go
import "github.com/seefs001/xox/xd"

// Create a new container
container := xd.NewContainer()

// Or use the default container
defaultContainer := xd.GetDefaultContainer()
```

### Registering Services

```go
// Simple service registration
xd.Provide(container, func(c *xd.Container) (*MyService, error) {
    return &MyService{}, nil
})

// Named service registration
xd.ProvideNamed(container, "custom", func(c *xd.Container) (*MyService, error) {
    return &MyService{Name: "Custom"}, nil
})
```

### Retrieving Services

```go
// Retrieve a service
service, err := xd.Invoke[*MyService](container)
if err != nil {
    // Handle error
}

// Retrieve a named service
namedService, err := xd.InvokeNamed[*MyService](container, "custom")
if err != nil {
    // Handle error
}

// Retrieve a service with panic on error
service := xd.MustInvoke[*MyService](container)
```

### Struct Injection

```go
type MyStruct struct {
    Service *MyService `xd:"-"`
}

myStruct := &MyStruct{}
err := xd.InjectStruct(container, myStruct)
if err != nil {
    // Handle error
}
```

### Named Struct Injection

```go
type MyStruct struct {
    CustomService *MyService `xd:"-"`
}

myStruct := &MyStruct{}
err := xd.InjectStructNamed(container, myStruct, map[string]string{
    "CustomService": "custom",
})
if err != nil {
    // Handle error
}
```

### Container Management

```go
// List all services
services := xd.ListServices(container)

// Remove a service
container.RemoveService(reflect.TypeOf((*MyService)(nil)).Elem())

// Remove a named service
container.RemoveNamedService(reflect.TypeOf((*MyService)(nil)).Elem(), "custom")

// Check if a service exists
exists := container.HasService(reflect.TypeOf((*MyService)(nil)).Elem())

// Clear all services
container.Clear()

// Clone a container
newContainer := container.Clone()
```

### Direct Service Management

```go
// Set a service directly
container.SetService(&MyService{})

// Set a named service directly
container.SetNamedService("custom", &MyService{Name: "Custom"})
```

### Logging

```go
// Set a custom logger
xd.SetLogger(func(format string, args ...any) {
    log.Printf(format, args...)
})
```

## Advanced Features

### Recursive Struct Injection

XD supports recursive struct injection, allowing you to inject dependencies into nested structs:

```go
type NestedStruct struct {
    Database *Database `xd:"-"`
}

type ParentStruct struct {
    Nested *NestedStruct
    Logger *Logger `xd:"-"`
}

parent := &ParentStruct{Nested: &NestedStruct{}}
err := xd.InjectStruct(container, parent)
// Both parent.Logger and parent.Nested.Database will be injected
```

### Circular Dependency Detection

XD automatically detects circular dependencies during service creation and injection, preventing infinite loops:

```go
xd.Provide(container, func(c *xd.Container) (*ServiceA, error) {
    b := xd.MustInvoke[*ServiceB](c)
    return &ServiceA{B: b}, nil
})

xd.Provide(container, func(c *xd.Container) (*ServiceB, error) {
    a := xd.MustInvoke[*ServiceA](c)
    return &ServiceB{A: a}, nil
})

// Invoking either ServiceA or ServiceB will result in an error
```

### Thread-Safe Operations

All operations in XD are thread-safe, allowing you to use the container in concurrent environments without additional synchronization.

## Best Practices

1. Use `Provide` and `Invoke` for most scenarios.
2. Utilize named services when you need multiple instances of the same type.
3. Implement proper error handling, especially when using `Invoke` and `InjectStruct`.
4. Use `MustInvoke` only when you're certain the service exists to avoid panics.
5. Clear the container or use a new one for each test to ensure isolation.
6. Use the `xd:"-"` tag to mark fields for injection.
7. Consider using interfaces for better decoupling and easier testing.
8. Avoid circular dependencies by redesigning your architecture if possible.

## Example

Here's a complete example demonstrating the usage of XD:

```go
package main

import (
    "fmt"
    "github.com/seefs001/xox/xd"
)

type Database struct {
    ConnectionString string
}

type Logger struct {
    Level string
}

type Service struct {
    DB  *Database `xd:"-"`
    Log *Logger   `xd:"-"`
}

func main() {
    container := xd.NewContainer()

    xd.Provide(container, func(c *xd.Container) (*Database, error) {
        return &Database{ConnectionString: "postgres://localhost:5432/mydb"}, nil
    })

    xd.Provide(container, func(c *xd.Container) (*Logger, error) {
        return &Logger{Level: "info"}, nil
    })

    xd.Provide(container, func(c *xd.Container) (*Service, error) {
        service := &Service{}
        err := xd.InjectStruct(c, service)
        return service, err
    })

    service := xd.MustInvoke[*Service](container)
    fmt.Printf("DB: %s, Log Level: %s\n", service.DB.ConnectionString, service.Log.Level)
}
```

This example demonstrates how to set up a container, register services, and use struct injection to create a fully configured `Service` instance.

## Contributing

Contributions to XD are welcome! Please feel free to submit issues, fork the repository and send pull requests!

## License

XD is released under the MIT License. See the LICENSE file for details.
