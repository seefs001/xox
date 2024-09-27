package x

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Must1 returns the result if no error occurred.
// If an error occurred, it panics with the error.
func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Must2 returns the results if no error occurred.
// If an error occurred, it panics with the error.
func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}

// Must3 returns the results if no error occurred.
// If an error occurred, it panics with the error.
func Must3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3) {
	if err != nil {
		panic(err)
	}
	return v1, v2, v3
}

// Must4 returns the results if no error occurred.
// If an error occurred, it panics with the error.
func Must4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, err error) (T1, T2, T3, T4) {
	if err != nil {
		panic(err)
	}
	return v1, v2, v3, v4
}

// Must0 panics if the error is not nil.
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Ignore1 returns the result, ignoring any error.
func Ignore1[T any](v T, _ error) T {
	return v
}

// Ignore2 returns the results, ignoring any error.
func Ignore2[T1, T2 any](v1 T1, v2 T2, _ error) (T1, T2) {
	return v1, v2
}

// Ignore3 returns the results, ignoring any error.
func Ignore3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, _ error) (T1, T2, T3) {
	return v1, v2, v3
}

// Ignore4 returns the results, ignoring any error.
func Ignore4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, _ error) (T1, T2, T3, T4) {
	return v1, v2, v3, v4
}

// Ignore0 ignores the error.
func Ignore0(_ error) {
}

