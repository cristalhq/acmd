package acmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"text/tabwriter"
)

var cmdNameRE = regexp.MustCompile("^[A-Za-z0-9-_]+$")

// Runner of the sub-commands.
type Runner struct {
	cfg     Config
	cmds    []Command
	errInit error

	ctx  context.Context
	args []string
}

// Command specifies a sub-command for a program's command-line interface.
type Command struct {
	// Name is just a one-word.
	Name string

	// Description of the command.
	Description string

	// Do will be invoked.
	Do func(ctx context.Context, args []string) error
}

// Config for the runner.
type Config struct {
	// AppName is an optional name for the app, if empty os.Args[0] will be used.
	AppName string

	// AppDescription is an optional description. default is empty.
	AppDescription string

	// Version of the application.
	Version string

	// Output is a destionation where result will be printed.
	// Exported for testing purpose only, if nil os.Stderr is used.
	Output io.Writer

	// Context for commands, if nil context based on os.Interrupt will be used.
	Context context.Context

	// Args passed to the executable, if nil os.Args[1:] will be used.
	Args []string

	// Usage of the application, if nil default will be used,
	Usage func(cfg Config, cmds []Command)
}

// RunnerOf creates a Runner.
func RunnerOf(cmds []Command, cfg Config) *Runner {
	r := &Runner{
		cmds: cmds,
		cfg:  cfg,
	}
	r.errInit = r.init()
	return r
}

func (r *Runner) init() error {
	if r.cfg.AppName == "" {
		r.cfg.AppName = os.Args[0]
	}

	r.args = r.cfg.Args
	if r.args == nil {
		r.args = os.Args[1:]
	}
	if len(r.args) == 0 {
		return errors.New("no args provided")
	}

	r.ctx = r.cfg.Context
	if r.ctx == nil {
		// ok to ignore cancel func because os.Interrupt is already almost os.Exit
		r.ctx, _ = signal.NotifyContext(context.Background(), os.Interrupt)
	}

	if r.cfg.Output == nil {
		r.cfg.Output = os.Stderr
	}

	if r.cfg.Usage == nil {
		r.cfg.Usage = defaultUsage(r.cfg.Output)
	}

	names := make(map[string]struct{})
	for _, cmd := range r.cmds {
		switch {
		case cmd.Do == nil:
			return fmt.Errorf("command %q function cannot be nil", cmd.Name)
		case cmd.Name == "help" || cmd.Name == "version":
			return fmt.Errorf("command %q is reserved", cmd.Name)
		case !cmdNameRE.MatchString(cmd.Name):
			return fmt.Errorf("command %q must contains only letters, digits, - and _", cmd.Name)
		}

		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("duplicate command %q", cmd.Name)
		}
		names[cmd.Name] = struct{}{}
	}

	r.cmds = append(r.cmds,
		Command{
			Name:        "help",
			Description: "shows help message",
			Do: func(ctx context.Context, args []string) error {
				r.cfg.Usage(r.cfg, r.cmds)
				return nil
			},
		},
		Command{
			Name:        "version",
			Description: "shows version of the application",
			Do: func(ctx context.Context, args []string) error {
				fmt.Fprintf(r.cfg.Output, "%s version: %s\n\n", r.cfg.AppName, r.cfg.Version)
				return nil
			},
		},
	)

	sort.Slice(r.cmds, func(i, j int) bool {
		return r.cmds[i].Name < r.cmds[j].Name
	})
	return nil
}

// Run ...
func (r *Runner) Run() error {
	if r.errInit != nil {
		return fmt.Errorf("acmd: cannot init runner: %w", r.errInit)
	}
	if err := r.run(); err != nil {
		return fmt.Errorf("acmd: cannot run command: %w", err)
	}
	return nil
}

func (r *Runner) run() error {
	cmd, params := r.args[0], r.args[1:]

	for _, c := range r.cmds {
		if c.Name == cmd {
			return c.Do(r.ctx, params)
		}
	}

	if suggestion := suggestCommand(cmd, r.cmds); suggestion != "" {
		fmt.Fprintf(r.cfg.Output, "%q is not a subcommand, did you mean %q?\n", cmd, suggestion)
	}
	return fmt.Errorf("no such command %q", cmd)
}

// suggestCommand for not found earlier command.
func suggestCommand(got string, cmds []Command) string {
	const maxMatchDist = 2
	minDist := maxMatchDist + 1
	match := ""

	for _, c := range cmds {
		dist := strDistance(got, c.Name)
		if dist < minDist {
			minDist = dist
			match = c.Name
		}
	}
	return match
}

var defaultUsage = func(w io.Writer) func(cfg Config, cmds []Command) {
	return func(cfg Config, cmds []Command) {
		if cfg.AppDescription != "" {
			fmt.Fprintf(w, "%s\n\n", cfg.AppDescription)
		}

		fmt.Fprintf(w, "Usage:\n\n    %s <command> [arguments...]\n\nThe commands are:\n\n", cfg.AppName)
		printCommands(w, cmds)

		if cfg.Version != "" {
			fmt.Fprintf(w, "Version: %s\n\n", cfg.Version)
		}
	}
}

// printCommands in a table form (Name and Description)
func printCommands(w io.Writer, cmds []Command) {
	minwidth, tabwidth, padding, padchar, flags := 0, 0, 11, byte(' '), uint(0)
	tw := tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, flags)
	for _, cmd := range cmds {
		desc := cmd.Description
		if desc == "" {
			desc = "<no description>"
		}
		fmt.Fprintf(tw, "    %s\t%s\n", cmd.Name, desc)
	}
	fmt.Fprint(tw, "\n")
	tw.Flush()
}