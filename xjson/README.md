# xjson

xjson is a powerful and flexible JSON manipulation package for Go, providing easy-to-use functions for parsing, querying, and transforming JSON data.

## Features

- Parse JSON strings into Go structures
- Query JSON data using dot notation and array indexing
- Type-safe value retrieval (string, int, float, bool, array)
- Support for nested objects and arrays
- Functional programming utilities (ForEach, Map, Filter, Reduce)
- Error handling for invalid paths or type mismatches

## Installation

```bash
go get github.com/seefs001/xox/xjson
```

## Usage

### Parsing JSON

```go
jsonStr := `{"name": "John", "age": 30, "city": "New York"}`
data, err := xjson.ParseJSON(jsonStr)
if err != nil {
    // Handle error
}
```

### Querying JSON

```go
// Get a string value
name, err := xjson.GetString(data, "name")

// Get an integer value
age, err := xjson.GetInt(data, "age")

// Get a nested value
street, err := xjson.GetString(data, "address.street")

// Get an array element
grade, err := xjson.GetFloat(data, "grades[1]")
```

### Working with JSON strings directly

```go
jsonStr := `{"user": {"name": "Alice", "age": 25}}`

// Get a value from a JSON string
name, err := xjson.GetStringFromString(jsonStr, "user.name")

// Get an integer from a JSON string
age, err := xjson.GetIntFromString(jsonStr, "user.age")
```

### Functional Programming Utilities

```go
// ForEach
err := xjson.ForEach(data["items"], func(key interface{}, value interface{}) error {
    fmt.Printf("Key: %v, Value: %v\n", key, value)
    return nil
})

// Map
result, err := xjson.Map(data["numbers"], func(key interface{}, value interface{}) (interface{}, error) {
    return value.(float64) * 2, nil
})

// Filter
result, err := xjson.Filter(data["users"], func(key interface{}, value interface{}) (bool, error) {
    user := value.(map[string]interface{})
    return user["age"].(float64) >= 18, nil
})

// Reduce
sum, err := xjson.Reduce(data["numbers"], func(accumulator, key, value interface{}) (interface{}, error) {
    return accumulator.(float64) + value.(float64), nil
}, 0.0)
```

## API Reference

### Parsing

- `ParseJSON(jsonStr string) (JSONObject, error)`

### Querying

- `Get(data JSONObject, path JSONPath) (interface{}, error)`
- `GetFromString(jsonStr string, path JSONPath) (interface{}, error)`
- `GetString(data JSONObject, path JSONPath) (string, error)`
- `GetStringFromString(jsonStr string, path JSONPath) (string, error)`
- `GetInt(data JSONObject, path JSONPath) (int, error)`
- `GetIntFromString(jsonStr string, path JSONPath) (int, error)`
- `GetFloat(data JSONObject, path JSONPath) (float64, error)`
- `GetFloatFromString(jsonStr string, path JSONPath) (float64, error)`
- `GetBool(data JSONObject, path JSONPath) (bool, error)`
- `GetBoolFromString(jsonStr string, path JSONPath) (bool, error)`
- `GetArray(data JSONObject, path JSONPath) (JSONArray, error)`
- `GetArrayFromString(jsonStr string, path JSONPath) (JSONArray, error)`

### Functional Programming

- `ForEach(data interface{}, fn func(key interface{}, value interface{}) error) error`
- `Map(data interface{}, fn func(key interface{}, value interface{}) (interface{}, error)) (interface{}, error)`
- `Filter(data interface{}, fn func(key interface{}, value interface{}) (bool, error)) (interface{}, error)`
- `Reduce(data interface{}, fn func(accumulator, key, value interface{}) (interface{}, error), initialValue interface{}) (interface{}, error)`

## Error Handling

All functions in the xjson package return errors when encountering issues such as:

- Invalid JSON strings
- Non-existent paths
- Type mismatches
- Array index out of bounds

Always check for errors and handle them appropriately in your code.

## Performance Considerations

The xjson package uses reflection and type assertions, which may impact performance for very large JSON structures or high-frequency operations. For performance-critical applications, consider using code generation or other specialized JSON libraries.

## Thread Safety

The xjson package does not provide built-in concurrency protection. When using xjson in concurrent environments, ensure proper synchronization is implemented at the application level.
