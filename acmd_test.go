package acmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
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
							Name: "for", ExecFunc: func(ctx context.Context, args []string) error {
								fmt.Fprint(buf, "for")
								return nil
							},
						},
					},
				},
				{
					Name: "bar",
					ExecFunc: func(ctx context.Context, args []string) error {
						fmt.Fprint(buf, "bar")
						return nil
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "status command gives status of the state",
			ExecFunc: func(ctx context.Context, args []string) error {
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

	failIfErr(t, r.Run())
	mustEqual(t, buf.String(), "for")
}

func TestRunnerMustSetDefaults(t *testing.T) {
	app := "./someapp"
	args := append([]string{app, "runner"}, os.Args[1:]...)
	cmds := []Command{{Name: "foo", ExecFunc: nopFunc}}
	r := RunnerOf(cmds, Config{
		Args:   args,
		Output: io.Discard,
		Usage:  nopUsage,
	})

	err := r.Run()
	failIfOk(t, err)

	if errStr := err.Error(); !strings.Contains(errStr, `no such command "runner"`) {
		t.Fatal(err)
	}

	mustEqual(t, r.cfg.AppName, app)
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
	cmds := []Command{{Name: "foo", ExecFunc: nopFunc}}
	r := RunnerOf(cmds, Config{
		Args:   []string{"./app"},
		Output: io.Discard,
		Usage:  nopUsage,
	})

	err := r.Run()
	failIfOk(t, err)
	mustEqual(t, err.Error(), "no args provided")
}

func TestRunnerMustSortCommands(t *testing.T) {
	cmds := []Command{
		{Name: "foo", ExecFunc: nopFunc},
		{Name: "xyz"},
		{Name: "cake", ExecFunc: nopFunc},
		{Name: "foo2", ExecFunc: nopFunc},
	}
	cmds[1].Subcommands = []Command{
		{Name: "a", ExecFunc: nopFunc},
		{Name: "c", ExecFunc: nopFunc},
		{Name: "b", ExecFunc: nopFunc},
	}

	r := RunnerOf(cmds, Config{
		Args: []string{"./someapp", "foo"},
	})

	failIfErr(t, r.Run())

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
	r := RunnerOf([]Command{{Name: "foo", ExecFunc: nopFunc}}, Config{
		AppName: "exit-test",
		Output:  buf,
	})
	r.Exit(nil)

	mustEqual(t, exitDone, true)
	exitDone = false

	r.Exit(errors.New("oops"))
	mustEqual(t, exitDone, true)

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
			cmds:       []Command{{Name: "app:cre.ate", ExecFunc: nopFunc}},
			wantErrStr: ``,
		},
		{
			cmds:       []Command{{Name: "", ExecFunc: nopFunc}},
			wantErrStr: `command "" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo%", ExecFunc: nopFunc}},
			wantErrStr: `command "foo%" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "%", ExecFunc: nopFunc}},
			wantErrStr: `command alias "%" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo%", ExecFunc: nil}},
			wantErrStr: `command "foo%" exec function cannot be nil OR must have subcommands`,
		},
		{
			cmds:       []Command{{Name: "foo", ExecFunc: nil}},
			wantErrStr: `command "foo" exec function cannot be nil OR must have subcommands`,
		},
		{
			cmds: []Command{{
				Name:        "foobar",
				ExecFunc:    nopFunc,
				Subcommands: []Command{{Name: "nested"}},
			}},
			wantErrStr: `command "foobar" exec function cannot be set AND have subcommands`,
		},
		{
			cmds: []Command{{Name: "foo", ExecFunc: nopFunc}},
			cfg: Config{
				Args: []string{},
			},
			wantErrStr: `no args provided`,
		},
		{
			cmds:       []Command{{Name: "help", ExecFunc: nopFunc}},
			wantErrStr: `command "help" is reserved`,
		},
		{
			cmds:       []Command{{Name: "version", ExecFunc: nopFunc}},
			wantErrStr: `command "version" is reserved`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "help", ExecFunc: nopFunc}},
			wantErrStr: `command alias "help" is reserved`,
		},
		{
			cmds:       []Command{{Name: "foo", Alias: "version", ExecFunc: nopFunc}},
			wantErrStr: `command alias "version" is reserved`,
		},
		{
			cmds:       []Command{{Name: "a", ExecFunc: nopFunc}, {Name: "a", ExecFunc: nopFunc}},
			wantErrStr: `duplicate command "a"`,
		},
		{
			cmds:       []Command{{Name: "aaa", ExecFunc: nopFunc}, {Name: "b", Alias: "aaa", ExecFunc: nopFunc}},
			wantErrStr: `duplicate command alias "aaa"`,
		},
		{
			cmds:       []Command{{Name: "aaa", Alias: "a", ExecFunc: nopFunc}, {Name: "bbb", Alias: "a", ExecFunc: nopFunc}},
			wantErrStr: `duplicate command alias "a"`,
		},
		{
			cmds:       []Command{{Name: "a", ExecFunc: nopFunc}, {Name: "b", Alias: "a", ExecFunc: nopFunc}},
			wantErrStr: `duplicate command alias "a"`,
		},
	}

	for _, tc := range testCases {
		tc.cfg.Output = io.Discard
		err := RunnerOf(tc.cmds, tc.cfg).Run()
		failIfOk(t, err)

		if got := err.Error(); tc.wantErrStr != "" && !strings.Contains(got, tc.wantErrStr) {
			t.Fatalf("\nhave: %+v\nwant: %+v\n", tc.wantErrStr, got)
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
				{Name: "for", ExecFunc: nopFunc},
				{Name: "foo", ExecFunc: nopFunc},
				{Name: "bar", ExecFunc: nopFunc},
			},
			args: []string{"./someapp", "fooo"},
			want: `"fooo" unknown command, did you mean "foo"?` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", ExecFunc: nopFunc}},
			args: []string{"./someapp", "hell"},
			want: `"hell" unknown command, did you mean "help"?` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", ExecFunc: nopFunc}},
			args: []string{"./someapp", "verZION"},
			want: `"verZION" unknown command` + "\n" + `Run "myapp help" for usage.` + "\n\n",
		},
		{
			cmds: []Command{{Name: "for", ExecFunc: nopFunc}},
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

		mustEqual(t, buf.String(), tc.want)
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
		mustEqual(t, HasHelpFlag(tc.args), tc.hasHelp)
	}
}

func TestCommand_IsHidden(t *testing.T) {
	buf := &bytes.Buffer{}
	cmds := []Command{
		{Name: "for", ExecFunc: nopFunc},
		{Name: "foo", ExecFunc: nopFunc, IsHidden: true},
		{Name: "bar", ExecFunc: nopFunc},
	}
	r := RunnerOf(cmds, Config{
		Args:    []string{"./someapp", "help"},
		AppName: "myapp",
		Output:  buf,
	})
	failIfErr(t, r.Run())

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
			ExecFunc: func(ctx context.Context, args []string) error {
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
	failIfOk(t, err)

	var gotStatus int
	doExitOld := func(code int) {
		gotStatus = code
	}
	defer func() { doExit = doExitOld }()

	doExitOld, doExit = doExit, doExitOld

	r.Exit(err)

	mustEqual(t, gotStatus, wantStatus)
	mustEqual(t, buf.String(), wantOutput)
}

func failIfOk(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fail()
	}
}

func failIfErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatal(err)
	}
}

func mustEqual(tb testing.TB, have, want interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(have, want) {
		tb.Fatalf("\nhave: %+v\nwant: %+v\n", have, want)
	}
}
