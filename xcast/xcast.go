package xcast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ToString converts various types to a string.
func ToString(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float()), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return string(v.Bytes()), nil
		}
		strs := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			str, err := ToString(v.Index(i).Interface())
			if err != nil {
				return "", err
			}
			strs[i] = str
		}
		return strings.Join(strs, ","), nil
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}

	return "", fmt.Errorf("unsupported type: %T", value)
}

// ToInt converts various types to an int.
func ToInt(value any) (int, error) {
	intVal, err := toInt64(value)
	if err != nil {
		return 0, err
	}
	return int(intVal), nil
}

// ToInt32 converts various types to an int32.
func ToInt32(value any) (int32, error) {
	intVal, err := toInt64(value)
	if err != nil {
		return 0, err
	}
	return int32(intVal), nil
}

// ToInt64 converts various types to an int64.
func ToInt64(value any) (int64, error) {
	return toInt64(value)
}

func toInt64(value any) (int64, error) {
	if value == nil {
		return 0, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		intVal, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string to int64: %v", err)
		}
		return intVal, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return 0, fmt.Errorf("cannot convert byte slice to int64")
		}
		if v.Len() == 0 {
			return 0, nil
		}
		return toInt64(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return 0, err
		}
		var result int64
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return 0, err
		}
		return result, nil
	}

	return 0, fmt.Errorf("unsupported type: %T", value)
}

// ToFloat64 converts various types to a float64.
func ToFloat64(value any) (float64, error) {
	if value == nil {
		return 0.0, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0.0, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		floatVal, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0.0, fmt.Errorf("cannot convert string to float64: %v", err)
		}
		return floatVal, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Bool:
		if v.Bool() {
			return 1.0, nil
		}
		return 0.0, nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return 0.0, fmt.Errorf("cannot convert byte slice to float64")
		}
		if v.Len() == 0 {
			return 0.0, nil
		}
		return ToFloat64(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return 0.0, err
		}
		var result float64
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return 0.0, err
		}
		return result, nil
	}

	return 0.0, fmt.Errorf("unsupported type: %T", value)
}

// ToBool converts various types to a bool.
func ToBool(value any) (bool, error) {
	if value == nil {
		return false, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		boolVal, err := strconv.ParseBool(v.String())
		if err != nil {
			return false, fmt.Errorf("cannot convert string to bool: %v", err)
		}
		return boolVal, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0.0, nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return false, fmt.Errorf("cannot convert byte slice to bool")
		}
		if v.Len() == 0 {
			return false, nil
		}
		return ToBool(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return false, err
		}
		var result bool
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return false, err
		}
		return result, nil
	}

	return false, fmt.Errorf("unsupported type: %T", value)
}

// ToMap converts various types to a map.
func ToMap(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		result := make(map[string]any)
		for _, key := range v.MapKeys() {
			strKey, err := ToString(key.Interface())
			if err != nil {
				return nil, err
			}
			result[strKey] = v.MapIndex(key).Interface()
		}
		return result, nil
	case reflect.Struct:
		result := make(map[string]any)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			result[field.Name] = v.Field(i).Interface()
		}
		return result, nil
	case reflect.Slice:
		result := make(map[string]any)
		for i := 0; i < v.Len(); i++ {
			strKey := fmt.Sprintf("%d", i)
			result[strKey] = v.Index(i).Interface()
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported type: %T", value)
}

// ToSlice converts various types to a slice.
func ToSlice(value any) ([]any, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result, nil
	case reflect.Map:
		result := make([]any, 0, v.Len())
		for _, key := range v.MapKeys() {
			result = append(result, v.MapIndex(key).Interface())
		}
		return result, nil
	case reflect.Struct:
		result := make([]any, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			result[i] = v.Field(i).Interface()
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported type: %T", value)
}

// ConvertStruct converts one struct to another, matching fields by name (case-insensitive).
func ConvertStruct(src any, dst any) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem()

	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return fmt.Errorf("both src and dst must be structs")
	}

	srcType := srcVal.Type()
	dstType := dstVal.Type()

	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		srcFieldName := strings.ToLower(srcField.Name)

		for j := 0; j < dstType.NumField(); j++ {
			dstField := dstType.Field(j)
			dstFieldName := strings.ToLower(dstField.Name)

			if srcFieldName == dstFieldName {
				dstFieldVal := dstVal.FieldByName(dstField.Name)
				if dstFieldVal.CanSet() {
					dstFieldVal.Set(srcVal.Field(i))
				}
				break
			}
		}
	}

	return nil
}
