package acmd

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed autocomplete/*
var shellScriptsFS embed.FS

func (r *Runner) completeInstallCmd(_ context.Context, args []string) error {
	var shell, binary, installDir, installFile string

	fset := flag.NewFlagSet("complete-install", flag.ContinueOnError)
	fset.StringVar(&shell, "shell", getShell(), "shell type")
	fset.StringVar(&binary, "binary", r.cfg.AppName, "binary name")
	fset.StringVar(&installDir, "dir", "", "dir to install")
	fset.StringVar(&installFile, "file", "", "file to install")
	if err := fset.Parse(args); err != nil {
		return err
	}

	script, err := r.completeScript(shell)
	if err != nil {
		return err
	}

	switch shell {
	case "bash":
		installDir = firstOrDef(installDir, "/etc/bash_completion.d")
		installFile = firstOrDef(installFile, binary+".bash")
	case "fish":
		installDir = firstOrDef(installDir, "/etc/fish/completions")
		installFile = firstOrDef(installFile, binary+".fish")
	case "power":
		// TODO
	case "zsh":
		installDir = firstOrDef(installDir, "/usr/local/share/zsh/site-functions")
		installFile = firstOrDef(installFile, "_"+binary)
	default:
		return fmt.Errorf("unknown shell: %s (want: bash, fish, power, zsh)", shell)
	}

	if err := os.MkdirAll(installDir, 0o700); err != nil {
		return err
	}
	filename := path.Join(installDir, installFile)
	return r.writeAutocompleteScript(filename, script)
}

func (r *Runner) writeAutocompleteScript(filename string, script []byte) error {
	fileFlags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	f, err := os.OpenFile(filename, fileFlags, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(script)
	return err
}

func (r *Runner) completeScriptCmd(_ context.Context, args []string) error {
	var shell string

	fset := flag.NewFlagSet("complete-script", flag.ContinueOnError)
	fset.StringVar(&shell, "shell", getShell(), "shell type")
	if err := fset.Parse(args); err != nil {
		return err
	}

	script, err := r.completeScript(shell)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.cfg.Output, string(script), r.cfg.AppName)
	return err
}

func (r *Runner) completeScript(shell string) ([]byte, error) {
	scriptFilename := filepath.Join("autocomplete", shell+".sh")
	return shellScriptsFS.ReadFile(scriptFilename)
}

func (r *Runner) completeQueryCmd(ctx context.Context, args []string) error {
	shell := getShell()

	w := r.cfg.Output
	opts := r.completeFor(Command{}, args)

	switch shell {
	case "bash":
		for _, opt := range opts {
			if opt.Alias != "" {
				fmt.Fprintln(w, opt.Alias)
			}
			fmt.Fprintln(w, opt.Name)
		}

	case "fish":
		for _, opt := range opts {
			if opt.Alias != "" {
				fmt.Fprintln(w, opt.Alias+"\t"+opt.Descr)
			}
			fmt.Fprintln(w, opt.Name+"\t"+opt.Descr)
		}

	case "power":
		// TODO

	case "zsh":
		fmt.Fprint(w, "(")
		for i, opt := range opts {
			if i != 0 {
				fmt.Fprint(w, " ")
			}
			r := strings.NewReplacer(
				"[", "(",
				"]", ")",
				"'", `"`,
			)
			descr := r.Replace(opt.Descr)

			if opt.Alias != "" {
				fmt.Fprintf(w, "'%s:%s' ", opt.Alias, descr)
			}
			fmt.Fprintf(w, "'%s:%s'", opt.Name, descr)
		}
		fmt.Fprint(w, ")")

	default:
		return fmt.Errorf("unknown shell: %s (want: bash, fish, power, zsh)", shell)
	}
	return nil
}

func (r *Runner) completeFor(cmd Command, args []string) []autocompleteEntry {
	fmt.Printf("# completeFor: %+v\n", args)
	flagSet := map[string]struct{}{}

	if cmd.Name == "" {
		for i, arg := range args {
			if strings.HasPrefix(arg, "-") {
				flagSet[arg] = struct{}{}
			} else {
				for _, cmd := range r.cmds {
					if cmd.Name == arg {
						return r.completeFor(cmd, args[i+1:])
					}
				}
			}
		}
	}

	var opts, ret []autocompleteEntry
	_ = opts
	return ret
}

type valueKind int

const (
	valueKindBool valueKind = iota
	valueKindSingle
	valueKindMulti
	valueKindFile
)

func (v valueKind) ExpectsValue() bool {
	return v == valueKindSingle || v == valueKindMulti
}

type autocompleteEntry struct {
	Name       string
	Alias      string
	Descr      string
	ValueKind  valueKind
	IsFilename bool
}

func (a autocompleteEntry) Equal(arg string) bool {
	return a.Name == arg ||
		a.Alias != "" && a.Alias == arg
}

func (a autocompleteEntry) Matches(arg string) bool {
	return strings.HasPrefix(a.Name, arg) ||
		a.Alias != "" && strings.HasPrefix(a.Alias, arg)
}

func getShell() string {
	shell := path.Base(os.Getenv("SHELL"))
	if shell == "sh" {
		return "bash"
	}
	return shell
}

func firstOrDef(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

type valueAliased interface {
	Alias() string
}
