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
)

var (
	debugMode = false
)

// Command represents a CLI command with its flags and subcommands.
type Command struct {
	Name        string
	Description string
	Flags       *flag.FlagSet
	Subcommands map[string]*Command
	Run         func(ctx context.Context, cmdCtx *CommandContext) error
	DisableHelp bool
	CustomHelp  func(*Command)
	Aliases     []string
	Hidden      bool
	Version     string
}

// CommandContext holds the context of the current command execution.
type CommandContext struct {
	Command *Command
	Flags   *flag.FlagSet
	Args    []string
}

// NewCommand creates a new Command.
func NewCommand(name, description string, run func(ctx context.Context, cmdCtx *CommandContext) error) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Flags:       flag.NewFlagSet(name, flag.ExitOnError),
		Subcommands: make(map[string]*Command),
		Run:         run,
	}
}

// AddSubcommand adds a subcommand to the command.
func (c *Command) AddSubcommand(sub *Command) {
	if _, exists := c.Subcommands[sub.Name]; !exists {
		c.Subcommands[sub.Name] = sub
		for _, alias := range sub.Aliases {
			if _, aliasExists := c.Subcommands[alias]; !aliasExists {
				c.Subcommands[alias] = sub
			}
		}
	}
}

// Execute parses the flags and executes the command or its subcommands.
func (c *Command) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--help" || args[0] == "-h")) {
		if !c.DisableHelp {
			c.PrintHelp()
		}
		return nil
	}

	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("%s version %s\n", c.Name, c.Version)
		return nil
	}

	cmdName := args[0]
	if subcommand, exists := c.Subcommands[cmdName]; exists {
		return subcommand.Execute(ctx, args[1:])
	}

	if err := c.Flags.Parse(args); err != nil {
		return xerror.Wrap(err, "failed to parse flags")
	}

	cmdCtx := &CommandContext{
		Command: c,
		Flags:   c.Flags,
		Args:    c.Flags.Args(),
	}

	if c.Run != nil {
		startTime := time.Now()
		err := c.Run(ctx, cmdCtx)
		if debugMode {
			fmt.Printf("Command execution time: %v\n", time.Since(startTime))
		}
		return xerror.Wrap(err, "command execution failed")
	}

	return xerror.Errorf("unknown command: %s", cmdName)
}

// PrintHelp prints the help message for the command and its subcommands.
func (c *Command) PrintHelp() {
	if c.CustomHelp != nil {
		c.CustomHelp(c)
		return
	}

	xcolor.Println(xcolor.Bold, "\n%s", c.Name)
	xcolor.Println(xcolor.Green, "%s\n", c.Description)

	xcolor.Println(xcolor.Yellow, "Usage:")
	xcolor.Println(xcolor.Cyan, "  %s [flags] [subcommand]", c.Name)
	fmt.Println()

	if c.Flags.NFlag() > 0 {
		xcolor.Println(xcolor.Yellow, "Flags:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		c.Flags.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "  -%s\t%s\t(default: %s)\n", f.Name, f.Usage, f.DefValue)
		})
		x.Must0(w.Flush())
		fmt.Println()
	}

	if len(c.Subcommands) > 0 {
		xcolor.Println(xcolor.Yellow, "Subcommands:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		uniqueCommands := make(map[string]*Command)
		for name, sub := range c.Subcommands {
			if !sub.Hidden && name == sub.Name {
				uniqueCommands[name] = sub
			}
		}
		var visibleCommands []*Command
		for _, sub := range uniqueCommands {
			visibleCommands = append(visibleCommands, sub)
		}
		sort.Slice(visibleCommands, func(i, j int) bool {
			return visibleCommands[i].Name < visibleCommands[j].Name
		})
		for _, sub := range visibleCommands {
			aliases := ""
			if len(sub.Aliases) > 0 {
				aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(sub.Aliases, ", "))
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", sub.Name, sub.Description, aliases)
		}
		x.Must0(w.Flush())
		fmt.Println()
	}
}

// Execute runs the root command with the provided arguments.
func Execute(rootCmd *Command, args []string) {
	ctx := context.Background()
	if err := rootCmd.Execute(ctx, args); err != nil {
		xcolor.Println(xcolor.Red, "Error: %s", err)
		os.Exit(1)
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

// SetVersion sets the version of the command
func (c *Command) SetVersion(version string) {
	c.Version = version
}

// SetAliases sets the aliases for the command
func (c *Command) SetAliases(aliases ...string) {
	c.Aliases = aliases
}

// SetHidden sets whether the command should be hidden from help output
func (c *Command) SetHidden(hidden bool) {
	c.Hidden = hidden
}
