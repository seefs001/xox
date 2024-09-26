package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/seefs001/xox/xcli"
)

func main() {
	// Create the root command
	rootCmd := xcli.NewCommand("app", "A comprehensive CLI application example", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		fmt.Println("Welcome to the CLI application. Use 'hello', 'goodbye', 'add', or 'version' subcommands.")
		return nil
	})

	// Set version for the root command
	rootCmd.SetVersion("1.0.0")

	// Create a hello subcommand
	helloCmd := xcli.NewCommand("hello", "Prints a greeting message", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		greeting := cmdCtx.Flags.Lookup("greeting").Value.String()
		if len(cmdCtx.Args) < 1 {
			return fmt.Errorf("name is required")
		}
		name := cmdCtx.Args[0]
		fmt.Printf("%s, %s!\n", greeting, name)
		return nil
	})

	// Add flags to the hello command
	helloCmd.Flags.String("greeting", "Hello", "Custom greeting message")

	// Set aliases for the hello command
	helloCmd.SetAliases("hi", "greet")

	// Add the hello subcommand to the root command
	rootCmd.AddSubcommand(helloCmd)

	// Create a goodbye subcommand
	goodbyeCmd := xcli.NewCommand("goodbye", "Prints a farewell message", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		farewell := cmdCtx.Flags.Lookup("farewell").Value.String()
		if len(cmdCtx.Args) < 1 {
			return fmt.Errorf("name is required")
		}
		name := cmdCtx.Args[0]
		fmt.Printf("%s, %s!\n", farewell, name)
		return nil
	})

	// Add flags to the goodbye command
	goodbyeCmd.Flags.String("farewell", "Goodbye", "Custom farewell message")

	// Add the goodbye subcommand to the root command
	rootCmd.AddSubcommand(goodbyeCmd)

	// Create an add subcommand for adding numbers
	addCmd := xcli.NewCommand("add", "Adds two or more numbers", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		if len(cmdCtx.Args) < 2 {
			return fmt.Errorf("at least two numbers are required")
		}
		sum := 0
		for _, arg := range cmdCtx.Args {
			num, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid number: %s", arg)
			}
			sum += num
		}
		fmt.Printf("The sum is %d\n", sum)
		return nil
	})

	// Add the add subcommand to the root command
	rootCmd.AddSubcommand(addCmd)

	// Create a hidden debug subcommand
	debugCmd := xcli.NewCommand("debug", "Prints debug information", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		fmt.Println("Debug mode activated")
		return nil
	})
	debugCmd.SetHidden(true)
	rootCmd.AddSubcommand(debugCmd)

	// Create a version subcommand
	versionCmd := xcli.NewCommand("version", "Prints the application version", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		fmt.Printf("App version: %s\n", rootCmd.Version)
		return nil
	})
	rootCmd.AddSubcommand(versionCmd)

	// Enable color output
	xcli.EnableColor(true)

	// Enable debug mode (for demonstration purposes)
	xcli.EnableDebug(true)

	// Execute the root command
	xcli.Execute(rootCmd, os.Args[1:])
}
