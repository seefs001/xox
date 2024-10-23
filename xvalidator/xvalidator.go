package xvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Validator represents a validation function
type Validator func(interface{}) error

// ValidationError represents an error that occurred during validation
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate performs validation on a struct or pointer to struct
func Validate(v interface{}) []error {
	if v == nil {
		return []error{fmt.Errorf("validation target is nil")}
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return []error{fmt.Errorf("validation target is nil pointer")}
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return []error{fmt.Errorf("validation target must be a struct or pointer to struct, got %v", val.Kind())}
	}

	var errors []error = make([]error, 0)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		value := val.Field(i)
		tag := field.Tag.Get("xv")

		if tag == "" {
			continue
		}

		validators := strings.Split(tag, ",")
		for _, validator := range validators {
			validator = strings.TrimSpace(validator)
			if validator == "" {
				continue
			}
			parts := strings.SplitN(validator, "=", 2)
			validatorName := strings.TrimSpace(parts[0])
			var validatorArg string
			if len(parts) > 1 {
				validatorArg = strings.TrimSpace(parts[1])
			}

			if err := applyValidator(validatorName, validatorArg, field.Name, value.Interface()); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func applyValidator(name, arg, fieldName string, value interface{}) error {
	if name == "" {
		return fmt.Errorf("empty validator name for field %s", fieldName)
	}

	switch name {
	case "required":
		return required(fieldName, value)
	case "min":
		return min(fieldName, value, arg)
	case "max":
		return max(fieldName, value, arg)
	case "email":
		return email(fieldName, value)
	case "regexp":
		return regexpValidator(fieldName, value, arg)
	case "len":
		return length(fieldName, value, arg)
	case "in":
		return in(fieldName, value, arg)
	case "notin":
		return notIn(fieldName, value, arg)
	case "alpha":
		return alpha(fieldName, value)
	case "alphanum":
		return alphanumeric(fieldName, value)
	case "numeric":
		return numeric(fieldName, value)
	case "datetime":
		return datetime(fieldName, value, arg)
	default:
		return fmt.Errorf("unknown validator: %s for field %s", name, fieldName)
	}
}

func required(field string, value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() == 0 {
			return ValidationError{Field: field, Message: "This field is required"}
		}
	case reflect.Bool:
		return nil
	default:
		if v.IsZero() {
			return ValidationError{Field: field, Message: "This field is required"}
		}
	}
	return nil
}

func min(field string, value interface{}, arg string) error {
	if arg == "" {
		return fmt.Errorf("min validator requires an argument for field %s", field)
	}

	num, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return fmt.Errorf("invalid min value: %s for field %s", arg, field)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at least %v", num)}
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		if v.Len() < int(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at least %v characters long", num)}
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() < int(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must have at least %v elements", num)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < int64(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at least %v", num)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() < uint64(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at least %v", num)}
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < num {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at least %v", num)}
		}
	default:
		return fmt.Errorf("min validator not supported for type %v in field %s", v.Kind(), field)
	}
	return nil
}

func max(field string, value interface{}, arg string) error {
	if arg == "" {
		return fmt.Errorf("max validator requires an argument for field %s", field)
	}

	num, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return fmt.Errorf("invalid max value: %s for field %s", arg, field)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at most %v", num)}
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() > int(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must have at most %v elements", num)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > int64(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at most %v", num)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() > uint64(num) {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at most %v", num)}
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > num {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must be at most %v", num)}
		}
	default:
		return fmt.Errorf("max validator not supported for type %v in field %s", v.Kind(), field)
	}
	return nil
}

func email(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && !v.IsNil() && v.Elem().Kind() == reflect.String {
			str = v.Elem().String()
		} else {
			return fmt.Errorf("email validator requires a string for field %s", field)
		}
	}
	if str == "" {
		return nil
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Invalid email format"}
	}
	return nil
}

func regexpValidator(field string, value interface{}, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("regexp validator requires a pattern for field %s", field)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("regexp validator requires a string for field %s", field)
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regexp pattern: %s for field %s", pattern, field)
	}
	if !re.MatchString(str) {
		return ValidationError{Field: field, Message: fmt.Sprintf("Does not match pattern %s", pattern)}
	}
	return nil
}

func length(field string, value interface{}, arg string) error {
	if arg == "" {
		return fmt.Errorf("length validator requires an argument for field %s", field)
	}

	length, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid length value: %s for field %s", arg, field)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must have exactly %d elements", length)}
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() != length {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must have exactly %d elements", length)}
		}
	default:
		return fmt.Errorf("length validator not supported for type %v in field %s", v.Kind(), field)
	}
	return nil
}

func in(field string, value interface{}, arg string) error {
	if arg == "" {
		return fmt.Errorf("in validator requires an argument for field %s", field)
	}

	options := strings.Split(arg, "|")
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	strValue := fmt.Sprintf("%v", v.Interface())
	if strValue == "" {
		return nil
	}

	for _, option := range options {
		if strValue == option {
			return nil
		}
	}
	return ValidationError{Field: field, Message: fmt.Sprintf("Must be one of: %s", arg)}
}

func notIn(field string, value interface{}, arg string) error {
	if arg == "" {
		return fmt.Errorf("notin validator requires an argument for field %s", field)
	}

	options := strings.Split(arg, "|")
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	strValue := fmt.Sprintf("%v", v.Interface())
	if strValue == "" {
		return nil
	}

	for _, option := range options {
		if strValue == option {
			return ValidationError{Field: field, Message: fmt.Sprintf("Must not be one of: %s", arg)}
		}
	}
	return nil
}

func alpha(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alpha validator requires a string for field %s", field)
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Must contain only alphabetic characters"}
	}
	return nil
}

func alphanumeric(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alphanum validator requires a string for field %s", field)
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	alphanumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Must contain only alphanumeric characters"}
	}
	return nil
}

func numeric(field string, value interface{}) error {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.String:
		if v.String() == "" {
			return nil // Allow empty string unless required is also specified
		}
		if _, err := strconv.ParseFloat(v.String(), 64); err != nil {
			return ValidationError{Field: field, Message: "Must be a numeric value"}
		}
		return nil
	default:
		return ValidationError{Field: field, Message: "Must be a numeric value"}
	}
}

func datetime(field string, value interface{}, layout string) error {
	if layout == "" {
		return fmt.Errorf("datetime validator requires a layout argument for field %s", field)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("datetime validator requires a string for field %s", field)
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	if _, err := time.Parse(layout, str); err != nil {
		return ValidationError{Field: field, Message: fmt.Sprintf("Must be a valid datetime in the format %s", layout)}
	}
	return nil
}
