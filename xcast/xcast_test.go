package xcast_test

import (
	"testing"
	"time"

	"github.com/seefs001/xox/xcast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
		hasError bool
	}{
		{nil, "", false},
		{"hello", "hello", false},
		{123, "123", false},
		{int32(123), "123", false},
		{int64(123), "123", false},
		{123.45, "123.450000", false},
		{true, "true", false},
		{[]byte("hello"), "hello", false},
		{[]int{1, 2, 3}, "1,2,3", false},
		{map[string]int{"a": 1}, `{"a":1}`, false},
		{struct{ Name string }{"John"}, `{"Name":"John"}`, false},
	}

	for i, tt := range tests {
		result, err := xcast.ToString(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int
		hasError bool
	}{
		{nil, 0, false},
		{"123", 123, false},
		{123, 123, false},
		{int32(123), 123, false},
		{int64(123), 123, false},
		{123.45, 123, true},
		{true, 1, false},
		{[]int{1, 2, 3}, 1, false},
		{map[string]int{"a": 1}, 0, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToInt(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToInt32(t *testing.T) {
	tests := []struct {
		input    any
		expected int32
		hasError bool
	}{
		{nil, 0, false},
		{"123", 123, false},
		{123, 123, false},
		{int32(123), 123, false},
		{int64(123), 123, false},
		{123.45, 0, true},
		{true, 1, false},
		{[]int{1, 2, 3}, 1, false},
		{map[string]int{"a": 1}, 0, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToInt32(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
		hasError bool
	}{
		{nil, 0, false},
		{"123", 123, false},
		{123, 123, false},
		{int32(123), 123, false},
		{int64(123), 123, false},
		{123.45, 0, true},
		{true, 1, false},
		{[]int{1, 2, 3}, 1, false},
		{map[string]int{"a": 1}, 0, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToInt64(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		hasError bool
	}{
		{nil, 0.0, false},
		{"123.45", 123.45, false},
		{123, 123.0, false},
		{int32(123), 123.0, false},
		{int64(123), 123.0, false},
		{123.45, 123.45, false},
		{true, 1.0, false},
		{[]float64{1.23, 4.56}, 1.23, false},
		{map[string]float64{"a": 1.23}, 0.0, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToFloat64(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
		hasError bool
	}{
		{nil, false, false},
		{"true", true, false},
		{1, true, false},
		{int32(1), true, false},
		{int64(1), true, false},
		{0, false, false},
		{123.45, true, false},
		{false, false, false},
		{[]bool{true, false}, true, false},
		{map[string]bool{"a": true}, false, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToBool(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToMap(t *testing.T) {
	tests := []struct {
		input    any
		expected map[string]any
		hasError bool
	}{
		{nil, nil, false},
		{map[string]int{"a": 1}, map[string]any{"a": 1}, false},
		{struct{ Name string }{"John"}, map[string]any{"Name": "John"}, false},
		{[]int{1, 2, 3}, map[string]any{"0": 1, "1": 2, "2": 3}, false},
		{123, nil, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToMap(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestToSlice(t *testing.T) {
	tests := []struct {
		input    any
		expected []any
		hasError bool
	}{
		{nil, nil, false},
		{[]int{1, 2, 3}, []any{1, 2, 3}, false},
		{map[string]int{"a": 1}, []any{1}, false},
		{struct{ Name string }{"John"}, []any{"John"}, false},
		{123, nil, true},
	}

	for i, tt := range tests {
		result, err := xcast.ToSlice(tt.input)
		if tt.hasError {
			require.Errorf(t, err, "Test case %d failed: expected error but got none", i)
		} else {
			require.NoErrorf(t, err, "Test case %d failed: unexpected error %v", i, err)
			assert.Equalf(t, tt.expected, result, "Test case %d failed: expected %v but got %v", i, tt.expected, result)
		}
	}
}

func TestConvertStruct(t *testing.T) {
	type SrcStruct struct {
		Name  string
		Value int
	}
	type DstStruct struct {
		Name  string
		Value int
	}

	src := SrcStruct{Name: "John", Value: 42}
	var dst DstStruct

	err := xcast.ConvertStruct(src, &dst)
	require.NoError(t, err)
	assert.Equal(t, src.Name, dst.Name)
	assert.Equal(t, src.Value, dst.Value)
}

func TestStringToStruct(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name     string
		input    string
		expected Person
		hasError bool
	}{
		{
			name:     "Valid JSON",
			input:    `{"name":"Alice","age":30}`,
			expected: Person{Name: "Alice", Age: 30},
			hasError: false,
		},
		{
			name:     "Empty JSON",
			input:    `{}`,
			expected: Person{},
			hasError: false,
		},
		{
			name:     "Invalid JSON",
			input:    `{"name":"Bob","age":}`,
			expected: Person{},
			hasError: true,
		},
		{
			name:     "Extra fields in JSON",
			input:    `{"name":"Charlie","age":35,"city":"New York"}`,
			expected: Person{Name: "Charlie", Age: 35},
			hasError: false,
		},
		{
			name:     "Missing fields in JSON",
			input:    `{"name":"David"}`,
			expected: Person{Name: "David", Age: 0},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := xcast.StringToStruct[Person](tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStructToString(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name     string
		input    Person
		expected string
		hasError bool
	}{
		{
			name:     "Valid struct",
			input:    Person{Name: "Alice", Age: 30},
			expected: `{"name":"Alice","age":30}`,
			hasError: false,
		},
		{
			name:     "Empty struct",
			input:    Person{},
			expected: `{"name":"","age":0}`,
			hasError: false,
		},
		{
			name:     "Struct with zero values",
			input:    Person{Name: "", Age: 0},
			expected: `{"name":"","age":0}`,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := xcast.StructToString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expected, result)
			}
		})
	}
}

// Test for edge cases and error handling
func TestStringToStructEdgeCases(t *testing.T) {
	type ComplexStruct struct {
		IntValue    int               `json:"int_value"`
		FloatValue  float64           `json:"float_value"`
		BoolValue   bool              `json:"bool_value"`
		StringValue string            `json:"string_value"`
		ArrayValue  []int             `json:"array_value"`
		MapValue    map[string]string `json:"map_value"`
	}

	tests := []struct {
		name     string
		input    string
		expected ComplexStruct
		hasError bool
	}{
		{
			name:     "Complex valid JSON",
			input:    `{"int_value":42,"float_value":3.14,"bool_value":true,"string_value":"test","array_value":[1,2,3],"map_value":{"key":"value"}}`,
			expected: ComplexStruct{IntValue: 42, FloatValue: 3.14, BoolValue: true, StringValue: "test", ArrayValue: []int{1, 2, 3}, MapValue: map[string]string{"key": "value"}},
			hasError: false,
		},
		{
			name:     "Invalid JSON syntax",
			input:    `{"int_value":42,}`,
			expected: ComplexStruct{},
			hasError: true,
		},
		{
			name:     "Invalid type for field",
			input:    `{"int_value":"not an int"}`,
			expected: ComplexStruct{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := xcast.StringToStruct[ComplexStruct](tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test for custom types and nested structs
func TestStructToStringComplexCases(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name     string
		input    Person
		expected string
		hasError bool
	}{
		{
			name: "Nested struct",
			input: Person{
				Name: "Alice",
				Age:  30,
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
				},
			},
			expected: `{"name":"Alice","age":30,"address":{"street":"123 Main St","city":"Anytown"}}`,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := xcast.StructToString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expected, result)
			}
		})
	}
}

func TestStringToBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"t", true},
		{"true", true},
		{"yes", true},
		{"y", true},
		{"on", true},
		{"0", false},
		{"f", false},
		{"false", false},
		{"no", false},
		{"n", false},
		{"off", false},
		{"", false},
		{"random", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := xcast.StringToBool(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStringToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"42", 42, false},
		{"-42", -42, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToInt(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"9223372036854775807", 9223372036854775807, false},
		{"-9223372036854775808", -9223372036854775808, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToInt64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToUint(t *testing.T) {
	tests := []struct {
		input    string
		expected uint
		hasError bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToUint(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"18446744073709551615", 18446744073709551615, false},
		{"0", 0, false},
		{"", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToUint64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"3.14", 3.14, false},
		{"-2.5", -2.5, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToFloat64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"5s", 5 * time.Second, false},
		{"10m", 10 * time.Minute, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := xcast.StringToDuration(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToMap(t *testing.T) {
	tests := []struct {
		input    string
		pairSep  string
		kvSep    string
		expected map[string]string
	}{
		{"key1=value1,key2=value2", ",", "=", map[string]string{"key1": "value1", "key2": "value2"}},
		{"k1:v1;k2:v2", ";", ":", map[string]string{"k1": "v1", "k2": "v2"}},
		{"", ",", "=", map[string]string{}},
		{"invalid", ",", "=", map[string]string{}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := xcast.StringToMap(test.input, test.pairSep, test.kvSep)
			assert.Equal(t, test.expected, result)
		})
	}
}
