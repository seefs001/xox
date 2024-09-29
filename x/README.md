# X Package

The `x` package provides a collection of utility functions and types to enhance Go programming productivity. It includes various helper functions for error handling, functional programming, string manipulation, and more.

## Installation

```bash
go get github.com/seefs001/xox/x
```

## Usage

Import the package in your Go code:

```go
import "github.com/seefs001/xox/x"
```

## API Reference

### Error Handling

#### OnlyErr

```go
func OnlyErr(values ...interface{}) error
```

Returns only the error from a function that returns multiple values, where the last value is an error. It discards all other return values.

Example:
```go
err := x.OnlyErr(SomeFunction())
if err != nil {
    // handle error
}
```

#### Must1, Must2, Must3, Must4

```go
func Must1[T any](v T, err error) T
func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2)
func Must3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3)
func Must4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, err error) (T1, T2, T3, T4)
```

These functions return the result(s) if no error occurred. If an error occurred, they panic with the error.

Example:
```go
result := x.Must1(SomeFunction())
v1, v2 := x.Must2(SomeFunctionReturningTwoValues())
```

#### Must0

```go
func Must0(err error)
```

Panics if the error is not nil.

Example:
```go
x.Must0(SomeFunctionReturningOnlyError())
```

#### Ignore1, Ignore2, Ignore3, Ignore4

```go
func Ignore1[T any](v T, _ error) T
func Ignore2[T1, T2 any](v1 T1, v2 T2, _ error) (T1, T2)
func Ignore3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, _ error) (T1, T2, T3)
func Ignore4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4, _ error) (T1, T2, T3, T4)
```

These functions return the result(s), ignoring any error.

Example:
```go
result := x.Ignore1(SomeFunction())
v1, v2 := x.Ignore2(SomeFunctionReturningTwoValues())
```

#### Ignore0

```go
func Ignore0(_ error)
```

Ignores the error.

Example:
```go
x.Ignore0(SomeFunctionReturningOnlyError())
```

### Functional Programming

#### Where

```go
func Where[T any](collection []T, predicate func(T) bool) []T
```

Returns a new slice containing all elements of the collection that satisfy the predicate.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
evenNumbers := x.Where(numbers, func(n int) bool { return n%2 == 0 })
// evenNumbers will be [2, 4]
```

#### Select

```go
func Select[T any, U any](collection []T, selector func(T) U) []U
```

Returns a new slice containing the results of applying the function to each element of the original slice.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
squares := x.Select(numbers, func(n int) int { return n * n })
// squares will be [1, 4, 9, 16, 25]
```

#### Aggregate

```go
func Aggregate[T any, U any](collection []T, seed U, accumulator func(U, T) U) U
```

Reduces collection to a single value using a reduction function.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
sum := x.Aggregate(numbers, 0, func(acc, n int) int { return acc + n })
// sum will be 15
```

#### ForEach

```go
func ForEach[T any](collection []T, action func(T))
```

Iterates over the collection and applies the action to each element.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
x.ForEach(numbers, func(n int) { fmt.Println(n) })
// This will print each number on a new line
```

#### Range

```go
func Range[T int](start T, count int) []T
```

Generates a sequence of integers.

Example:
```go
numbers := x.Range(1, 5)
// numbers will be [1, 2, 3, 4, 5]
```

#### Count

```go
func Count[T any](collection []T, predicate func(T) bool) int
```

Returns the number of elements in the collection that satisfy the predicate.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
evenCount := x.Count(numbers, func(n int) bool { return n%2 == 0 })
// evenCount will be 2
```

#### GroupBy

```go
func GroupBy[T any, K comparable](collection []T, keySelector func(T) K) map[K][]T
```

Groups elements of the collection by keys returned by the keySelector.

Example:
```go
type Person struct {
    Name string
    Age  int
}
people := []Person{
    {"Alice", 25},
    {"Bob", 30},
    {"Charlie", 25},
}
groupedByAge := x.GroupBy(people, func(p Person) int { return p.Age })
// groupedByAge will be map[int][]Person{
//   25: {{"Alice", 25}, {"Charlie", 25}},
//   30: {{"Bob", 30}},
// }
```

#### First

```go
func First[T any](collection []T, predicate func(T) bool) (T, bool)
```

Returns the first element of the collection that satisfies the predicate.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
firstEven, found := x.First(numbers, func(n int) bool { return n%2 == 0 })
// firstEven will be 2, found will be true
```

#### Last

```go
func Last[T any](collection []T, predicate func(T) bool) (T, bool)
```

