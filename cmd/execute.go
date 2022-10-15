package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/sflag"
)

func fatal(w io.Writer, g *sflag.Group, msg string) (exitCode int) {
	fmt.Fprintf(w, "Error %s\n", msg)
	fmt.Fprint(w, g.Usage())
	return 1
}

type fset struct {
	*sflag.FlagSet
}

func Execute(stdout, stderr io.Writer, args []string) (exitCode int) {
	var out, thin string
	remove, extract, extractFamily, verifyArch := []string{}, []string{}, []string{}, []string{}
	replace, segAligns, arch := [][2]string{}, [][2]string{}, [][2]string{}
	create, archs, info, detailedInfo := false, false, false, false

	fset := &fset{sflag.NewFlagSet("lipo")}
	createGroup := fset.NewGroup("create").AddDescription(createDescription)
	thinGroup := fset.NewGroup("thin").AddDescription(thinDescription)
	extractGroup := fset.NewGroup("extract").AddDescription(extractDescription)
	extractFamilyGroup := fset.NewGroup("extract_family").AddDescription(extractFamilyDescription)
	removeGroup := fset.NewGroup("remove").AddDescription(removeDescription)
	replaceGroup := fset.NewGroup("replace").AddDescription(replaceDescription)
	archsGroup := fset.NewGroup("archs").AddDescription(archsDescription)
	verifyArchGroup := fset.NewGroup("verify_arch").AddDescription(verifyArchDescription)
	infoGroup := fset.NewGroup("info").AddDescription(infoDescription)
	detailedInfoGroup := fset.NewGroup("detailed_info").AddDescription(detailedInfoDescription)
	groups := []*sflag.Group{createGroup, thinGroup, extractGroup,
		extractFamilyGroup, removeGroup, replaceGroup,
		archsGroup, verifyArchGroup, infoGroup,
		detailedInfoGroup,
	}
	fset.Usage = sflag.UsageFunc(groups...)
	fset.String(&out, "output",
		"-output <output_file>",
		sflag.WithGroup(createGroup, sflag.TypeRequire),
		sflag.WithGroup(thinGroup, sflag.TypeRequire),
		sflag.WithGroup(extractGroup, sflag.TypeRequire),
		sflag.WithGroup(removeGroup, sflag.TypeRequire),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeRequire),
		sflag.WithGroup(replaceGroup, sflag.TypeRequire),
	)
	fset.MultipleFlagFixedStrings(&segAligns, "segalign", "-segalign <arch_type> <alignment>",
		sflag.WithGroup(createGroup, sflag.TypeOption),
		sflag.WithGroup(thinGroup, sflag.TypeOption), // apple lipo does not raise error if -thin with -segalign
		sflag.WithGroup(extractGroup, sflag.TypeOption),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeOption),
		sflag.WithGroup(removeGroup, sflag.TypeOption),
		sflag.WithGroup(replaceGroup, sflag.TypeOption),
	)
	fset.MultipleFlagFixedStrings(&arch, "arch", "-arch <arch_type> <input_file>",
		sflag.WithGroup(createGroup, sflag.TypeOption),
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
		sflag.WithGroup(extractGroup, sflag.TypeRequire),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeOption)) // if specified, apple lipo regard values as family
	fset.MultipleFlagString(&extractFamily, "extract_family",
		"-extract_family <arch_type> [-extract_family <arch_type> ...]",
		sflag.WithGroup(extractFamilyGroup, sflag.TypeRequire))
	fset.MultipleFlagString(&remove, "remove",
		"-remove <arch_type> [-remove <arch_type> ...]",
		sflag.WithGroup(removeGroup, sflag.TypeRequire))
	fset.MultipleFlagFixedStrings(&replace, "replace",
		"-replace <arch_type> <file_name> [-replace <arch_type> <file_name> ...]",
		sflag.WithGroup(replaceGroup, sflag.TypeRequire))
	fset.Bool(&archs, "archs",
		"-archs",
		sflag.WithGroup(archsGroup, sflag.TypeRequire))
	fset.FlexStrings(&verifyArch, "verify_arch",
		"-verify_arch <arch_type> ...",
		sflag.WithGroup(verifyArchGroup, sflag.TypeRequire))
	fset.Bool(&info, "info",
		"-info",
		sflag.WithGroup(infoGroup, sflag.TypeRequire))
	fset.Bool(&detailedInfo, "detailed_info",
		"-detailed_info",
		sflag.WithGroup(detailedInfoGroup, sflag.TypeRequire))
	if err := fset.Parse(args); err != nil {
		fmt.Fprint(stderr, fset.Usage())
		return 1
	}

	if len(args) == 0 {
		fmt.Fprint(stderr, fset.Usage())
		return 1
	}

	group, err := sflag.LookupGroup(groups...)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		if group != nil {
			fmt.Fprint(stderr, group.Usage())
		} else {
			fmt.Fprint(stderr, fset.Usage())
		}
		return 1
	}

	in := fset.Args()
	l := lipo.New(
		lipo.WithOutput(out),
		lipo.WithInputs(in...),
		lipo.WithArch(conv(arch, newArch)),
		lipo.WithSegAlign(conv(segAligns, newSegAlign)),
	)
	switch group.Name {
	case "create":
		if err := l.Create(); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "thin":
		if err := l.Thin(thin); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "remove":
		if err := l.Remove(remove...); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "extract":
		if err := l.Extract(extract...); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "extract_family":
		extractFamily = append(extractFamily, extract...)
		if err := l.ExtractFamily(extractFamily...); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "replace":
		if err := l.Replace(conv(replace, newReplace)); err != nil {
			return fatal(stderr, group, err.Error())
		}
		return
	case "archs":
		arches, err := l.Archs()
		if err != nil {
			return fatal(stderr, group, err.Error())
		}
		fmt.Fprintln(stdout, strings.Join(arches, " "))
		return
	case "info":
		v, err := l.Info()
		if err != nil {
			return fatal(stderr, group, err.Error())
		}
		fmt.Fprintln(stdout, strings.Join(v, "\n"))
		return
	case "detailed_info":
		v, err := l.DetailedInfo()
		if err != nil {
			return fatal(stderr, group, err.Error())
		}
		fmt.Fprint(stdout, v)
		return
	case "verify_arch":
		ok, err := l.VerifyArch(verifyArch...)
		if err != nil {
			return fatal(stderr, group, err.Error())
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

func newArch(r [2]string) *lipo.ArchInput {
	return &lipo.ArchInput{Arch: r[0], Bin: r[1]}
}

func conv[T any](raw [][2]string, f func([2]string) T) []T {
	ret := make([]T, 0, len(raw))
	for _, r := range raw {
		ret = append(ret, f(r))
	}
	return ret
}
