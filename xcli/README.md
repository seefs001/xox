# xcli

`xcli` is a powerful and flexible CLI application framework for Go. It provides a simple way to create command-line interfaces with support for commands, subcommands, flags, and error handling.

## Features

- Easy-to-use API for creating CLI applications
- Support for commands, subcommands, and aliases
- Customizable flags for both the application and individual commands
- Built-in help and version commands
- Error handling and recovery with custom error handlers
- Colorized output
- Debug mode for performance analysis
- Before and after run hooks
- Custom handlers for unknown commands
- Command suggestions for misspelled commands

## Installation

To install `xcli`, use the following command:

```bash
go get github.com/seefs001/xox/xcli
```

## Basic Usage

Here's a basic example of how to use `xcli`:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/seefs001/xox/xcli"
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

#### (a *App) SetUnknownCommandHandler(handler func(ctx context.Context, cmdName string, args []string) error)

Sets a custom handler for unknown commands.

```go
app.SetUnknownCommandHandler(func(ctx context.Context, cmdName string, args []string) error {
    fmt.Printf("Unknown command: %s\n", cmdName)
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

## Advanced Features

### Subcommands

You can create nested command structures using subcommands:

```go
mathCmd := &xcli.Command{
    Name:        "math",
    Description: "Perform mathematical operations",
    SubCommands: make(map[string]*xcli.Command),
}

mathCmd.SubCommands["add"] = &xcli.Command{
    Name:        "add",
    Description: "Add two numbers",
    Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
        // Implementation
        return nil
    },
}

app.AddCommand(mathCmd)
```

### Custom Output

You can customize help and error output:

```go
app.CustomHelpPrinter = func(app *xcli.App) {
    // Custom help printing logic
}

app.CustomErrorPrinter = func(err error) {
    // Custom error printing logic
}

app.CustomCommandHelpPrinter = func(app *xcli.App, cmdName string) {
    // Custom command help printing logic
}
```

## Example

For a more comprehensive example demonstrating various features of `xcli`, please refer to the `examples/xcli_example/main.go` file in the repository.

## Contributing

Contributions to `xcli` are welcome! Please feel free to submit issues, fork the repository and send pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
