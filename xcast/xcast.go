package xcast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/xerror"
)

// ToString converts various types to a string.
func ToString(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return "", nil
	}

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
				return "", xerror.Wrap(err, "failed to convert slice element to string")
			}
			strs[i] = str
		}
		return strings.Join(strs, ","), nil
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return "", xerror.Wrap(err, "failed to marshal value to JSON")
		}
		return string(jsonBytes), nil
	}

	return "", xerror.Errorf("unsupported type: %T", value)
}

// ToInt converts various types to an int.
func ToInt(value any) (int, error) {
	intVal, err := toInt64(value)
	if err != nil {
		return 0, xerror.Wrap(err, "failed to convert to int64")
	}
	return int(intVal), nil
}

// ToInt32 converts various types to an int32.
func ToInt32(value any) (int32, error) {
	intVal, err := toInt64(value)
	if err != nil {
		return 0, xerror.Wrap(err, "failed to convert to int64")
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
	if !v.IsValid() {
		return 0, nil
	}

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
			return 0, xerror.Wrap(err, "failed to parse string as int64")
		}
		return intVal, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		floatVal := v.Float()
		intVal := int64(floatVal)
		if floatVal != float64(intVal) {
			return 0, xerror.Errorf("cannot convert %v to int64 without loss of precision", floatVal)
		}
		return intVal, nil
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return 0, xerror.New("cannot convert byte slice to int64")
		}
		if v.Len() == 0 {
			return 0, nil
		}
		return toInt64(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return 0, xerror.Wrap(err, "failed to marshal value to JSON")
		}
		var result int64
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return 0, xerror.Wrap(err, "failed to unmarshal JSON to int64")
		}
		return result, nil
	}

	return 0, xerror.Errorf("unsupported type: %T", value)
}

// ToFloat64 converts various types to a float64.
func ToFloat64(value any) (float64, error) {
	if value == nil {
		return 0.0, nil
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return 0.0, nil
	}

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
			return 0.0, xerror.Wrap(err, "failed to parse string as float64")
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
			return 0.0, xerror.New("cannot convert byte slice to float64")
		}
		if v.Len() == 0 {
			return 0.0, nil
		}
		return ToFloat64(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return 0.0, xerror.Wrap(err, "failed to marshal value to JSON")
		}
		var result float64
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return 0.0, xerror.Wrap(err, "failed to unmarshal JSON to float64")
		}
		return result, nil
	}

	return 0.0, xerror.Errorf("unsupported type: %T", value)
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
			return false, xerror.Wrap(err, "failed to parse string as bool")
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
			return false, xerror.New("cannot convert byte slice to bool")
		}
		if v.Len() == 0 {
			return false, nil
		}
		return ToBool(v.Index(0).Interface())
	case reflect.Map, reflect.Struct:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return false, xerror.Wrap(err, "failed to marshal value to JSON")
		}
		var result bool
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return false, xerror.Wrap(err, "failed to unmarshal JSON to bool")
		}
		return result, nil
	}

	return false, xerror.Errorf("unsupported type: %T", value)
}

// ToMap converts various types to a map.
func ToMap(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil, nil
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		result := make(map[string]any, v.Len())
		for _, key := range v.MapKeys() {
			strKey, err := ToString(key.Interface())
			if err != nil {
				return nil, xerror.Wrap(err, "failed to convert map key to string")
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

	return nil, xerror.Errorf("unsupported type: %T", value)
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

	return nil, xerror.Errorf("unsupported type: %T", value)
}

// ConvertStruct converts one struct to another, matching fields by name (case-insensitive).
func ConvertStruct(src any, dst any) error {
	if src == nil || dst == nil {
		return xerror.New("src and dst cannot be nil")
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if !dstVal.IsValid() || dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return xerror.New("dst must be a non-nil pointer to struct")
	}
	dstVal = dstVal.Elem()

	if srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return xerror.New("src cannot be nil pointer")
		}
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return xerror.New("both src and dst must be structs")
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

// StringToStruct converts a string to a struct of type T.
// It uses json.Unmarshal to perform the conversion.
//
// Example:
//
//	type Person struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//
//	jsonStr := `{"name":"Alice","age":30}`
//	person, err := StringToStruct[Person](jsonStr)
//	if err != nil {
//		// handle error
//	}
//	fmt.Printf("%+v\n", person) // Output: {Name:Alice Age:30}
func StringToStruct[T any](s string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return result, xerror.Wrap(err, "failed to convert string to struct")
	}
	return result, nil
}

// StructToString converts a struct of type T to a string.
// It uses json.Marshal to perform the conversion.
//
// Example:
//
//	type Person struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//
//	person := Person{Name: "Bob", Age: 25}
//	jsonStr, err := StructToString(person)
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(jsonStr) // Output: {"name":"Bob","age":25}
func StructToString[T any](v T) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", xerror.Wrap(err, "failed to convert struct to string")
	}
	return string(bytes), nil
}

func MustToString(value any) string {
	str, err := ToString(value)
	if err != nil {
		return ""
	}
	return str
}

func MustToInt64(value any) int64 {
	i, err := ToInt64(value)
	if err != nil {
		return 0
	}
	return i
}

func MustToFloat64(value any) float64 {
	f, err := ToFloat64(value)
	if err != nil {
		return 0.0
	}
	return f
}

func MustToBool(value any) bool {
	b, err := ToBool(value)
	if err != nil {
		return false
	}
	return b
}

// StringToBool converts a string to a boolean
func StringToBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "t" || s == "true" || s == "yes" || s == "y" || s == "on"
}

// StringToInt converts a string to an int
func StringToInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// StringToInt64 converts a string to an int64
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// StringToUint converts a string to a uint
func StringToUint(s string) (uint, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// StringToUint64 converts a string to a uint64
func StringToUint64(s string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), 10, 64)
}

// StringToFloat64 converts a string to a float64
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// StringToDuration converts a string to a time.Duration
func StringToDuration(s string) (time.Duration, error) {
	return time.ParseDuration(strings.TrimSpace(s))
}

// StringToMap converts a string to a map[string]string
func StringToMap(s, pairSep, kvSep string) map[string]string {
	m := make(map[string]string)
	pairs := strings.Split(s, pairSep)
	for _, pair := range pairs {
		kv := strings.SplitN(pair, kvSep, 2)
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return m
}
