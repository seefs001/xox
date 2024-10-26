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
	UserMsg   string // User-friendly error message
}

// Error returns the error message
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// String implements the fmt.Stringer interface
func (e *Error) String() string {
	return e.Error()
}

// ToJSON converts the error to a JSON string
// Usage: jsonStr, err := xerror.ToJSON(err)
func (e *Error) ToJSON() (string, error) {
	if e == nil {
		return "", nil
	}

	type jsonError struct {
		Message   string                 `json:"message"`
		UserMsg   string                 `json:"user_message"`
		Code      int                    `json:"code"`
		Timestamp string                 `json:"timestamp"`
		Stack     string                 `json:"stack"`
		Context   map[string]interface{} `json:"context,omitempty"`
	}

	jErr := jsonError{
		Code:      e.Code,
		Timestamp: e.Timestamp.Format(time.RFC3339),
		Stack:     e.Stack,
		UserMsg:   e.GetUserMsg(),
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
		Code:      1,
		Context:   make(map[string]interface{}),
		UserMsg:   msg, // Default to technical message
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
		Code:      1,
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
		Code:      1,
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
func Is(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}

	// Handle when target is xerror type
	if t, ok := target.(*Error); ok {
		if xe, ok := err.(*Error); ok {
			return xe.Is(t)
		}
		return errors.Is(err, t.Err)
	}

	// Handle when err is xerror type
	if xe, ok := err.(*Error); ok {
		return errors.Is(xe.Err, target)
	}

	// Fallback to standard errors.Is
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}

	// Handle xerror type
	if xe, ok := err.(*Error); ok {
		if t, ok := target.(**Error); ok {
			*t = xe
			return true
		}
		return errors.As(xe.Err, target)
	}

	// Try to convert standard error to xerror
	if t, ok := target.(**Error); ok {
		*t = &Error{
			Err:       err,
			Stack:     getStack(),
			Timestamp: time.Now(),
			Code:      int(GetErrorCode(err)),
			Context:   make(map[string]interface{}),
		}
		return true
	}

	// Fallback to standard errors.As
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
func Cause(err error) error {
	if err == nil {
		return nil
	}

	current := err
	for current != nil {
		switch e := current.(type) {
		case *Error:
			if e.Err == nil {
				return e
			}
			current = e.Err
		default:
			if u, ok := current.(interface{ Unwrap() error }); ok {
				current = u.Unwrap()
				if current == nil {
					return e
				}
			} else {
				return e
			}
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

// NewWithUserMsg creates a new Error with separate technical and user-friendly messages
func NewWithUserMsg(msg string, userMsg string) *Error {
	return &Error{
		Err:       errors.New(msg),
		Stack:     getStack(),
		Timestamp: time.Now(),
		Code:      1,
		Context:   make(map[string]interface{}),
		UserMsg:   userMsg,
	}
}

// SetUserMsg sets a user-friendly error message
func (e *Error) SetUserMsg(msg string) *Error {
	if e != nil {
		e.UserMsg = msg
	}
	return e
}

// GetUserMsg returns the user-friendly error message or the technical message if not set
func (e *Error) GetUserMsg() string {
	if e == nil {
		return ""
	}
	if e.UserMsg != "" {
		return e.UserMsg
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// Unwrap returns the underlying error
func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	// Handle xerror type
	if xe, ok := err.(*Error); ok {
		return xe.Err
	}

	// Handle standard error unwrapping
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return u.Unwrap()
	}

	// Return the error itself if it can't be unwrapped
	return err
}

// Unwrap implements the errors.Unwrap interface
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is implements the errors.Is interface
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}

	// Check if target is *Error
	if t, ok := target.(*Error); ok {
		return errors.Is(e.Err, t.Err)
	}

	// Compare with standard error
	return errors.Is(e.Err, target)
}

// As implements the errors.As interface
func (e *Error) As(target interface{}) bool {
	if e == nil {
		return false
	}

	// Try to convert to *Error first
	if t, ok := target.(**Error); ok {
		*t = e
		return true
	}

	// Fallback to standard error handling
	return errors.As(e.Err, target)
}

// Predefined errors with xerror type
var (
	ErrNotFound = &Error{
		Err:       errors.New("resource not found"),
		Code:      int(CodeNotFound),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}

	ErrUnauthorized = &Error{
		Err:       errors.New("unauthorized access"),
		Code:      int(CodeUnauthorized),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}

	ErrInvalidInput = &Error{
		Err:       errors.New("invalid input"),
		Code:      int(CodeInvalidInput),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}

	ErrTimeout = &Error{
		Err:       errors.New("operation timed out"),
		Code:      int(CodeTimeout),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}

	ErrDatabaseConnection = &Error{
		Err:       errors.New("database connection error"),
		Code:      int(CodeDatabaseConnection),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}

	ErrInternalServer = &Error{
		Err:       errors.New("internal server error"),
		Code:      int(CodeInternalServer),
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context:   make(map[string]interface{}),
	}
)
