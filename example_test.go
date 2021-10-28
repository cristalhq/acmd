package acmd_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cristalhq/acmd"
)

func ExampleRunner() {
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
		Context:        nil,
		Args:           []string{"now"},
		Usage:          nil,
	})

	if err := r.Run(); err != nil {
		panic(err)
	}

	// Output: now: 10:20:30
}