Returns the last element of the collection that satisfies the predicate.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
lastEven, found := x.Last(numbers, func(n int) bool { return n%2 == 0 })
// lastEven will be 4, found will be true
```

#### Any

```go
func Any[T any](collection []T, predicate func(T) bool) bool
```

Returns true if any element in the collection satisfies the predicate.

Example:
```go
numbers := []int{1, 2, 3, 4, 5}
hasEven := x.Any(numbers, func(n int) bool { return n%2 == 0 })
// hasEven will be true
```

#### All

```go
func All[T any](collection []T, predicate func(T) bool) bool
```

Returns true if all elements in the collection satisfy the predicate.

Example:
```go
numbers := []int{2, 4, 6, 8, 10}
allEven := x.All(numbers, func(n int) bool { return n%2 == 0 })
// allEven will be true
```

### String Manipulation

#### RandomString

```go
func RandomString(length int, mode RandomStringMode) (string, error)
```

Generates a random string of the specified length using the given mode.

Example:
```go
randomStr, err := x.RandomString(10, x.ModeAlphanumeric)
if err != nil {
    log.Fatal(err)
}
fmt.Println(randomStr) // Outputs a random 10-character alphanumeric string
```

#### RandomStringCustom

```go
func RandomStringCustom(length int, charset string) (string, error)
```

Generates a random string of the specified length using a custom character set.

Example:
```go
customCharset := "ABC123"
randomStr, err := x.RandomStringCustom(5, customCharset)
if err != nil {
    log.Fatal(err)
}
fmt.Println(randomStr) // Outputs a random 5-character string using only 'A', 'B', 'C', '1', '2', '3'
```

#### IsImageURL

```go
func IsImageURL(s string) bool
```

Checks if a string is a valid image URL.

Example:
```go
fmt.Println(x.IsImageURL("https://example.com/image.jpg"))  // true
fmt.Println(x.IsImageURL("https://example.com/file.txt"))   // false
```

#### IsBase64

```go
func IsBase64(s string) bool
```

Checks if a string is a valid base64 encoded string.

Example:
```go
fmt.Println(x.IsBase64("SGVsbG8gV29ybGQ="))  // true
fmt.Println(x.IsBase64("Not base64"))        // false
```

#### TrimSuffixes

```go
func TrimSuffixes(s string, suffixes ...string) string
```

Removes all specified suffixes from the given string.

Example:
```go
input := "example.tar.gz"
trimmed := x.TrimSuffixes(input, ".gz", ".tar")
fmt.Println(trimmed) // Output: "example"
```

### Data Structures

#### Tuple

```go
type Tuple[T1, T2 any] struct {
    First  T1
    Second T2
}

func NewTuple[T1, T2 any](first T1, second T2) Tuple[T1, T2]
func (t Tuple[T1, T2]) Unpack() (T1, T2)
func (t Tuple[T1, T2]) Swap() Tuple[T2, T1]
func (t Tuple[T1, T2]) Map(f func(T1, T2) (T1, T2)) Tuple[T1, T2]
```

Represents a generic tuple type with two elements.

Example:
```go
t := x.NewTuple("hello", 42)
fmt.Println(t.First)  // Outputs: hello
fmt.Println(t.Second) // Outputs: 42

str, num := t.Unpack()
fmt.Println(str) // Outputs: hello
fmt.Println(num) // Outputs: 42

swapped := t.Swap()
fmt.Println(swapped.First)  // Outputs: 42
fmt.Println(swapped.Second) // Outputs: hello

doubled := t.Map(func(a string, b int) (string, int) { return a + a, b * 2 })
fmt.Println(doubled.First)  // Outputs: hellohello
fmt.Println(doubled.Second) // Outputs: 84
```

#### Triple

```go
type Triple[T1, T2, T3 any] struct {
    First  T1
    Second T2
    Third  T3
}

func NewTriple[T1, T2, T3 any](first T1, second T2, third T3) Triple[T1, T2, T3]
func (t Triple[T1, T2, T3]) Unpack() (T1, T2, T3)
func (t Triple[T1, T2, T3]) Rotate() Triple[T2, T3, T1]
func (t Triple[T1, T2, T3]) Map(f func(T1, T2, T3) (T1, T2, T3)) Triple[T1, T2, T3]
```

Represents a generic tuple type with three elements.

Example:
```go
t := x.NewTriple("hello", 42, true)
fmt.Println(t.First)  // Outputs: hello
fmt.Println(t.Second) // Outputs: 42
fmt.Println(t.Third)  // Outputs: true

