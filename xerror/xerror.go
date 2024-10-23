package xerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
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
	if e == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Error occurred at: %s\n", e.Timestamp.Format(time.RFC3339)))
	if e.Err != nil {
		sb.WriteString(fmt.Sprintf("Error message: %s\n", e.Err.Error()))
	}
	sb.WriteString(fmt.Sprintf("Error code: %d\n", e.Code))
	if e.Context != nil && len(e.Context) > 0 {
		sb.WriteString("Context:\n")
		for key, value := range e.Context {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}
	if e.Stack != "" {
		sb.WriteString("Stack trace:\n")
		sb.WriteString(e.Stack)
	}
	return sb.String()
}

// ToJSON converts the error to a JSON string
// Usage: jsonStr, err := xerror.ToJSON(err)
func (e *Error) ToJSON() (string, error) {
	if e == nil {
		return "", nil
	}

	type jsonError struct {
		Message   string                 `json:"message"`
		Code      int                    `json:"code"`
		Timestamp string                 `json:"timestamp"`
		Stack     string                 `json:"stack"`
		Context   map[string]interface{} `json:"context,omitempty"`
	}

	jErr := jsonError{
		Code:      e.Code,
		Timestamp: e.Timestamp.Format(time.RFC3339),
		Stack:     e.Stack,
	}

	if e.Err != nil {
		jErr.Message = e.Err.Error()
	}
	if e.Context != nil {
		jErr.Context = e.Context
	}

	jsonBytes, err := json.Marshal(jErr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal error: %w", err)
	}
	return string(jsonBytes), nil
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

// Wrapf wraps an existing error with additional context and captures the stack trace
// Usage: err = xerror.Wrapf(err, "failed to process request")
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:       fmt.Errorf(format+": %w", append(args, err)...),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      0,
		Context:   make(map[string]interface{}),
	}
}

// Newf creates a new Error with a formatted error message
// Usage: err := xerror.Newf("failed to process item %d: %v", itemID, err)
func Newf(format string, args ...interface{}) *Error {
	return New(fmt.Sprintf(format, args...))
}

// Errorf creates a new Error with a formatted error message
// Usage: err := xerror.Errorf("failed to process item %d: %v", itemID, err)
func Errorf(format string, args ...interface{}) *Error {
	return New(fmt.Sprintf(format, args...))
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
	if err == nil {
		return nil
	}
	if key == "" {
		return Wrap(err, "empty context key")
	}

	if xerr, ok := err.(*Error); ok {
		if xerr.Context == nil {
			xerr.Context = make(map[string]interface{})
		}
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
		xerr, ok := err.(*Error)
		if ok {
			err = xerr.Err
		} else {
			cause, ok := err.(interface{ Unwrap() error })
			if !ok {
				break
			}
			err = cause.Unwrap()
		}
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

// MustWrap wraps an error if it's not nil, otherwise returns nil
// Usage: err = xerror.MustWrap(someFunction())
func MustWrap(err error, msg string) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err, msg)
}

// MustWrapf wraps an error with a formatted message if it's not nil, otherwise returns nil
// Usage: err = xerror.MustWrapf(someFunction(), "failed to process %s", item)
func MustWrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return Wrapf(err, format, args...)
}

// WrapWithStackTrace wraps an error and includes a full stack trace
// Usage: err = xerror.WrapWithStackTrace(err, "operation failed")
func WrapWithStackTrace(err error, msg string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:       fmt.Errorf("%s: %w", msg, err),
		Stack:     string(debug.Stack()),
		Timestamp: time.Now(),
		Code:      0,
		Context:   make(map[string]interface{}),
	}
}

// NewWithStackTrace creates a new error with a full stack trace
// Usage: err := xerror.NewWithStackTrace("operation failed")
func NewWithStackTrace(msg string) *Error {
	return &Error{
		Err:       errors.New(msg),
		Stack:     string(debug.Stack()),
		Timestamp: time.Now(),
		Code:      0,
		Context:   make(map[string]interface{}),
	}
}

// WrapIfNotNil wraps an error only if it's not nil, otherwise returns nil
// Usage: err = xerror.WrapIfNotNil(err, "operation failed")
func WrapIfNotNil(err error, msg string) error {
	if err == nil {
		return nil
	}
	return Wrap(err, msg)
}

// IsType checks if an error is of a specific type
// Usage: if xerror.IsType[*CustomError](err) { ... }
func IsType[T error](err error) bool {
	var target T
	return As(err, &target)
}

// MapError applies a function to an error if it's not nil
// Usage: err = xerror.MapError(err, func(e error) error { return fmt.Errorf("wrapped: %w", e) })
func MapError(err error, f func(error) error) error {
	if err == nil {
		return nil
	}
	return f(err)
}

// CombineErrors combines multiple errors into a single error
// Usage: err := xerror.CombineErrors(err1, err2, err3)
func CombineErrors(errs ...error) error {
	var nonNilErrs []error
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return nil
	}
	return Join(nonNilErrs...)
}

// WithFields adds multiple context fields to an error
// Usage: err = xerror.WithFields(err, map[string]interface{}{"user_id": 123, "action": "login"})
func WithFields(err error, fields map[string]interface{}) *Error {
	if err == nil {
		return nil
	}
	if fields == nil {
		return Wrap(err, "nil fields map")
	}

	xerr, ok := err.(*Error)
	if !ok {
		xerr = &Error{
			Err:       err,
			Stack:     getStack(),
			Timestamp: time.Now(),
			Code:      int(GetErrorCode(err)),
			Context:   make(map[string]interface{}),
		}
	}

	if xerr.Context == nil {
		xerr.Context = make(map[string]interface{})
	}

	for k, v := range fields {
		if k != "" {
			xerr.Context[k] = v
		}
	}
	return xerr
}

func GetContext(err error) map[string]interface{} {
	if xerr, ok := err.(*Error); ok && xerr != nil {
		return xerr.Context
	}
	return nil
}

func GetCode(err error) int {
	if xerr, ok := err.(*Error); ok && xerr != nil {
		return xerr.Code
	}
	return int(CodeUnknown)
}

func GetStack(err error) string {
	if xerr, ok := err.(*Error); ok && xerr != nil {
		return xerr.Stack
	}
	return ""
}
