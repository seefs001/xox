package x

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seefs001/xox/xerror"
)

// OnlyErr returns only the error from a function that returns multiple values,
// where the last value is an error. It discards all other return values.
//
// Example:
//
//	err := OnlyErr(SomeFunction())
//	if err != nil {
//		// handle error
//	}
func OnlyErr(values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	lastValue := values[len(values)-1]
	if err, ok := lastValue.(error); ok {
		return err
	}

	return nil
}

// Must1 returns the result if no error occurred.
// If an error occurred, it panics with the error.
//
// Example:
//
//	result := Must1(SomeFunction())
//	// If SomeFunction() returns an error, this will panic
func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Must2 returns the results if no error occurred.
// If an error occurred, it panics with the error.
//
// Example:
//
//	v1, v2 := Must2(SomeFunctionReturningTwoValues())
//	// If SomeFunctionReturningTwoValues() returns an error, this will panic
func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}

// Must3 returns the results if no error occurred.
// If an error occurred, it panics with the error.
//
// Example:
//
//	v1, v2, v3 := Must3(SomeFunctionReturningThreeValues())
//	// If SomeFunctionReturningThreeValues() returns an error, this will panic
func Must3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3) {
	if err != nil {
		panic(err)
	}
	return v1, v2, v3
}

// Must4 returns the results if no error occurred.
// If an error occurred, it panics with the error.
//
// Example:
//
//	v1, v2, v3, v4 := Must4(SomeFunctionReturningFourValues())
//	// If SomeFunctionReturningFourValues() returns an error, this will panic
func Must4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, err error) (T1, T2, T3, T4) {
	if err != nil {
		panic(err)
	}
	return v1, v2, v3, v4
}

// Must0 panics if the error is not nil.
//
// Example:
//
//	Must0(SomeFunctionReturningOnlyError())
//	// If SomeFunctionReturningOnlyError() returns an error, this will panic
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Ignore1 returns the result, ignoring any error.
//
// Example:
//
//	result := Ignore1(SomeFunction())
//	// Even if SomeFunction() returns an error, it will be ignored
func Ignore1[T any](v T, _ error) T {
	return v
}

// Ignore2 returns the results, ignoring any error.
//
// Example:
//
//	v1, v2 := Ignore2(SomeFunctionReturningTwoValues())
//	// Even if SomeFunctionReturningTwoValues() returns an error, it will be ignored
func Ignore2[T1, T2 any](v1 T1, v2 T2, _ error) (T1, T2) {
	return v1, v2
}

// Ignore3 returns the results, ignoring any error.
//
// Example:
//
//	v1, v2, v3 := Ignore3(SomeFunctionReturningThreeValues())
//	// Even if SomeFunctionReturningThreeValues() returns an error, it will be ignored
func Ignore3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, _ error) (T1, T2, T3) {
	return v1, v2, v3
}

// Ignore4 returns the results, ignoring any error.
//
// Example:
//
//	v1, v2, v3, v4 := Ignore4(SomeFunctionReturningFourValues())
//	// Even if SomeFunctionReturningFourValues() returns an error, it will be ignored
func Ignore4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, _ error) (T1, T2, T3, T4) {
	return v1, v2, v3, v4
}

// Ignore0 ignores the error.
//
// Example:
//
//	Ignore0(SomeFunctionReturningOnlyError())
//	// The error returned by SomeFunctionReturningOnlyError() will be ignored
func Ignore0(_ error) {
}

// Where returns a new slice containing all elements of the collection
// that satisfy the predicate f.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	evenNumbers := Where(numbers, func(n int) bool { return n%2 == 0 })
//	// evenNumbers will be [2, 4]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	squares := Select(numbers, func(n int) int { return n * n })
//	// squares will be [1, 4, 9, 16, 25]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	sum := Aggregate(numbers, 0, func(acc, n int) int { return acc + n })
//	// sum will be 15
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	ForEach(numbers, func(n int) { fmt.Println(n) })
//	// This will print each number on a new line
func ForEach[T any](collection []T, action func(T)) {
	if collection == nil || action == nil {
		return
	}
	for _, item := range collection {
		action(item)
	}
}

