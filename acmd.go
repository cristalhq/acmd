package acmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
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
	// AppName is an optional nal for the app, if empty os.Args[0] will be used.
	AppName string

	// Context for commands, if nil context based on os.Interrupt will be used.
	Context context.Context

	// Args passed to the executable, if nil os.Args[1:] will be used.
	Args []string

	// Usage of the application, if nil flag.Usage will be used,
	Usage func()
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
	args := r.cfg.Args
	if len(args) == 0 {
		args = os.Args[1:]
	}
	if len(r.cfg.Args) == 0 {
		return errors.New("acmd: no args provided")
	}

	ctx := r.cfg.Context
	if ctx == nil {
		var cancel context.CancelFunc
		r.cfg.Context, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
		// TODO: fix
		defer cancel()
	}

	names := make(map[string]struct{})
	for _, cmd := range r.cmds {
		if cmd.Name == "help" {
			return errors.New("acmd: command name 'help' is reserved")
		}
		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("acmd: duplicate command %q", cmd.Name)
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
	if cmd == "help" {
		r.cfg.Usage()
		return nil
	}

	for _, c := range r.cmds {
		if c.Name == cmd {
			return c.Do(r.ctx, params)
		}
	}
	return fmt.Errorf("acmd: no such command %q", cmd)
}
