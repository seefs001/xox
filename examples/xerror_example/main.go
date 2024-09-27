package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xlog"
)

func main() {
	// Set up logging configuration
	xlog.SetLogConfig(xlog.LogConfig{IncludeFileAndLine: true, Level: slog.LevelDebug})

	// Basic error creation
	err := xerror.New("This is a basic error")
	xlog.Error("Basic error example", "error", err)

	// Error with code
	errWithCode := xerror.NewWithCode("This is an error with code", 500)
	xlog.Error("Error with code example", "error", errWithCode)

	// Wrapping an error
	originalErr := errors.New("original error")
	wrappedErr := xerror.Wrap(originalErr, "wrapped error")
	xlog.Error("Wrapped error example", "error", wrappedErr)

	// Error with context
	errWithContext := xerror.WithContext(err, "user_id", 12345)
	xlog.Error("Error with context example", "error", errWithContext)

	// Formatted error
	formattedErr := xerror.NewErrorf("Formatted error: %d", 42)
	xlog.Error("Formatted error example", "error", formattedErr)

	// Using Is and As
	if xerror.Is(wrappedErr, originalErr) {
		xlog.Info("xerror.Is example: wrappedErr is originalErr")
	}

	var xerr *xerror.Error
	if xerror.As(wrappedErr, &xerr) {
		xlog.Info("xerror.As example: successfully cast to *xerror.Error")
	}

	// Demonstrating error codes
	notFoundErr := xerror.WrapWithCode(xerror.ErrNotFound, "User not found", int(xerror.CodeNotFound))
	xlog.Error("Error with specific code", "error", notFoundErr)

	if xerror.IsErrorCode(notFoundErr, xerror.CodeNotFound) {
		xlog.Info("IsErrorCode example: notFoundErr has CodeNotFound")
	}

	// Retry mechanism
	retryFunc := func() error {
		return xerror.ErrTimeout
	}

	retryErr := xerror.Retry(3, time.Second, retryFunc)
	xlog.Error("Retry mechanism example", "error", retryErr)

	// Panic and recover
	defer func() {
		if r := recover(); r != nil {
			xlog.Error("Recovered from panic", "panic", r)
		}
	}()

	xerror.PanicIfError(xerror.New("This will cause a panic"))

	// This line will not be reached due to the panic above
	fmt.Println("This line will not be executed")
}
