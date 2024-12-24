package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/seefs001/xox/xcli"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.SetDefaultLogLevel(slog.LevelError)
	xlog.Info("Starting CLI application example")

	// Create the app
	app := xcli.NewApp("myapp", "A comprehensive CLI application example", "1.0.0")

	// Add global flags
	verbose := app.Flags.Bool("verbose", false, "Enable verbose output")
	logFile := app.Flags.String("log-file", "app.log", "Path to log file")
	debug := app.Flags.Bool("debug", false, "Enable debug mode")

	// Set custom error handler
	app.SetErrorHandler(func(err error) {
		xlog.Errorf("Application error: %v", err)
		os.Exit(1)
	})

	// Set default run function
	app.SetDefaultRun(func(ctx context.Context, app *xcli.App) error {
		fmt.Println("Welcome to the CLI application. Use 'hello', 'goodbye', 'add', or 'version' commands.")
		fmt.Println("For more information, use the --help flag.")
		if *verbose {
			fmt.Println("Verbose mode is enabled.")
			fmt.Printf("Log file: %s\n", *logFile)
		}
		return nil
	})

	// Add hello command
	app.AddCommand(&xcli.Command{
		Name:        "hello",
		Description: "Prints a greeting message",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			name := cmd.Flags.String("name", "World", "Name to greet")
			uppercase := cmd.Flags.Bool("uppercase", false, "Print greeting in uppercase")
			if err := cmd.Flags.Parse(args); err != nil {
				return err
			}

			greeting := fmt.Sprintf("Hello, %s!", *name)
			if *uppercase {
				greeting = strings.ToUpper(greeting)
			}
			fmt.Println(greeting)

			if *verbose {
				xlog.Infof("Executed hello command with name: %s, uppercase: %v", *name, *uppercase)
			}
			return nil
		},
		Aliases: []string{"hi", "greet"},
	})

	// Add goodbye command
	app.AddCommand(&xcli.Command{
		Name:        "goodbye",
		Description: "Prints a farewell message",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			name := cmd.Flags.String("name", "World", "Name to bid farewell")
			if err := cmd.Flags.Parse(args); err != nil {
				return err
			}

			fmt.Printf("Goodbye, %s!\n", *name)

			if *verbose {
				xlog.Infof("Executed goodbye command with name: %s", *name)
			}
			return nil
		},
	})

	// Add add command
	app.AddCommand(&xcli.Command{
		Name:        "add",
		Description: "Adds two or more numbers",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			numbers := cmd.Flags.String("numbers", "", "Comma-separated list of numbers to add")
			if err := cmd.Flags.Parse(args); err != nil {
				return err
			}

			var nums []string
			if *numbers != "" {
				nums = strings.Split(*numbers, ",")
			} else {
				nums = cmd.Flags.Args()
			}

			if len(nums) < 2 {
				return fmt.Errorf("please provide at least two numbers to add")
			}

			sum := 0
			for _, num := range nums {
				n, err := strconv.Atoi(strings.TrimSpace(num))
				if err != nil {
					return fmt.Errorf("invalid number: %s", num)
				}
				sum += n
			}

			fmt.Printf("The sum is %d\n", sum)

			if *verbose {
				xlog.Infof("Executed add command with numbers: %v, sum: %d", nums, sum)
			}
			return nil
		},
	})

	// Add version command
	app.AddCommand(&xcli.Command{
		Name:        "version",
		Description: "Prints the application version",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			fmt.Printf("App version: %s\n", app.Version)
			if *verbose {
				xlog.Info("Executed version command")
			}
			return nil
		},
	})

	// Enable color output
	xcli.EnableColor(true)

	// Set debug mode based on flag
	xcli.EnableDebug(*debug)
	// Run the application
	xlog.Debug("Running the application with arguments", "args", os.Args[1:])
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		xlog.Errorf("Application error: %v", err)
		os.Exit(1)
	}

	xlog.Info("CLI application example completed")

	// Add a new "math" command with subcommands
	mathCmd := &xcli.Command{
		Name:        "math",
		Description: "Perform various mathematical operations",
		SubCommands: make(map[string]*xcli.Command),
	}

	// Add "add" subcommand
	mathCmd.SubCommands["add"] = &xcli.Command{
		Name:        "add",
		Description: "Add two or more numbers",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			numbers := cmd.Flags.String("numbers", "", "Comma-separated list of numbers to add")
			if err := cmd.Flags.Parse(args); err != nil {
				return err
			}

			var nums []string
			if *numbers != "" {
				nums = strings.Split(*numbers, ",")
			} else {
				nums = cmd.Flags.Args()
			}

			if len(nums) < 2 {
				return fmt.Errorf("please provide at least two numbers to add")
			}

			sum := 0
			for _, num := range nums {
				n, err := strconv.Atoi(strings.TrimSpace(num))
				if err != nil {
					return fmt.Errorf("invalid number: %s", num)
				}
				sum += n
			}

			fmt.Printf("The sum is %d\n", sum)

			if *verbose {
				xlog.Infof("Executed math add command with numbers: %v, sum: %d", nums, sum)
			}
			return nil
		},
	}

	// Add "multiply" subcommand
	mathCmd.SubCommands["multiply"] = &xcli.Command{
		Name:        "multiply",
		Description: "Multiply two or more numbers",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			numbers := cmd.Flags.String("numbers", "", "Comma-separated list of numbers to multiply")
			if err := cmd.Flags.Parse(args); err != nil {
				return err
			}

			var nums []string
			if *numbers != "" {
				nums = strings.Split(*numbers, ",")
			} else {
				nums = cmd.Flags.Args()
			}

			if len(nums) < 2 {
				return fmt.Errorf("please provide at least two numbers to multiply")
			}

			product := 1
			for _, num := range nums {
				n, err := strconv.Atoi(strings.TrimSpace(num))
				if err != nil {
					return fmt.Errorf("invalid number: %s", num)
				}
				product *= n
			}

			fmt.Printf("The product is %d\n", product)

			if *verbose {
				xlog.Infof("Executed math multiply command with numbers: %v, product: %d", nums, product)
			}
			return nil
		},
	}

	app.AddCommand(mathCmd)

	xlog.Info("CLI application example completed")
}
