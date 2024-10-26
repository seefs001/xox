package xerror_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/seefs001/xox/xerror"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := xerror.New("test error")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test error")
	assert.NotEmpty(t, err.Stack)
}

func TestNewWithCode(t *testing.T) {
	err := xerror.NewWithCode("test error", 500)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test error")
	assert.Equal(t, 500, err.Code)
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := xerror.Wrap(originalErr, "wrapped error")
	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "wrapped error: original error")
}

func TestWrapWithCode(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := xerror.WrapWithCode(originalErr, "wrapped error", 400)
	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "wrapped error: original error")
	assert.Equal(t, 400, wrappedErr.Code)
}

func TestIs(t *testing.T) {
	// Test standard error comparison
	stdErr := errors.New("standard error")
	xerr := xerror.Wrap(stdErr, "wrapped")
	assert.True(t, xerror.Is(xerr, stdErr))

	// Test xerror comparison
	targetXerr := xerror.New("target error")
	wrappedXerr := xerror.Wrap(targetXerr, "wrapped")
	assert.True(t, xerror.Is(wrappedXerr, targetXerr))

	// Test nil cases
	assert.True(t, xerror.Is(nil, nil))
	assert.False(t, xerror.Is(xerr, nil))
	assert.False(t, xerror.Is(nil, xerr))
}

func TestAs(t *testing.T) {
	// Test conversion from xerror to xerror
	originalErr := xerror.New("test error")
	var xerr *xerror.Error
	assert.True(t, xerror.As(originalErr, &xerr))
	assert.Equal(t, originalErr, xerr)

	// Test conversion from standard error to xerror
	stdErr := errors.New("standard error")
	var convertedXerr *xerror.Error
	assert.True(t, xerror.As(stdErr, &convertedXerr))
	assert.Equal(t, stdErr.Error(), convertedXerr.Err.Error())

	// Test conversion to different error type
	var stdErrTarget error
	assert.True(t, xerror.As(originalErr, &stdErrTarget))

	// Test nil case
	assert.False(t, xerror.As(nil, &xerr))
}

func TestUnwrap(t *testing.T) {
	// Test unwrapping xerror
	innerErr := errors.New("inner error")
	xerr := xerror.Wrap(innerErr, "wrapped")
	unwrapped := xerror.Unwrap(xerr)
	assert.Equal(t, "wrapped: inner error", unwrapped.Error())

	// Test unwrapping standard error
	stdErr := fmt.Errorf("outer: %w", errors.New("inner"))
	unwrappedStd := xerror.Unwrap(stdErr)
	assert.Equal(t, "inner", unwrappedStd.Error())

	// Test unwrapping nil
	assert.Nil(t, xerror.Unwrap(nil))

	// Test unwrapping non-wrapped error
	plainErr := errors.New("plain error")
	assert.Equal(t, plainErr, xerror.Unwrap(plainErr))

	// Test multiple levels of wrapping
	multiWrap := xerror.Wrap(xerror.Wrap(innerErr, "level 1"), "level 2")
	unwrappedMulti := xerror.Unwrap(multiWrap)
	assert.Equal(t, "level 2: level 1: inner error", unwrappedMulti.Error())
}

func TestWithContext(t *testing.T) {
	err := xerror.New("test error")
	contextErr := xerror.WithContext(err, "key", "value")
	assert.Equal(t, "value", contextErr.Context["key"])
}

func TestIsTemporary(t *testing.T) {
	tempErr := &temporaryError{}
	assert.True(t, xerror.IsTemporary(tempErr))
	assert.False(t, xerror.IsTemporary(errors.New("non-temporary error")))
}

