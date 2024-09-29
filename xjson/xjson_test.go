package xjson_test

import (
	"fmt"
	"testing"

	"github.com/seefs001/xox/xjson"
	"github.com/seefs001/xox/xlog"
	"github.com/stretchr/testify/assert"
)

func TestXJSON(t *testing.T) {
	jsonStr := `{
		"name": "John Doe",
		"age": 30,
		"isStudent": false,
		"grades": [85, 90, 78],
		"address": {
			"street": "123 Main St",
			"city": "Anytown"
		},
		"contacts": [
			{"type": "email", "value": "john@example.com"},
			{"type": "phone", "value": "555-1234"}
		]
	}`

	t.Run("ParseJSON", func(t *testing.T) {
		data, err := xjson.ParseJSON(jsonStr)
		assert.NoError(t, err)
		assert.IsType(t, xjson.JSONObject{}, data)
		xlog.Info("Successfully parsed JSON", "data", data)
	})

	t.Run("Get", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.Get(data, "name")
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", value)
		xlog.Info("Retrieved value using Get", "key", "name", "value", value)
	})

	t.Run("GetFromString", func(t *testing.T) {
		value, err := xjson.GetFromString(jsonStr, "age")
		assert.NoError(t, err)
		assert.Equal(t, float64(30), value)
		xlog.Info("Retrieved value using GetFromString", "key", "age", "value", value)
	})

	t.Run("GetString", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.GetString(data, "name")
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", value)
		xlog.Info("Retrieved string value", "key", "name", "value", value)
	})

	t.Run("GetStringFromString", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "address.street")
		assert.NoError(t, err)
		assert.Equal(t, "123 Main St", value)
		xlog.Info("Retrieved nested string value", "path", "address.street", "value", value)
	})

	t.Run("GetInt", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.GetInt(data, "age")
		assert.NoError(t, err)
		assert.Equal(t, 30, value)
		xlog.Info("Retrieved integer value", "key", "age", "value", value)
	})

	t.Run("GetIntFromString", func(t *testing.T) {
		value, err := xjson.GetIntFromString(jsonStr, "age")
		assert.NoError(t, err)
		assert.Equal(t, 30, value)
		xlog.Info("Retrieved integer value from string", "key", "age", "value", value)
	})

	t.Run("GetFloat", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.GetFloat(data, "age")
		assert.NoError(t, err)
		assert.Equal(t, float64(30), value)
		xlog.Info("Retrieved float value", "key", "age", "value", value)
	})

	t.Run("GetFloatFromString", func(t *testing.T) {
		value, err := xjson.GetFloatFromString(jsonStr, "age")
		assert.NoError(t, err)
		assert.Equal(t, float64(30), value)
		xlog.Info("Retrieved float value from string", "key", "age", "value", value)
	})

	t.Run("GetBool", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.GetBool(data, "isStudent")
		assert.NoError(t, err)
		assert.Equal(t, false, value)
		xlog.Info("Retrieved boolean value", "key", "isStudent", "value", value)
	})

	t.Run("GetBoolFromString", func(t *testing.T) {
		value, err := xjson.GetBoolFromString(jsonStr, "isStudent")
		assert.NoError(t, err)
		assert.Equal(t, false, value)
		xlog.Info("Retrieved boolean value from string", "key", "isStudent", "value", value)
	})

	t.Run("GetArray", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		value, err := xjson.GetArray(data, "grades")
		assert.NoError(t, err)
		assert.Equal(t, xjson.JSONArray{float64(85), float64(90), float64(78)}, value)
		xlog.Info("Retrieved array value", "key", "grades", "value", value)
	})

	t.Run("GetArrayFromString", func(t *testing.T) {
		value, err := xjson.GetArrayFromString(jsonStr, "grades")
		assert.NoError(t, err)
		assert.Equal(t, xjson.JSONArray{float64(85), float64(90), float64(78)}, value)
		xlog.Info("Retrieved array value from string", "key", "grades", "value", value)
	})

	t.Run("NestedAccess", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "address.city")
		assert.NoError(t, err)
		assert.Equal(t, "Anytown", value)
		xlog.Info("Retrieved nested value", "path", "address.city", "value", value)
	})

	t.Run("ArrayAccess", func(t *testing.T) {
		value, err := xjson.GetFromString(jsonStr, "grades[1]")
		assert.NoError(t, err)
		assert.Equal(t, float64(90), value)
		xlog.Info("Retrieved array element", "path", "grades[1]", "value", value)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		_, err := xjson.GetStringFromString(jsonStr, "nonexistent")
		assert.Error(t, err)
		xlog.Error("Error handling for non-existent key", "error", err)

		_, err = xjson.GetIntFromString(jsonStr, "name")
		assert.Error(t, err)
		xlog.Error("Error handling for type mismatch", "error", err)

		_, err = xjson.GetFromString(jsonStr, "grades[10]")
		assert.Error(t, err)
		xlog.Error("Error handling for out of bounds array access", "error", err)
	})

	t.Run("ComplexNestedAccess", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "contacts[0].value")
		assert.NoError(t, err)
		assert.Equal(t, "john@example.com", value)
		xlog.Info("Retrieved complex nested value", "path", "contacts[0].value", "value", value)
	})

	t.Run("ForEach", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		err := xjson.ForEach(data["contacts"], func(key interface{}, value interface{}) error {
			xlog.Info("ForEach iteration", "key", key, "value", value)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("Map", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		result, err := xjson.Map(data["grades"], func(key interface{}, value interface{}) (interface{}, error) {
			return float64(value.(float64) + 5), nil
		})
		assert.NoError(t, err)
		xlog.Info("Mapped grades", "original", data["grades"], "mapped", result)
	})

	t.Run("Filter", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		result, err := xjson.Filter(data["grades"], func(key interface{}, value interface{}) (bool, error) {
			return value.(float64) >= 85, nil
		})
		assert.NoError(t, err)
		xlog.Info("Filtered grades", "original", data["grades"], "filtered", result)
	})

	t.Run("Reduce", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		result, err := xjson.Reduce(data["grades"], func(accumulator, key, value interface{}) (interface{}, error) {
			return accumulator.(float64) + value.(float64), nil
		}, 0.0)
		assert.NoError(t, err)
		xlog.Info("Reduced grades", "original", data["grades"], "sum", result)
	})

	t.Run("NestedArrayAccess", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "contacts[1].value")
		assert.NoError(t, err)
		assert.Equal(t, "555-1234", value)
		xlog.Info("Retrieved nested array value", "path", "contacts[1].value", "value", value)
	})

	t.Run("InvalidArrayAccess", func(t *testing.T) {
		_, err := xjson.GetFromString(jsonStr, "grades[5]")
		assert.Error(t, err)
		xlog.Error("Error handling for invalid array access", "error", err)
	})

	t.Run("InvalidNestedAccess", func(t *testing.T) {
		_, err := xjson.GetFromString(jsonStr, "address.street.name")
		assert.Error(t, err)
		xlog.Error("Error handling for invalid nested access", "error", err)
	})

	t.Run("GetIntFromFloat", func(t *testing.T) {
		value, err := xjson.GetIntFromString(jsonStr, "age")
		assert.NoError(t, err)
		assert.Equal(t, 30, value)
		xlog.Info("Retrieved int value from float", "key", "age", "value", value)
	})

	t.Run("GetStringFromNumber", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "age")
		assert.NoError(t, err)
		assert.Equal(t, "30", value)
		xlog.Info("Retrieved string value from number", "key", "age", "value", value)
	})

	t.Run("GetStringFromBoolean", func(t *testing.T) {
		value, err := xjson.GetStringFromString(jsonStr, "isStudent")
		assert.NoError(t, err)
		assert.Equal(t, "false", value)
		xlog.Info("Retrieved string value from boolean", "key", "isStudent", "value", value)
	})

	t.Run("MapWithError", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		_, err := xjson.Map(data["grades"], func(key interface{}, value interface{}) (interface{}, error) {
			return nil, fmt.Errorf("intentional error")
		})
		assert.Error(t, err)
		xlog.Error("Error handling in Map function", "error", err)
	})

	t.Run("FilterWithError", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		_, err := xjson.Filter(data["grades"], func(key interface{}, value interface{}) (bool, error) {
			return false, fmt.Errorf("intentional error")
		})
		assert.Error(t, err)
		xlog.Error("Error handling in Filter function", "error", err)
	})

	t.Run("ReduceWithError", func(t *testing.T) {
		data, _ := xjson.ParseJSON(jsonStr)
		_, err := xjson.Reduce(data["grades"], func(accumulator, key, value interface{}) (interface{}, error) {
			return nil, fmt.Errorf("intentional error")
		}, 0.0)
		assert.Error(t, err)
		xlog.Error("Error handling in Reduce function", "error", err)
	})

	t.Run("InvalidJSONParse", func(t *testing.T) {
		invalidJSON := `{"name": "John", "age": 30,}`
		_, err := xjson.ParseJSON(invalidJSON)
		assert.Error(t, err)
		xlog.Error("Error handling for invalid JSON parsing", "error", err)
	})
}
