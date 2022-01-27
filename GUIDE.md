# acmd guides

## Flag
Example with command like that `./dummy server ./openapi.yml -port=8080` from [dummy](https://github.com/go-dummy/dummy/blob/main/cmd/dummy/main.go)
```go
func run() error {
	cmds := []acmd.Command{
		{
			// Command
			Name:        "server",
			Description: "run mock server",
			Do: func(ctx context.Context, args []string) error {
				cfg := config.NewConfig()
                
				// Argument ./openapi
				cfg.Server.Path = args[0]
                
				// Flags after argument
				fs := flag.NewFlagSet("dummy", flag.ContinueOnError)
				fs.StringVar(&cfg.Server.Port, "port", "8080", "")
				fs.StringVar(&cfg.Logger.Level, "logger-level", "INFO", "")
				if err := fs.Parse(args[1:]); err != nil {
					return err
				}

				...
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "Dummy",
		Version: version,
	})

	return r.Run()
}
```
