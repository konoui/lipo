package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/sflag"
)

var osExit = os.Exit

func fatal(msg string) {
	fmt.Printf("Error %s\n", msg)
	osExit(1)
}

func main() {
	var out string
	remove, extract, verifyArch := []string{}, []string{}, []string{}
	replace := [][]string{make([]string, 2)}
	create := false
	archs := false

	fset := sflag.NewFlagSet("lipo")
	fset.String(&out, "output", "-output <file>")
	fset.Bool(&create, "create", "-create")
	fset.MultipleFlagFixedStrings(&replace, "replace", "-replace <arch> <file>")
	fset.MultipleFlagStrings(&extract, "extract", "-extract <arch>")
	fset.MultipleFlagStrings(&remove, "remove", "-remove <arch>")
	fset.Bool(&archs, "archs", "-archs <arch> ...")
	fset.FlexStrings(&verifyArch, "verify_arch", "verify_arch <arch> ...")

	if err := fset.Parse(os.Args[1:]); err != nil {
		fatal(err.Error())
	}

	in := fset.Args()
	if create {
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Create(); err != nil {
			fatal(err.Error())
		}
		return
	}

	if len(remove) != 0 {
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Remove(remove...); err != nil {
			fatal(err.Error())
		}
		return
	}

	if len(extract) != 0 {
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Extract(extract...); err != nil {
			fatal(err.Error())
		}
		return
	}

	if len(replace) != 0 {
		l := lipo.New(lipo.WithInputs(in...), lipo.WithOutput(out))
		ris, err := lipo.ReplaceInputs(replace)
		if err != nil {
			fatal(err.Error())
		}
		if err := l.Replace(ris); err != nil {
			fatal(err.Error())
		}
		return
	}

	if archs {
		l := lipo.New(lipo.WithInputs(in...))
		arches, err := l.Archs()
		if err != nil {
			fatal(err.Error())
		}
		fmt.Fprintln(os.Stdout, strings.Join(arches, " "))
		return
	}

	if len(verifyArch) != 0 {
		l := lipo.New(lipo.WithInputs(in...))
		ok, err := l.VerifyArch(verifyArch...)
		if err != nil {
			fatal(err.Error())
		}
		if !ok {
			os.Exit(1)
		}
		return
	}
}
