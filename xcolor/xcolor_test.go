package xcolor_test

import (
	"io"
	"os"
	"testing"

	"github.com/seefs001/xox/xcolor"
	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	xcolor.EnableColor(true)
	assert.Equal(t, "\033[31mHello\033[0m", xcolor.Colorize(xcolor.Red, "Hello"))

	xcolor.EnableColor(false)
	assert.Equal(t, "Hello", xcolor.Colorize(xcolor.Red, "Hello"))
}

func TestPrint(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.Print(xcolor.Blue, "Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	assert.Equal(t, "\033[34mHello World\033[0m", string(out))
}

func TestPrintln(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.Println(xcolor.Green, "Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	assert.Equal(t, "\033[32mHello World\033[0m\n", string(out))
}

func TestSprint(t *testing.T) {
	xcolor.EnableColor(true)
	assert.Equal(t, "\033[33mHello World\033[0m", xcolor.Sprint(xcolor.Yellow, "Hello %s", "World"))

	xcolor.EnableColor(false)
	assert.Equal(t, "Hello World", xcolor.Sprint(xcolor.Yellow, "Hello %s", "World"))
}

func TestIsTerminal(t *testing.T) {
	assert.True(t, xcolor.IsTerminal(os.Stdout.Fd()))
	assert.True(t, xcolor.IsTerminal(os.Stdin.Fd()))
	assert.True(t, xcolor.IsTerminal(os.Stderr.Fd()))

	f, _ := os.Create("test.txt")
	defer f.Close()
	defer os.Remove("test.txt")
	assert.False(t, xcolor.IsTerminal(f.Fd()))
}

func TestAutoEnableColor(t *testing.T) {
	xcolor.AutoEnableColor()
	assert.Equal(t, xcolor.IsTerminal(os.Stdout.Fd()), xcolor.IsColorEnabled())
}

func TestStripColor(t *testing.T) {
	colored := "\033[31mRed\033[0m \033[32mGreen\033[0m"
	assert.Equal(t, "Red Green", xcolor.StripColor(colored))
}

func TestColorizeMulti(t *testing.T) {
	xcolor.EnableColor(true)
	assert.Equal(t, "\033[31m\033[1mBold Red\033[0m", xcolor.ColorizeMulti([]xcolor.ColorCode{xcolor.Red, xcolor.Bold}, "Bold Red"))

	xcolor.EnableColor(false)
	assert.Equal(t, "Bold Red", xcolor.ColorizeMulti([]xcolor.ColorCode{xcolor.Red, xcolor.Bold}, "Bold Red"))
}

func TestPrintMulti(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.PrintMulti([]xcolor.ColorCode{xcolor.Blue, xcolor.Bold}, "Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	assert.Equal(t, "\033[34m\033[1mHello World\033[0m", string(out))
}

func TestPrintlnMulti(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.PrintlnMulti([]xcolor.ColorCode{xcolor.Green, xcolor.Italic}, "Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	assert.Equal(t, "\033[32m\033[3mHello World\033[0m\n", string(out))
}

func TestSprintMulti(t *testing.T) {
	xcolor.EnableColor(true)
	assert.Equal(t, "\033[33m\033[1mHello World\033[0m", xcolor.SprintMulti([]xcolor.ColorCode{xcolor.Yellow, xcolor.Bold}, "Hello %s", "World"))

	xcolor.EnableColor(false)
	assert.Equal(t, "Hello World", xcolor.SprintMulti([]xcolor.ColorCode{xcolor.Yellow, xcolor.Bold}, "Hello %s", "World"))
}

func TestRainbow(t *testing.T) {
	xcolor.EnableColor(true)
	expected := "\033[31mH\033[33me\033[32ml\033[36ml\033[34mo\033[35m \033[31mW\033[33mo\033[32mr\033[36ml\033[34md\033[0m"
	assert.Equal(t, expected, xcolor.Rainbow("Hello World"))

	xcolor.EnableColor(false)
	assert.Equal(t, "Hello World", xcolor.Rainbow("Hello World"))
}

func TestPrintRainbow(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.PrintRainbow("Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	expected := "\033[31mH\033[33me\033[32ml\033[36ml\033[34mo\033[35m \033[31mW\033[33mo\033[32mr\033[36ml\033[34md\033[0m"
	assert.Equal(t, expected, string(out))
}

func TestPrintlnRainbow(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	xcolor.EnableColor(true)
	xcolor.PrintlnRainbow("Hello %s", "World")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	expected := "\033[31mH\033[33me\033[32ml\033[36ml\033[34mo\033[35m \033[31mW\033[33mo\033[32mr\033[36ml\033[34md\033[0m\n"
	assert.Equal(t, expected, string(out))
}

func TestEnableColor(t *testing.T) {
	xcolor.EnableColor(true)
	assert.True(t, xcolor.IsColorEnabled())

	xcolor.EnableColor(false)
	assert.False(t, xcolor.IsColorEnabled())
}

func TestIsColorEnabled(t *testing.T) {
	xcolor.EnableColor(true)
	assert.True(t, xcolor.IsColorEnabled())

	xcolor.EnableColor(false)
	assert.False(t, xcolor.IsColorEnabled())
}
