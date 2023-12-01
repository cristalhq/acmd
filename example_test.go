package acmd_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cristalhq/acmd"
)

var (
	nopFunc  = func(context.Context, []string) error { return nil }
	nopUsage = func(cfg acmd.Config, cmds []acmd.Command) {}
)

func Example() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "now", "--times", "3"}

	const format = "15:04:05"
	now, _ := time.Parse(format, "10:20:30")

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			ExecFunc: func(ctx context.Context, args []string) error {
				fs := flag.NewFlagSet("some name for help", flag.ContinueOnError)
				times := fs.Int("times", 1, "how many times to print time")
				if err := fs.Parse(args); err != nil {
					return err
				}

				for i := 0; i < *times; i++ {
					fmt.Printf("now: %s\n", now.Format(format))
				}
				return nil
			},
			FlagSet: &commandFlags{},
		},
		{
			Name:        "status",
			Description: "prints status of the system",
			ExecFunc: func(ctx context.Context, args []string) error {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.githubstatus.com/", http.NoBody)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				// TODO: parse response, I don't know
				return nil
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:         "acmd-example",
		AppDescription:  "Example of acmd package",
		PostDescription: "Best place to add examples",
		Version:         "the best v0.x.y",
		Output:          testOut,
		Args:            testArgs,
		Usage:           nopUsage,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output:
	// now: 10:20:30
	// now: 10:20:30
	// now: 10:20:30
}

func Example_verboseHelp() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "help"}

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			ExecFunc:    nopFunc,
		},
		{
			Name:        "status",
			Description: "prints status of the system",
			ExecFunc:    nopFunc,
		},
		{
			Name:     "boom",
			ExecFunc: nopFunc,
			FlagSet:  &generalFlags{},
		},
		{
			Name: "time", Subcommands: []acmd.Command{
				{Name: "next", ExecFunc: nopFunc, Description: "next time subcommand"},
				{Name: "curr", ExecFunc: nopFunc, Description: "curr time subcommand"},
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:         "acmd-example",
		AppDescription:  "Example of acmd package",
		PostDescription: "Best place to add examples.",
		Version:         "the best v0.x.y",
		Output:          testOut,
		Args:            testArgs,
		VerboseHelp:     !true, // TODO(cristaloleg): fix this
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output:
	// Example of acmd package
	//
	// Usage:
	//
	//     acmd-example <command> [arguments...]
	//
	// The commands are:
	//
	//     boom                <no description>
	//     help                shows help message
	//     now                 prints current time
	//     status              prints status of the system
	//     time curr           curr time subcommand
	//     time next           next time subcommand
	//     version             shows version of the application
	//
	// Best place to add examples.
	//
	// Version: the best v0.x.y
}

func Example_version() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "version"}

	cmds := []acmd.Command{
		{Name: "foo", ExecFunc: nopFunc},
		{Name: "bar", ExecFunc: nopFunc},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "acmd-example",
		AppDescription: "Example of acmd package",
		Version:        "the best v0.x.y",
		Output:         testOut,
		Args:           testArgs,
		Usage:          nopUsage,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output: acmd-example version: the best v0.x.y
}

func Example_alias() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "f"}

	cmds := []acmd.Command{
		{
			Name:  "foo",
			Alias: "f",
			ExecFunc: func(ctx context.Context, args []string) error {
				fmt.Fprint(testOut, "foo")
				return nil
			},
		},
		{
			Name:  "bar",
			Alias: "b",
			ExecFunc: func(ctx context.Context, args []string) error {
				fmt.Fprint(testOut, "bar")
				return nil
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "acmd-example",
		AppDescription: "Example of acmd package",
		Version:        "the best v0.x.y",
		Output:         testOut,
		Args:           testArgs,
		Usage:          nopUsage,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output: foo
}

func Example_autosuggestion() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "baz"}

	cmds := []acmd.Command{
		{Name: "foo", ExecFunc: nopFunc},
		{Name: "bar", ExecFunc: nopFunc},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "acmd-example",
		AppDescription: "Example of acmd package",
		Version:        "the best v0.x.y",
		Output:         testOut,
		Args:           testArgs,
		Usage:          nopUsage,
	})

	if err := r.Run(); err == nil {
		panic("must fail with command not found")
	}

	// Output:
	// "baz" unknown command, did you mean "bar"?
	// Run "acmd-example help" for usage.
}

func Example_nestedCommands() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "foo", "qux"}

	cmds := []acmd.Command{
		{
			Name: "foo",
			Subcommands: []acmd.Command{
				{Name: "bar", ExecFunc: nopFunc},
				{Name: "baz", ExecFunc: nopFunc},
				{
					Name: "qux",
					ExecFunc: func(ctx context.Context, args []string) error {
						fmt.Fprint(testOut, "qux")
						return nil
					},
				},
			},
		},
		{Name: "boom", ExecFunc: nopFunc},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "acmd-example",
		AppDescription: "Example of acmd package",
		Version:        "the best v0.x.y",
		Output:         testOut,
		Args:           testArgs,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output: qux
}

type myCommand struct {
	ErrToReturn error
}

func (mc *myCommand) ExecCommand(ctx context.Context, args []string) error {
	return mc.ErrToReturn
}

func Example_execStruct() {
	myErr := errors.New("everything is ok")
	myCmd := &myCommand{ErrToReturn: myErr}

	cmds := []acmd.Command{
		{
			Name:        "what",
			Description: "does something",

			// ExecFunc:    myCmd.ExecCommand,
			// NOTE: line below is literally line above
			Exec: myCmd,
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:         "acmd-example",
		AppDescription:  "Example of acmd package",
		PostDescription: "Best place to add examples",
		Output:          io.Discard,
		Args:            []string{"someapp", "what"},
		Usage:           nopUsage,
	})

	err := r.Run()
	if !errors.Is(err, myErr) {
		panic(fmt.Sprintf("\ngot : %+v\nwant: %+v\n", err, myErr))
	}

	// Output:
}

func Example_propagateFlags() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "foo", "-dir=test-dir", "--verbose"}
	buf := &bytes.Buffer{}

	cmds := []acmd.Command{
		{
			Name: "foo", ExecFunc: func(ctx context.Context, args []string) error {
				var cfg generalFlags
				if err := cfg.Flags().Parse(args); err != nil {
					return err
				}
				if cfg.IsVerbose {
					fmt.Fprintf(buf, "TODO: dir %q, is verbose = %v\n", cfg.Dir, cfg.IsVerbose)
				}
				return nil
			},
		},
		{
			Name: "bar", ExecFunc: func(ctx context.Context, args []string) error {
				var cfg commandFlags
				if err := cfg.Flags().Parse(args); err != nil {
					return err
				}
				if cfg.IsVerbose {
					fmt.Fprintf(buf, "TODO: dir %q\n", cfg.Dir)
				}
				return nil
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "acmd-example",
		AppDescription: "Example of acmd package",
		Version:        "the best v0.x.y",
		Output:         testOut,
		Args:           testArgs,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
	fmt.Println(buf.String())

	// Output: TODO: dir "test-dir", is verbose = true
}

type generalFlags struct {
	IsVerbose bool
	Dir       string
}

func (c *generalFlags) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&c.IsVerbose, "verbose", false, "should app be verbose")
	fs.StringVar(&c.Dir, "dir", ".", "directory to process")
	return fs
}

type commandFlags struct {
	generalFlags
	File string
}

func (c *commandFlags) Flags() *flag.FlagSet {
	fs := c.generalFlags.Flags()
	fs.StringVar(&c.File, "file", "input.txt", "file to process")
	return fs
}
