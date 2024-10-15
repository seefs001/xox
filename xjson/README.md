# xjson

xjson is a powerful and flexible JSON manipulation package for Go, providing easy-to-use functions for parsing, querying, and transforming JSON data.

## Features

- Parse JSON strings into Go structures
- Query JSON data using dot notation and array indexing
- Type-safe value retrieval (string, int, float, bool, array)
- Support for nested objects and arrays
- Functional programming utilities (ForEach, Map, Filter, Reduce)
- Error handling for invalid paths or type mismatches
- JSON Schema generation for Go structs

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

### Generating JSON Schema

```go
type Person struct {
    Name string `json:"name" description:"The person's name"`
    Age  int    `json:"age" description:"The person's age"`
}

schema, err := xjson.GenerateJSONSchema(Person{})
if err != nil {
    // Handle error
}
fmt.Printf("JSON Schema: %+v\n", schema)
```

## API Reference

### Parsing

- `ParseJSON(jsonStr string) (JSONObject, error)`
  Parses a JSON string into a JSONObject.

### Querying

- `Get(data JSONObject, path JSONPath) (interface{}, error)`
  Retrieves a value from a JSON object using a JSON path.

- `GetFromString(jsonStr string, path JSONPath) (interface{}, error)`
  Retrieves a value from a JSON string using a JSON path.

- `GetString(data JSONObject, path JSONPath) (string, error)`
  Retrieves a string value from a JSON object using a JSON path.

- `GetStringFromString(jsonStr string, path JSONPath) (string, error)`
  Retrieves a string value from a JSON string using a JSON path.

- `GetInt(data JSONObject, path JSONPath) (int, error)`
  Retrieves an integer value from a JSON object using a JSON path.

- `GetIntFromString(jsonStr string, path JSONPath) (int, error)`
  Retrieves an integer value from a JSON string using a JSON path.

- `GetFloat(data JSONObject, path JSONPath) (float64, error)`
  Retrieves a float value from a JSON object using a JSON path.

- `GetFloatFromString(jsonStr string, path JSONPath) (float64, error)`
  Retrieves a float value from a JSON string using a JSON path.

- `GetBool(data JSONObject, path JSONPath) (bool, error)`
  Retrieves a boolean value from a JSON object using a JSON path.

- `GetBoolFromString(jsonStr string, path JSONPath) (bool, error)`
  Retrieves a boolean value from a JSON string using a JSON path.

- `GetArray(data JSONObject, path JSONPath) (JSONArray, error)`
  Retrieves an array value from a JSON object using a JSON path.

- `GetArrayFromString(jsonStr string, path JSONPath) (JSONArray, error)`
  Retrieves an array value from a JSON string using a JSON path.

### Functional Programming

- `ForEach(data interface{}, fn func(key interface{}, value interface{}) error) error`
  Applies a function to each element in an array or object.

- `Map(data interface{}, fn func(key interface{}, value interface{}) (interface{}, error)) (interface{}, error)`
  Applies a function to each element in an array or object and returns a new array or object.

- `Filter(data interface{}, fn func(key interface{}, value interface{}) (bool, error)) (interface{}, error)`
  Returns a new array or object with elements that pass the test implemented by the provided function.

- `Reduce(data interface{}, fn func(accumulator, key, value interface{}) (interface{}, error), initialValue interface{}) (interface{}, error)`
  Applies a function against an accumulator and each element in the array or object to reduce it to a single value.

### JSON Schema Generation

- `GenerateJSONSchema(v interface{}) (map[string]interface{}, error)`
  Generates a JSON schema for the given struct.

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

## Contributing

Contributions to xjson are welcome! Please submit issues and pull requests on the GitHub repository.

## License

xjson is released under the MIT License. See the LICENSE file for details.
