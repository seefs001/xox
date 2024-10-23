package xjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// JSONPath represents a JSON path query
type JSONPath string

// JSONObject represents a JSON object
type JSONObject map[string]interface{}

// JSONArray represents a JSON array
type JSONArray []interface{}

// Get retrieves a value from a JSON object using a JSON path
func Get(data JSONObject, path JSONPath) (interface{}, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return nil, err
	}

	// Convert json.Number to float64 for consistency
	if num, ok := value.(json.Number); ok {
		return num.Float64()
	}
	return value, nil
}

// GetFromString retrieves a value from a JSON string using a JSON path
func GetFromString(jsonStr string, path JSONPath) (interface{}, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON string")
	}
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("parse JSON error: %w", err)
	}
	return Get(data, path)
}

// GetString retrieves a string value from a JSON object using a JSON path
func GetString(data JSONObject, path JSONPath) (string, error) {
	return getString(data, path)
}

// GetStringFromString retrieves a string value from a JSON string using a JSON path
func GetStringFromString(jsonStr string, path JSONPath) (string, error) {
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return "", err
	}
	return getString(data, path)
}

// GetInt retrieves an integer value from a JSON object using a JSON path
func GetInt(data JSONObject, path JSONPath) (int, error) {
	return getInt(data, path)
}

// GetIntFromString retrieves an integer value from a JSON string using a JSON path
func GetIntFromString(jsonStr string, path JSONPath) (int, error) {
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return 0, err
	}
	return getInt(data, path)
}

// GetFloat retrieves a float value from a JSON object using a JSON path
func GetFloat(data JSONObject, path JSONPath) (float64, error) {
	return getFloat(data, path)
}

// GetFloatFromString retrieves a float value from a JSON string using a JSON path
func GetFloatFromString(jsonStr string, path JSONPath) (float64, error) {
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return 0, err
	}
	return getFloat(data, path)
}

// GetBool retrieves a boolean value from a JSON object using a JSON path
func GetBool(data JSONObject, path JSONPath) (bool, error) {
	return getBool(data, path)
}

// GetBoolFromString retrieves a boolean value from a JSON string using a JSON path
func GetBoolFromString(jsonStr string, path JSONPath) (bool, error) {
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return false, err
	}
	return getBool(data, path)
}

// GetArray retrieves an array value from a JSON object using a JSON path
func GetArray(data JSONObject, path JSONPath) (JSONArray, error) {
	return getArray(data, path)
}

// GetArrayFromString retrieves an array value from a JSON string using a JSON path
func GetArrayFromString(jsonStr string, path JSONPath) (JSONArray, error) {
	data, err := ParseJSON(jsonStr)
	if err != nil {
		return nil, err
	}
	return getArray(data, path)
}

// ForEach applies a function to each element in an array or object
func ForEach(data interface{}, fn func(key interface{}, value interface{}) error) error {
	switch v := data.(type) {
	case JSONObject:
		for key, value := range v {
			if err := fn(key, value); err != nil {
				return err
			}
		}
	case JSONArray:
		for i, value := range v {
			if err := fn(i, value); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, value := range v {
			if err := fn(i, value); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("ForEach can only be applied to JSONObject, JSONArray, or []interface{}")
	}
	return nil
}

// Map applies a function to each element in an array or object and returns a new array or object
func Map(data interface{}, fn func(key interface{}, value interface{}) (interface{}, error)) (interface{}, error) {
	switch v := data.(type) {
	case JSONObject:
		result := make(JSONObject)
		for key, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			mappedValue, err := fn(key, value)
			if err != nil {
				return nil, err
			}
			result[key] = mappedValue
		}
		return result, nil
	case JSONArray:
		result := make(JSONArray, len(v))
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			mappedValue, err := fn(i, value)
			if err != nil {
				return nil, err
			}
			result[i] = mappedValue
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			mappedValue, err := fn(i, value)
			if err != nil {
				return nil, err
			}
			result[i] = mappedValue
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Map can only be applied to JSONObject, JSONArray, or []interface{}")
	}
}

// Filter returns a new array or object with elements that pass the test implemented by the provided function
func Filter(data interface{}, fn func(key interface{}, value interface{}) (bool, error)) (interface{}, error) {
	switch v := data.(type) {
	case JSONObject:
		result := make(JSONObject)
		for key, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			include, err := fn(key, value)
			if err != nil {
				return nil, err
			}
			if include {
				result[key] = value
			}
		}
		return result, nil
	case JSONArray:
		result := make(JSONArray, 0)
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			include, err := fn(i, value)
			if err != nil {
				return nil, err
			}
			if include {
				result = append(result, value)
			}
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, 0)
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			include, err := fn(i, value)
			if err != nil {
				return nil, err
			}
			if include {
				result = append(result, value)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Filter can only be applied to JSONObject, JSONArray, or []interface{}")
	}
}

// Reduce applies a function against an accumulator and each element in the array/object
func Reduce(data interface{}, fn func(accumulator, key, value interface{}) (interface{}, error), initial interface{}) (interface{}, error) {
	accumulator := initial

	switch v := data.(type) {
	case JSONObject:
		for key, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			result, err := fn(accumulator, key, value)
			if err != nil {
				return nil, err
			}
			accumulator = result
		}
	case JSONArray:
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			result, err := fn(accumulator, i, value)
			if err != nil {
				return nil, err
			}
			accumulator = result
		}
	case []interface{}:
		for i, value := range v {
			if num, ok := value.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				value = f
			}
			result, err := fn(accumulator, i, value)
			if err != nil {
				return nil, err
			}
			accumulator = result
		}
	default:
		return nil, fmt.Errorf("Reduce can only be applied to JSONObject, JSONArray, or []interface{}")
	}

	return accumulator, nil
}

