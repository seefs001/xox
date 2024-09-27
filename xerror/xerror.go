package xerror

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/seefs001/xox/xlog"
)

// Error represents a custom error type that includes stack trace information
type Error struct {
	Err       error
	Stack     string
	Timestamp time.Time
	Code      int
	Context   map[string]interface{}
}

// Error returns the error message
func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s (Code: %d)", e.Timestamp.Format(time.RFC3339), e.Err.Error(), e.Code)
}

// New creates a new Error with the given error message and captures the stack trace
// Usage: err := xerror.New("something went wrong")
func New(msg string) *Error {
	return &Error{
		Err:       errors.New(msg),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      0,
		Context:   make(map[string]interface{}),
	}
}

// NewWithCode creates a new Error with the given error message, error code, and captures the stack trace
// Usage: err := xerror.NewWithCode("something went wrong", 500)
func NewWithCode(msg string, code int) *Error {
	return &Error{
		Err:       errors.New(msg),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      code,
		Context:   make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context and captures the stack trace
// Usage: err = xerror.Wrap(err, "failed to process request")
func Wrap(err error, msg string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:       fmt.Errorf("%s: %w", msg, err),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      0,
		Context:   make(map[string]interface{}),
	}
}

// WrapWithCode wraps an existing error with additional context, an error code, and captures the stack trace
// Usage: err = xerror.WrapWithCode(err, "failed to process request", 500)
func WrapWithCode(err error, msg string, code int) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:       fmt.Errorf("%s: %w", msg, err),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      code,
		Context:   make(map[string]interface{}),
	}
}

// Is reports whether any error in err's chain matches target
// Usage: if xerror.Is(err, ErrNotFound) { ... }
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
// Usage: var customErr *CustomError; if xerror.As(err, &customErr) { ... }
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// getStack returns the stack trace as a string
func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	var builder strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return builder.String()
}

// PanicIfError panics if the given error is not nil
// Usage: xerror.PanicIfError(err)
func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

// RecoverError recovers from a panic and returns it as an error
// Usage: defer xerror.RecoverError(&err)
func RecoverError(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%v", r)
	}
}

// ErrNotFound is a sentinel error for when a resource is not found
var ErrNotFound = errors.New("resource not found")

// ErrUnauthorized is a sentinel error for when a user is not authorized
var ErrUnauthorized = errors.New("unauthorized access")

// ErrInvalidInput is a sentinel error for when input is invalid
var ErrInvalidInput = errors.New("invalid input")

// ErrTimeout is a sentinel error for when an operation times out
var ErrTimeout = errors.New("operation timed out")

// ErrDatabaseConnection is a sentinel error for database connection issues
var ErrDatabaseConnection = errors.New("database connection error")

// ErrInternalServer is a sentinel error for internal server errors
var ErrInternalServer = errors.New("internal server error")

// ErrorCode represents error codes for different types of errors
type ErrorCode int

const (
	CodeUnknown ErrorCode = iota
	CodeNotFound
	CodeUnauthorized
	CodeInvalidInput
	CodeTimeout
	CodeDatabaseConnection
	CodeInternalServer
)

// GetErrorCode returns the appropriate error code based on the error type
func GetErrorCode(err error) ErrorCode {
	switch {
	case errors.Is(err, ErrNotFound):
		return CodeNotFound
	case errors.Is(err, ErrUnauthorized):
		return CodeUnauthorized
	case errors.Is(err, ErrInvalidInput):
		return CodeInvalidInput
	case errors.Is(err, ErrTimeout):
		return CodeTimeout
	case errors.Is(err, ErrDatabaseConnection):
		return CodeDatabaseConnection
	case errors.Is(err, ErrInternalServer):
		return CodeInternalServer
	default:
		return CodeUnknown
	}
}

// WithContext adds context to an existing error
// Usage: err = xerror.WithContext(err, "user_id", userId)
func WithContext(err error, key string, value interface{}) *Error {
	if xerr, ok := err.(*Error); ok {
		xerr.Context[key] = value
		return xerr
	}
	return &Error{
		Err:       err,
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      int(GetErrorCode(err)),
		Context:   map[string]interface{}{key: value},
	}
}

// IsTemporary checks if the error is temporary and can be retried
func IsTemporary(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

// Retry executes the given function and retries it if it returns a temporary error
// Usage: err := xerror.Retry(3, time.Second, func() error { return someFunction() })
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}
		if !IsTemporary(err) || i >= attempts-1 {
			return fmt.Errorf("after %d attempts, last error: %s", i+1, err)
		}
		time.Sleep(sleep)
		sleep *= 2
	}
}

// LogError logs the error with its stack trace using xlog
// Usage: xerror.LogError(err)
func LogError(err error) {
	if xerr, ok := err.(*Error); ok {
		xlog.Error("Error occurred",
			"message", xerr.Err.Error(),
			"code", xerr.Code,
			"timestamp", xerr.Timestamp,
			"stack", xerr.Stack,
			"context", xerr.Context,
		)
	} else {
		xlog.Error("Error occurred", "error", err.Error())
	}
}

// MustNoError panics if the given error is not nil, otherwise returns the value
// Usage: result := xerror.MustNoError(someFunction())
func MustNoError[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Cause returns the root cause of the error
// Usage: rootErr := xerror.Cause(err)
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(interface{ Cause() error })
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

// Join creates a new error joining multiple errors
// Usage: err := xerror.Join(err1, err2, err3)
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// WithStack adds a stack trace to an error if it doesn't already have one
// Usage: err = xerror.WithStack(err)
func WithStack(err error) error {
	if _, ok := err.(*Error); ok {
		return err
	}
	return &Error{
		Err:       err,
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      int(GetErrorCode(err)),
		Context:   make(map[string]interface{}),
	}
}

// FormatError returns a formatted string representation of the error
// Usage: formattedErr := xerror.FormatError(err)
func FormatError(err error) string {
	if xerr, ok := err.(*Error); ok {
		return fmt.Sprintf("Error: %s\nCode: %d\nTimestamp: %s\nStack Trace:\n%s\nContext: %v",
			xerr.Err.Error(), xerr.Code, xerr.Timestamp, xerr.Stack, xerr.Context)
	}
	return err.Error()
}

// NewErrorf creates a new Error with a formatted error message
// Usage: err := xerror.NewErrorf("failed to process item %d: %v", itemID, err)
func NewErrorf(format string, args ...interface{}) *Error {
	return New(fmt.Sprintf(format, args...))
}

// WrapWithContextf wraps an error with a formatted message and additional context
// Usage: err = xerror.WrapWithContextf(err, "user_id", userID, "failed to process user %d", userID)
func WrapWithContextf(err error, key string, value interface{}, format string, args ...interface{}) *Error {
	wrappedErr := Wrap(err, fmt.Sprintf(format, args...))
	return WithContext(wrappedErr, key, value)
}

// IsErrorCode checks if the error has a specific error code
// Usage: if xerror.IsErrorCode(err, xerror.CodeNotFound) { ... }
func IsErrorCode(err error, code ErrorCode) bool {
	if xerr, ok := err.(*Error); ok {
		return xerr.Code == int(code)
	}
	return false
}

// LogErrorAndReturn logs the error and returns it
// Usage: return xerror.LogErrorAndReturn(err)
func LogErrorAndReturn(err error) error {
	LogError(err)
	return err
}