str, num, bool := t.Unpack()
fmt.Println(str)  // Outputs: hello
fmt.Println(num)  // Outputs: 42
fmt.Println(bool) // Outputs: true

rotated := t.Rotate()
fmt.Println(rotated.First)  // Outputs: 42
fmt.Println(rotated.Second) // Outputs: true
fmt.Println(rotated.Third)  // Outputs: hello

transformed := t.Map(func(a string, b int, c bool) (string, int, bool) {
    return a + "world", b * 2, !c
})
fmt.Println(transformed.First)  // Outputs: helloworld
fmt.Println(transformed.Second) // Outputs: 84
fmt.Println(transformed.Third)  // Outputs: false
```

### Concurrency

#### SafePool

```go
type SafePool struct {}

func NewSafePool(size int) *SafePool
func (p *SafePool) SafeGo(f func()) <-chan error
func (p *SafePool) SafeGoWithContext(ctx context.Context, f func()) <-chan error
func (p *SafePool) SafeGoNoError(f func())
func (p *SafePool) SafeGoWithContextNoError(ctx context.Context, f func())
```

SafePool represents a pool of goroutines with panic recovery.

Example:
```go
pool := x.NewSafePool(5) // Creates a pool with a maximum of 5 concurrent workers

errChan := pool.SafeGo(func() {
    // Some potentially panicking code
    panic("Something went wrong")
})

if err := <-errChan; err != nil {
    fmt.Printf("Error occurred: %v\n", err)
}
// Output: Error occurred: panic recovered: Something went wrong
```

#### AsyncTask

```go
type AsyncTask[T any] struct {}

func NewAsyncTask[T any](f func() (T, error)) *AsyncTask[T]
func (t *AsyncTask[T]) Wait() (T, error)
func (t *AsyncTask[T]) WaitWithTimeout(timeout time.Duration) (T, error, bool)
func (t *AsyncTask[T]) WaitWithContext(ctx context.Context) (T, error, bool)
```

AsyncTask represents an asynchronous task that can be executed and waited upon.

Example:
```go
task := x.NewAsyncTask(func() (int, error) {
    time.Sleep(2 * time.Second)
    return 42, nil
})

result, err := task.Wait()
fmt.Printf("Result: %d, Error: %v\n", result, err)
// Output: Result: 42, Error: <nil>
```

#### WaitGroup

```go
type WaitGroup struct {}

func NewWaitGroup() *WaitGroup
func (wg *WaitGroup) Go(f func() error)
func (wg *WaitGroup) GoWithContext(ctx context.Context, f func(context.Context) error)
func (wg *WaitGroup) Wait() error
func (wg *WaitGroup) WaitWithTimeout(timeout time.Duration) (error, bool)
```

WaitGroup is a safer version of sync.WaitGroup with panic recovery and error handling.

Example:
```go
wg := x.NewWaitGroup()
wg.Go(func() error {
    // Some work here
    return nil
})
wg.Go(func() error {
    return errors.New("an error occurred")
})
if err := wg.Wait(); err != nil {
    fmt.Printf("Error occurred: %v\n", err)
}
// Output: Error occurred: an error occurred
```

### Miscellaneous

#### ToJSON

```go
func ToJSON(v interface{}) (string, error)
```

Converts a struct or map to a JSON string.

Example:
```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
p := Person{Name: "John", Age: 30}
jsonStr, err := x.ToJSON(p)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("JSON: %s\n", jsonStr)
}
// Output: JSON: {"name":"John","age":30}
```

#### MustToJSON

```go
func MustToJSON(v interface{}) string
```

Converts a struct or map to a JSON string, panicking on error.

Example:
```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
p := Person{Name: "John", Age: 30}
jsonStr := x.MustToJSON(p)
fmt.Printf("JSON: %s\n", jsonStr)
// Output: JSON: {"name":"John","age":30}
```

#### Ternary

```go
func Ternary[T any](condition bool, ifTrue, ifFalse T) T
```

A generic function that implements the ternary operator.

Example:
```go
result := x.Ternary(1 > 0, "Greater", "Less")
fmt.Println(result)
// Output: Greater
```

#### TernaryF

```go
func TernaryF[T any](condition bool, ifTrue, ifFalse func() T) T
```

A generic function that implements the ternary operator with lazy evaluation.

Example:
```go
result := x.TernaryF(1 > 0, func() string { return "Greater" }, func() string { return "Less" })
fmt.Println(result)
// Output: Greater
```

#### If

```go
func If[T any](condition bool, then T) *ifChain[T]
```

A generic function that implements an if-else chain.

Example:
```go
result := x.If(x > 0, "Positive").
    ElseIf(x < 0, "Negative").
    Else("Zero")
