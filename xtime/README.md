# xtime

xtime is a Go package that provides extended time-related functionality, building upon the standard `time` package. It offers various utility functions for working with dates, times, and durations.

## Installation

To install xtime, use `go get`:

```bash
go get github.com/seefs001/xox/xtime
```

## Usage

Import the package in your Go code:

```go
import "github.com/seefs001/xox/xtime"
```

## Features

### Date Calculations

#### IsLeapYear

Checks if a given year is a leap year.

```go
isLeap := xtime.IsLeapYear(2024) // true
```

#### DaysInMonth

Returns the number of days in a given month and year.

```go
days := xtime.DaysInMonth(2023, 2) // 28
```

#### DaysBetween

Calculates the number of days between two dates.

```go
date1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
date2 := time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC)
days := xtime.DaysBetween(date1, date2) // 9
```

#### IsSameDay

Checks if two times are on the same day.

```go
date1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
date2 := time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC)
same := xtime.IsSameDay(date1, date2) // true
```

### Time Manipulation

#### StartOfDay

Returns the start of the day for a given time.

```go
now := time.Now()
start := xtime.StartOfDay(now) // 00:00:00.000000000
```

#### EndOfDay

Returns the end of the day for a given time.

```go
now := time.Now()
end := xtime.EndOfDay(now) // 23:59:59.999999999
```

#### StartOfWeek

Returns the start of the week for a given time. The week start day is configurable.

```go
now := time.Now()
start := xtime.StartOfWeek(now, time.Sunday)
```

#### EndOfWeek

Returns the end of the week for a given time. The week end day is configurable.

```go
now := time.Now()
end := xtime.EndOfWeek(now, time.Saturday)
```

#### StartOfMonth

Returns the start of the month for a given time.

```go
now := time.Now()
start := xtime.StartOfMonth(now)
```

#### EndOfMonth

Returns the end of the month for a given time.

```go
now := time.Now()
end := xtime.EndOfMonth(now)
```

#### StartOfYear

Returns the start of the year for a given time.

```go
now := time.Now()
start := xtime.StartOfYear(now)
```

#### EndOfYear

Returns the end of the year for a given time.

```go
now := time.Now()
end := xtime.EndOfYear(now)
```

#### AddDate

Adds the specified number of years, months, and days to a given time.

```go
now := time.Now()
future := xtime.AddDate(now, 1, 2, 3) // 1 year, 2 months, 3 days from now
```

#### TimeIn

Returns the time in the specified timezone.

```go
now := time.Now()
nyTime, err := xtime.TimeIn(now, "America/New_York")
if err != nil {
    // handle error
}
```

### Duration Handling

#### FormatDuration

Formats a duration into a human-readable string.

```go
d := 36*time.Hour + 15*time.Minute + 30*time.Second
formatted := xtime.FormatDuration(d) // "1 day 12 hours 15 minutes 30 seconds"
```

#### ParseDuration

Parses a duration string and returns the time.Duration. Supports years (y), months (M), weeks (w), days (d), hours (h), minutes (m), and seconds (s).

```go
d, err := xtime.ParseDuration("1y2M3w4d5h6m7s")
if err != nil {
    // handle error
}
```

## Constants

xtime provides constants for common durations:

- `xtime.Nanosecond`
- `xtime.Microsecond`
- `xtime.Millisecond`
- `xtime.Second`
- `xtime.Minute`
- `xtime.Hour`
- `xtime.Day`
- `xtime.Week`
- `xtime.Month` (approximate)
- `xtime.Year` (approximate)

These can be used in duration calculations and comparisons.

## Error Handling

Most functions in xtime return an error as a second return value when applicable. Always check for errors when using functions that may fail, such as `ParseDuration` or `TimeIn`.

## Thread Safety

The xtime package is designed to be thread-safe. All functions can be safely called from multiple goroutines concurrently.

## Performance Considerations

While xtime provides convenient functions for time manipulations, some operations (like `DaysBetween` for large date ranges) may be computationally expensive. Consider caching results or using more efficient algorithms for performance-critical applications.

## Examples

Here are some additional examples to demonstrate the usage of xtime functions:

### Working with Weeks

```go
now := time.Now()
weekStart := xtime.StartOfWeek(now, time.Monday)
weekEnd := xtime.EndOfWeek(now, time.Sunday)
fmt.Printf("Week starts on %s and ends on %s\n", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
```

### Calculating Age

```go
birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
now := time.Now()
age := now.Year() - birthDate.Year()
if now.YearDay() < birthDate.YearDay() {
    age--
}
fmt.Printf("Age: %d years\n", age)
```

### Working with Different Timezones

```go
now := time.Now()
locations := []string{"America/New_York", "Europe/London", "Asia/Tokyo"}

for _, loc := range locations {
    t, err := xtime.TimeIn(now, loc)
    if err != nil {
        fmt.Printf("Error for %s: %v\n", loc, err)
        continue
    }
    fmt.Printf("Time in %s: %s\n", loc, t.Format("2006-01-02 15:04:05"))
}
```

### Parsing Complex Durations

```go
durations := []string{
    "1y2M3w4d",
    "5h30m",
    "1d12h",
    "2w3d4h5m6s",
}

for _, d := range durations {
    parsed, err := xtime.ParseDuration(d)
    if err != nil {
        fmt.Printf("Error parsing %s: %v\n", d, err)
        continue
    }
    fmt.Printf("Parsed %s as %s\n", d, xtime.FormatDuration(parsed))
}
```

## Contributing

Contributions to xtime are welcome! Please feel free to submit issues, fork the repository and send pull requests!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
