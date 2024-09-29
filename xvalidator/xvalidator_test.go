package xvalidator_test

import (
	"testing"
	"time"

	"github.com/seefs001/xox/xvalidator"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name     string    `xv:"min=2,max=50"`
	Age      int       `xv:"required,min=18,max=120"`
	Email    string    `xv:"email"`
	Password string    `xv:"min=8,regexp=^[a-zA-Z0-9]+$"`
	Role     string    `xv:"in=admin|user|guest"`
	Score    float64   `xv:"min=0,max=100"`
	Tags     []string  `xv:"min=1,max=5"`
	Active   bool      `xv:"required"`
	Created  time.Time `xv:"required,datetime=2006-01-02"`
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		input         TestStruct
		expectedError bool
	}{
		{
			name: "Valid struct",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name: "Invalid name (too short)",
			input: TestStruct{
				Name:     "J",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid name (too long)",
			input: TestStruct{
				Name:     string(make([]byte, 51)),
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid age (too young)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      17,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid age (too old)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      121,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid email",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "invalid-email",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid password (too short)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "pass",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid password (non-alphanumeric)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password!@#",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid role",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "manager",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid score (negative)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    -1,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid score (too high)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    100.1,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid tags (empty)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid tags (too many)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Invalid created date",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    85.5,
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 13, 1, 0, 0, 0, 0, time.UTC), // Invalid month
			},
			expectedError: true,
		},
		{
			name: "Edge case: max values",
			input: TestStruct{
				Name:     string(make([]byte, 50)),
				Age:      120,
				Email:    "a@b.cd",
				Password: "12345678",
				Role:     "admin",
				Score:    100,
				Tags:     []string{"1", "2", "3", "4", "5"},
				Active:   true,
				Created:  time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name: "Edge case: min values",
			input: TestStruct{
				Name:     "ab",
				Age:      18,
				Email:    "a@b.co",
				Password: "12345678",
				Role:     "guest",
				Score:    0,
				Tags:     []string{"1"},
				Active:   false,
				Created:  time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: false,
		},
		{
			name: "Edge case: float precision",
			input: TestStruct{
				Name:     "John Doe",
				Age:      30,
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				Score:    100.00000000000001, // Using a literal value instead of math.NextAfter
				Tags:     []string{"tag1", "tag2"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: true,
		},
		{
			name: "Empty string fields",
			input: TestStruct{
				Name:     "",
				Age:      30,
				Email:    "",
				Password: "",
				Role:     "",
				Score:    85.5,
				Tags:     []string{"tag1"},
				Active:   true,
				Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := xvalidator.Validate(tt.input)
			if tt.expectedError {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := xvalidator.ValidationError{
		Field:   "TestField",
		Message: "Test error message",
	}

	assert.Equal(t, "TestField: Test error message", err.Error())
}

func TestSpecificValidators(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		type RequiredTest struct {
			Field string `xv:"required"`
		}

		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: "not empty"}))
		assert.NotEmpty(t, xvalidator.Validate(RequiredTest{Field: ""}))
		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: " "}))

		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: "\t"}))
		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: "\n"}))
		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: "0"}))
		assert.Empty(t, xvalidator.Validate(RequiredTest{Field: "false"}))
	})

	t.Run("Min/Max", func(t *testing.T) {
		type MinMaxTest struct {
			IntField    int     `xv:"min=5,max=10"`
			FloatField  float64 `xv:"min=0.5,max=1.5"`
			StringField string  `xv:"min=3,max=5"`
		}

		assert.Empty(t, xvalidator.Validate(MinMaxTest{IntField: 7, FloatField: 1.0, StringField: "abcd"}))
		assert.NotEmpty(t, xvalidator.Validate(MinMaxTest{IntField: 4, FloatField: 0.4, StringField: "ab"}))
		assert.NotEmpty(t, xvalidator.Validate(MinMaxTest{IntField: 11, FloatField: 1.6, StringField: "abcdef"}))
		assert.Empty(t, xvalidator.Validate(MinMaxTest{IntField: 5, FloatField: 0.5, StringField: "abc"}))
		assert.Empty(t, xvalidator.Validate(MinMaxTest{IntField: 10, FloatField: 1.5, StringField: "abcde"}))
		assert.NotEmpty(t, xvalidator.Validate(MinMaxTest{IntField: 0, FloatField: 0, StringField: ""}))
	})

	t.Run("Email", func(t *testing.T) {
		type EmailTest struct {
			Email string `xv:"email"`
		}

		assert.Empty(t, xvalidator.Validate(EmailTest{Email: "test@example.com"}))
		assert.Empty(t, xvalidator.Validate(EmailTest{Email: "test+alias@example.co.uk"}))
		assert.NotEmpty(t, xvalidator.Validate(EmailTest{Email: "invalid-email"}))
		assert.NotEmpty(t, xvalidator.Validate(EmailTest{Email: "test@example"}))
		assert.NotEmpty(t, xvalidator.Validate(EmailTest{Email: "@example.com"}))
		assert.NotEmpty(t, xvalidator.Validate(EmailTest{Email: "test@.com"}))
		assert.Empty(t, xvalidator.Validate(EmailTest{Email: ""}))
	})

	t.Run("Regexp", func(t *testing.T) {
		type RegexpTest struct {
			Field string `xv:"regexp=^[a-z]+$"`
		}

		assert.Empty(t, xvalidator.Validate(RegexpTest{Field: "abcdef"}))
		assert.NotEmpty(t, xvalidator.Validate(RegexpTest{Field: "abc123"}))
		assert.NotEmpty(t, xvalidator.Validate(RegexpTest{Field: "ABC"}))
		assert.Empty(t, xvalidator.Validate(RegexpTest{Field: ""}))
		assert.NotEmpty(t, xvalidator.Validate(RegexpTest{Field: "a b c"}))
	})

	t.Run("In", func(t *testing.T) {
		type InTest struct {
			Field string `xv:"in=apple|banana|cherry"`
		}

		assert.Empty(t, xvalidator.Validate(InTest{Field: "banana"}))
		assert.NotEmpty(t, xvalidator.Validate(InTest{Field: "grape"}))
		assert.Empty(t, xvalidator.Validate(InTest{Field: ""}))
		assert.NotEmpty(t, xvalidator.Validate(InTest{Field: "APPLE"}))
		assert.NotEmpty(t, xvalidator.Validate(InTest{Field: "apple "}))
	})

	t.Run("NotIn", func(t *testing.T) {
		type NotInTest struct {
			Field string `xv:"notin=apple|banana|cherry"`
		}

		assert.Empty(t, xvalidator.Validate(NotInTest{Field: "grape"}))
		assert.NotEmpty(t, xvalidator.Validate(NotInTest{Field: "banana"}))
		assert.Empty(t, xvalidator.Validate(NotInTest{Field: ""}))
		assert.Empty(t, xvalidator.Validate(NotInTest{Field: "APPLE"}))
		assert.Empty(t, xvalidator.Validate(NotInTest{Field: "apple "}))
	})

	t.Run("Alpha", func(t *testing.T) {
		type AlphaTest struct {
			Field string `xv:"alpha"`
		}

		assert.Empty(t, xvalidator.Validate(AlphaTest{Field: "abcDEF"}))
		assert.NotEmpty(t, xvalidator.Validate(AlphaTest{Field: "abc123"}))
		assert.Empty(t, xvalidator.Validate(AlphaTest{Field: ""}))
		assert.NotEmpty(t, xvalidator.Validate(AlphaTest{Field: "abc def"}))
		assert.NotEmpty(t, xvalidator.Validate(AlphaTest{Field: "abc-def"}))
	})

	t.Run("Alphanumeric", func(t *testing.T) {
		type AlphanumTest struct {
			Field string `xv:"alphanum"`
		}

		assert.Empty(t, xvalidator.Validate(AlphanumTest{Field: "abc123"}))
		assert.NotEmpty(t, xvalidator.Validate(AlphanumTest{Field: "abc_123"}))
		assert.Empty(t, xvalidator.Validate(AlphanumTest{Field: ""}))
		assert.NotEmpty(t, xvalidator.Validate(AlphanumTest{Field: "abc 123"}))
		assert.NotEmpty(t, xvalidator.Validate(AlphanumTest{Field: "abc-123"}))
	})

	t.Run("Numeric", func(t *testing.T) {
		type NumericTest struct {
			IntField    int     `xv:"numeric"`
			FloatField  float64 `xv:"numeric"`
			StringField string  `xv:"numeric"`
		}

		assert.Empty(t, xvalidator.Validate(NumericTest{IntField: 123, FloatField: 123.45, StringField: "123"}))
		assert.Empty(t, xvalidator.Validate(NumericTest{IntField: -123, FloatField: -123.45, StringField: "-123.45"}))
		assert.NotEmpty(t, xvalidator.Validate(NumericTest{IntField: 123, FloatField: 123.45, StringField: "12a3"}))
		assert.Empty(t, xvalidator.Validate(NumericTest{IntField: 0, FloatField: 0, StringField: ""}))
		assert.NotEmpty(t, xvalidator.Validate(NumericTest{IntField: 123, FloatField: 123.45, StringField: "12.34.56"}))
	})

	t.Run("Datetime", func(t *testing.T) {
		type DatetimeTest struct {
			Field string `xv:"datetime=2006-01-02"`
		}

		assert.Empty(t, xvalidator.Validate(DatetimeTest{Field: "2023-05-01"}))
		assert.NotEmpty(t, xvalidator.Validate(DatetimeTest{Field: "2023/05/01"}))
		assert.NotEmpty(t, xvalidator.Validate(DatetimeTest{Field: "2023-13-01"}))
		assert.NotEmpty(t, xvalidator.Validate(DatetimeTest{Field: "2023-05-32"}))
		assert.Empty(t, xvalidator.Validate(DatetimeTest{Field: ""}))
		assert.NotEmpty(t, xvalidator.Validate(DatetimeTest{Field: "2023-05-01 12:00:00"}))
	})

	t.Run("Length", func(t *testing.T) {
		type LengthTest struct {
			StringField string `xv:"len=5"`
			SliceField  []int  `xv:"len=3"`
		}

		assert.Empty(t, xvalidator.Validate(LengthTest{StringField: "12345", SliceField: []int{1, 2, 3}}))
		assert.NotEmpty(t, xvalidator.Validate(LengthTest{StringField: "1234", SliceField: []int{1, 2}}))
		assert.NotEmpty(t, xvalidator.Validate(LengthTest{StringField: "123456", SliceField: []int{1, 2, 3, 4}}))
	})

	t.Run("Invalid validator", func(t *testing.T) {
		type InvalidTest struct {
			Field string `xv:"invalid_validator"`
		}

		errors := xvalidator.Validate(InvalidTest{Field: "test"})
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "unknown validator")
	})

	t.Run("Multiple validators", func(t *testing.T) {
		type MultiTest struct {
			Field string `xv:"required,min=3,max=10,alphanum"`
		}

		assert.Empty(t, xvalidator.Validate(MultiTest{Field: "abc123"}))
		assert.NotEmpty(t, xvalidator.Validate(MultiTest{Field: "ab"}))
		assert.NotEmpty(t, xvalidator.Validate(MultiTest{Field: "abcdefghijk"}))
		assert.NotEmpty(t, xvalidator.Validate(MultiTest{Field: "abc-123"}))
		assert.NotEmpty(t, xvalidator.Validate(MultiTest{Field: ""}))
	})

	t.Run("Nested structs", func(t *testing.T) {
		type NestedStruct struct {
			InnerField string `xv:"required,min=3"`
		}
		type OuterStruct struct {
			OuterField string `xv:"max=5"`
			Nested     NestedStruct
		}

		assert.Empty(t, xvalidator.Validate(OuterStruct{OuterField: "abc", Nested: NestedStruct{InnerField: "1234"}}))
		assert.NotEmpty(t, xvalidator.Validate(OuterStruct{OuterField: "abcdef", Nested: NestedStruct{InnerField: "12"}}))
	})

	t.Run("Pointer fields", func(t *testing.T) {
		type PointerTest struct {
			IntPtr    *int    `xv:"required,min=0"`
			StringPtr *string `xv:"required,min=3"`
		}

		intVal := 5
		strVal := "test"
		assert.Empty(t, xvalidator.Validate(PointerTest{IntPtr: &intVal, StringPtr: &strVal}))

		intZero := 0
		strShort := "ab"
		assert.NotEmpty(t, xvalidator.Validate(PointerTest{IntPtr: &intZero, StringPtr: &strShort}))

		assert.NotEmpty(t, xvalidator.Validate(PointerTest{IntPtr: nil, StringPtr: nil}))
	})

	t.Run("Custom error messages", func(t *testing.T) {
		type CustomErrorTest struct {
			Field string `xv:"required,min=3"`
		}

		errors := xvalidator.Validate(CustomErrorTest{Field: ""})
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "Field:")
		assert.Contains(t, errors[0].Error(), "This field is required")

		errors = xvalidator.Validate(CustomErrorTest{Field: "a"})
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "Field:")
		assert.Contains(t, errors[0].Error(), "Must be at least 3")
	})
}

func TestValidateNilAndNonStruct(t *testing.T) {
	assert.NotEmpty(t, xvalidator.Validate(nil))
	assert.NotEmpty(t, xvalidator.Validate("not a struct"))
	assert.NotEmpty(t, xvalidator.Validate(123))

	var nilPtr *struct{}
	assert.NotEmpty(t, xvalidator.Validate(nilPtr))
}
