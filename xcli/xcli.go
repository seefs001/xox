package xcli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xcolor"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xlog"
)

var (
	debugMode = false
)

// App represents the main CLI application
type App struct {
	Name         string
	Description  string
	Version      string
	Flags        *flag.FlagSet
	Commands     map[string]*Command
	DefaultRun   func(ctx context.Context, app *App) error
	ErrorHandler func(error)                               // Error handling function
	BeforeRun    func(ctx context.Context, app *App) error // Function to run before command execution
	AfterRun     func(ctx context.Context, app *App) error // Function to run after command execution
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Flags       *flag.FlagSet
	Run         func(ctx context.Context, cmd *Command, args []string) error
	Aliases     []string
	Hidden      bool
	SubCommands map[string]*Command // Support for subcommands
}

// NewApp creates a new CLI application
func NewApp(name, description, version string) *App {
	return &App{
		Name:        name,
		Description: description,
		Version:     version,
		Flags:       flag.NewFlagSet(name, flag.ExitOnError),
		Commands:    make(map[string]*Command),
	}
}

// AddCommand adds a new command to the application
func (a *App) AddCommand(cmd *Command) {
	if _, exists := a.Commands[cmd.Name]; !exists {
		a.Commands[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			if _, exists := a.Commands[alias]; !exists {
				a.Commands[alias] = cmd
			}
		}
	}
}

// SetDefaultRun sets the default run function for the application
func (a *App) SetDefaultRun(run func(ctx context.Context, app *App) error) {
	a.DefaultRun = run
}

// SetErrorHandler sets a custom error handling function
func (a *App) SetErrorHandler(handler func(error)) {
	a.ErrorHandler = handler
}

// SetBeforeRun sets a function to run before command execution
func (a *App) SetBeforeRun(before func(ctx context.Context, app *App) error) {
	a.BeforeRun = before
}

// SetAfterRun sets a function to run after command execution
func (a *App) SetAfterRun(after func(ctx context.Context, app *App) error) {
	a.AfterRun = after
}

// Run executes the application
func (a *App) Run(ctx context.Context, args []string) error {
	defer func() {
		if r := recover(); r != nil {
			xlog.Errorf("Panic recovered: %v", r)
			if a.ErrorHandler != nil {
				a.ErrorHandler(fmt.Errorf("panic: %v", r))
			}
		}
	}()

	if len(args) > 1 && args[1] == "--help" {
		a.PrintCommandHelp(args[0])
		return nil
	}

	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		a.PrintHelp()
		return nil
	}

	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("%s version %s\n", a.Name, a.Version)
		return nil
	}

	if err := a.Flags.Parse(args); err != nil {
		return a.handleError(xerror.Wrap(err, "failed to parse flags"))
	}

	// Run BeforeRun function if set
	if a.BeforeRun != nil {
		if err := a.BeforeRun(ctx, a); err != nil {
			return a.handleError(err)
		}
	}

	var err error
	if len(a.Flags.Args()) > 0 {
		// Skip the first argument (program name)
		cmdArgs := a.Flags.Args()[1:]
		if len(cmdArgs) > 0 {
			if cmd, exists := a.Commands[cmdArgs[0]]; exists {
				err = a.runCommand(ctx, cmd, cmdArgs[1:])
			} else {
				err = fmt.Errorf("unknown command: %s", cmdArgs[0])
			}
		} else if a.DefaultRun != nil {
			startTime := time.Now()
			err = a.DefaultRun(ctx, a)
			if debugMode {
				xlog.Infof("App execution time: %v", time.Since(startTime))
			}
		} else {
			// If no DefaultRun is set, print help
			a.PrintHelp()
		}
	} else if a.DefaultRun != nil {
		startTime := time.Now()
		err = a.DefaultRun(ctx, a)
		if debugMode {
			xlog.Infof("App execution time: %v", time.Since(startTime))
		}
	} else {
		// If no DefaultRun is set, print help
		a.PrintHelp()
	}

	// Run AfterRun function if set
	if a.AfterRun != nil {
		if afterErr := a.AfterRun(ctx, a); afterErr != nil {
			if err == nil {
				err = afterErr
			} else {
				xlog.Errorf("Error in AfterRun: %v", afterErr)
			}
		}
	}

	return a.handleError(err)
}

// runCommand executes a command and its subcommands if any
func (a *App) runCommand(ctx context.Context, cmd *Command, args []string) error {
	// Parse flags for the current command
	if cmd.Flags == nil {
		cmd.Flags = flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	}
	if err := cmd.Flags.Parse(args); err != nil {
		return err
	}

	// Check for subcommands recursively
	if len(cmd.Flags.Args()) > 0 && cmd.SubCommands != nil {
		subCmdName := cmd.Flags.Arg(0)
		if subCmd, exists := cmd.SubCommands[subCmdName]; exists {
			return a.runCommand(ctx, subCmd, cmd.Flags.Args()[1:])
		}
	}

	// Execute the command's Run function
	return cmd.Run(ctx, cmd, cmd.Flags.Args())
}

