package xcolor

import (
	"fmt"
	"os"
	"strings"
)

// ColorCode represents ANSI color codes
type ColorCode string

const (
	Reset  ColorCode = "\033[0m"
	Red    ColorCode = "\033[31m"
	Green  ColorCode = "\033[32m"
	Yellow ColorCode = "\033[33m"
	Blue   ColorCode = "\033[34m"
	Purple ColorCode = "\033[35m"
	Cyan   ColorCode = "\033[36m"
	White  ColorCode = "\033[37m"
	Bold   ColorCode = "\033[1m"
	Italic ColorCode = "\033[3m"
)

var (
	colorEnabled = true
)

// EnableColor enables or disables color output
func EnableColor(enable bool) {
	colorEnabled = enable
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return colorEnabled
}

// Colorize applies the given color to the text if color is enabled
func Colorize(color ColorCode, text string) string {
	if colorEnabled {
		return string(color) + text + string(Reset)
	}
	return text
}

// Print prints text with the specified color if color is enabled
func Print(color ColorCode, format string, a ...interface{}) {
	if colorEnabled {
		fmt.Printf(string(color)+format+string(Reset), a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// Println prints text with the specified color if color is enabled, followed by a newline
func Println(color ColorCode, format string, a ...interface{}) {
	Print(color, format+"\n", a...)
}

// Sprint returns a string with the specified color if color is enabled
func Sprint(color ColorCode, format string, a ...interface{}) string {
	return Colorize(color, fmt.Sprintf(format, a...))
}

// IsTerminal returns true if the given file descriptor is a terminal
func IsTerminal(fd uintptr) bool {
	switch fd {
	case os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd():
		return true
	}
	return false
}

// AutoEnableColor automatically enables color if the output is a terminal
func AutoEnableColor() {
	EnableColor(IsTerminal(os.Stdout.Fd()))
}

// StripColor removes ANSI color codes from the given string
func StripColor(s string) string {
	return strings.ReplaceAll(s, "\033[", "")
}

// ColorizeMulti applies multiple colors to the text
func ColorizeMulti(colors []ColorCode, text string) string {
	if !colorEnabled {
		return text
	}
	colorStr := ""
	for _, color := range colors {
		colorStr += string(color)
	}
	return colorStr + text + string(Reset)
}

// PrintMulti prints text with multiple specified colors if color is enabled
func PrintMulti(colors []ColorCode, format string, a ...interface{}) {
	if colorEnabled {
		colorStr := ""
		for _, color := range colors {
			colorStr += string(color)
		}
		fmt.Printf(colorStr+format+string(Reset), a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// PrintlnMulti prints text with multiple specified colors if color is enabled, followed by a newline
func PrintlnMulti(colors []ColorCode, format string, a ...interface{}) {
	PrintMulti(colors, format+"\n", a...)
}

// SprintMulti returns a string with multiple specified colors if color is enabled
func SprintMulti(colors []ColorCode, format string, a ...interface{}) string {
	return ColorizeMulti(colors, fmt.Sprintf(format, a...))
}

// Rainbow returns a string with each character in a different color
func Rainbow(text string) string {
	if !colorEnabled {
		return text
	}
	colors := []ColorCode{Red, Yellow, Green, Cyan, Blue, Purple}
	result := ""
	for i, char := range text {
		result += string(colors[i%len(colors)]) + string(char)
	}
	return result + string(Reset)
}

// PrintRainbow prints text with each character in a different color
func PrintRainbow(format string, a ...interface{}) {
	fmt.Print(Rainbow(fmt.Sprintf(format, a...)))
}

// PrintlnRainbow prints text with each character in a different color, followed by a newline
func PrintlnRainbow(format string, a ...interface{}) {
	PrintRainbow(format+"\n", a...)
}
