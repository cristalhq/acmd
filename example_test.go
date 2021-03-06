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

func ExampleRunner() {
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

func ExampleHelp() {
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
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:         "acmd-example",
		AppDescription:  "Example of acmd package",
		PostDescription: "Best place to add examples.",
		Version:         "the best v0.x.y",
		Output:          testOut,
		Args:            testArgs,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output: Example of acmd package
	//
	// Usage:
	//
	//     acmd-example <command> [arguments...]
	//
	// The commands are:
	//
	//     boom              <no description>
	//     help              shows help message
	//     now               prints current time
	//     status            prints status of the system
	//     version           shows version of the application
	//
	// Best place to add examples.
	//
	// Version: the best v0.x.y
}

func ExampleVersion() {
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

func ExampleAlias() {
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

func ExampleAutosuggestion() {
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

func ExampleNestedCommands() {
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

func ExampleExecStruct() {
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

func ExamplePropagateFlags() {
	testOut := os.Stdout
	testArgs := []string{"someapp", "foo", "-dir=test-dir", "--verbose"}
	buf := &bytes.Buffer{}

	cmds := []acmd.Command{
		{
			Name: "foo", ExecFunc: func(ctx context.Context, args []string) error {
				fs := flag.NewFlagSet("foo", flag.ContinueOnError)
				isRecursive := fs.Bool("r", false, "should file list be recursive")
				common := withCommonFlags(fs)
				if err := fs.Parse(args); err != nil {
					return err
				}
				if common.IsVerbose {
					fmt.Fprintf(buf, "TODO: dir %q, is recursive = %v\n", common.Dir, *isRecursive)
				}
				return nil
			},
		},
		{
			Name: "bar", ExecFunc: func(ctx context.Context, args []string) error {
				fs := flag.NewFlagSet("bar", flag.ContinueOnError)
				common := withCommonFlags(fs)
				if err := fs.Parse(args); err != nil {
					return err
				}
				if common.IsVerbose {
					fmt.Fprintf(buf, "TODO: dir %q\n", common.Dir)
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

	// Output: TODO: dir "test-dir", is recursive = false
}

type commonFlags struct {
	IsVerbose bool
	Dir       string
}

// NOTE: should be added before flag.FlagSet method Parse().
func withCommonFlags(fs *flag.FlagSet) *commonFlags {
	c := &commonFlags{}
	fs.BoolVar(&c.IsVerbose, "verbose", false, "should app be verbose")
	fs.StringVar(&c.Dir, "dir", ".", "directory to process")
	return c
}