func TestRetry(t *testing.T) {
	count := 0
	err := xerror.Retry(3, time.Millisecond, func() error {
		count++
		if count < 3 {
			return &temporaryError{}
		}
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
}

func TestGetErrorCode(t *testing.T) {
	assert.Equal(t, xerror.CodeNotFound, xerror.GetErrorCode(xerror.ErrNotFound))
	assert.Equal(t, xerror.CodeUnauthorized, xerror.GetErrorCode(xerror.ErrUnauthorized))
	assert.Equal(t, xerror.CodeInvalidInput, xerror.GetErrorCode(xerror.ErrInvalidInput))
	assert.Equal(t, xerror.CodeUnknown, xerror.GetErrorCode(errors.New("unknown error")))
}

func TestFormatError(t *testing.T) {
	err := xerror.NewWithCode("test error", 500)
	formatted := xerror.FormatError(err)
	assert.Contains(t, formatted, "test error")
	assert.Contains(t, formatted, "Code: 500")
	assert.Contains(t, formatted, "Stack Trace:")
}

func TestNewErrorf(t *testing.T) {
	err := xerror.NewErrorf("error %d: %s", 1, "test")
	assert.Contains(t, err.Error(), "error 1: test")
}

func TestWrapWithContextf(t *testing.T) {
	originalErr := errors.New("original error")
	err := xerror.WrapWithContextf(originalErr, "key", "value", "wrapped error %d", 1)
	assert.Contains(t, err.Error(), "wrapped error 1: original error")
	assert.Equal(t, "value", err.Context["key"])
}

func TestIsErrorCode(t *testing.T) {
	err := xerror.NewWithCode("test error", int(xerror.CodeNotFound))
	assert.True(t, xerror.IsErrorCode(err, xerror.CodeNotFound))
	assert.False(t, xerror.IsErrorCode(err, xerror.CodeUnauthorized))
}

// Helper struct for testing IsTemporary
type temporaryError struct{}

func (e *temporaryError) Error() string   { return "temporary error" }
func (e *temporaryError) Temporary() bool { return true }

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := xerror.Wrapf(originalErr, "wrapped error %d", 1)
	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Err.Error(), "wrapped error 1: original error")
}

func TestNewf(t *testing.T) {
	err := xerror.Newf("error %d: %s", 1, "test")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error 1: test")
}

func TestErrorf(t *testing.T) {
	err := xerror.Errorf("error %d: %s", 1, "test")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error 1: test")
}

func TestPanicIfError(t *testing.T) {
	assert.NotPanics(t, func() {
		xerror.PanicIfError(nil)
	})
	assert.Panics(t, func() {
		xerror.PanicIfError(errors.New("test error"))
	})
}

func TestRecoverError(t *testing.T) {
	var err error
	func() {
		defer xerror.RecoverError(&err)
		panic("test panic")
	}()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test panic")
}

func TestMustNoError(t *testing.T) {
	result := xerror.MustNoError(42, nil)
	assert.Equal(t, 42, result)
	assert.Panics(t, func() {
		xerror.MustNoError(0, errors.New("test error"))
	})
}

func TestCause(t *testing.T) {
	// Test with multiple wrapping levels
	root := errors.New("root error")
	level1 := xerror.Wrap(root, "level 1")
	level2 := xerror.Wrap(level1, "level 2")
	cause := xerror.Cause(level2)
	assert.Equal(t, root, cause)

	// Test with standard error wrapping
	stdRoot := errors.New("std root")
	stdWrapped := fmt.Errorf("wrapped: %w", stdRoot)
	stdCause := xerror.Cause(stdWrapped)
	assert.Equal(t, stdRoot, stdCause)

	// Test with xerror at root
	xerrRoot := xerror.New("xerror root")
	xerrWrapped := xerror.Wrap(xerrRoot, "wrapped")
	xerrCause := xerror.Cause(xerrWrapped)
	assert.Equal(t, xerrRoot.Err, xerrCause)

	// Test nil case
	assert.Nil(t, xerror.Cause(nil))
}

func TestJoin(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	joinedErr := xerror.Join(err1, err2)
	assert.NotNil(t, joinedErr)
	assert.Contains(t, joinedErr.Error(), "error 1")
	assert.Contains(t, joinedErr.Error(), "error 2")
}

func TestWithStack(t *testing.T) {
	err := errors.New("test error")
	stackErr := xerror.WithStack(err)
	assert.NotNil(t, stackErr)
	xerr, ok := stackErr.(*xerror.Error)
	assert.True(t, ok)
	assert.NotEmpty(t, xerr.Stack)
}

func TestErrorChainInteraction(t *testing.T) {
	// Create a mixed error chain
	rootErr := errors.New("root error")
	xerr := xerror.Wrap(rootErr, "xerror wrap")
	stdWrap := fmt.Errorf("std wrap: %w", xerr)
	finalXerr := xerror.Wrap(stdWrap, "final wrap")

	// Test Is through the chain
	assert.True(t, xerror.Is(finalXerr, rootErr))

	// Test As through the chain
	var targetXerr *xerror.Error
	assert.True(t, xerror.As(finalXerr, &targetXerr))

	// Test Cause through the chain
	cause := xerror.Cause(finalXerr)
	assert.Equal(t, rootErr, cause)
}
