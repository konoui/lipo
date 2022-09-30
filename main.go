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

var _ sflag.Values = &replaceInputs{}

type replaceInputs struct {
	slice *[]lipo.ReplaceInput
	idx   int
	cur   int
}

func (ri *replaceInputs) Set(v string) error {
	if len(*ri.slice) <= ri.idx {
		*ri.slice = append(*ri.slice, lipo.ReplaceInput{})
	}
	if ri.cur == 0 {
		(*ri.slice)[ri.idx].Arch = v
	} else if ri.cur == 1 {
		(*ri.slice)[ri.idx].Bin = v
	} else {
		return fmt.Errorf("fill error. cur %d, slice %v", ri.cur, ri.slice)
	}
	return nil
}

func (ri *replaceInputs) Cap() int {
	cap := 2 - ri.cur
	if cap == 0 {
		ri.cur = 0
		ri.idx++
	}
	return cap
}

func MultipleFlagReplaceInputs(fset *sflag.FlagSet, inputs *[]lipo.ReplaceInput, name, usage string) {
	fset.Var(&replaceInputs{slice: &[]lipo.ReplaceInput{}}, name, usage)
}

func main() {
	var out string
	remove, extract, verifyArch := []string{}, []string{}, []string{}
	replace := []lipo.ReplaceInput{}
	create := false
	archs := false

	fset := sflag.NewFlagSet("lipo")
	fset.String(&out, "output", "-output <file>")
	fset.Bool(&create, "create", "-create")
	fset.MultipleFlagString(&extract, "extract", "-extract <arch>")
	fset.MultipleFlagString(&remove, "remove", "-remove <arch>")
	fset.Bool(&archs, "archs", "-archs <arch> ...")
	fset.FlexStrings(&verifyArch, "verify_arch", "verify_arch <arch> ...")
	MultipleFlagReplaceInputs(fset, &replace, "replace", "-replace <arch> <file>")
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
		if err := l.Replace(replace); err != nil {
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