// getPath is a helper function to traverse the JSON object using the path
func getPath(data interface{}, path string) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("nil data input")
	}
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("null value at path segment: %s", part)
		}

		if strings.HasSuffix(part, "]") {
			arrayParts := strings.Split(strings.TrimSpace(part[:len(part)-1]), "[")
			if len(arrayParts) != 2 {
				return nil, fmt.Errorf("invalid array access syntax: %s", part)
			}

			key := strings.TrimSpace(arrayParts[0])
			index, err := strconv.Atoi(strings.TrimSpace(arrayParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", arrayParts[1])
			}

			if key != "" {
				var ok bool
				switch v := current.(type) {
				case JSONObject:
					current, ok = v[key]
				case map[string]interface{}:
					current, ok = v[key]
				default:
					return nil, fmt.Errorf("cannot navigate through non-object type at: %s", key)
				}
				if !ok {
					return nil, fmt.Errorf("key not found: %s", key)
				}
			}

			switch v := current.(type) {
			case []interface{}:
				if index < 0 || index >= len(v) {
					return nil, fmt.Errorf("array index out of bounds: %d, length: %d", index, len(v))
				}
				if num, ok := v[index].(json.Number); ok {
					f, err := num.Float64()
					if err != nil {
						return nil, err
					}
					current = f
				} else {
					current = v[index]
				}
			case JSONArray:
				if index < 0 || index >= len(v) {
					return nil, fmt.Errorf("array index out of bounds: %d, length: %d", index, len(v))
				}
				if num, ok := v[index].(json.Number); ok {
					f, err := num.Float64()
					if err != nil {
						return nil, err
					}
					current = f
				} else {
					current = v[index]
				}
			default:
				return nil, fmt.Errorf("cannot access index on non-array type at: %s", part)
			}
		} else {
			var ok bool
			switch v := current.(type) {
			case JSONObject:
				current, ok = v[part]
			case map[string]interface{}:
				current, ok = v[part]
			default:
				return nil, fmt.Errorf("cannot access property on non-object type at: %s", part)
			}
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
		}
	}

	return current, nil
}

// ParseJSON parses a JSON string into a JSONObject
func ParseJSON(jsonStr string) (JSONObject, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON string")
	}
	var result JSONObject
	d := json.NewDecoder(strings.NewReader(jsonStr))
	d.UseNumber()
	if err := d.Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}

// Helper functions for type-specific retrieval

func getString(data JSONObject, path JSONPath) (string, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return "", err
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case int:
		return strconv.Itoa(v), nil
	case nil:
		return "", fmt.Errorf("value at path %s is null", path)
	default:
		return "", fmt.Errorf("value at path %s is not a string", path)
	}
}

func getInt(data JSONObject, path JSONPath) (int, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return 0, err
	}
	switch v := value.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return int(i), nil
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("value at path %s is not an integer", path)
	}
}

func getFloat(data JSONObject, path JSONPath) (float64, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return 0, err
	}
	switch v := value.(type) {
	case json.Number:
		return v.Float64()
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("value at path %s is not a float", path)
	}
}

func getBool(data JSONObject, path JSONPath) (bool, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return false, err
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("value at path %s is not a boolean", path)
	}
}

func getArray(data JSONObject, path JSONPath) (JSONArray, error) {
	value, err := getPath(data, string(path))
	if err != nil {
		return nil, err
	}
	switch arr := value.(type) {
	case JSONArray:
		result := make(JSONArray, len(arr))
		for i, v := range arr {
			if num, ok := v.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				result[i] = f
			} else {
				result[i] = v
			}
		}
		return result, nil
	case []interface{}:
		result := make(JSONArray, len(arr))
		for i, v := range arr {
			if num, ok := v.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return nil, err
				}
				result[i] = f
			} else {
				result[i] = v
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value at path %s is not an array", path)
	}
}

// GenerateJSONSchema generates a JSON schema for the given struct
func GenerateJSONSchema(v interface{}) (map[string]interface{}, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		fieldSchema := map[string]interface{}{}
		switch field.Type.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			fieldSchema["type"] = "number"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		}

		if description := field.Tag.Get("description"); description != "" {
			fieldSchema["description"] = description
		}

		schema["properties"].(map[string]interface{})[fieldName] = fieldSchema

		if !strings.Contains(jsonTag, "omitempty") {
			schema["required"] = append(schema["required"].([]string), fieldName)
		}
	}

	return schema, nil
}

// MustGet retrieves a value and panics on error
func MustGet(data JSONObject, path JSONPath) interface{} {
	v, err := Get(data, path)
	if err != nil {
		panic(err)
	}
	return v
}

// ToJSON converts any value to a JSON string
func ToJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}
	return string(b), nil
}

// MustToJSON converts any value to a JSON string and panics on error
func MustToJSON(v interface{}) string {
	s, err := ToJSON(v)
	if err != nil {
		panic(err)
	}
	return s
}
