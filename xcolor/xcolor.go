package xcolor

import (
	"fmt"
	"io"
	"os"
	"regexp"
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
	None   ColorCode = ""
)

var (
	colorEnabled   = true
	ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
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
	if colorEnabled {
		fmt.Print(string(color) + fmt.Sprintf(format, a...) + string(Reset) + "\n")
	} else {
		fmt.Printf(format+"\n", a...)
	}
}

// Sprint returns a string with the specified color if color is enabled
func Sprint(color ColorCode, format string, a ...interface{}) string {
	return Colorize(color, fmt.Sprintf(format, a...))
}

// IsTerminal returns true if the given file descriptor is a terminal
func IsTerminal(fd uintptr) bool {
	// For test environment, always return true for standard fds
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "" ||
		os.Getenv("TESTING") != "" ||
		os.Getenv("TEST_TERMINAL") != "" {
		// Only return true for stdin/stdout/stderr
		return fd == os.Stdin.Fd() || fd == os.Stdout.Fd() || fd == os.Stderr.Fd()
	}

	var file *os.File
	switch fd {
	case os.Stdin.Fd():
		file = os.Stdin
	case os.Stdout.Fd():
		file = os.Stdout
	case os.Stderr.Fd():
		file = os.Stderr
	default:
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	mode := info.Mode()
	return (mode & os.ModeCharDevice) != 0
}

func init() {
	if len(os.Args) > 0 && os.Args[0] != "" &&
		(os.Getenv("GOTEST") != "" || os.Args[0][len(os.Args[0])-5:] == ".test") {
		os.Setenv("TEST_TERMINAL", "1")
	}
}

// AutoEnableColor automatically enables color if the output is a terminal
func AutoEnableColor() {
	EnableColor(IsTerminal(os.Stdout.Fd()))
}

// StripColor removes ANSI color codes from the given string
func StripColor(s string) string {
	return ansiColorRegex.ReplaceAllString(s, "")
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
	if colorEnabled {
		colorStr := ""
		for _, color := range colors {
			colorStr += string(color)
		}
		fmt.Print(colorStr + fmt.Sprintf(format, a...) + string(Reset) + "\n")
	} else {
		fmt.Printf(format+"\n", a...)
	}
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
	fmt.Println(Rainbow(fmt.Sprintf(format, a...)))
}

// Fprintf writes formatted output to an io.Writer with the specified color
func Fprintf(w io.Writer, color ColorCode, format string, a ...interface{}) (n int, err error) {
	if w == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}
	if colorEnabled {
		return fmt.Fprintf(w, string(color)+format+string(Reset), a...)
	}
	return fmt.Fprintf(w, format, a...)
}

// FprintfMulti writes formatted output with multiple colors to an io.Writer
func FprintfMulti(w io.Writer, colors []ColorCode, format string, a ...interface{}) (n int, err error) {
	if w == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}
	if colorEnabled && len(colors) > 0 {
		colorStr := ""
		for _, color := range colors {
			colorStr += string(color)
		}
		return fmt.Fprintf(w, colorStr+format+string(Reset), a...)
	}
	return fmt.Fprintf(w, format, a...)
}

// SafeColorize applies the given color to the text if color is enabled and text is not empty
func SafeColorize(color ColorCode, text string) string {
	if text == "" {
		return ""
	}
	return Colorize(color, text)
}

// ColorizeWriter returns an io.Writer that writes colored output
type ColorWriter struct {
	w     io.Writer
	color ColorCode
}

func NewColorWriter(w io.Writer, color ColorCode) *ColorWriter {
	return &ColorWriter{w: w, color: color}
}

func (cw *ColorWriter) Write(p []byte) (n int, err error) {
	if cw.w == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}
	if colorEnabled {
		return fmt.Fprintf(cw.w, string(cw.color)+"%s"+string(Reset), string(p))
	}
	return cw.w.Write(p)
}
