package acmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

var (
	nopFunc  = func(context.Context, []string) error { return nil }
	nopUsage = func(cfg Config, cmds []Command) {}
)

func TestRunner(t *testing.T) {
	buf := &bytes.Buffer{}

	cmds := []Command{
		{
			Name:        "test",
			Description: "some test command",
			Subcommands: []Command{
				{
					Name: "foo",
					Subcommands: []Command{
						{
							Name: "for", Do: func(ctx context.Context, args []string) error {
								fmt.Fprint(buf, "for")
								return nil
							},
						},
					},
				},
				{
					Name: "bar",
					Do: func(ctx context.Context, args []string) error {
						fmt.Fprint(buf, "bar")
						return nil
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "status command gives status of the state",
			Do: func(ctx context.Context, args []string) error {
				return nil
			},
		},
	}
	r := RunnerOf(cmds, Config{
		Args:           []string{"./someapp", "test", "foo", "for"},
		AppName:        "myapp",
		AppDescription: "myapp is a test application.",
		Version:        time.Now().String(),
		Output:         buf,
	})

	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "for" {
		t.Fatalf("want %q got %q", "for", got)
	}
}

func TestRunnerMustSetDefaults(t *testing.T) {
	app := "./someapp"
	args := append([]string{app, "runner"}, os.Args[1:]...)
	cmds := []Command{{Name: "foo", Do: nopFunc}}
	r := RunnerOf(cmds, Config{
		Args:   args,
		Output: io.Discard,
		Usage:  nopUsage,
	})

	err := r.Run()
	if err == nil {
		t.Fatal()
	}
	if errStr := err.Error(); !strings.Contains(errStr, `no such command "runner"`) {
		t.Fatal(err)
	}

	if r.cfg.AppName != app {
		t.Fatalf("want %q got %q", app, r.cfg.AppName)
	}
	if r.ctx == nil {
		t.Fatal("context must be set")
	}
	if r.cfg.Usage == nil {
		t.Fatal("usage must be set")
	}

	gotCmds := map[string]struct{}{}
	for _, c := range r.cmds {
		gotCmds[c.Name] = struct{}{}
	}
	if _, ok := gotCmds["help"]; !ok {
		t.Fatal("help command not found")
	}
	if _, ok := gotCmds["version"]; !ok {
		t.Fatal("version command not found")
	}
}

func TestRunnerWithoutArgs(t *testing.T) {
	cmds := []Command{{Name: "foo", Do: nopFunc}}
	r := RunnerOf(cmds, Config{
		Args:   []string{"./app"},
		Output: io.Discard,
		Usage:  nopUsage,
	})

	err := r.Run()
	if err == nil {
		t.Fatal("must be error")
	}
	want := "init error: no args provided"
	if got := err.Error(); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRunnerMustSortCommands(t *testing.T) {
	cmds := []Command{
		{Name: "foo", Do: nopFunc},
		{Name: "xyz"},
		{Name: "cake", Do: nopFunc},
		{Name: "foo2", Do: nopFunc},
	}
	cmds[1].Subcommands = []Command{
		{Name: "a", Do: nopFunc},
		{Name: "c", Do: nopFunc},
		{Name: "b", Do: nopFunc},
	}

	r := RunnerOf(cmds, Config{
		Args: []string{"./someapp", "foo"},
	})

	if err := r.Run(); err != nil {
		t.Fatal(err)
	}

	sort.SliceIsSorted(r.cmds, func(i, j int) bool {
		return r.cmds[i].Name < r.cmds[j].Name
	})

	subc := r.cmds[len(r.cmds)-1].Subcommands
	sort.SliceIsSorted(subc, func(i, j int) bool {
		return subc[i].Name < subc[j].Name
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

func TestRunnerJustExit(t *testing.T) {
	var exitDone bool
	doExitOld := func(_ int) {
		exitDone = true
	}
	defer func() { doExit = doExitOld }()
	doExitOld, doExit = doExit, doExitOld

	buf := &bytes.Buffer{}
	r := RunnerOf([]Command{{Name: "foo", Do: nopFunc}}, Config{
		AppName: "exit-test",
		Output:  buf,
	})
	r.Exit(nil)

	if !exitDone {
		t.Fatal("must be done")
	}
	exitDone = false

	r.Exit(errors.New("oops"))
	if !exitDone {
		t.Fatal("must be done")
	}
	got := buf.String()
	if !strings.Contains(got, "exit-test: oops") {
		t.Fatal(got)
	}
}

func TestRunnerInit(t *testing.T) {
	testCases := []struct {
		cmds       []Command
		cfg        Config
		wantErrStr string
	}{
		{
			cmds:       []Command{{Name: "app:cre.ate", Do: nopFunc}},
			wantErrStr: ``,
		},
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
				Subcommands: []Command{{Name: "nested"}},
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

		if got := err.Error(); tc.wantErrStr != "" && !strings.Contains(got, tc.wantErrStr) {
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
			args: []string{"./someapp", "fooo"},
			want: `"fooo" unknown command, did you mean "foo"?` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"./someapp", "hell"},
			want: `"hell" unknown command, did you mean "help"?` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"./someapp", "verZION"},
			want: `"verZION" unknown command` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", Do: nopFunc}},
			args: []string{"./someapp", "verZion"},
			want: `"verZion" unknown command, did you mean "version"?` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
	}

	for _, tc := range testCases {
		buf := &bytes.Buffer{}
		r := RunnerOf(tc.cmds, Config{
			Args:    tc.args,
			AppName: "myapp",
			Output:  buf,
			Usage:   nopUsage,
		})
		if err := r.Run(); err != nil && !strings.Contains(err.Error(), "no such command") {
			t.Fatal(err)
		}

		if got := buf.String(); got != tc.want {
			t.Fatalf("want %q got %q", tc.want, got)
		}
	}
}

func TestHasHelpFlag(t *testing.T) {
	testCases := []struct {
		args    []string
		hasHelp bool
	}{
		{[]string{"foo", "bar"}, false},
		{[]string{"foo", "-help"}, true},
		{[]string{"foo", "-h", "baz"}, true},
		{[]string{"--help", "-h", "baz"}, true},
	}
	for _, tc := range testCases {
		if got := HasHelpFlag(tc.args); got != tc.hasHelp {
			t.Fatalf("got %v, want %v", got, tc.hasHelp)
		}
	}
}

func TestCommand_IsHidden(t *testing.T) {
	buf := &bytes.Buffer{}
	cmds := []Command{
		{Name: "for", Do: nopFunc},
		{Name: "foo", Do: nopFunc, IsHidden: true},
		{Name: "bar", Do: nopFunc},
	}
	r := RunnerOf(cmds, Config{
		Args:    []string{"./someapp", "help"},
		AppName: "myapp",
		Output:  buf,
	})
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "foo") {
		t.Fatal("should not show foo")
	}
}

func TestExit(t *testing.T) {
	wantStatus := 42
	wantOutput := "myapp: code 42\n"

	cmds := []Command{
		{
			Name: "for",
			Do: func(ctx context.Context, args []string) error {
				return ErrCode(wantStatus)
			},
		},
	}

	buf := &bytes.Buffer{}
	r := RunnerOf(cmds, Config{
		AppName: "myapp",
		Args:    []string{"./someapp", "for"},
		Output:  buf,
	})

	err := r.Run()
	if err == nil {
		t.Fatal("must not be nil")
	}

	var gotStatus int
	doExitOld := func(code int) {
		gotStatus = code
	}
	defer func() { doExit = doExitOld }()

	doExitOld, doExit = doExit, doExitOld

	r.Exit(err)

	if gotStatus != wantStatus {
		t.Fatalf("got %d want %d", gotStatus, wantStatus)
	}
	if got := buf.String(); got != wantOutput {
		t.Fatalf("got %q want %q", got, wantOutput)
	}
}