// Range generates a sequence of integers.
//
// Example:
//
//	numbers := Range(1, 5)
//	// numbers will be [1, 2, 3, 4, 5]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	evenCount := Count(numbers, func(n int) bool { return n%2 == 0 })
//	// evenCount will be 2
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
//
// Example:
//
//	type Person struct {
//		Name string
//		Age  int
//	}
//	people := []Person{
//		{"Alice", 25},
//		{"Bob", 30},
//		{"Charlie", 25},
//	}
//	groupedByAge := GroupBy(people, func(p Person) int { return p.Age })
//	// groupedByAge will be map[int][]Person{
//	//   25: {{"Alice", 25}, {"Charlie", 25}},
//	//   30: {{"Bob", 30}},
//	// }
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	firstEven, found := First(numbers, func(n int) bool { return n%2 == 0 })
//	// firstEven will be 2, found will be true
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	lastEven, found := Last(numbers, func(n int) bool { return n%2 == 0 })
//	// lastEven will be 4, found will be true
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	hasEven := Any(numbers, func(n int) bool { return n%2 == 0 })
//	// hasEven will be true
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
//
// Example:
//
//	numbers := []int{2, 4, 6, 8, 10}
//	allEven := All(numbers, func(n int) bool { return n%2 == 0 })
//	// allEven will be true
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
//
// Example:
//
//	randomStr, err := RandomString(10, ModeAlphanumeric)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(randomStr) // Outputs a random 10-character alphanumeric string
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
//
// Example:
//
//	customCharset := "ABC123"
//	randomStr, err := RandomStringCustom(5, customCharset)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(randomStr) // Outputs a random 5-character string using only 'A', 'B', 'C', '1', '2', '3'
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
		randomIndex, err := RandomInt(0, len(charset)-1)
		if err != nil {
			return "", xerror.Wrap(err, "failed to generate random index")
		}
		result[i] = charset[randomIndex]
	}
	return string(result), nil
}

// RandomInt generates a random integer between min and max (inclusive).
func RandomInt(min, max int) (int, error) {
	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}
	return rand.Intn(max-min+1) + min, nil
}

// Tuple represents a generic tuple type with two elements.
type Tuple[T1, T2 any] struct {
	First  T1
	Second T2
}

// NewTuple creates a new Tuple with the given values.
//
// Example:
//
//	t := NewTuple("hello", 42)
//	fmt.Println(t.First)  // Outputs: hello
//	fmt.Println(t.Second) // Outputs: 42
func NewTuple[T1, T2 any](first T1, second T2) Tuple[T1, T2] {
	return Tuple[T1, T2]{
		First:  first,
		Second: second,
	}
}

// Unpack returns the two elements of the Tuple.
//
// Example:
//
//	t := NewTuple("hello", 42)
//	str, num := t.Unpack()
//	fmt.Println(str) // Outputs: hello
//	fmt.Println(num) // Outputs: 42
func (t Tuple[T1, T2]) Unpack() (T1, T2) {
	return t.First, t.Second
}

// Swap returns a new Tuple with the elements in reverse order.
//
// Example:
//
//	t := NewTuple("hello", 42)
//	swapped := t.Swap()
//	fmt.Println(swapped.First)  // Outputs: 42
//	fmt.Println(swapped.Second) // Outputs: hello
func (t Tuple[T1, T2]) Swap() Tuple[T2, T1] {
	return NewTuple(t.Second, t.First)
}

// Map applies the given function to both elements of the Tuple and returns a new Tuple.
//
// Example:
//
//	t := NewTuple(2, 3)
//	doubled := t.Map(func(a, b int) (int, int) { return a * 2, b * 2 })
//	fmt.Println(doubled.First)  // Outputs: 4
//	fmt.Println(doubled.Second) // Outputs: 6
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
//
// Example:
//
//	t := NewTriple("hello", 42, true)
//	fmt.Println(t.First)  // Outputs: hello
//	fmt.Println(t.Second) // Outputs: 42
//	fmt.Println(t.Third)  // Outputs: true
func NewTriple[T1, T2, T3 any](first T1, second T2, third T3) Triple[T1, T2, T3] {
	return Triple[T1, T2, T3]{
		First:  first,
		Second: second,
		Third:  third,
	}
}

