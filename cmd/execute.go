package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/sflag"
)

func fatal(w io.Writer, fset *fset, msg string) (exitCode int) {
	fmt.Fprintf(w, "Error %s\n", msg)
	fmt.Fprint(w, fset.Usage())
	return 1
}

type fset struct {
	*sflag.FlagSet
}

func (f *fset) MultipleFlagReplaceInput(p *[]*lipo.ReplaceInput, name, usage string, opts ...sflag.FlagOption) {
	var idx, cur int
	from := func(v string) ([]*lipo.ReplaceInput, error) {
		if len(*p) <= idx {
			*p = append(*p, &lipo.ReplaceInput{})
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
	f.Var(sflag.FlagValues(p, from, cap), name, usage, opts...)
}

func Execute(stdout, stderr io.Writer, args []string) (exitCode int) {
	var out, thin string
	remove, extract, verifyArch := []string{}, []string{}, []string{}
	replace := []*lipo.ReplaceInput{}
	create := false
	archs := false

	fset := &fset{sflag.NewFlagSet("lipo")}
	createGroup := fset.NewGroup("create")
	thinGroup := fset.NewGroup("thin")
	extractGroup := fset.NewGroup("extract")
	removeGroup := fset.NewGroup("remove")
	replaceGroup := fset.NewGroup("replace")
	archsGroup := fset.NewGroup("archs")
	verifyArchGroup := fset.NewGroup("verify_arch")
	fset.String(&out, "output",
		"-output <output_file>",
		sflag.WithGroup(createGroup, sflag.TypeRequire),
		sflag.WithGroup(thinGroup, sflag.TypeRequire),
		sflag.WithGroup(extractGroup, sflag.TypeRequire),
		sflag.WithGroup(removeGroup, sflag.TypeRequire),
		sflag.WithGroup(replaceGroup, sflag.TypeRequire),
	)
	fset.Bool(&create, "create",
		"-create",
		sflag.WithGroup(createGroup, sflag.TypeRequire))
	fset.String(&thin, "thin",
		"-thin <arch_type>",
		sflag.WithGroup(thinGroup, sflag.TypeRequire))
	fset.MultipleFlagString(&extract, "extract",
		"-extract <arch_type> [-extract <arch_type> ...]",
		sflag.WithGroup(extractGroup, sflag.TypeRequire))
	fset.MultipleFlagString(&remove, "remove",
		"-remove <arch_type> [-remove <arch_type> ...]",
		sflag.WithGroup(removeGroup, sflag.TypeRequire))
	fset.MultipleFlagReplaceInput(&replace, "replace",
		"-replace <arch> <file>",
		sflag.WithGroup(replaceGroup, sflag.TypeRequire))
	fset.Bool(&archs, "archs",
		"-archs",
		sflag.WithGroup(archsGroup, sflag.TypeRequire))
	fset.FlexStrings(&verifyArch, "verify_arch",
		"-verify_arch <arch_type> ...",
		sflag.WithGroup(verifyArchGroup, sflag.TypeRequire))

	if err := fset.Parse(args); err != nil {
		return fatal(stderr, fset, err.Error())
	}

	group, err := sflag.LookupGroup(
		createGroup, thinGroup, extractGroup,
		removeGroup, replaceGroup, archsGroup,
		verifyArchGroup)
	if err != nil {
		return fatal(stderr, fset, err.Error())
	}

	in := fset.Args()
	l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
	switch group.Name {
	case "create":
		if err := l.Create(); err != nil {
			return fatal(stderr, fset, err.Error())
		}
		return
	case "thin":
		if err := l.Thin(thin); err != nil {
			return fatal(stderr, fset, err.Error())
		}
		return
	case "remove":
		if err := l.Remove(remove...); err != nil {
			return fatal(stderr, fset, err.Error())
		}
		return
	case "extract":
		if err := l.Extract(extract...); err != nil {
			return fatal(stderr, fset, err.Error())
		}
		return
	case "replace":
		if err := l.Replace(replace); err != nil {
			return fatal(stderr, fset, err.Error())
		}
		return
	case "archs":
		arches, err := l.Archs()
		if err != nil {
			return fatal(stderr, fset, err.Error())
		}
		fmt.Fprintln(stdout, strings.Join(arches, " "))
		return
	case "verify_arch":
		ok, err := l.VerifyArch(verifyArch...)
		if err != nil {
			return fatal(stderr, fset, err.Error())
		}
		if !ok {
			return 1
		}
		return
	default:
		fset.Usage()
		return 1
	}
}
