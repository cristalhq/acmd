# Guide for acmd

## Flag

Example with command like that `./dummy server ./openapi.yml -port=8080` from [dummy](https://github.com/neotoolkit/dummy/blob/main/cmd/dummy/main.go)

```go
func run() error {
	cmds := []acmd.Command{
		{
			// Command "server"
			Name:        "server",
			Description: "run mock server",
			Do: func(ctx context.Context, args []string) error {
				cfg := config.NewConfig()
                
				// Argument ./openapi.yml
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

## Flags propagation

There is no special methods, config fields to propagate flags to subcommands. However it's not hard to make this, because every command can access predefined flags, which are shared across handlers.

```go
type commonFlags struct {
	IsVerbose bool
}

// NOTE: should be added before flag.FlagSet method Parse().
func withCommonFlags(fs *flag.FlagSet) *commonFlags {
	c := &commonFlags{}
	fs.BoolVar(&c.IsVerbose, "verbose", false, "should app be verbose")
	return c
}

func cmdFoo(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("foo", flag.ContinueOnError)
	// NOTE: here add flags for cmdBar as always

	// add common flags, make sure it's before Parse but after all defined flags
	common := withCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	// use commonFlags fields or any other flags that you have defined
	return nil
}

func cmdBar(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("bar", flag.ContinueOnError)
	// NOTE: here add flags for cmdFoo as always

	// add common flags, make sure it's before Parse but after all defined flags
	common := withCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	// use commonFlags fields or any other flags that you have defined
	return nil
}
```

Also see `ExamplePropagateFlags` test.
