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

func Execute(stdout, stderr io.Writer, args []string) (exitCode int) {
	var out, thin string
	remove, extract, verifyArch := []string{}, []string{}, []string{}
	replace := [][2]string{}
	segAligns := [][2]string{}
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
	fset.MultipleFlagFixedStrings(&segAligns, "segalign", "<arch_type> <alignment>",
		sflag.WithGroup(createGroup, sflag.TypeOption),
		sflag.WithGroup(thinGroup, sflag.TypeOption), // apple lipo does not raise error if -thin with -segalign
		sflag.WithGroup(extractGroup, sflag.TypeOption),
		sflag.WithGroup(removeGroup, sflag.TypeOption),
		sflag.WithGroup(replaceGroup, sflag.TypeOption),
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
	fset.MultipleFlagFixedStrings(&replace, "replace",
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
	l := lipo.New(
		lipo.WithOutput(out),
		lipo.WithInputs(in...),
		lipo.WithSegAlign(conv(segAligns, newSegAlign)))
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
		if err := l.Replace(conv(replace, newReplace)); err != nil {
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

func newSegAlign(r [2]string) *lipo.SegAlignInput {
	return &lipo.SegAlignInput{Arch: r[0], AlignHex: r[1]}
}

func newReplace(r [2]string) *lipo.ReplaceInput {
	return &lipo.ReplaceInput{Arch: r[0], Bin: r[1]}
}

func conv[T any](raw [][2]string, f func([2]string) T) []T {
	ret := make([]T, 0, len(raw))
	for _, r := range raw {
		ret = append(ret, f(r))
	}
	return ret
}
