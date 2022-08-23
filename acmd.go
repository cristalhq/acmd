package acmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"text/tabwriter"
)

// changed only in tests.
var doExit = os.Exit

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
	// Name of the command, ex: `init`
	Name string

	// Alias is an optional short second name, ex: `i`.
	Alias string

	// Description of the command.
	Description string

	// ExecFunc represents the command function.
	// Use Exec if you have struct implementing this function.
	ExecFunc func(ctx context.Context, args []string) error

	// Exec represents the command function.
	// Will be used only if ExecFunc is nil.
	Exec Exec

	// Subcommands of the command.
	Subcommands []Command

	// IsHidden reports whether command should not be show in help. Default false.
	IsHidden bool
}

// simple way to get exec function
func (cmd *Command) getExec() func(ctx context.Context, args []string) error {
	switch {
	case cmd.ExecFunc != nil:
		return cmd.ExecFunc
	case cmd.Exec != nil:
		return cmd.Exec.ExecCommand
	default:
		return nil
	}
}

// Exec represents a command to run.
type Exec interface {
	ExecCommand(ctx context.Context, args []string) error
}

// Config for the runner.
type Config struct {
	// AppName is an optional name for the app, if empty os.Args[0] will be used.
	AppName string

	// AppDescription is an optional description. default is empty.
	AppDescription string

	// PostDescription of the command. Is shown after a help.
	PostDescription string

	// Version of the application.
	Version string

	// Output is a destionation where result will be printed.
	// Exported for testing purpose only, if nil os.Stdout is used.
	Output io.Writer

	// Context for commands, if nil context based on os.Interrupt and syscall.SIGTERM will be used.
	Context context.Context

	// Args passed to the executable, if nil os.Args[1:] will be used.
	Args []string

	// Usage of the application, if nil default will be used.
	Usage func(cfg Config, cmds []Command)
}

// HasHelpFlag reports whether help flag is presented in args.
func HasHelpFlag(flags []string) bool {
	for _, f := range flags {
		switch f {
		case "-h", "-help", "--help":
			return true
		}
	}
	return false
}

// RunnerOf creates a Runner.
func RunnerOf(cmds []Command, cfg Config) *Runner {
	if len(cmds) == 0 {
		panic("acmd: cannot run without commands")
	}

	r := &Runner{
		cmds: cmds,
		cfg:  cfg,
	}
	r.errInit = r.init()
	return r
}

// Exit the application depending on the error.
// If err is nil, so successful/no error exit is done: os.Exit(0)
// If err is of type ErrCode: code from the error is returned: os.Exit(code)
// Otherwise: os.Exit(1).
func (r *Runner) Exit(err error) {
	if err == nil {
		doExit(0)
		return
	}
	errCode := ErrCode(1)
	errors.As(err, &errCode)

	fmt.Fprintf(r.cfg.Output, "%s: %s\n", r.cfg.AppName, err.Error())
	doExit(int(errCode))
}