// Where returns a new slice containing all elements of the collection
// that satisfy the predicate f.
func Where[T any](collection []T, predicate func(T) bool) []T {
	if collection == nil || predicate == nil {
		return nil
	}
	result := make([]T, 0)
	for _, item := range collection {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Select returns a new slice containing the results of applying
// the function f to each element of the original slice.
func Select[T any, U any](collection []T, selector func(T) U) []U {
	if collection == nil || selector == nil {
		return nil
	}
	result := make([]U, len(collection))
	for i, item := range collection {
		result[i] = selector(item)
	}
	return result
}

// Aggregate reduces collection to a single value using a reduction function.
func Aggregate[T any, U any](collection []T, seed U, accumulator func(U, T) U) U {
	if collection == nil || accumulator == nil {
		return seed
	}
	result := seed
	for _, item := range collection {
		result = accumulator(result, item)
	}
	return result
}

// ForEach iterates over the collection and applies the action to each element.
func ForEach[T any](collection []T, action func(T)) {
	if collection == nil || action == nil {
		return
	}
	for _, item := range collection {
		action(item)
	}
}

// Range generates a sequence of integers.
func Range[T int](start T, count int) []T {
	if count < 0 {
		return nil
	}
	result := make([]T, count)
	for i := 0; i < count; i++ {
		result[i] = start + T(i)
	}
	return result
}

// Count returns the number of elements in the collection that satisfy the predicate.
func Count[T any](collection []T, predicate func(T) bool) int {
	if collection == nil || predicate == nil {
		return 0
	}
	count := 0
	for _, item := range collection {
		if predicate(item) {
			count++
		}
	}
	return count
}

// GroupBy groups elements of the collection by keys returned by the keySelector.
func GroupBy[T any, K comparable](collection []T, keySelector func(T) K) map[K][]T {
	if collection == nil || keySelector == nil {
		return nil
	}
	result := make(map[K][]T)
	for _, item := range collection {
		key := keySelector(item)
		result[key] = append(result[key], item)
	}
	return result
}

// First returns the first element of the collection that satisfies the predicate.
func First[T any](collection []T, predicate func(T) bool) (T, bool) {
	if collection == nil || predicate == nil {
		var zero T
		return zero, false
	}
	for _, item := range collection {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Last returns the last element of the collection that satisfies the predicate.
func Last[T any](collection []T, predicate func(T) bool) (T, bool) {
	if collection == nil || predicate == nil {
		var zero T
		return zero, false
	}
	for i := len(collection) - 1; i >= 0; i-- {
		if predicate(collection[i]) {
			return collection[i], true
		}
	}
	var zero T
	return zero, false
}

// Any returns true if any element in the collection satisfies the predicate.
func Any[T any](collection []T, predicate func(T) bool) bool {
	if collection == nil || predicate == nil {
		return false
	}
	for _, item := range collection {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all elements in the collection satisfy the predicate.
func All[T any](collection []T, predicate func(T) bool) bool {
	if collection == nil || predicate == nil {
		return false
	}
	for _, item := range collection {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// RandomStringMode defines the mode for random string generation
type RandomStringMode int

const (
	// ModeAlphanumeric includes lowercase and uppercase letters, and digits
	ModeAlphanumeric RandomStringMode = iota
	// ModeAlpha includes only lowercase and uppercase letters
	ModeAlpha
	// ModeNumeric includes only digits
	ModeNumeric
	// ModeLowercase includes only lowercase letters
	ModeLowercase
	// ModeUppercase includes only uppercase letters
	ModeUppercase
)

// charsets for different modes
var (
	alphanumeric = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	alpha        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numeric      = []rune("0123456789")
	lowercase    = []rune("abcdefghijklmnopqrstuvwxyz")
	uppercase    = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandomString generates a random string of the specified length using the given mode.
func RandomString(length int, mode RandomStringMode) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	var charset []rune
	switch mode {
	case ModeAlphanumeric:
		charset = alphanumeric
	case ModeAlpha:
		charset = alpha
	case ModeNumeric:
		charset = numeric
	case ModeLowercase:
		charset = lowercase
	case ModeUppercase:
		charset = uppercase
	default:
		return "", errors.New("invalid mode")
	}

	return randomStringWithCharset(length, charset)
}

// RandomStringCustom generates a random string of the specified length using a custom character set.
func RandomStringCustom(length int, charset string) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	if charset == "" {
		return "", errors.New("charset must not be empty")
	}
	return randomStringWithCharset(length, []rune(charset))
}

// randomStringWithCharset is a helper function to generate a random string using the given charset.
func randomStringWithCharset(length int, charset []rune) (string, error) {
	result := make([]rune, length)
	for i := range result {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[randomIndex.Int64()]
	}
	return string(result), nil
}

// Tuple represents a generic tuple type with two elements.
type Tuple[T1, T2 any] struct {
	First  T1
	Second T2
}

// NewTuple creates a new Tuple with the given values.
func NewTuple[T1, T2 any](first T1, second T2) Tuple[T1, T2] {
	return Tuple[T1, T2]{
		First:  first,
		Second: second,
	}
}

// Unpack returns the two elements of the Tuple.
func (t Tuple[T1, T2]) Unpack() (T1, T2) {
	return t.First, t.Second
}

// Swap returns a new Tuple with the elements in reverse order.
func (t Tuple[T1, T2]) Swap() Tuple[T2, T1] {
	return NewTuple(t.Second, t.First)
}

// Map applies the given function to both elements of the Tuple and returns a new Tuple.
func (t Tuple[T1, T2]) Map(f func(T1, T2) (T1, T2)) Tuple[T1, T2] {
	newFirst, newSecond := f(t.First, t.Second)
	return NewTuple(newFirst, newSecond)
}

// Triple represents a generic tuple type with three elements.
type Triple[T1, T2, T3 any] struct {
	First  T1
	Second T2
	Third  T3
}

// NewTriple creates a new Triple with the given values.
func NewTriple[T1, T2, T3 any](first T1, second T2, third T3) Triple[T1, T2, T3] {
	return Triple[T1, T2, T3]{
		First:  first,
		Second: second,
		Third:  third,
	}
}

// Unpack returns the three elements of the Triple.
func (t Triple[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return t.First, t.Second, t.Third
}

// Rotate returns a new Triple with the elements rotated one position to the left.
func (t Triple[T1, T2, T3]) Rotate() Triple[T2, T3, T1] {
	return NewTriple(t.Second, t.Third, t.First)
}

// Map applies the given function to all elements of the Triple and returns a new Triple.
func (t Triple[T1, T2, T3]) Map(f func(T1, T2, T3) (T1, T2, T3)) Triple[T1, T2, T3] {
	newFirst, newSecond, newThird := f(t.First, t.Second, t.Third)
	return NewTriple(newFirst, newSecond, newThird)
}

// Map applies a given function to each element of a slice and returns a new slice with the results.
func Map[T any, U any](collection []T, f func(T) U) []U {
	if collection == nil || f == nil {
		return nil
	}
	result := make([]U, len(collection))
	for i, v := range collection {
		result[i] = f(v)
	}
	return result
}

// Filter returns a new slice containing all elements of the collection that satisfy the predicate f.
func Filter[T any](collection []T, f func(T) bool) []T {
	if collection == nil || f == nil {
		return nil
	}
	var result []T
	for _, v := range collection {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce applies a function against an accumulator and each element in the slice to reduce it to a single value.
func Reduce[T any, U any](collection []T, initialValue U, f func(U, T) U) U {
	if collection == nil || f == nil {
		return initialValue
	}
	result := initialValue
	for _, v := range collection {
		result = f(result, v)
	}
	return result
}

// Contains checks if an element is present in a slice.
func Contains[T comparable](collection []T, element T) bool {
	if collection == nil {
		return false
	}
	for _, v := range collection {
		if v == element {
			return true
		}
	}
	return false
}

// Unique returns a new slice with duplicate elements removed.
func Unique[T comparable](collection []T) []T {
	if collection == nil {
		return nil
	}
	seen := make(map[T]struct{})
	var result []T
	for _, v := range collection {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Chunk splits a slice into chunks of specified size.
func Chunk[T any](collection []T, size int) [][]T {
	if collection == nil || size <= 0 {
		return nil
	}
	var chunks [][]T
	for {
		if len(collection) == 0 {
			break
		}
		if len(collection) < size {
			size = len(collection)
		}
		chunks = append(chunks, collection[0:size])
		collection = collection[size:]
	}
	return chunks
}

// Flatten flattens a slice of slices into a single slice.
func Flatten[T any](collection [][]T) []T {
	if collection == nil {
		return nil
	}
	var result []T
	for _, inner := range collection {
		result = append(result, inner...)
	}
	return result
}

// Zip returns a slice of Tuples where each Tuple contains the i-th element from each of the input slices.
func Zip[T1, T2 any](slice1 []T1, slice2 []T2) []Tuple[T1, T2] {
	if slice1 == nil || slice2 == nil {
		return nil
	}
	minLen := len(slice1)
	if len(slice2) < minLen {
		minLen = len(slice2)
	}
	result := make([]Tuple[T1, T2], minLen)
	for i := 0; i < minLen; i++ {
		result[i] = NewTuple(slice1[i], slice2[i])
	}
	return result
}

// Reverse returns a new slice with the elements in reverse order.
func Reverse[T any](collection []T) []T {
	if collection == nil {
		return nil
	}
	result := make([]T, len(collection))
	for i, v := range collection {
		result[len(collection)-1-i] = v
	}
	return result
}

// ParallelFor executes the given function in parallel for each item in the collection.
func ParallelFor[T any](collection []T, f func(T)) {
	if collection == nil || f == nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(collection))
	for _, item := range collection {
		go func(v T) {
			defer wg.Done()
			f(v)
		}(item)
	}
	wg.Wait()
}

// ParallelMap applies the given function to each element of the slice in parallel
// and returns a new slice with the results.
func ParallelMap[T any, U any](collection []T, f func(T) U) []U {
	if collection == nil || f == nil {
		return nil
	}
	result := make([]U, len(collection))
	var wg sync.WaitGroup
	wg.Add(len(collection))
	for i, item := range collection {
		go func(index int, value T) {
			defer wg.Done()
			result[index] = f(value)
		}(i, item)
	}
	wg.Wait()
	return result
}

// Debounce returns a function that delays invoking `f` until after `wait` duration
// has elapsed since the last time the debounced function was invoked.
func Debounce(f func(), wait time.Duration) func() {
	if f == nil || wait <= 0 {
		return func() {}
	}
	var mutex sync.Mutex
	var timer *time.Timer

	return func() {
		mutex.Lock()
		defer mutex.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(wait, f)
	}
}

// Throttle returns a function that, when invoked repeatedly,
// will only actually call the original function at most once per every `wait` duration.
func Throttle(f func(), wait time.Duration) func() {
	if f == nil || wait <= 0 {
		return func() {}
	}
	var mutex sync.Mutex
	var lastTime time.Time

	return func() {
		mutex.Lock()
		defer mutex.Unlock()

		now := time.Now()
		if now.Sub(lastTime) >= wait {
			f()
			lastTime = now
		}
	}
}

// AsyncTask represents an asynchronous task that can be executed and waited upon.
type AsyncTask[T any] struct {
	result T
	err    error
	done   chan struct{}
}

// NewAsyncTask creates a new AsyncTask and starts executing the given function.
func NewAsyncTask[T any](f func() (T, error)) *AsyncTask[T] {
	if f == nil {
		return nil
	}
	task := &AsyncTask[T]{
		done: make(chan struct{}),
	}

	go func() {
		task.result, task.err = f()
		close(task.done)
	}()

	return task
}

// Wait blocks until the task is complete and returns the result and any error.
func (t *AsyncTask[T]) Wait() (T, error) {
	if t == nil {
		var zero T
		return zero, errors.New("nil AsyncTask")
	}
	<-t.done
	return t.result, t.err
}

// WaitWithTimeout blocks until the task is complete or the timeout is reached.
func (t *AsyncTask[T]) WaitWithTimeout(timeout time.Duration) (T, error, bool) {
	if t == nil {
		var zero T
		return zero, errors.New("nil AsyncTask"), false
	}
	select {
	case <-t.done:
		return t.result, t.err, true
	case <-time.After(timeout):
		var zero T
		return zero, nil, false
	}
}

// WaitWithContext blocks until the task is complete or the context is done.
func (t *AsyncTask[T]) WaitWithContext(ctx context.Context) (T, error, bool) {
	if t == nil {
		var zero T
		return zero, errors.New("nil AsyncTask"), false
	}
	select {
	case <-t.done:
		return t.result, t.err, true
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err(), false
	}
}

// SafePool represents a pool of goroutines with panic recovery
type SafePool struct {
	workers chan struct{}
}

// NewSafePool creates a new SafePool with the specified number of workers
// If size is 0 or negative, it creates an unbounded pool
func NewSafePool(size int) *SafePool {
	if size <= 0 {
		return &SafePool{}
	}
	return &SafePool{
		workers: make(chan struct{}, size),
	}
}

// SafeGo runs the given function in a goroutine with panic recovery.
// It returns a channel that will receive an error if a panic occurs,
// or nil if the function completes successfully.
func (p *SafePool) SafeGo(f func()) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		if p.workers != nil {
			p.workers <- struct{}{}
			defer func() { <-p.workers }()
		}

		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic recovered: %v", r)
			}
			close(errChan)
		}()

		f()
	}()

	return errChan
}

// SafeGoWithContext runs the given function in a goroutine with panic recovery and context cancellation.
// It returns a channel that will receive an error if a panic occurs or the context is cancelled,
// or nil if the function completes successfully.
func (p *SafePool) SafeGoWithContext(ctx context.Context, f func()) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		if p.workers != nil {
			p.workers <- struct{}{}
			defer func() { <-p.workers }()
		}

		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic recovered: %v", r)
			}
			close(errChan)
		}()

		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
		default:
			f()
		}
	}()

	return errChan
}

// SafeGoNoError runs the given function in a goroutine with panic recovery.
// It does not return any error information.
func (p *SafePool) SafeGoNoError(f func()) {
	go func() {
		if p.workers != nil {
			p.workers <- struct{}{}
			defer func() { <-p.workers }()
		}

		defer func() {
			if r := recover(); r != nil {
				// Log the panic instead of sending it to a channel
				fmt.Printf("Panic recovered: %v\n", r)
			}
		}()

		f()
	}()
}

// SafeGoWithContextNoError runs the given function in a goroutine with panic recovery and context cancellation.
// It does not return any error information.
func (p *SafePool) SafeGoWithContextNoError(ctx context.Context, f func()) {
	go func() {
		if p.workers != nil {
			p.workers <- struct{}{}
			defer func() { <-p.workers }()
		}

		defer func() {
			if r := recover(); r != nil {
				// Log the panic instead of sending it to a channel
				fmt.Printf("Panic recovered: %v\n", r)
			}
		}()

		select {
		case <-ctx.Done():
			// Log context cancellation instead of sending it to a channel
			fmt.Printf("Context cancelled: %v\n", ctx.Err())
		default:
			f()
		}
	}()
}

// For backwards compatibility, keep the global functions
var defaultPool = NewSafePool(0)

func SafeGo(f func()) <-chan error {
	return defaultPool.SafeGo(f)
}

func SafeGoWithContext(ctx context.Context, f func()) <-chan error {
	return defaultPool.SafeGoWithContext(ctx, f)
}

func SafeGoNoError(f func()) {
	defaultPool.SafeGoNoError(f)
}

func SafeGoWithContextNoError(ctx context.Context, f func()) {
	defaultPool.SafeGoWithContextNoError(ctx, f)
}

// WaitGroup is a safer version of sync.WaitGroup with panic recovery and error handling
type WaitGroup struct {
	wg     sync.WaitGroup
	errMu  sync.Mutex
	errors []error
}

// Go runs the given function in a new goroutine and adds it to the WaitGroup
func (wg *WaitGroup) Go(f func() error) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				wg.errMu.Lock()
				wg.errors = append(wg.errors, fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack()))
				wg.errMu.Unlock()
			}
		}()
		if err := f(); err != nil {
			wg.errMu.Lock()
			wg.errors = append(wg.errors, err)
			wg.errMu.Unlock()
		}
	}()
}

