package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/sflag"
)

func fatal(fset *fset, msg string) (exitCode int) {
	fmt.Fprintf(fset.Out(), "Error %s\n", msg)
	return 1
}

type fset struct {
	*sflag.FlagSet
}

func (f *fset) MultipleFlagReplaceInputs(p *[]lipo.ReplaceInput, name, usage string) {
	var idx, cur int
	from := func(v string) ([]lipo.ReplaceInput, error) {
		if len(*p) <= idx {
			*p = append(*p, lipo.ReplaceInput{})
		}
		if cur == 0 {
			(*p)[idx].Arch = v
		} else if cur == 1 {
			(*p)[idx].Bin = v
		} else {
			return nil, fmt.Errorf("out of index %d", cur)
		}
		cur++
		return *p, nil
	}
	cap := func() int {
		cap := 2 - cur
		if cap == 0 {
			cur = 0
			idx++
		}
		return cap
	}
	f.Var(sflag.FlagValues(p, from, cap), name, usage)
}

func Execute(w io.Writer, args []string) (exitCode int) {
	var out, thin string
	remove, extract, verifyArch := []string{}, []string{}, []string{}
	replace := []lipo.ReplaceInput{}
	create := false
	archs := false

	fset := &fset{sflag.NewFlagSet("lipo", sflag.WithOut(w))}
	fset.String(&out, "output", "-output <output_file>")
	fset.String(&thin, "thin", "-thin <arch_type>")
	fset.Bool(&create, "create", "-create")
	fset.MultipleFlagString(&extract, "extract", "-extract <arch_type> [-extract <arch_type> ...]")
	fset.MultipleFlagString(&remove, "remove", "-remove <arch_type> [-remove <arch_type> ...]")
	fset.Bool(&archs, "archs", "-archs")
	fset.FlexStrings(&verifyArch, "verify_arch", "-verify_arch <arch_type> ...")
	fset.MultipleFlagReplaceInputs(&replace, "replace", "-replace <arch> <file>")
	if err := fset.Parse(args); err != nil {
		return fatal(fset, err.Error())
	}

	in := fset.Args()
	if create {
		if out == "" {
			return fatal(fset, "-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Create(); err != nil {
			return fatal(fset, err.Error())
		}
		return
	}

	if thin != "" {
		if out == "" {
			return fatal(fset, "-output flag is required")
		}
		l := lipo.New(lipo.WithInputs(out), lipo.WithInputs(in...))
		if err := l.Thin(thin); err != nil {
			return fatal(fset, err.Error())
		}
		return
	}

	if len(remove) != 0 {
		if out == "" {
			return fatal(fset, "-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Remove(remove...); err != nil {
			return fatal(fset, err.Error())
		}
		return
	}

	if len(extract) != 0 {
		if out == "" {
			return fatal(fset, "-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Extract(extract...); err != nil {
			return fatal(fset, err.Error())
		}
	}

	if len(replace) != 0 {
		l := lipo.New(lipo.WithInputs(in...), lipo.WithOutput(out))
		if err := l.Replace(replace); err != nil {
			return fatal(fset, err.Error())
		}
		return
	}

	if archs {
		l := lipo.New(lipo.WithInputs(in...))
		arches, err := l.Archs()
		if err != nil {
			return fatal(fset, err.Error())
		}
		fmt.Fprintln(fset.Out(), strings.Join(arches, " "))
		return
	}

	if len(verifyArch) != 0 {
		l := lipo.New(lipo.WithInputs(in...))
		ok, err := l.VerifyArch(verifyArch...)
		if err != nil {
			return fatal(fset, err.Error())
		}
		if !ok {
			return 1
		}
		return
	}

	fset.Usage()
	return 1
}
