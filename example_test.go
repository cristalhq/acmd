package acmd_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cristalhq/acmd"
)

var nopFunc = func(context.Context, []string) error { return nil }

func ExampleRunner() {
	testOut := os.Stdout
	testArgs := []string{"now"}

	const format = "15:04:05"
	now, _ := time.Parse("15:04:05", "10:20:30")

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			Do: func(ctx context.Context, args []string) error {
				fmt.Printf("now: %s\n", now.Format(format))
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
				fmt.Print()
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

	// Output: now: 10:20:30
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
	})

	if err := r.Run(); err == nil {
		panic("must fail with command not found")
	}

	// Output: "baz" is not a subcommand, did you mean "bar"?
}
