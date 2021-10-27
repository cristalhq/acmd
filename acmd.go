package acmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"text/tabwriter"
)

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
	r.args = r.cfg.Args
	if len(r.args) == 0 {
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

	if r.cfg.Usage == nil {
		r.cfg.Usage = defaultUsage
	}

	names := make(map[string]struct{})
	for _, cmd := range r.cmds {
		switch {
		case cmd.Do == nil:
			return fmt.Errorf("command %q function cannot be nil", cmd.Name)
		case cmd.Name == "help" || cmd.Name == "version":
			return fmt.Errorf("command %q is reserved", cmd.Name)
		}

		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("duplicate command %q", cmd.Name)
		}
		names[cmd.Name] = struct{}{}
	}

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
	switch {
	case cmd == "help":
		r.cfg.Usage(r.cfg, r.cmds)
		return nil
	case cmd == "version":
		fmt.Printf("%s version: %s\n\n", r.cfg.AppName, r.cfg.Version)
		return nil
	}

	for _, c := range r.cmds {
		if c.Name == cmd {
			return c.Do(r.ctx, params)
		}
	}
	return fmt.Errorf("no such command %q", cmd)
}

var defaultUsage = func(cfg Config, cmds []Command) {
	if cfg.AppDescription != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", cfg.AppDescription)
	}

	fmt.Fprintf(os.Stderr, "Usage:\n\n    %s <command> [arguments]\n\nThe commands are:\n\n", cfg.AppName)
	printCommands(cmds)

	if cfg.Version != "" {
		fmt.Fprintf(os.Stderr, "Version: %s\n\n", cfg.Version)
	}
}

// printCommands in a table form (Name and Description)
func printCommands(cmds []Command) {
	minwidth, tabwidth, padding, padchar, flags := 0, 0, 11, byte(' '), uint(0)
	tw := tabwriter.NewWriter(os.Stderr, minwidth, tabwidth, padding, padchar, flags)
	for _, cmd := range cmds {
		fmt.Fprintf(tw, "    %s\t%s\n", cmd.Name, cmd.Description)
	}
	fmt.Fprint(tw, "\n")
	tw.Flush()
}
