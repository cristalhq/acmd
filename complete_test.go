package acmd

import (
	"context"
	"testing"
)

func Test_completeInstallCmd(t *testing.T) {
	cmds := []Command{{
		Name:     "foo",
		ExecFunc: nopFunc,
	}}
	r := RunnerOf(cmds, Config{
		AutoComplete: true,
	})

	testCases := []struct {
		name string
		args []string
	}{
		{
			args: []string{"./app", "__complete"},
		},
	}

	for _, tc := range testCases {
		r.completeInstallCmd(context.Background(), tc.args)
	}
}
