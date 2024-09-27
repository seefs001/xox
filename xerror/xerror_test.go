package xerror_test

import (
	"errors"
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
	err := xerror.New("test error")
	assert.True(t, xerror.Is(err, err))
	assert.False(t, xerror.Is(err, errors.New("different error")))
}

func TestAs(t *testing.T) {
	err := xerror.New("test error")
	var xerr *xerror.Error
	assert.True(t, xerror.As(err, &xerr))
	assert.NotNil(t, xerr)
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