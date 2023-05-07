package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/sflag"
)

func fatal(w io.Writer, msg string) (exitCode int) {
	fmt.Fprintf(w, "Error: %s\n", msg)
	return 1
}

func Execute(stdout, stderr io.Writer, args []string) (exitCode int) {
	fset := sflag.NewFlagSet("lipo")
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
	out := fset.String("output", "-output <output_file>", sflag.WithShortName("o"))
	segAligns := fset.FixedStringFlags("segalign", "-segalign <arch_type> <alignment>", sflag.WithShortName("s"))
	arch := fset.FixedStringFlags("arch", "-arch <arch_type> <input_file>", sflag.WithShortName("a"))
	create := fset.Bool("create", "-create", sflag.WithShortName("c"))
	thin := fset.String("thin", "-thin <arch_type>", sflag.WithShortName("t"))
	extract := fset.StringFlags("extract", "-extract <arch_type> [-extract <arch_type> ...]", sflag.WithShortName("e"))
	extractFamily := fset.StringFlags("extract_family", "-extract_family <arch_type> [-extract_family <arch_type> ...]")
	remove := fset.StringFlags("remove", "-remove <arch_type> [-remove <arch_type> ...]", sflag.WithShortName("rem"))
	replace := fset.FixedStringFlags("replace", "-replace <arch_type> <file_name> [-replace <arch_type> <file_name> ...]", sflag.WithShortName("rep"))
	archs := fset.Bool("archs", "-archs")
	verifyArch := fset.Strings("verify_arch", "-verify_arch <arch_type> ...")
	info := fset.Bool("info", "-info", sflag.WithShortName("i"))
	detailedInfo := fset.Bool("detailed_info", "-detailed_info", sflag.WithShortName("d"))
	hideArm64 := fset.Bool("hideARM64", "-hideARM64")
	fat64 := fset.Bool("fat64", "-fat64")

	createGroup.
		AddRequired(create).
		AddRequired(out).
		AddOptional(segAligns).
		AddOptional(arch).
		AddOptional(hideArm64).
		AddOptional(fat64)
	thinGroup.
		// apple lipo does not raise error if -thin with -segalign but this this lipo will raise an error
		AddRequired(thin).
		AddRequired(out)
	extractGroup.
		AddRequired(extract).
		AddRequired(out).
		AddOptional(segAligns).
		AddOptional(fat64)
	extractFamilyGroup.
		AddRequired(extractFamily).
		AddRequired(out).
		// if extract is specified, apple lipo regard values as family
		AddOptional(extract).
		AddOptional(segAligns).
		AddOptional(fat64)
	removeGroup.
		AddRequired(remove).
		AddRequired(out).
		AddOptional(segAligns).
		AddOptional(hideArm64).
		AddOptional(fat64)
	replaceGroup.
		AddRequired(replace).
		AddRequired(out).
		AddOptional(segAligns).
		AddOptional(arch).
		AddOptional(hideArm64).
		AddOptional(fat64)
	archsGroup.
		AddRequired(archs)
	verifyArchGroup.
		AddRequired(verifyArch)
	infoGroup.
		AddRequired(info)
	detailedInfoGroup.
		AddRequired(detailedInfo)

	if err := fset.Parse(args); err != nil {
		fmt.Fprintf(stderr, "ParseError: %s\n", err.Error())
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
		}
		return 1
	}

	in := fset.Args()
	opts := []lipo.Option{
		lipo.WithOutput(out.Get()),
		lipo.WithInputs(in...),
		lipo.WithArch(conv(arch.Get(), newArch)...),
		lipo.WithSegAlign(conv(segAligns.Get(), newSegAlign)...),
	}
	if hideArm64.Get() {
		opts = append(opts, lipo.WithHideArm64())
	}
	if fat64.Get() {
		opts = append(opts, lipo.WithFat64())
	}
	l := lipo.New(opts...)
	switch group.Name {
	case "create":
		if err := l.Create(); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "thin":
		if err := l.Thin(thin.Get()); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "remove":
		if err := l.Remove(remove.Get()...); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "extract":
		if err := l.Extract(extract.Get()...); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "extract_family":
		extractFamily := extractFamily.Get()
		extractFamily = append(extractFamily, extract.Get()...)
		if err := l.ExtractFamily(extractFamily...); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "replace":
		if err := l.Replace(conv(replace.Get(), newArch)); err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "archs":
		arches, err := l.Archs()
		if err != nil {
			return fatal(stderr, err.Error())
		}
		fmt.Fprintln(stdout, strings.Join(arches, " "))
		return
	case "info":
		v, err := l.Info()
		if err != nil {
			return fatal(stderr, err.Error())
		}
		fmt.Fprintln(stdout, strings.Join(v, "\n"))
		return
	case "detailed_info":
		err := l.DetailedInfo(stdout)
		if err != nil {
			return fatal(stderr, err.Error())
		}
		return
	case "verify_arch":
		ok, err := l.VerifyArch(verifyArch.Get()...)
		if err != nil {
			return fatal(stderr, err.Error())
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
