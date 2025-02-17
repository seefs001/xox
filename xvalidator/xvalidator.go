package xvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex  = regexp.MustCompile(`^[0-9]+$`)
	regexpCache   = &sync.Map{}
)

// ValidationError represents an error that occurred during validation
type ValidationError struct {
	Field      string
	Message    string
	NestedPath string
}

func (e ValidationError) Error() string {
	if e.NestedPath != "" {
		return fmt.Sprintf("%s.%s: %s", e.NestedPath, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateOptions contains options for validation
type ValidateOptions struct {
	StopOnFirst bool
	NestedPath  string
}

// Validate performs validation on a struct or pointer to struct
func Validate(v interface{}) []error {
	return ValidateWithOptions(v, ValidateOptions{})
}

// ValidateWithOptions performs validation with custom options
func ValidateWithOptions(v interface{}, opts ValidateOptions) []error {
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
			// Check for nested struct
			if value.Kind() == reflect.Struct || (value.Kind() == reflect.Ptr && !value.IsNil() && value.Elem().Kind() == reflect.Struct) {
				nestedPath := field.Name
				if opts.NestedPath != "" {
					nestedPath = opts.NestedPath + "." + nestedPath
				}
				if nestedErrors := ValidateWithOptions(value.Interface(), ValidateOptions{
					StopOnFirst: opts.StopOnFirst,
					NestedPath:  nestedPath,
				}); nestedErrors != nil {
					errors = append(errors, nestedErrors...)
					if opts.StopOnFirst && len(errors) > 0 {
						return errors
					}
				}
			}
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

			if err := applyValidator(validatorName, validatorArg, field.Name, value.Interface(), opts.NestedPath); err != nil {
				errors = append(errors, err)
				if opts.StopOnFirst {
					return errors
				}
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func applyValidator(name, arg, fieldName string, value interface{}, nestedPath string) error {
	if name == "" {
		return fmt.Errorf("empty validator name for field %s", fieldName)
	}

	var err error
	switch name {
	case "required":
		err = required(fieldName, value)
	case "min":
		err = min(fieldName, value, arg)
	case "max":
		err = max(fieldName, value, arg)
	case "email":
		err = email(fieldName, value)
	case "regexp":
		err = regexpValidator(fieldName, value, arg)
	case "len":
		err = length(fieldName, value, arg)
	case "in":
		err = in(fieldName, value, arg)
	case "notin":
		err = notIn(fieldName, value, arg)
	case "alpha":
		err = alpha(fieldName, value)
	case "alphanum":
		err = alphanumeric(fieldName, value)
	case "numeric":
		err = numeric(fieldName, value)
	case "datetime":
		err = datetime(fieldName, value, arg)
	case "url":
		err = url(fieldName, value)
	case "uuid":
		err = uuid(fieldName, value)
	case "ipv4":
		err = ipv4(fieldName, value)
	case "ipv6":
		err = ipv6(fieldName, value)
	default:
		err = fmt.Errorf("unknown validator: %s for field %s", name, fieldName)
	}

	if err != nil {
		if ve, ok := err.(ValidationError); ok && nestedPath != "" {
			ve.NestedPath = nestedPath
			return ve
		}
	}
	return err
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
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	if !emailRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Invalid email format"}
	}
	return nil
}

func regexpValidator(field string, value interface{}, pattern string) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	var regex *regexp.Regexp
	if cached, ok := regexpCache.Load(pattern); ok {
		regex = cached.(*regexp.Regexp)
	} else {
		var err error
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regexp pattern: %s for field %s: %v", pattern, field, err)
		}
		regexpCache.Store(pattern, regex)
	}

	if !regex.MatchString(str) {
		return ValidationError{Field: field, Message: fmt.Sprintf("Does not match pattern: %s", pattern)}
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
		return ValidationError{Field: field, Message: "Must be a string"}
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	if !alphaRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Must contain only alphabetic characters"}
	}
	return nil
}

func alphanumeric(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
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
		return ValidationError{Field: field, Message: "Must be a string"}
	}
	if str == "" {
		return nil // Allow empty string unless required is also specified
	}
	if _, err := time.Parse(layout, str); err != nil {
		return ValidationError{Field: field, Message: fmt.Sprintf("Must be a valid datetime in the format %s", layout)}
	}
	return nil
}

func url(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
	if !urlRegex.MatchString(str) {
		return ValidationError{Field: field, Message: "Invalid URL format"}
	}
	return nil
}

func uuid(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(strings.ToLower(str)) {
		return ValidationError{Field: field, Message: "Invalid UUID format"}
	}
	return nil
}

func ipv4(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	ipv4Regex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipv4Regex.MatchString(str) {
		return ValidationError{Field: field, Message: "Invalid IPv4 format"}
	}

	parts := strings.Split(str, ".")
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return ValidationError{Field: field, Message: "Invalid IPv4 format"}
		}
	}
	return nil
}

func ipv6(field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ValidationError{Field: field, Message: "Must be a string"}
	}

	if str == "" {
		return nil
	}

	ipv6Regex := regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))$`)
	if !ipv6Regex.MatchString(str) {
		return ValidationError{Field: field, Message: "Invalid IPv6 format"}
	}
	return nil
}
