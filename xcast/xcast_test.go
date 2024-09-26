package xcast_test

import (
	"testing"

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

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		hasError bool
	}{
		{nil, 0.0, false},
		{"123.45", 123.45, false},
		{123, 123.0, false},
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
