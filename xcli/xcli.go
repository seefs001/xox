package xcli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
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
	Name                  string
	Description           string
	Version               string
	Flags                 *flag.FlagSet
	Commands              map[string]*Command
	DefaultRun            func(ctx context.Context, app *App) error
	ErrorHandler          func(error)
	BeforeRun             func(ctx context.Context, app *App) error
	AfterRun              func(ctx context.Context, app *App) error
	UnknownCommandHandler func(ctx context.Context, cmdName string, args []string) error // Handler for unsupported commands

	// Added fields for customizing output messages
	CustomErrorPrinter       func(err error)
	CustomHelpPrinter        func(app *App)
	CustomCommandHelpPrinter func(app *App, cmdName string)
	initialized              bool
	mutex                    sync.RWMutex
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
	if cmd == nil {
		return
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.Commands[cmd.Name]; !exists {
		if a.Commands == nil {
			a.Commands = make(map[string]*Command)
		}
		a.Commands[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			if alias != "" && alias != cmd.Name {
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

// SetUnknownCommandHandler sets a custom handler for unsupported commands
func (a *App) SetUnknownCommandHandler(handler func(ctx context.Context, cmdName string, args []string) error) {
	a.UnknownCommandHandler = handler
}

// Run executes the application
func (a *App) Run(ctx context.Context, args []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	if !a.initialized {
		a.Initialize()
	}

	defer func() {
		if r := recover(); r != nil {
			xlog.Errorf("Panic recovered: %v", r)
			if a.ErrorHandler != nil {
				a.ErrorHandler(fmt.Errorf("panic: %v", r))
			}
		}
	}()

	// Ensure Flags is initialized
	if a.Flags == nil {
		a.Initialize()
	}

	// Parse global flags
	if err := a.Flags.Parse(args[1:]); err != nil {
		return a.handleError(xerror.Wrap(err, "failed to parse flags"))
	}
	parsedArgs := a.Flags.Args()

	// Check for help or version flags
	helpFlag := a.Flags.Lookup("help")
	versionFlag := a.Flags.Lookup("version")

	if (helpFlag != nil && helpFlag.Value.(flag.Getter).Get().(bool)) ||
		(a.Flags.Lookup("h") != nil && a.Flags.Lookup("h").Value.(flag.Getter).Get().(bool)) {
		a.PrintHelp()
		return nil
	}
	if (versionFlag != nil && versionFlag.Value.(flag.Getter).Get().(bool)) ||
		(a.Flags.Lookup("v") != nil && a.Flags.Lookup("v").Value.(flag.Getter).Get().(bool)) {
		xcolor.Println(xcolor.Cyan, "%s version %s", a.Name, a.Version)
		return nil
	}

	// Run BeforeRun function if set
	if a.BeforeRun != nil {
		if err := a.BeforeRun(ctx, a); err != nil {
			return a.handleError(err)
		}
	}

	var err error
	if len(parsedArgs) > 0 {
		if cmd, exists := a.Commands[parsedArgs[0]]; exists {
			err = a.runCommand(ctx, cmd, parsedArgs[1:])
		} else {
			if a.UnknownCommandHandler != nil {
				err = a.UnknownCommandHandler(ctx, parsedArgs[0], parsedArgs[1:])
			} else {
				err = fmt.Errorf("unknown command: %s", parsedArgs[0])
			}
		}
	} else if a.DefaultRun != nil {
		startTime := time.Now()
		err = a.DefaultRun(ctx, a)
		if debugMode {
			xlog.Infof("App execution time: %v", time.Since(startTime))
		}
	} else {
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
	if cmd == nil {
		return fmt.Errorf("nil command provided")
	}

	if cmd.Run == nil {
		return fmt.Errorf("command %s has no run function", cmd.Name)
	}

	if cmd.Flags == nil {
		cmd.Flags = flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
		cmd.Flags.SetOutput(io.Discard)
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
		if a.CustomErrorPrinter != nil {
			a.CustomErrorPrinter(err)
		} else {
			xcolor.Println(xcolor.Red, "Error: %v", err)
		}
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
						xcolor.Println(xcolor.Cyan, "  %s", suggestion)
					}
				}
			}
		}
	}
	return err
}

// PrintHelp prints the help message for the application and its commands
func (a *App) PrintHelp() {
	if a.CustomHelpPrinter != nil {
		a.CustomHelpPrinter(a)
		return
	}
	xcolor.Println(xcolor.Bold, "\n%s", a.Name)
	xcolor.Println(xcolor.Green, "%s\n", a.Description)

	xcolor.Println(xcolor.Yellow, "Usage:")
	xcolor.Println(xcolor.Cyan, "  %s [flags] [command]", a.Name)
	fmt.Println()

	if a.Flags.NFlag() > 0 {
		xcolor.Println(xcolor.Yellow, "Flags:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		a.Flags.VisitAll(func(f *flag.Flag) {
			xcolor.Fprintf(w, xcolor.Cyan, "  -%s", f.Name)
			xcolor.Fprintf(w, xcolor.None, "\t%s\t(default: %s)\n", f.Usage, f.DefValue)
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
			xcolor.Fprintf(w, xcolor.Cyan, "  %s", cmd.Name)
			xcolor.Fprintf(w, xcolor.None, "\t%s%s\n", cmd.Description, aliases)
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
	if a.CustomCommandHelpPrinter != nil {
		a.CustomCommandHelpPrinter(a, cmdName)
		return
	}
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
				xcolor.Fprintf(w, xcolor.Cyan, "  -%s", f.Name)
				xcolor.Fprintf(w, xcolor.None, "\t%s\t(default: %s)\n", f.Usage, f.DefValue)
			})
			x.Must0(w.Flush())
			fmt.Println()
		}

		if len(cmd.SubCommands) > 0 {
			xcolor.Println(xcolor.Yellow, "Subcommands:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, subCmd := range cmd.SubCommands {
				xcolor.Fprintf(w, xcolor.Cyan, "  %s", subCmd.Name)
				xcolor.Fprintf(w, xcolor.None, "\t%s\n", subCmd.Description)
			}
			x.Must0(w.Flush())
			fmt.Println()
		}
	} else {
		xcolor.Println(xcolor.Red, "Unknown command: %s", cmdName)
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

// Initialize sets up the global flags
func (a *App) Initialize() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.initialized {
		return
	}

	if a.Flags == nil {
		a.Flags = flag.NewFlagSet(a.Name, flag.ExitOnError)
		a.Flags.SetOutput(io.Discard)
		a.Flags.Usage = func() { a.PrintHelp() }
		a.Flags.Bool("help", false, "Display help information")
		a.Flags.Bool("h", false, "Display help information")
		a.Flags.Bool("version", false, "Display version information")
		a.Flags.Bool("v", false, "Display version information")
	}

	a.initialized = true
}

func (a *App) GetCommand(name string) (*Command, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	cmd, exists := a.Commands[name]
	return cmd, exists
}

func (a *App) RemoveCommand(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if cmd, exists := a.Commands[name]; exists {
		for _, alias := range cmd.Aliases {
			delete(a.Commands, alias)
		}
		delete(a.Commands, name)
	}
}
