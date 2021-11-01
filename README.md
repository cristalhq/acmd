# acmd

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]

Simple, useful and opinionated CLI package in Go. For config loader see [aconfig](https://github.com/cristalhq/aconfig)

## Rationale

Popular CLI libraries (or better frameworks) have too large and unclear API, in most cases, you just want to define commands for your CLI application and run them without additional work. This package does this by providing a small API, good defaults and clear code.

## Features

* Simple API.
* Dependency-free.
* Clean and tested code.
* Command aliases.
* Auto suggesting command.
* Builtin `help` and `version` commands.

## Install

Go version 1.17+

```
go get github.com/cristalhq/acmd
```

## Example

```go
cmds := []acmd.Command{
	{
		Name:        "now",
		Description: "prints current time",
		Do: func(ctx context.Context, args []string) error {
			fmt.Printf("now: %s\n", now.Format("15:04:05"))
			return nil
		},
	},
	{
		Name:        "status",
		Description: "prints status of the system",
		Do: func(ctx context.Context, args []string) error {
			// do something with ctx :)
			return nil
		},
	},
}

// all the acmd.Config fields are optional
r := acmd.RunnerOf(cmds, acmd.Config{
	AppName:        "acmd-example",
	AppDescription: "Example of acmd package",
	Version:        "the best v0.x.y",
	// Context - if nil `signal.Notify` will be used
	// Args - if nil `os.Args[1:]` will be used
	// Usage - if nil default print will be used
})

if err := r.Run(); err != nil {
	panic(err)
}
```

Also see examples: [examples_test.go](https://github.com/cristalhq/acmd/blob/main/example_test.go).

## Documentation

See [these docs][pkg-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristalhq/acmd/workflows/build/badge.svg
[build-url]: https://github.com/cristalhq/acmd/actions
[pkg-img]: https://pkg.go.dev/badge/cristalhq/acmd
[pkg-url]: https://pkg.go.dev/github.com/cristalhq/acmd
[reportcard-img]: https://goreportcard.com/badge/cristalhq/acmd
[reportcard-url]: https://goreportcard.com/report/cristalhq/acmd
[coverage-img]: https://codecov.io/gh/cristalhq/acmd/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristalhq/acmd
