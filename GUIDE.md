# Guide for acmd

## Default command

Sometimes the app should work without any commands (ex `./app`), `acmd` doesn't have a feature for that, if command is not passed `help` will be used.

But if you really want to have default command for the app you can do:

```go
func main() {
	cmds := []acmd.Command{...}

	if len(os.Args) <= 1 { // no command is passed?
	    os.Args = []string{"", cmds[0].Name} // change the app args
	}

	// just create acmd.Runner as usual
}
```

Note: this solution doesn't work with flags, consider to create a command even if you have just 1.

## Flag

Example with command like that `./dummy server ./openapi.yml -port=8080` from [dummy](https://github.com/neotoolkit/dummy/blob/main/cmd/dummy/main.go)

```go
func run() error {
	cmds := []acmd.Command{
		{
			// Command "server"
			Name:        "server",
			Description: "run mock server",
			ExecFunc: func(ctx context.Context, args []string) error {
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
// generalFlags can be used as flags for all command
type generalFlags struct {
	IsVerbose bool
	Dir       string
}

func (c *generalFlags) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&c.IsVerbose, "verbose", false, "should app be verbose")
	fs.StringVar(&c.Dir, "dir", ".", "directory to process")
	return fs
}

// commandFlags is a flags for a command
// using struct embedding we can inherit other flags
type commandFlags struct {
	generalFlags
	File string
}

func (c *commandFlags) Flags() *flag.FlagSet {
	fs := c.generalFlags.Flags()
	fs.StringVar(&c.File, "file", "input.txt", "file to process")
	return fs
}

func cmdFoo(ctx context.Context, args []string) error {
	var cfg generalFlags
	if fs := cfg.Flags().Parse(args); err != nil{
		return err
	}

	// use cfg fields or any other flags that you have defined
	return nil
}

func cmdBar(ctx context.Context, args []string) error {
	var cfg commandFlags
	if fs := cfg.Flags().Parse(args); err != nil{
		return err
	}
	
	// use cfg fields or any other flags that you have defined
	return nil
}
```

Also see `ExamplePropagateFlags` test.

## Build version

Let's assume you have `var Version string` in `main` package. To fill this variable and then use as `acmd.Config.Version` you can do:

```sh
$ go build -ldflags="-X 'main.Version=(local)'" -o my_binary .

$ ./my_binary version
./my_binary version: (local)
```

Starting from Go 1.18 this information is avaliable in `runtime/debug.BuildInfo`, see: https://github.com/golang/go/issues/37475

To help with the wordy API to get these values you can use `acmd.GetBuildInfo()` function, which returns struct with `revision`, `last commit` and `is dirty build` fields.

Note: there is a chance that you will build application and will find that data is not populated, probably you're observing https://github.com/golang/go/issues/51279. Consider to use `go install <your module>`.
