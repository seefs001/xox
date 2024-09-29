# xerror

`xerror` is a comprehensive error handling package for Go, providing enhanced error creation, wrapping, and manipulation capabilities. It offers a rich set of features to improve error management in your Go applications.

## Features

- Error creation with stack traces
- Error wrapping with additional context
- Error code support
- Temporary error handling and retry mechanism
- Panic recovery
- Error cause extraction
- Error joining
- Context addition to errors

## Installation

```bash
go get github.com/seefs001/xox/xerror
```

## Usage

### Basic Error Creation

```go
import "github.com/seefs001/xox/xerror"

// Create a new error
err := xerror.New("something went wrong")

// Create a new error with a specific error code
err := xerror.NewWithCode("unauthorized access", 401)

// Create a new error with formatted message
err := xerror.Newf("failed to process item %d: %v", itemID, err)
```

### Error Wrapping

```go
// Wrap an existing error
err = xerror.Wrap(err, "failed to process request")

// Wrap with formatting
err = xerror.Wrapf(err, "failed to process item %d", itemID)

// Wrap with error code
err = xerror.WrapWithCode(err, "failed to process request", 500)
```

### Error Handling

```go
// Check if an error matches a specific error
if xerror.Is(err, xerror.ErrNotFound) {
    // Handle not found error
}

// Type assertion
var customErr *CustomError
if xerror.As(err, &customErr) {
    // Handle custom error
}

// Get error code
code := xerror.GetErrorCode(err)

// Check if error is of a specific type
if xerror.IsType[*CustomError](err) {
    // Handle custom error type
}
```

### Context and Fields

```go
// Add context to an error
err = xerror.WithContext(err, "user_id", userID)

// Add multiple fields
err = xerror.WithFields(err, map[string]interface{}{
    "user_id": 123,
    "action": "login",
})
```

### Temporary Errors and Retry

```go
// Retry a function with temporary errors
err := xerror.Retry(3, time.Second, func() error {
    return someFunction()
})
```

### Panic Handling

```go
// Panic if error is not nil
xerror.PanicIfError(err)

// Recover from panic
var err error
defer xerror.RecoverError(&err)
```

### Error Combination

```go
// Join multiple errors
err := xerror.Join(err1, err2, err3)

// Combine errors, ignoring nil errors
err := xerror.CombineErrors(err1, err2, err3)
```

### Stack Traces

```go
// Add stack trace to an error
err = xerror.WithStack(err)

// Create a new error with full stack trace
err := xerror.NewWithStackTrace("operation failed")
```

### Error Formatting

```go
// Format error with details
formattedErr := xerror.FormatError(err)
```

## Advanced Usage

### Custom Error Types

You can create custom error types that work seamlessly with `xerror`:

```go
type CustomError struct {
    xerror.Error
    CustomField string
}

func NewCustomError(msg string, customField string) *CustomError {
    return &CustomError{
        Error:       *xerror.New(msg),
        CustomField: customField,
    }
}
```

### Error Code Handling

```go
// Define custom error codes
const (
    ErrCodeNotFound xerror.ErrorCode = iota + 1000
    ErrCodeUnauthorized
    ErrCodeInvalidInput
)

// Check for specific error codes
if xerror.IsErrorCode(err, ErrCodeNotFound) {
    // Handle not found error
}
```

## Best Practices

1. Use `xerror.Wrap` or `xerror.Wrapf` to add context to errors as they propagate up the call stack.
2. Utilize error codes for consistent error handling across your application.
3. Add relevant context to errors using `xerror.WithContext` or `xerror.WithFields`.
4. Implement custom error types for domain-specific errors.
5. Use `xerror.Retry` for operations that may experience temporary failures.

By following these practices and utilizing the `xerror` package, you can significantly improve error handling and debugging in your Go applications.
