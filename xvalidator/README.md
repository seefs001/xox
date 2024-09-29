# xvalidator

xvalidator is a powerful and flexible struct validation library for Go. It provides a simple way to validate struct fields using tags, supporting a wide range of validation rules.

## Features

- Tag-based validation
- Supports various data types including strings, numbers, slices, and time.Time
- Customizable error messages
- Easy to extend with custom validators

## Installation

```bash
go get github.com/seefs001/xox/xvalidator
```

## Usage

To use xvalidator, add `xv` tags to your struct fields and call the `Validate` function.

```go
import "github.com/seefs001/xox/xvalidator"

type User struct {
    Name     string    `xv:"required,min=2,max=50"`
    Age      int       `xv:"required,min=18,max=120"`
    Email    string    `xv:"email"`
    Password string    `xv:"min=8,regexp=^[a-zA-Z0-9]+$"`
    Role     string    `xv:"in=admin|user|guest"`
    Score    float64   `xv:"min=0,max=100"`
    Tags     []string  `xv:"min=1,max=5"`
    Active   bool      `xv:"required"`
    Created  time.Time `xv:"required,datetime=2006-01-02"`
}

user := User{
    Name:     "John Doe",
    Age:      30,
    Email:    "john@example.com",
    Password: "password123",
    Role:     "user",
    Score:    85.5,
    Tags:     []string{"tag1", "tag2"},
    Active:   true,
    Created:  time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
}

errors := xvalidator.Validate(user)
if errors != nil {
    for _, err := range errors {
        fmt.Println(err)
    }
}
```

## Validation Rules

xvalidator supports the following validation rules:

- `required`: Field must not be empty
- `min`: Minimum value (for numbers) or length (for strings, slices)
- `max`: Maximum value (for numbers) or length (for strings, slices)
- `email`: Valid email format
- `regexp`: Match a regular expression
- `in`: Value must be one of the specified options
- `notin`: Value must not be one of the specified options
- `len`: Exact length (for strings, slices)
- `alpha`: Only alphabetic characters
- `alphanum`: Only alphanumeric characters
- `numeric`: Only numeric values
- `datetime`: Valid date/time format

## API Reference

### func Validate(v interface{}) []error

Validate performs validation on a struct or pointer to struct. It returns a slice of errors if any validation fails.

```go
errors := xvalidator.Validate(user)
```

### type ValidationError

ValidationError represents an error that occurred during validation.

```go
type ValidationError struct {
    Field   string
    Message string
}
```

#### func (e ValidationError) Error() string

Error returns a string representation of the validation error.

## Examples

### Basic Usage

```go
type Product struct {
    Name  string  `xv:"required,min=3,max=50"`
    Price float64 `xv:"required,min=0"`
    SKU   string  `xv:"required,alphanum,len=10"`
}

product := Product{
    Name:  "Widget",
    Price: 9.99,
    SKU:   "WIDGET1234",
}

errors := xvalidator.Validate(product)
if errors != nil {
    for _, err := range errors {
        fmt.Println(err)
    }
}
```

### Custom Validation

You can combine multiple validation rules for more complex validations:

```go
type Account struct {
    Username string `xv:"required,min=5,max=20,alphanum"`
    Password string `xv:"required,min=8,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$"`
}

account := Account{
    Username: "johndoe123",
    Password: "StrongP@ss1",
}

errors := xvalidator.Validate(account)
if errors != nil {
    for _, err := range errors {
        fmt.Println(err)
    }
}
```

This example demonstrates a more complex password validation using a regular expression to ensure it contains at least one uppercase letter, one lowercase letter, one number, and one special character.
