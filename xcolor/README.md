# xcolor

xcolor is a Go package that provides easy-to-use color output functionality for terminal applications. It supports ANSI color codes and offers various methods for colorizing text, printing colored output, and managing color settings.

## Features

- Simple API for adding color to terminal output
- Support for basic ANSI colors and text styles
- Automatic color detection for terminals
- Color stripping functionality
- Rainbow text output

## Installation

To install xcolor, use `go get`:

```bash
go get github.com/seefs001/xox/xcolor
```

## Usage

### Basic Coloring

```go
import "github.com/seefs001/xox/xcolor"

// Print colored text
xcolor.Print(xcolor.Red, "This is red text")

// Print colored text with a newline
xcolor.Println(xcolor.Blue, "This is blue text")

// Get colored string
coloredText := xcolor.Sprint(xcolor.Green, "This is green text")

// Colorize text
colorizedText := xcolor.Colorize(xcolor.Yellow, "This is yellow text")
```

### Multiple Colors and Styles

```go
// Apply multiple colors/styles
xcolor.PrintMulti([]xcolor.ColorCode{xcolor.Red, xcolor.Bold}, "Bold red text")

// Get string with multiple colors/styles
multiColored := xcolor.SprintMulti([]xcolor.ColorCode{xcolor.Blue, xcolor.Italic}, "Italic blue text")
```

### Rainbow Text

```go
// Print rainbow text
xcolor.PrintRainbow("This text is a rainbow")

// Print rainbow text with newline
xcolor.PrintlnRainbow("This rainbow text has a newline")

// Get rainbow string
rainbowText := xcolor.Rainbow("Another rainbow text")
```

### Color Management

```go
// Enable or disable color output
xcolor.EnableColor(true)

// Check if color is enabled
isEnabled := xcolor.IsColorEnabled()

// Automatically enable color based on terminal capability
xcolor.AutoEnableColor()

// Strip color codes from a string
stripped := xcolor.StripColor("\033[31mRed text\033[0m")
```

### Terminal Detection

```go
// Check if stdout is a terminal
isTerm := xcolor.IsTerminal(os.Stdout.Fd())
```

## Available Colors and Styles

- `Reset`
- `Red`
- `Green`
- `Yellow`
- `Blue`
- `Purple`
- `Cyan`
- `White`
- `Bold`
- `Italic`

## Notes

- Color output can be globally enabled or disabled using `EnableColor(bool)`.
- `AutoEnableColor()` can be used to automatically enable color based on terminal capability.
- The `Rainbow` function cycles through colors for each character in the input string.
- Use `StripColor` to remove ANSI color codes from strings.

## Example

```go
package main

import (
    "fmt"
    "github.com/seefs001/xox/xcolor"
)

func main() {
    xcolor.AutoEnableColor()

    xcolor.Println(xcolor.Red, "This is red text")
    xcolor.Println(xcolor.Blue, "This is %s text", "blue")

    fmt.Println(xcolor.Sprint(xcolor.Green, "This is green text"))

    xcolor.PrintMulti([]xcolor.ColorCode{xcolor.Yellow, xcolor.Bold}, "This is bold yellow text")

    xcolor.PrintlnRainbow("This is a rainbow!")

    colorized := xcolor.Colorize(xcolor.Cyan, "Cyan text")
    stripped := xcolor.StripColor(colorized)
    fmt.Println("Stripped:", stripped)
}
```

This example demonstrates various features of the xcolor package, including basic coloring, multi-color output, rainbow text, and color stripping.