fmt.Println(result)
```

#### Switch

```go
func Switch[T comparable, R any](value T) *switchChain[T, R]
```

A generic function that implements a switch statement.

Example:
```go
result := x.Switch[string, int]("b").
    Case("a", 1).
    Case("b", 2).
    Default(3)
fmt.Println(result)
// Output: 2
```

#### IsEmpty, IsNil, IsZero

```go
func IsEmpty[T any](value T) bool
func IsNil(value interface{}) bool
func IsZero[T any](value T) bool
```

These functions check if a value is considered empty, nil, or zero, respectively.

Example:
```go
fmt.Println(x.IsEmpty(""))        // true
fmt.Println(x.IsNil((*int)(nil))) // true
fmt.Println(x.IsZero(0))          // true
```

#### SetNonZeroValues, SetNonZeroValuesWithKeys

```go
func SetNonZeroValues(dst map[string]interface{}, src map[string]interface{})
func SetNonZeroValuesWithKeys(dst map[string]interface{}, src map[string]interface{}, keys ...string)
```

These functions set non-zero values from src to dst.

Example:
```go
dst := map[string]interface{}{"a": 1, "b": ""}
src := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
x.SetNonZeroValues(dst, src)
fmt.Println(dst) // Output: map[a:2 b:hello c:3]

dst = map[string]interface{}{"a": 1, "b": ""}
x.SetNonZeroValuesWithKeys(dst, src, "a", "c")
fmt.Println(dst) // Output: map[a:2 b: c:3]
```

#### MapKeys, MapValues

```go
func MapKeys[K comparable, V any](m map[K]V) []K
func MapValues[K comparable, V any](m map[K]V) []V
```

These functions return a slice of keys or values from a map.

Example:
```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
keys := x.MapKeys(m)
values := x.MapValues(m)
fmt.Println(keys)   // Output: [a b c] (order may vary)
fmt.Println(values) // Output: [1 2 3] (order may vary)
```

#### Shuffle

```go
func Shuffle[T any](s []T)
```

Randomly shuffles the elements in a slice.

Example:
```go
s := []int{1, 2, 3, 4, 5}
x.Shuffle(s)
fmt.Println(s) // Output: [3 1 5 2 4] (order will be random)
```

#### FlattenMap

```go
func FlattenMap(data map[string]any, prefix string) map[string]any
```

Flattens a nested map into a single-level map with dot notation keys.

Example:
```go
nested := map[string]any{
    "a": 1,
    "b": map[string]any{
        "c": 2,
        "d": map[string]any{
            "e": 3,
        },
    },
}
flat := x.FlattenMap(nested, "")
fmt.Println(flat) // Output: map[a:1 b.c:2 b.d.e:3]
```

#### CopyMap

```go
func CopyMap[K comparable, V any](m map[K]V) map[K]V
```

Creates a deep copy of the given map.

Example:
```go
original := map[string]int{"a": 1, "b": 2}
copied := x.CopyMap(original)
copied["c"] = 3
fmt.Println(original) // Output: map[a:1 b:2]
fmt.Println(copied)   // Output: map[a:1 b:2 c:3]
```

#### FileExists

```go
func FileExists(filePath string) bool
```

Checks if a file exists.

Example:
```go
exists := x.FileExists("/path/to/file.txt")
fmt.Println(exists) // Output: true or false depending on file existence
```

#### GenerateUUID

```go
func GenerateUUID() (string, error)
```

Generates a new UUID (Universally Unique Identifier) as a string.

Example:
```go
uuid, err := x.GenerateUUID()
if err != nil {
    // handle error
}
fmt.Println(uuid) // Output: e.g., "f47ac10b-58cc-4372-a567-0e02b2c3d479"
```

#### DecodeUnicodeURL, EncodeUnicodeURL

```go
func DecodeUnicodeURL(encodedURL string) (string, error)
func EncodeUnicodeURL(originalURL string) (string, error)
```

These functions decode and encode URLs that contain Unicode characters.

Example:
```go
encoded := "https://example.com/path?q=%E4%BD%A0%E5%A5%BD"
decoded, err := x.DecodeUnicodeURL(encoded)
if err != nil {
    // handle error
}
fmt.Println(decoded) // Output: https://example.com/path?q=你好

original := "https://example.com/path?q=你好"
encoded, err := x.EncodeUnicodeURL(original)
if err != nil {
    // handle error
}
fmt.Println(encoded) // Output: https://example.com/path?q=%E4%BD%A0%E5%A5%BD
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