// handleError processes the error and calls the error handling function if set
func (a *App) handleError(err error) error {
	if err != nil {
		xlog.Errorf("Error: %v", err)
		if a.ErrorHandler != nil {
			a.ErrorHandler(err)
		} else {
			// Provide suggestions if the command is unknown
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "unknown command:") {
				cmdName := strings.TrimSpace(strings.TrimPrefix(errMsg, "unknown command:"))
				suggestions := a.suggestCommands(cmdName)
				if len(suggestions) > 0 {
					xcolor.Println(xcolor.Yellow, "\nDid you mean:")
					for _, suggestion := range suggestions {
						fmt.Printf("  %s\n", suggestion)
					}
				}
			}
		}
	}
	return err
}

// PrintHelp prints the help message for the application and its commands
func (a *App) PrintHelp() {
	xcolor.Println(xcolor.Bold, "\n%s", a.Name)
	xcolor.Println(xcolor.Green, "%s\n", a.Description)

	xcolor.Println(xcolor.Yellow, "Usage:")
	xcolor.Println(xcolor.Cyan, "  %s [flags] [command]", a.Name)
	fmt.Println()

	if a.Flags.NFlag() > 0 {
		xcolor.Println(xcolor.Yellow, "Flags:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		a.Flags.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "  -%s\t%s\t(default: %s)\n", f.Name, f.Usage, f.DefValue)
		})
		x.Must0(w.Flush())
		fmt.Println()
	}

	if len(a.Commands) > 0 {
		xcolor.Println(xcolor.Yellow, "Commands:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var visibleCommands []*Command
		for _, cmd := range a.Commands {
			if !cmd.Hidden && cmd.Name == cmd.Name {
				visibleCommands = append(visibleCommands, cmd)
			}
		}
		sort.Slice(visibleCommands, func(i, j int) bool {
			return visibleCommands[i].Name < visibleCommands[j].Name
		})
		for _, cmd := range visibleCommands {
			aliases := ""
			if len(cmd.Aliases) > 0 {
				aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(cmd.Aliases, ", "))
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", cmd.Name, cmd.Description, aliases)
		}
		x.Must0(w.Flush())
		fmt.Println()
	}
}

// EnableColor enables or disables color output
func EnableColor(enable bool) {
	xcolor.EnableColor(enable)
}

// EnableDebug enables or disables debug mode
func EnableDebug(enable bool) {
	debugMode = enable
}

// PrintCommandHelp prints help information for a specific command
func (a *App) PrintCommandHelp(cmdName string) {
	if cmd, exists := a.Commands[cmdName]; exists {
		xcolor.Println(xcolor.Bold, "\n%s", cmd.Name)
		xcolor.Println(xcolor.Green, "%s\n", cmd.Description)

		xcolor.Println(xcolor.Yellow, "Usage:")
		xcolor.Println(xcolor.Cyan, "  %s %s [flags] [subcommand]", a.Name, cmd.Name)
		fmt.Println()

		if cmd.Flags != nil && cmd.Flags.NFlag() > 0 {
			xcolor.Println(xcolor.Yellow, "Flags:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			cmd.Flags.VisitAll(func(f *flag.Flag) {
				fmt.Fprintf(w, "  -%s\t%s\t(default: %s)\n", f.Name, f.Usage, f.DefValue)
			})
			x.Must0(w.Flush())
			fmt.Println()
		}

		if len(cmd.SubCommands) > 0 {
			xcolor.Println(xcolor.Yellow, "Subcommands:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, subCmd := range cmd.SubCommands {
				fmt.Fprintf(w, "  %s\t%s\n", subCmd.Name, subCmd.Description)
			}
			x.Must0(w.Flush())
			fmt.Println()
		}
	} else {
		fmt.Printf("Unknown command: %s\n", cmdName)
	}
}

// Added method to support command suggestions
func (a *App) suggestCommands(input string) []string {
	// Collect all command names and aliases
	var allCommands []string
	for name, cmd := range a.Commands {
		if !cmd.Hidden {
			allCommands = append(allCommands, name)
			allCommands = append(allCommands, cmd.Aliases...)
		}
	}

	// Find commands that are close to the input
	var suggestions []string
	for _, name := range allCommands {
		if strings.HasPrefix(name, input) {
			suggestions = append(suggestions, name)
		}
	}
	return suggestions
}