func (r *Runner) init() error {
	if r.cfg.Output == nil {
		r.cfg.Output = os.Stdout
	}

	if r.cfg.Usage == nil {
		r.cfg.Usage = defaultUsage(r.cfg.Output)
	}

	r.args = r.cfg.Args
	if r.args == nil {
		r.args = os.Args
	} else if len(r.args) == 0 {
		return ErrNoArgs
	}

	if r.cfg.AppName == "" {
		r.cfg.AppName = r.args[0]
	}

	r.args = r.args[1:]
	if len(r.args) == 0 {
		return ErrNoArgs
	}

	r.ctx = r.cfg.Context
	if r.ctx == nil {
		// ok to ignore cancel func because os.Interrupt and syscall.SIGTERM is already almost os.Exit
		r.ctx, _ = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	}

	fakeRootCmd := Command{
		Name:        "root",
		Subcommands: r.cmds,
	}
	if err := validateCommand(fakeRootCmd); err != nil {
		return err
	}

	r.cmds = append(r.cmds,
		Command{
			Name:        "help",
			Description: "shows help message",
			ExecFunc: func(ctx context.Context, args []string) error {
				r.cfg.Usage(r.cfg, r.cmds)
				return nil
			},
		},
		Command{
			Name:        "version",
			Description: "shows version of the application",
			ExecFunc: func(ctx context.Context, args []string) error {
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

func validateCommand(cmd Command) error {
	cmds := cmd.Subcommands

	switch {
	case cmd.getExec() == nil && len(cmds) == 0:
		return fmt.Errorf("command %q exec function cannot be nil OR must have subcommands", cmd.Name)

	case cmd.getExec() != nil && len(cmds) != 0:
		return fmt.Errorf("command %q exec function cannot be set AND have subcommands", cmd.Name)

	case cmd.Name == "help" || cmd.Name == "version":
		return fmt.Errorf("command %q is reserved", cmd.Name)

	case cmd.Alias == "help" || cmd.Alias == "version":
		return fmt.Errorf("command alias %q is reserved", cmd.Alias)

	case !isStringValid(cmd.Name):
		return fmt.Errorf("command %q must contains only letters, digits, - and _", cmd.Name)

	case cmd.Alias != "" && !isStringValid(cmd.Alias):
		return fmt.Errorf("command alias %q must contains only letters, digits, - and _", cmd.Alias)

	case len(cmds) != 0:
		if err := validateSubcommands(cmds); err != nil {
			return err
		}
	}
	return nil
}

func validateSubcommands(cmds []Command) error {
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})

	names := make(map[string]struct{})
	for _, cmd := range cmds {
		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("duplicate command %q", cmd.Name)
		}
		if _, ok := names[cmd.Alias]; ok {
			return fmt.Errorf("duplicate command alias %q", cmd.Alias)
		}

		names[cmd.Name] = struct{}{}
		if cmd.Alias != "" {
			names[cmd.Alias] = struct{}{}
		}

		if err := validateCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func isStringValid(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !(('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') ||
			('0' <= c && c <= '9') || c == '-' || c == '_' || c == ':' || c == '.') {
			return false
		}
	}
	return true
}

// Run commands.
func (r *Runner) Run() error {
	if r.errInit != nil {
		return r.errInit
	}
	cmd, params, err := findCmd(r.cfg, r.cmds, r.args)
	if err != nil {
		return err
	}
	return cmd(r.ctx, params)
}

func findCmd(cfg Config, cmds []Command, args []string) (func(ctx context.Context, args []string) error, []string, error) {
	for {
		selected, params := args[0], args[1:]

		var found bool
		for _, c := range cmds {
			if selected != c.Name && selected != c.Alias {
				continue
			}

			// go deeper into subcommands
			if c.getExec() == nil {
				if len(params) == 0 {
					return nil, nil, errors.New("no args for command provided")
				}
				cmds, args = c.Subcommands, params
				found = true
				break
			}
			return c.getExec(), params, nil
		}

		if !found {
			return nil, nil, errNotFoundAndSuggest(cfg.Output, cfg.AppName, selected, cmds)
		}
	}
}

func errNotFoundAndSuggest(w io.Writer, appName, selected string, cmds []Command) error {
	suggestion := suggestCommand(selected, cmds)
	if suggestion != "" {
		fmt.Fprintf(w, "%q unknown command, did you mean %q?\n", selected, suggestion)
	} else {
		fmt.Fprintf(w, "%q unknown command\n", selected)
	}
	fmt.Fprintf(w, "Run %q for usage.\n\n", appName+" help")
	return fmt.Errorf("no such command %q", selected)
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

func defaultUsage(w io.Writer) func(cfg Config, cmds []Command) {
	return func(cfg Config, cmds []Command) {
		if cfg.AppDescription != "" {
			fmt.Fprintf(w, "%s\n\n", cfg.AppDescription)
		}

		fmt.Fprintf(w, "Usage:\n\n    %s <command> [arguments...]\n\nThe commands are:\n\n", cfg.AppName)
		printCommands(w, cmds)

		if cfg.PostDescription != "" {
			fmt.Fprintf(w, "%s\n\n", cfg.PostDescription)
		}
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
		if cmd.IsHidden {
			continue
		}
		desc := cmd.Description
		if desc == "" {
			desc = "<no description>"
		}
		fmt.Fprintf(tw, "    %s\t%s\n", cmd.Name, desc)
	}
	fmt.Fprint(tw, "\n")
	tw.Flush()
}