// WaitWithTimeout waits for all goroutines to complete or the timeout to expire
func (wg *WaitGroup) WaitWithTimeout(timeout time.Duration) (error, bool) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.wg.Wait()
	}()
	select {
	case <-c:
		return wg.Wait(), false
	case <-time.After(timeout):
		return nil, true
	}
}

// GoWithContext runs the given function in a new goroutine with context support and adds it to the WaitGroup
func (wg *WaitGroup) GoWithContext(ctx context.Context, f func(context.Context) error) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				wg.errMu.Lock()
				wg.errors = append(wg.errors, fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack()))
				wg.errMu.Unlock()
			}
		}()
		if err := f(ctx); err != nil {
			wg.errMu.Lock()
			wg.errors = append(wg.errors, err)
			wg.errMu.Unlock()
		}
	}()
}

// Wait waits for all goroutines to complete and returns any errors that occurred
func (wg *WaitGroup) Wait() error {
	wg.wg.Wait()
	if len(wg.errors) == 1 {
		return wg.errors[0]
	}
	if len(wg.errors) > 1 {
		return fmt.Errorf("multiple errors occurred: %v", wg.errors)
	}
	return nil
}

// NewWaitGroup creates a new WaitGroup
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// ToJSON converts a struct or map to a JSON string
func ToJSON(v interface{}) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// MustToJSON converts a struct or map to a JSON string, panicking on error
func MustToJSON(v interface{}) string {
	jsonStr, err := ToJSON(v)
	if err != nil {
		panic(err)
	}
	return jsonStr
}

