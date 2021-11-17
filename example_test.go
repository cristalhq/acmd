package acmd_test

import (
	"context"
	"flag"
	"fmt"
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
	testArgs := []string{"now", "--times", "3"}

	const format = "15:04:05"
	now, _ := time.Parse(format, "10:20:30")

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			Do: func(ctx context.Context, args []string) error {
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
			Do: func(ctx context.Context, args []string) error {
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

	// Output:
	// now: 10:20:30
	// now: 10:20:30
	// now: 10:20:30
}

func ExampleHelp() {
	testOut := os.Stdout
	testArgs := []string{"help"}

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			Do:          nopFunc,
		},
		{
			Name:        "status",
			Description: "prints status of the system",
			Do:          nopFunc,
		},
		{
			Name: "boom",
			Do:   nopFunc,
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
	// Version: the best v0.x.y
}

func ExampleVersion() {
	testOut := os.Stdout
	testArgs := []string{"version"}

	cmds := []acmd.Command{
		{Name: "foo", Do: nopFunc},
		{Name: "bar", Do: nopFunc},
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
	testArgs := []string{"f"}

	cmds := []acmd.Command{
		{
			Name:  "foo",
			Alias: "f",
			Do: func(ctx context.Context, args []string) error {
				fmt.Fprint(testOut, "foo")
				return nil
			},
		},
		{
			Name:  "bar",
			Alias: "b",
			Do: func(ctx context.Context, args []string) error {
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
	testArgs := []string{"baz"}

	cmds := []acmd.Command{
		{Name: "foo", Do: nopFunc},
		{Name: "bar", Do: nopFunc},
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
	testArgs := []string{"foo", "qux"}

	cmds := []acmd.Command{
		{
			Name: "foo",
			Subcommands: []acmd.Command{
				{Name: "bar", Do: nopFunc},
				{Name: "baz", Do: nopFunc},
				{
					Name: "qux",
					Do: func(ctx context.Context, args []string) error {
						fmt.Fprint(testOut, "qux")
						return nil
					},
				},
			},
		},
		{Name: "boom", Do: nopFunc},
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
