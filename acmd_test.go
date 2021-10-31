package acmd

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"testing"
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

func TestRunnerInit(t *testing.T) {
	testCases := []struct {
		cmds       []Command
		cfg        Config
		wantErrStr string
	}{
		{
			cmds:       []Command{{Name: "foo%", Do: nopFunc}},
			wantErrStr: `command "foo%" must contains only letters, digits, - and _`,
		},
		{
			cmds:       []Command{{Name: "foo%", Do: nil}},
			wantErrStr: `command "foo%" function cannot be nil`,
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
			cmds:       []Command{{Name: "a", Do: nopFunc}, {Name: "a", Do: nopFunc}},
			wantErrStr: `duplicate command "a"`,
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
			cmds: []Command{},
			args: []string{"hell"},
			want: `"hell" is not a subcommand, did you mean "help"?` + "\n",
		},
		{
			cmds: []Command{},
			args: []string{"verZION"},
			want: "",
		},
		{
			cmds: []Command{},
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
		if err := r.Run(); err == nil {
			t.Fatal()
		}

		if got := buf.String(); got != tc.want {
			t.Logf("want %q got %q", tc.want, got)
		}
	}
}