// Ternary is a generic function that implements the ternary operator
func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TernaryF is a generic function that implements the ternary operator with lazy evaluation
func TernaryF[T any](condition bool, ifTrue, ifFalse func() T) T {
	if condition {
		return ifTrue()
	}
	return ifFalse()
}

// If is a generic function that implements an if-else chain
func If[T any](condition bool, then T) *ifChain[T] {
	if condition {
		return &ifChain[T]{result: then, done: true}
	}
	return &ifChain[T]{}
}

type ifChain[T any] struct {
	result T
	done   bool
}

func (ic *ifChain[T]) ElseIf(condition bool, then T) *ifChain[T] {
	if !ic.done && condition {
		ic.result = then
		ic.done = true
	}
	return ic
}

func (ic *ifChain[T]) Else(otherwise T) T {
	if !ic.done {
		return otherwise
	}
	return ic.result
}

// Switch is a generic function that implements a switch statement
func Switch[T comparable, R any](value T) *switchChain[T, R] {
	return &switchChain[T, R]{value: value}
}

type switchChain[T comparable, R any] struct {
	value  T
	result R
	done   bool
}

func (sc *switchChain[T, R]) Case(caseValue T, result R) *switchChain[T, R] {
	if !sc.done && sc.value == caseValue {
		sc.result = result
		sc.done = true
	}
	return sc
}

func (sc *switchChain[T, R]) Default(defaultResult R) R {
	if !sc.done {
		return defaultResult
	}
	return sc.result
}

// IsEmpty checks if a value is considered empty
func IsEmpty[T any](value T) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Interface:
		return v.IsNil() || IsEmpty(v.Elem().Interface())
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	}
	return false
}

// IsNil checks if a value is nil
func IsNil(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// IsZero checks if a value is the zero value for its type
func IsZero[T any](value T) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !IsZero(v.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsZero(v.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.UnsafePointer:
		return v.Pointer() == 0
	}
	return false
}

// IsImageURL checks if a string is a valid image URL
func IsImageURL(s string) bool {
	// Check if the URL starts with http:// or https://
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return false
	}

	// Simple check for common image file extensions
	// You might want to implement a more robust check based on your requirements
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	lowered := strings.ToLower(s)
	for _, ext := range extensions {
		if strings.HasSuffix(lowered, ext) {
			return true
		}
	}
	return false
}

// IsBase64 checks if a string is a valid base64 encoded string
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