// Unpack returns the three elements of the Triple.
//
// Example:
//
//	t := NewTriple("hello", 42, true)
//	str, num, bool := t.Unpack()
//	fmt.Println(str)  // Outputs: hello
//	fmt.Println(num)  // Outputs: 42
//	fmt.Println(bool) // Outputs: true
func (t Triple[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return t.First, t.Second, t.Third
}

// Rotate returns a new Triple with the elements rotated one position to the left.
//
// Example:
//
//	t := NewTriple("hello", 42, true)
//	rotated := t.Rotate()
//	fmt.Println(rotated.First)  // Outputs: 42
//	fmt.Println(rotated.Second) // Outputs: true
//	fmt.Println(rotated.Third)  // Outputs: hello
func (t Triple[T1, T2, T3]) Rotate() Triple[T2, T3, T1] {
	return NewTriple(t.Second, t.Third, t.First)
}

// Map applies the given function to all elements of the Triple and returns a new Triple.
//
// Example:
//
//	t := NewTriple(2, 3, 4)
//	doubled := t.Map(func(a, b, c int) (int, int, int) { return a * 2, b * 2, c * 2 })
//	fmt.Println(doubled.First)  // Outputs: 4
//	fmt.Println(doubled.Second) // Outputs: 6
//	fmt.Println(doubled.Third)  // Outputs: 8
func (t Triple[T1, T2, T3]) Map(f func(T1, T2, T3) (T1, T2, T3)) Triple[T1, T2, T3] {
	newFirst, newSecond, newThird := f(t.First, t.Second, t.Third)
	return NewTriple(newFirst, newSecond, newThird)
}

// Map applies a given function to each element of a slice and returns a new slice with the results.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	squares := Map(numbers, func(n int) int { return n * n })
//	fmt.Println(squares) // Outputs: [1 4 9 16 25]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	evenNumbers := Filter(numbers, func(n int) bool { return n%2 == 0 })
//	fmt.Println(evenNumbers) // Outputs: [2 4]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	sum := Reduce(numbers, 0, func(acc, n int) int { return acc + n })
//	fmt.Println(sum) // Outputs: 15
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	fmt.Println(Contains(numbers, 3)) // Outputs: true
//	fmt.Println(Contains(numbers, 6)) // Outputs: false
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
//
// Example:
//
//	numbers := []int{1, 2, 2, 3, 3, 4, 5, 5}
//	uniqueNumbers := Unique(numbers)
//	fmt.Println(uniqueNumbers) // Outputs: [1 2 3 4 5]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5, 6, 7}
//	chunks := Chunk(numbers, 3)
//	fmt.Println(chunks) // Outputs: [[1 2 3] [4 5 6] [7]]
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
//
// Example:
//
//	nestedSlice := [][]int{{1, 2}, {3, 4}, {5, 6}}
//	flattened := Flatten(nestedSlice)
//	fmt.Println(flattened) // Outputs: [1 2 3 4 5 6]
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
//
// Example:
//
//	names := []string{"Alice", "Bob", "Charlie"}
//	ages := []int{25, 30, 35}
//	zipped := Zip(names, ages)
//	for _, t := range zipped {
//	    fmt.Printf("%s is %d years old\n", t.First, t.Second)
//	}
//	// Outputs:
//	// Alice is 25 years old
//	// Bob is 30 years old
//	// Charlie is 35 years old
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	reversed := Reverse(numbers)
//	fmt.Println(reversed) // Outputs: [5 4 3 2 1]
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	ParallelFor(numbers, func(n int) {
//	    fmt.Printf("Processing %d\n", n)
//	})
//	// Outputs (order may vary):
//	// Processing 1
//	// Processing 2
//	// Processing 3
//	// Processing 4
//	// Processing 5
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
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	squares := ParallelMap(numbers, func(n int) int {
//	    return n * n
//	})
//	fmt.Println(squares) // Outputs: [1 4 9 16 25] (order may vary)
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
//
// Example:
//
//	debouncedPrint := Debounce(func() {
//	    fmt.Println("Debounced function called")
//	}, 100*time.Millisecond)
//
//	for i := 0; i < 5; i++ {
//	    debouncedPrint()
//	    time.Sleep(50 * time.Millisecond)
//	}
//	time.Sleep(200 * time.Millisecond)
//	// Outputs: Debounced function called (only once, after the last call)
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
//
// Parameters:
// - f: The function to be throttled.
// - wait: The minimum duration between function calls.
//
// Returns:
// A new function that wraps the original function with throttling behavior.
//
// Example:
//
//	throttledPrint := Throttle(func() {
//	    fmt.Println("Throttled function called")
//	}, 1*time.Second)
//
//	for i := 0; i < 5; i++ {
//	    throttledPrint()
//	    time.Sleep(300 * time.Millisecond)
//	}
//	// Output: Throttled function called (printed only twice, at 0s and 1s)
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
// It encapsulates the result, error, and completion status of the task.
type AsyncTask[T any] struct {
	result T
	err    error
	done   chan struct{}
}

// NewAsyncTask creates a new AsyncTask and starts executing the given function.
//
// Parameters:
// - f: A function that returns a value of type T and an error.
//
// Returns:
// A pointer to the created AsyncTask.
//
// Example:
//
//	task := NewAsyncTask(func() (int, error) {
//	    time.Sleep(2 * time.Second)
//	    return 42, nil
//	})
//
//	result, err := task.Wait()
//	fmt.Printf("Result: %d, Error: %v\n", result, err)
//	// Output: Result: 42, Error: <nil>
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
//
// Returns:
// - The result of type T from the task execution.
// - Any error that occurred during the task execution.
//
// Example:
//
//	task := NewAsyncTask(func() (string, error) {
//	    time.Sleep(1 * time.Second)
//	    return "Hello, World!", nil
//	})
//
//	result, err := task.Wait()
//	fmt.Printf("Result: %s, Error: %v\n", result, err)
//	// Output: Result: Hello, World!, Error: <nil>
func (t *AsyncTask[T]) Wait() (T, error) {
	if t == nil {
		var zero T
		return zero, errors.New("nil AsyncTask")
	}
	<-t.done
	return t.result, t.err
}

// WaitWithTimeout blocks until the task is complete or the timeout is reached.
//
// Parameters:
// - timeout: The maximum duration to wait for the task to complete.
//
// Returns:
// - The result of type T from the task execution.
// - Any error that occurred during the task execution.
// - A boolean indicating whether the task completed (true) or timed out (false).
//
// Example:
//
//	task := NewAsyncTask(func() (int, error) {
//	    time.Sleep(2 * time.Second)
//	    return 42, nil
//	})
//
//	result, err, completed := task.WaitWithTimeout(1 * time.Second)
//	fmt.Printf("Result: %d, Error: %v, Completed: %v\n", result, err, completed)
//	// Output: Result: 0, Error: <nil>, Completed: false
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
//
// Parameters:
// - ctx: The context to control the waiting operation.
//
// Returns:
// - The result of type T from the task execution.
// - Any error that occurred during the task execution or context cancellation.
// - A boolean indicating whether the task completed (true) or the context was done (false).
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	task := NewAsyncTask(func() (int, error) {
//	    time.Sleep(2 * time.Second)
//	    return 42, nil
//	})
//
//	result, err, completed := task.WaitWithContext(ctx)
//	fmt.Printf("Result: %d, Error: %v, Completed: %v\n", result, err, completed)
//	// Output: Result: 0, Error: context deadline exceeded, Completed: false
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
//
// Parameters:
// - size: The maximum number of concurrent workers. Use 0 or a negative value for an unbounded pool.
//
// Returns:
// A pointer to the created SafePool.
//
// Example:
//
//	pool := NewSafePool(5) // Creates a pool with a maximum of 5 concurrent workers
//	unboundedPool := NewSafePool(0) // Creates an unbounded pool
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
//
// Parameters:
// - f: The function to be executed in a goroutine.
//
// Returns:
// A channel that will receive an error if a panic occurs, or nil if the function completes successfully.
//
// Example:
//
//	pool := NewSafePool(5)
//	errChan := pool.SafeGo(func() {
//	    // Some potentially panicking code
//	    panic("Something went wrong")
//	})
//
//	if err := <-errChan; err != nil {
//	    fmt.Printf("Error occurred: %v\n", err)
//	}
//	// Output: Error occurred: panic recovered: Something went wrong
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
//
// Parameters:
// - ctx: The context to control the execution of the function.
// - f: The function to be executed in a goroutine.
//
// Returns:
// A channel that will receive an error if a panic occurs or the context is cancelled,
// or nil if the function completes successfully.
//
// Example:
//
//	pool := NewSafePool(5)
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	errChan := pool.SafeGoWithContext(ctx, func() {
//	    time.Sleep(2 * time.Second)
//	    fmt.Println("This won't be printed due to context cancellation")
//	})
//
//	if err := <-errChan; err != nil {
//	    fmt.Printf("Error occurred: %v\n", err)
//	}
//	// Output: Error occurred: context deadline exceeded
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
//
// Parameters:
// - f: The function to be executed in a goroutine.
//
// Example:
//
//	pool := NewSafePool(5)
//	pool.SafeGoNoError(func() {
//	    // Some potentially panicking code
//	    panic("Something went wrong")
//	})
//	// Output: Panic recovered: Something went wrong (printed to stdout)
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
//
// Parameters:
// - ctx: The context to control the execution of the function.
// - f: The function to be executed in a goroutine.
//
// Example:
//
//	pool := NewSafePool(5)
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	pool.SafeGoWithContextNoError(ctx, func() {
//	    time.Sleep(2 * time.Second)
//	    fmt.Println("This won't be printed due to context cancellation")
//	})
//	// Output: Context cancelled: context deadline exceeded (printed to stdout)
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
//
// Parameters:
// - f: The function to be executed in a goroutine.
//
// Example:
//
//	wg := NewWaitGroup()
//	wg.Go(func() error {
//	    // Some work here
//	    return nil
//	})
//	wg.Go(func() error {
//	    return errors.New("an error occurred")
//	})
//	if err := wg.Wait(); err != nil {
//	    fmt.Printf("Error occurred: %v\n", err)
//	}
//	// Output: Error occurred: an error occurred
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
//
// Parameters:
// - timeout: The maximum duration to wait for all goroutines to complete.
//
// Returns:
// - error: Any errors that occurred during the execution of the goroutines.
// - bool: true if the timeout was reached, false otherwise.
//
// Example:
//
//	wg := NewWaitGroup()
//	wg.Go(func() error {
//	    time.Sleep(2 * time.Second)
//	    return nil
//	})
//	err, timedOut := wg.WaitWithTimeout(1 * time.Second)
//	fmt.Printf("Error: %v, Timed out: %v\n", err, timedOut)
//	// Output: Error: <nil>, Timed out: true
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
//
// Parameters:
// - ctx: The context to control the execution of the function.
// - f: The function to be executed in a goroutine.
//
// Example:
//
//	wg := NewWaitGroup()
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	wg.GoWithContext(ctx, func(ctx context.Context) error {
//	    select {
//	    case <-time.After(2 * time.Second):
//	        return nil
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    }
//	})
//
//	if err := wg.Wait(); err != nil {
//	    fmt.Printf("Error occurred: %v\n", err)
//	}
//	// Output: Error occurred: context deadline exceeded
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
//
// Returns:
// - error: Any errors that occurred during the execution of the goroutines.
//
// Example:
//
//	wg := NewWaitGroup()
//	wg.Go(func() error {
//	    return errors.New("error 1")
//	})
//	wg.Go(func() error {
//	    return errors.New("error 2")
//	})
//	if err := wg.Wait(); err != nil {
//	    fmt.Printf("Errors occurred: %v\n", err)
//	}
//	// Output: Errors occurred: multiple errors occurred: [error 1 error 2]
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
//
// Returns:
// A pointer to the created WaitGroup.
//
// Example:
//
//	wg := NewWaitGroup()
//	// Use wg.Go() or wg.GoWithContext() to add tasks
//	// ...
//	if err := wg.Wait(); err != nil {
//	    fmt.Printf("Error occurred: %v\n", err)
//	}
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// ToJSON converts a struct or map to a JSON string
//
// Parameters:
// - v: The value to be converted to JSON.
//
// Returns:
// - string: The JSON representation of the input value.
// - error: Any error that occurred during the conversion.
//
// Example:
//
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	p := Person{Name: "John", Age: 30}
//	jsonStr, err := ToJSON(p)
//	if err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	} else {
//	    fmt.Printf("JSON: %s\n", jsonStr)
//	}
//	// Output: JSON: {"name":"John","age":30}
func ToJSON(v interface{}) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// MustToJSON converts a struct or map to a JSON string, panicking on error
//
// Parameters:
// - v: The value to be converted to JSON.
//
// Returns:
// - string: The JSON representation of the input value.
//
// Example:
//
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	p := Person{Name: "John", Age: 30}
//	jsonStr := MustToJSON(p)
//	fmt.Printf("JSON: %s\n", jsonStr)
//	// Output: JSON: {"name":"John","age":30}
func MustToJSON(v interface{}) string {
	jsonStr, err := ToJSON(v)
	if err != nil {
		panic(err)
	}
	return jsonStr
}

// Ternary is a generic function that implements the ternary operator
//
// Parameters:
// - condition: The condition to evaluate.
// - ifTrue: The value to return if the condition is true.
// - ifFalse: The value to return if the condition is false.
//
// Returns:
// The value of ifTrue if the condition is true, otherwise the value of ifFalse.
//
// Example:
//
//	result := Ternary(1 > 0, "Greater", "Less")
//	fmt.Println(result)
//	// Output: Greater
func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TernaryF is a generic function that implements the ternary operator with lazy evaluation
//
// Parameters:
// - condition: The condition to evaluate.
// - ifTrue: A function that returns the value if the condition is true.
// - ifFalse: A function that returns the value if the condition is false.
//
// Returns:
// The result of calling ifTrue if the condition is true, otherwise the result of calling ifFalse.
//
// Example:
//
//	result := TernaryF(1 > 0, func() string { return "Greater" }, func() string { return "Less" })
//	fmt.Println(result)
//	// Output: Greater
func TernaryF[T any](condition bool, ifTrue, ifFalse func() T) T {
	if condition {
		return ifTrue()
	}
	return ifFalse()
}

// If is a generic function that implements an if-else chain
//
// Parameters:
// - condition: The condition to evaluate.
// - then: The value to return if the condition is true.
//
// Returns:
// An ifChain object that can be used to chain additional conditions.
//
// Example:
//
//	result := If(x > 0, "Positive").
//	    ElseIf(x < 0, "Negative").
//	    Else("Zero")
//	fmt.Println(result)
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
//
// Parameters:
// - value: The value to switch on.
//
// Returns:
// A switchChain object that can be used to define cases.
//
// Example:
//
//	result := Switch(dayOfWeek).
//	    Case("Monday", "Start of the week").
//	    Case("Friday", "TGIF").
//	    Default("Regular day")
//	fmt.Println(result)
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
//
// Parameters:
// - value: The value to check for emptiness.
//
// Returns:
// true if the value is considered empty, false otherwise.
//
// Example:
//
//	fmt.Println(IsEmpty(""))        // true
//	fmt.Println(IsEmpty("hello"))   // false
//	fmt.Println(IsEmpty([]int{}))   // true
//	fmt.Println(IsEmpty([]int{1}))  // false
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
//
// Parameters:
// - value: The value to check for nil.
//
// Returns:
// true if the value is nil, false otherwise.
//
// Example:
//
//	var p *int
//	fmt.Println(IsNil(p))  // true
//	fmt.Println(IsNil(42)) // false
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
//
// Parameters:
// - value: The value to check for zero.
//
// Returns:
// true if the value is the zero value for its type, false otherwise.
//
// Example:
//
//	fmt.Println(IsZero(0))        // true
//	fmt.Println(IsZero(""))       // true
//	fmt.Println(IsZero(42))       // false
//	fmt.Println(IsZero("hello"))  // false
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
//
// Parameters:
// - s: The string to check.
//
// Returns:
// true if the string is a valid image URL, false otherwise.
//
// Example:
//
//	fmt.Println(IsImageURL("https://example.com/image.jpg"))  // true
//	fmt.Println(IsImageURL("https://example.com/file.txt"))   // false
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
//
// Parameters:
// - s: The string to check.
//
// Returns:
// true if the string is a valid base64 encoded string, false otherwise.
//
// Example:
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// TrimSuffixes removes all specified suffixes from the given string.
//
// Parameters:
// - s: The input string to trim.
// - suffixes: A variadic list of suffixes to remove from the input string.
//
// Returns:
// The input string with all specified suffixes removed.
//
// Example:
//
//	input := "example.tar.gz"
//	trimmed := TrimSuffixes(input, ".gz", ".tar")
//	fmt.Println(trimmed) // Output: "example"
func TrimSuffixes(s string, suffixes ...string) string {
	for _, suffix := range suffixes {
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}

// SetNonZeroValues sets non-zero values from src to dst.
// It updates the dst map with values from src that are not zero.
//
// Parameters:
// - dst: The destination map to be updated.
// - src: The source map containing values to be copied.
//
// Example:
//
//	dst := map[string]interface{}{"a": 1, "b": ""}
//	src := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
//	SetNonZeroValues(dst, src)
//	fmt.Println(dst) // Output: map[a:2 b:hello c:3]
func SetNonZeroValues(dst map[string]interface{}, src map[string]interface{}) {
	for key, value := range src {
		if !IsZero(value) {
			dst[key] = value
		}
	}
}

// SetNonZeroValuesWithKeys sets non-zero values from src to dst for specified keys.
// It updates the dst map with values from src that are not zero, but only for the specified keys.
//
// Parameters:
// - dst: The destination map to be updated.
// - src: The source map containing values to be copied.
// - keys: A variadic list of keys to consider when copying values.
//
// Example:
//
//	dst := map[string]interface{}{"a": 1, "b": ""}
//	src := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
//	SetNonZeroValuesWithKeys(dst, src, "a", "c")
//	fmt.Println(dst) // Output: map[a:2 b: c:3]
func SetNonZeroValuesWithKeys(dst map[string]interface{}, src map[string]interface{}, keys ...string) {
	for _, key := range keys {
		if value, ok := src[key]; ok && !IsZero(value) {
			dst[key] = value
		}
	}
}

// MapKeys returns a slice of keys from a map.
//
// Example:
//
//	m := map[string]int{"a": 1, "b": 2, "c": 3}
//	keys := MapKeys(m)
//	fmt.Println(keys) // Output: [a b c] (order may vary)
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues returns a slice of values from a map.
//
// Example:
//
//	m := map[string]int{"a": 1, "b": 2, "c": 3}
//	values := MapValues(m)
//	fmt.Println(values) // Output: [1 2 3] (order may vary)
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Shuffle randomly shuffles the elements in a slice.
//
// Example:
//
//	s := []int{1, 2, 3, 4, 5}
//	Shuffle(s)
//	fmt.Println(s) // Output: [3 1 5 2 4] (order will be random)
func Shuffle[T any](s []T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(s) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

// FlattenMap flattens a nested map into a single-level map with dot notation keys.
func FlattenMap(data map[string]any, prefix string) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		if subMap, ok := value.(map[string]any); ok {
			subResult := FlattenMap(subMap, fullKey)
			for subKey, subValue := range subResult {
				result[subKey] = subValue
			}
		} else {
			result[fullKey] = value
		}
	}
	return result
}

// CopyMap creates a deep copy of the given map.
//
// Example:
//
//	original := map[string]int{"a": 1, "b": 2}
//	copied := CopyMap(original)
//	copied["c"] = 3
//	fmt.Println(original) // Output: map[a:1 b:2]
//	fmt.Println(copied)   // Output: map[a:1 b:2 c:3]
func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}

	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// GenerateUUID generates a new UUID (Universally Unique Identifier) as a string.
// It uses the crypto/rand package to ensure cryptographically secure random numbers.
//
// Example:
//
//	uuid, err := GenerateUUID()
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(uuid) // Output: e.g., "f47ac10b-58cc-4372-a567-0e02b2c3d479"
func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", xerror.Wrap(err, "failed to generate UUID")
	}

	// Set version (4) and variant (2) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// DecodeUnicodeURL decodes a URL string that contains Unicode escape sequences
// back to its original form.
//
// Example:
//
//	encoded := "https://example.com/path?q=%E4%BD%A0%E5%A5%BD"
//	decoded, err := DecodeUnicodeURL(encoded)
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(decoded) // Output: https://example.com/path?q=你好
func DecodeUnicodeURL(encodedURL string) (string, error) {
	decodedURL, err := url.QueryUnescape(encodedURL)
	if err != nil {
		return "", xerror.Wrap(err, "failed to decode Unicode URL")
	}
	return decodedURL, nil
}

// EncodeUnicodeURL encodes a URL string to include Unicode escape sequences
// for non-ASCII characters.
//
// Example:
//
//	original := "https://example.com/path?q=你好"
//	encoded, err := EncodeUnicodeURL(original)
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(encoded) // Output: https://example.com/path?q=%E4%BD%A0%E5%A5%BD
func EncodeUnicodeURL(originalURL string) (string, error) {
	return url.QueryEscape(originalURL), nil
}

// JSONToURLValues converts a JSON string to url.Values.
//
// Example:
//
//	jsonStr := `{"key1": "value1", "key2": ["value2", "value3"]}`
//	values, err := JSONToURLValues(jsonStr)
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(values.Encode()) // Output: key1=value1&key2=value2&key2=value3
func JSONToURLValues(jsonStr string) (url.Values, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	for key, value := range data {
		switch v := value.(type) {
		case string:
			values.Add(key, v)
		case []interface{}:
			for _, item := range v {
				values.Add(key, fmt.Sprint(item))
			}
		case float64:
			if v == float64(int64(v)) {
				values.Add(key, strconv.FormatInt(int64(v), 10))
			} else {
				values.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
			}
		case int64:
			values.Add(key, strconv.FormatInt(v, 10))
		default:
			values.Add(key, fmt.Sprint(v))
		}
	}

	return values, nil
}

// MapToSlice transforms a map into a slice based on a specific iteratee function.
// It uses three generic type parameters:
// K: the type of the map keys (must be comparable)
// V: the type of the map values
// R: the type of the resulting slice elements
//
// Parameters:
// - data: the input map to be transformed
// - iteratee: a function that takes a key-value pair from the map and returns a value of type R
//
// Returns:
// - A slice of type []R containing the transformed elements
// - An error if any issues occur during the transformation
//
// Example with type inference:
//
//	data := map[string]int{"a": 1, "b": 2, "c": 3}
//	result, err := MapToSlice(data, func(k string, v int) string {
//	    return fmt.Sprintf("%s:%d", k, v)
//	})
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("%v\n", result) // Output: [a:1 b:2 c:3] (order may vary)
//
// Example with explicit type parameters:
//
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//
//	data := map[string]map[string]interface{}{
//	    "person1": {"name": "Alice", "age": 30},
//	    "person2": {"name": "Bob", "age": 25},
//	}
//
//	result, err := MapToSlice[string, map[string]interface{}, Person](data, func(k string, v map[string]interface{}) Person {
//	    return Person{
//	        Name: v["name"].(string),
//	        Age:  v["age"].(int),
//	    }
//	})
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("%+v\n", result) // Output: [{Name:Alice Age:30} {Name:Bob Age:25}] (order may vary)
func MapToSlice[K comparable, V any, R any](data map[K]V, iteratee func(K, V) R) ([]R, error) {
	var result []R

	if iteratee == nil {
		// Direct conversion using JSON marshaling/unmarshaling
		for _, v := range data {
			var item R
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, xerror.Wrap(err, "failed to marshal map value")
			}

			err = json.Unmarshal(jsonData, &item)
			if err != nil {
				return nil, xerror.Wrap(err, "failed to unmarshal into struct")
			}

			result = append(result, item)
		}
	} else {
		// Custom transformation using the provided iteratee function
		for k, v := range data {
			result = append(result, iteratee(k, v))
		}
	}

	return result, nil
}

// Ptr returns a pointer to the given value.
//
// This function is useful when you need to pass a pointer to a value,
// especially for optional fields in structs or when working with APIs
// that expect pointers.
//
// Example:
//
//	type Config struct {
//	    Debug *bool
//	}
//
//	cfg := Config{
//	    Debug: Ptr(true),
//	}
func Ptr[T any](v T) *T {
	return &v
}

// Deref dereferences a pointer and returns the value it points to.
// If the pointer is nil, it returns the zero value for the type.
//
// This function is useful when you want to safely dereference a pointer
// without checking for nil, or when working with APIs that return pointers.
//
// Example:
//
//	type Config struct {
//	    Debug *bool
//	}
//
//	cfg := Config{
//	    Debug: Ptr(true),
//	}
//
//	debug := Deref(cfg.Debug) // debug is true
//
//	var nilPtr *bool
//	defaultValue := Deref(nilPtr) // defaultValue is false
func Deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
