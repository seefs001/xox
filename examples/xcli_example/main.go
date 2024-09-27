package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/seefs001/xox/xcli"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.Info("Starting CLI application example")

	// Create the root command
	rootCmd := xcli.NewCommand("app", "A comprehensive CLI application example", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing root command")
		fmt.Println("Welcome to the CLI application. Use 'hello', 'goodbye', 'add', or 'version' subcommands.")
		return nil
	})

	// Set version for the root command
	rootCmd.SetVersion("1.0.0")
	xlog.Info("Root command version set", "version", rootCmd.Version)

	// Create a hello subcommand
	helloCmd := xcli.NewCommand("hello", "Prints a greeting message", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing hello subcommand")
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
	xlog.Info("Added greeting flag to hello command")

	// Set aliases for the hello command
	helloCmd.SetAliases("hi", "greet")
	xlog.Info("Set aliases for hello command", "aliases", []string{"hi", "greet"})

	// Add the hello subcommand to the root command
	rootCmd.AddSubcommand(helloCmd)
	xlog.Info("Added hello subcommand to root command")

	// Create a goodbye subcommand
	goodbyeCmd := xcli.NewCommand("goodbye", "Prints a farewell message", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing goodbye subcommand")
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
	xlog.Info("Added farewell flag to goodbye command")

	// Add the goodbye subcommand to the root command
	rootCmd.AddSubcommand(goodbyeCmd)
	xlog.Info("Added goodbye subcommand to root command")

	// Create an add subcommand for adding numbers
	addCmd := xcli.NewCommand("add", "Adds two or more numbers", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing add subcommand")
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
	xlog.Info("Added add subcommand to root command")

	// Create a hidden debug subcommand
	debugCmd := xcli.NewCommand("debug", "Prints debug information", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing debug subcommand")
		fmt.Println("Debug mode activated")
		return nil
	})
	debugCmd.SetHidden(true)
	rootCmd.AddSubcommand(debugCmd)
	xlog.Info("Added hidden debug subcommand to root command")

	// Create a version subcommand
	versionCmd := xcli.NewCommand("version", "Prints the application version", func(ctx context.Context, cmdCtx *xcli.CommandContext) error {
		xlog.Info("Executing version subcommand")
		fmt.Printf("App version: %s\n", rootCmd.Version)
		return nil
	})
	rootCmd.AddSubcommand(versionCmd)
	xlog.Info("Added version subcommand to root command")

	// Enable color output
	xcli.EnableColor(true)
	xlog.Info("Enabled color output")

	// Enable debug mode (for demonstration purposes)
	xcli.EnableDebug(true)
	xlog.Info("Enabled debug mode")

	// Execute the root command
	xlog.Info("Executing root command with arguments", "args", os.Args[1:])
	xcli.Execute(rootCmd, os.Args[1:])

	xlog.Info("CLI application example completed")
}
