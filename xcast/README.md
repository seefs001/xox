# xcast

xcast is a Go package that provides type conversion utilities for various data types. It offers a set of functions to convert between strings, integers, floats, booleans, maps, slices, and structs.

## Installation

To install xcast, use `go get`:

```bash
go get github.com/seefs001/xox/xcast
```

## Usage

Import the package in your Go code:

```go
import "github.com/seefs001/xox/xcast"
```

## Functions

### ToString

Converts various types to a string.

```go
func ToString(value any) (string, error)
```

Example:
```go
str, err := xcast.ToString(123)
if err != nil {
    // handle error
}
fmt.Println(str) // Output: "123"

// Converting a slice
slice := []int{1, 2, 3}
str, err = xcast.ToString(slice)
if err != nil {
    // handle error
}
fmt.Println(str) // Output: "1,2,3"
```

### ToInt

Converts various types to an int.

```go
func ToInt(value any) (int, error)
```

Example:
```go
num, err := xcast.ToInt("123")
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 123

// Converting a boolean
num, err = xcast.ToInt(true)
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 1
```

### ToInt32

Converts various types to an int32.

```go
func ToInt32(value any) (int32, error)
```

Example:
```go
num, err := xcast.ToInt32(123.45)
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 123
```

### ToInt64

Converts various types to an int64.

```go
func ToInt64(value any) (int64, error)
```

Example:
```go
num, err := xcast.ToInt64("9223372036854775807")
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 9223372036854775807
```

### ToFloat64

Converts various types to a float64.

```go
func ToFloat64(value any) (float64, error)
```

Example:
```go
num, err := xcast.ToFloat64("123.45")
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 123.45

// Converting an integer
num, err = xcast.ToFloat64(123)
if err != nil {
    // handle error
}
fmt.Println(num) // Output: 123.0
```

### ToBool

Converts various types to a bool.

```go
func ToBool(value any) (bool, error)
```

Example:
```go
b, err := xcast.ToBool("true")
if err != nil {
    // handle error
}
fmt.Println(b) // Output: true

// Converting an integer
b, err = xcast.ToBool(1)
if err != nil {
    // handle error
}
fmt.Println(b) // Output: true
```

### ToMap

Converts various types to a map[string]any.

```go
func ToMap(value any) (map[string]any, error)
```

Example:
```go
type Person struct {
    Name string
    Age  int
}

person := Person{Name: "John", Age: 30}
m, err := xcast.ToMap(person)
if err != nil {
    // handle error
}
fmt.Println(m) // Output: map[Name:John Age:30]

// Converting a slice
slice := []int{1, 2, 3}
m, err = xcast.ToMap(slice)
if err != nil {
    // handle error
}
fmt.Println(m) // Output: map[0:1 1:2 2:3]
```

### ToSlice

Converts various types to a []any.

```go
func ToSlice(value any) ([]any, error)
```

Example:
```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
s, err := xcast.ToSlice(m)
if err != nil {
    // handle error
}
fmt.Println(s) // Output: [1 2 3]

// Converting a struct
type Person struct {
    Name string
    Age  int
}
person := Person{Name: "John", Age: 30}
s, err = xcast.ToSlice(person)
if err != nil {
    // handle error
}
fmt.Println(s) // Output: [John 30]
```

### ConvertStruct

Converts one struct to another, matching fields by name (case-insensitive).

```go
func ConvertStruct(src any, dst any) error
```

Example:
```go
type Source struct {
    Name string
    Age  int
}
type Destination struct {
    Name string
    Age  int
}

src := Source{Name: "John", Age: 30}
var dst Destination
err := xcast.ConvertStruct(src, &dst)
if err != nil {
    // handle error
}
fmt.Printf("%+v\n", dst) // Output: {Name:John Age:30}
```

### StringToStruct

Converts a string to a struct of type T using JSON unmarshaling.

```go
func StringToStruct[T any](s string) (T, error)
```

Example:
```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

jsonStr := `{"name":"Alice","age":30}`
person, err := xcast.StringToStruct[Person](jsonStr)
if err != nil {
    // handle error
}
fmt.Printf("%+v\n", person) // Output: {Name:Alice Age:30}
```

### StructToString

Converts a struct of type T to a string using JSON marshaling.

```go
func StructToString[T any](v T) (string, error)
```

Example:
```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

person := Person{Name: "Bob", Age: 25}
jsonStr, err := xcast.StructToString(person)
if err != nil {
    // handle error
}
fmt.Println(jsonStr) // Output: {"name":"Bob","age":25}
```

## Error Handling

All functions in this package return an error as the second return value. Always check for errors when using these functions to ensure proper error handling in your application.

## Note

This package uses reflection and type assertions internally. Be aware that this might have performance implications for large-scale operations or in performance-critical parts of your application.

## Contributing

Contributions to the xcast package are welcome! Please submit issues and pull requests on the GitHub repository.

## License

This package is released under the MIT License. See the LICENSE file for details.
