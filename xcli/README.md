# xcli

`xcli` is a powerful and flexible CLI application framework for Go. It provides a simple way to create command-line interfaces with support for commands, subcommands, flags, and error handling.

## Features

- Easy-to-use API for creating CLI applications
- Support for commands and subcommands
- Customizable flags for both the application and individual commands
- Built-in help and version commands
- Error handling and recovery
- Colorized output
- Debug mode for performance analysis

## Installation

To install `xcli`, use the following command:

```bash
go get github.com/seefs001/xox/seefspkg/xcli
```

## Usage

Here's a basic example of how to use `xcli`:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/seefs001/xox/seefspkg/xcli"
)

func main() {
    app := xcli.NewApp("myapp", "A sample CLI application", "1.0.0")

    app.AddCommand(&xcli.Command{
        Name:        "greet",
        Description: "Greet the user",
        Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
            fmt.Println("Hello, user!")
            return nil
        },
    })

    if err := app.Run(context.Background(), os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## API Reference

### App

#### NewApp(name, description, version string) *App

Creates a new CLI application.

```go
app := xcli.NewApp("myapp", "A sample CLI application", "1.0.0")
```

#### (a *App) AddCommand(cmd *Command)

Adds a new command to the application.

```go
app.AddCommand(&xcli.Command{
    Name:        "greet",
    Description: "Greet the user",
    Run:         greetCommand,
})
```

#### (a *App) SetDefaultRun(run func(ctx context.Context, app *App) error)

Sets the default run function for the application.

```go
app.SetDefaultRun(func(ctx context.Context, app *App) error {
    fmt.Println("Default action")
    return nil
})
```

#### (a *App) SetErrorHandler(handler func(error))

Sets a custom error handling function.

```go
app.SetErrorHandler(func(err error) {
    fmt.Fprintf(os.Stderr, "Custom error handler: %v\n", err)
})
```

#### (a *App) SetBeforeRun(before func(ctx context.Context, app *App) error)

Sets a function to run before command execution.

```go
app.SetBeforeRun(func(ctx context.Context, app *App) error {
    fmt.Println("Before run")
    return nil
})
```

#### (a *App) SetAfterRun(after func(ctx context.Context, app *App) error)

Sets a function to run after command execution.

```go
app.SetAfterRun(func(ctx context.Context, app *App) error {
    fmt.Println("After run")
    return nil
})
```

#### (a *App) Run(ctx context.Context, args []string) error

Executes the application.

```go
err := app.Run(context.Background(), os.Args)
```

### Command

```go
type Command struct {
    Name        string
    Description string
    Flags       *flag.FlagSet
    Run         func(ctx context.Context, cmd *Command, args []string) error
    Aliases     []string
    Hidden      bool
    SubCommands map[string]*Command
}
```

### Utility Functions

#### EnableColor(enable bool)

Enables or disables color output.

```go
xcli.EnableColor(true)
```

#### EnableDebug(enable bool)

Enables or disables debug mode.

```go
xcli.EnableDebug(true)
```

## Example

Here's a more comprehensive example demonstrating various features of `xcli`:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/seefs001/xox/seefspkg/xcli"
)

func main() {
    app := xcli.NewApp("myapp", "A sample CLI application", "1.0.0")

    // Add a flag to the main application
    verbose := app.Flags.Bool("verbose", false, "Enable verbose output")

    // Add a command
    greetCmd := &xcli.Command{
        Name:        "greet",
        Description: "Greet the user",
        Aliases:     []string{"g", "hello"},
        Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
            name := "user"
            if len(args) > 0 {
                name = args[0]
            }
            fmt.Printf("Hello, %s!\n", name)
            if *verbose {
                fmt.Println("Verbose mode enabled")
            }
            return nil
        },
    }

    // Add a flag to the greet command
    greetCmd.Flags = flag.NewFlagSet("greet", flag.ExitOnError)
    uppercase := greetCmd.Flags.Bool("uppercase", false, "Print greeting in uppercase")

    app.AddCommand(greetCmd)

    // Set custom error handler
    app.SetErrorHandler(func(err error) {
        fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
    })

    // Enable debug mode
    xcli.EnableDebug(true)

    // Run the application
    if err := app.Run(context.Background(), os.Args); err != nil {
        os.Exit(1)
    }
}
```

This example demonstrates:
- Creating an application with a description and version
- Adding a global flag
- Creating a command with aliases and a local flag
- Using a custom error handler
- Enabling debug mode

To run the application:

```bash
# Run the greet command
./myapp greet John

# Use an alias
./myapp g John

# Use flags
./myapp --verbose greet John --uppercase
```
