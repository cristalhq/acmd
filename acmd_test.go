package acmd

import (
	"bytes"
	"context"
	"testing"
)

var nopFunc = func(context.Context, []string) error { return nil }

func TestRunner_init(t *testing.T) {
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
		err := RunnerOf(tc.cmds, tc.cfg).init()

		if got := err.Error(); got != tc.wantErrStr {
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
