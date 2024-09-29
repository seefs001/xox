package x

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/xerror"
)

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

// BindData binds data from a map to a struct based on tags
func BindData(v interface{}, data map[string][]string) error {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return xerror.New("v must be a pointer to a struct")
	}

	val := reflect.ValueOf(v).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)

		if !field.CanSet() {
			continue
		}

		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = strings.ToLower(typeField.Name)
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			continue
		}

		if err := setField(field, inputValue); err != nil {
			return xerror.Wrapf(err, "error setting field %s", typeField.Name)
		}
	}

	return nil
}

func setField(field reflect.Value, values []string) error {
	if len(values) == 0 {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(values[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			return xerror.Wrap(err, "error parsing int value")
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(values[0], 10, 64)
		if err != nil {
			return xerror.Wrap(err, "error parsing uint value")
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return xerror.Wrap(err, "error parsing float value")
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(values[0])
		if err != nil {
			return xerror.Wrap(err, "error parsing bool value")
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		slice := reflect.MakeSlice(field.Type(), len(values), len(values))
		for i, value := range values {
			if err := setField(slice.Index(i), []string{value}); err != nil {
				return xerror.Wrapf(err, "error setting slice element %d", i)
			}
		}
		field.Set(slice)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			timeValue, err := time.Parse(time.RFC3339, values[0])
			if err != nil {
				return xerror.Wrap(err, "error parsing time value")
			}
			field.Set(reflect.ValueOf(timeValue))
		}
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), values)
	case reflect.Interface:
		if field.IsNil() {
			field.Set(reflect.ValueOf(values[0]))
		} else {
			return setField(field.Elem(), values)
		}
	default:
		return xerror.Newf("unsupported field type: %v", field.Kind())
	}
	return nil
}
