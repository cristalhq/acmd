package acmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestRunnerMustSetDefaults(t *testing.T) {
	cmds := []Command{{Name: "foo", Do: nopFunc}}
	r := RunnerOf(cmds, Config{})

	err := r.Run()
	if err == nil {
		t.Fatal()
	}
	if errStr := err.Error(); !strings.Contains(errStr, "acmd: cannot run command: no such command") {
		t.Fatal(err)
	}

	if r.cfg.AppName != os.Args[0] {
		t.Fatalf("want %q got %q", os.Args[0], r.cfg.AppName)
	}
	if r.ctx == nil {
		t.Fatal("context must be set")
	}
	if r.cfg.Output != os.Stderr {
		t.Fatal("incorrect output")
	}
	if r.cfg.Usage == nil {
		t.Fatal("usage nust be set")
	}

	gotCmds := map[string]struct{}{}
	for _, c := range r.rootCmd.subcommands {
		gotCmds[c.Name] = struct{}{}
	}
	if _, ok := gotCmds["help"]; !ok {
		t.Fatal("help command not found")
	}
	if _, ok := gotCmds["version"]; !ok {
		t.Fatal("version command not found")
	}
}

func TestRunnerMustSortCommands(t *testing.T) {
	cmds := []Command{
		{Name: "foo", Do: nopFunc},
		{Name: "xyz", Do: nopFunc},
		{Name: "cake", Do: nopFunc},
		{Name: "foo2", Do: nopFunc},
	}
	r := RunnerOf(cmds, Config{
		Args: []string{"foo"},
	})

	if err := r.Run(); err != nil {
		t.Fatal(err)
	}

	sort.SliceIsSorted(r.cmds, func(i, j int) bool {
		return r.cmds[i].Name < r.cmds[j].Name
	})
}
func TestRunnerPanicWithoutCommands(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("must be panic")
		}
	}()
	RunnerOf(nil, Config{})
}

func TestRunnerInit(t *testing.T) {
	testCases := []struct {
		cmds       []Command
		cfg        Config
		wantErrStr string
	}{
		{
			cmds:       []Command{{Name: "", Do: nopFunc}},
			wantErrStr: `command "" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo%", Do: nopFunc}},
			wantErrStr: `command "foo%" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "%", Do: nopFunc}},
			wantErrStr: `command alias "%" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo%", Do: nil}},
			wantErrStr: `command "foo%" function cannot be nil`,
		},
		{
			cmds:       []Command{{Name: "foo", Do: nil}},
			wantErrStr: `command "foo" function cannot be nil or must have subcommands`,
		},
		{
			cmds: []Command{{
				Name:        "foobar",
				Do:          nopFunc,
				subcommands: []Command{{Name: "nested"}},
			}},
			wantErrStr: `command "foobar" function cannot be set and have subcommands`,
		},
		{
			cmds: []Command{{Name: "foo", Do: nopFunc}},
			cfg: Config{
				Args: []string{},
			},
			wantErrStr: `no args provided`,
		},
		{
			cmds:       []Command{{Name: "help", Do: nopFunc}},
			wantErrStr: `command "help" is reserved`,
		},
		{
			cmds:       []Command{{Name: "version", Do: nopFunc}},
			wantErrStr: `command "version" is reserved`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "help", Do: nopFunc}},
			wantErrStr: `command alias "help" is reserved`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "version", Do: nopFunc}},
			wantErrStr: `command alias "version" is reserved`,
		},
		{
			cmds:       []Command{{Name: "a", Do: nopFunc}, {Name: "a", Do: nopFunc}},
			wantErrStr: `duplicate command "a"`,
		},
		{
			cmds:       []Command{{Name: "aaa", Do: nopFunc}, {Name: "b", Alias: "aaa", Do: nopFunc}},
			wantErrStr: `duplicate command alias "aaa"`,
		},
		{
			cmds:       []Command{{Name: "aaa", Alias: "a", Do: nopFunc}, {Name: "bbb", Alias: "a", Do: nopFunc}},
			wantErrStr: `duplicate command alias "a"`,
		},
		{
			cmds:       []Command{{Name: "a", Do: nopFunc}, {Name: "b", Alias: "a", Do: nopFunc}},
			wantErrStr: `duplicate command alias "a"`,
		},
	}

	for _, tc := range testCases {
		err := RunnerOf(tc.cmds, tc.cfg).Run()

		if got := err.Error(); !strings.Contains(got, tc.wantErrStr) {
			t.Fatalf("want %q got %q", tc.wantErrStr, got)
		}
	}
}

func TestRunner_suggestCommand(t *testing.T) {
	testCases := []struct {
		cmds []Command
		args []string
		want string
	}{
		{
			cmds: []Command{
				{Name: "for", Do: nopFunc},
				{Name: "foo", Do: nopFunc},
				{Name: "bar", Do: nopFunc},
			},
			args: []string{"fooo"},
			want: `"fooo" is not a subcommand, did you mean "foo"?` + "\n",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"hell"},
			want: `"hell" is not a subcommand, did you mean "help"?` + "\n",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"verZION"},
			want: "",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"verZion"},
			want: `"verZion" is not a subcommand, did you mean "version"?` + "\n",
		},
	}

	for _, tc := range testCases {
		buf := &bytes.Buffer{}
		r := RunnerOf(tc.cmds, Config{
			Args:   tc.args,
			Output: buf,
		})
		if err := r.Run(); err != nil && !strings.Contains(err.Error(), "no such command") {
			t.Fatal(err)
		}

		if got := buf.String(); got != tc.want {
			t.Fatalf("want %q got %q", tc.want, got)
		}
	}
}

func TestRunner(t *testing.T) {
	r := RunnerOf([]Command{
		{
			Name:        "test",
			Description: "some test command",
			// Do: func(ctx context.Context, args []string) error {
			// 	return nil
			// },
			subcommands: []Command{
				{
					Name: "foo",
					// Do: func(ctx context.Context, args []string) error {
					// 	fmt.Fprint(os.Stderr, "foo")
					// 	return nil
					// },
					subcommands: []Command{
						{
							Name: "for", Do: func(ctx context.Context, args []string) error {
								fmt.Fprint(os.Stderr, "for")
								return nil
							},
						},
					},
				},
				{Name: "bar", Do: func(ctx context.Context, args []string) error {
					fmt.Fprint(os.Stderr, "bar")
					return nil
				}},
			},
		},
		{
			Name:        "status",
			Description: "status command gives status of the state",
			Do: func(ctx context.Context, args []string) error {
				return nil
			},
		},
	}, Config{
		Args:           []string{"test", "foo", "for"},
		AppName:        "acmd_test_app",
		AppDescription: "acmd_test_app is a test application.",
		Version:        time.Now().String(),
	})

	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
}
