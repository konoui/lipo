package sflag_test

import (
	"testing"

	"github.com/konoui/lipo/pkg/sflag"
)

var (
	output, thin                               string
	remove, extract, extractFamily, verifyArch = []string{}, []string{}, []string{}, []string{}
	replace, segAligns                         = [][2]string{}, [][2]string{}
	create, archs                              = false, false
)

func register() (*sflag.FlagSet, []*sflag.Group) {
	f := sflag.NewFlagSet("lipo")
	// init
	output, thin = "", ""
	remove, extract, extractFamily, verifyArch = []string{}, []string{}, []string{}, []string{}
	replace, segAligns = [][2]string{}, [][2]string{}
	create, archs = false, false

	createGroup := f.NewGroup("create")
	thinGroup := f.NewGroup("thin")
	extractGroup := f.NewGroup("extract")
	extractFamilyGroup := f.NewGroup("extract_family")
	removeGroup := f.NewGroup("remove")
	replaceGroup := f.NewGroup("replace")
	archsGroup := f.NewGroup("archs")
	verifyArchGroup := f.NewGroup("verify_arch")
	f.String(&output, "output", "-output <file>",
		sflag.WithGroup(createGroup, sflag.TypeRequired),
		sflag.WithGroup(thinGroup, sflag.TypeRequired),
		sflag.WithGroup(extractGroup, sflag.TypeRequired),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeRequired),
		sflag.WithGroup(removeGroup, sflag.TypeRequired),
		sflag.WithGroup(replaceGroup, sflag.TypeRequired))
	f.MultipleFlagFixedStrings(&segAligns, "segalign", "<arch_type> <alignment>",
		sflag.WithGroup(createGroup, sflag.TypeOption),
		sflag.WithGroup(thinGroup, sflag.TypeOption),
		sflag.WithGroup(extractGroup, sflag.TypeOption),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeOption),
		sflag.WithGroup(removeGroup, sflag.TypeOption),
		sflag.WithGroup(replaceGroup, sflag.TypeOption),
	)
	f.Bool(&create, "create", "-create",
		sflag.WithGroup(createGroup, sflag.TypeRequired))
	f.String(&thin, "thin", "thin <arch>",
		sflag.WithGroup(thinGroup, sflag.TypeRequired))
	f.MultipleFlagFixedStrings(&replace, "replace", "-replace <arch> <file>",
		sflag.WithGroup(replaceGroup, sflag.TypeRequired))
	f.MultipleFlagString(&extract, "extract", "-extract <arch>",
		sflag.WithGroup(extractGroup, sflag.TypeRequired),
		sflag.WithGroup(extractFamilyGroup, sflag.TypeOption)) // if specified, apple lipo regard values as family
	f.MultipleFlagString(&extractFamily, "extract_family",
		"-extract_family <arch>",
		sflag.WithGroup(extractFamilyGroup, sflag.TypeRequired))
	f.MultipleFlagString(&remove, "remove", "-remove <arch>",
		sflag.WithGroup(removeGroup, sflag.TypeRequired))
	f.Bool(&archs, "archs", "-archs <arch> ...",
		sflag.WithGroup(archsGroup, sflag.TypeRequired))
	f.FlexStrings(&verifyArch, "verify_arch", "verify_arch <arch>",
		sflag.WithGroup(verifyArchGroup, sflag.TypeRequired))
	return f, []*sflag.Group{
		createGroup, thinGroup, extractGroup, extractFamilyGroup,
		removeGroup, replaceGroup, archsGroup,
		verifyArchGroup}
}

func fset(t *testing.T, in []string) (*sflag.FlagSet, *sflag.Group) {
	f, groups := register()
	if err := f.Parse(in); err != nil {
		t.Fatal(err)
	}

	group, err := sflag.LookupGroup(groups...)
	if err != nil {
		t.Fatal(err)
	}
	return f, group
}

func TestFlagSet_Parse(t *testing.T) {
	t.Run("create fat", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"path/to/in2"},
			{"path/to/in3"},
			{"path/to/in4"},
			{"-output", "path/to/out"},
			{"-create"},
		}
		gotInput := []string{"path/to/in1", "path/to/in2", "path/to/in3", "path/to/in4"}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "create")
			equal(t, gotInput, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			if !create {
				t.Errorf("create is not true")
			}
		}
	})
	t.Run("replace", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-replace", "x86_64", "path/to/target1"},
			{"-replace", "arm64", "path/to/target2"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "replace")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			got1 := replace[0]
			got2 := replace[1]

			if len(replace) != 2 {
				t.Fatalf("len() is not equal. want: %v, got: %v", dataSet[2:], replace)
			}

			if replace[0][0] == "arm64" {
				got1, got2 = got2, got1
			}
			equal(t, []string{"x86_64", "path/to/target1"}, got1[:])
			equal(t, []string{"arm64", "path/to/target2"}, got2[:])
		}
	})
	t.Run("extract", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-extract", "arm64"},
			{"-extract", "arm64e"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "extract")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"arm64", "arm64e"}, extract)
		}
	})
	t.Run("extract_family", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-extract_family", "arm64e"},
			{"-extract", "x86_64"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "extract_family")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"arm64e"}, extractFamily)
			equal(t, []string{"x86_64"}, extract)
		}
	})
	t.Run("remove", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-remove", "x86_64"},
			{"-remove", "x86_64h"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "remove")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"x86_64h", "x86_64"}, remove)
		}
	})
	t.Run("thin", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-thin", "x86_64"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "thin")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"x86_64"}, []string{thin})
		}
	})
	t.Run("archs", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-archs"},
		}
		for _, in := range shuffle(dataSet) {
			f, g := fset(t, in)
			eq(t, g.Name, "archs")
			equal(t, []string{"path/to/in1"}, f.Args())
			if !archs {
				t.Errorf("archs is false")
			}
		}
	})
	t.Run("verify_arch", func(t *testing.T) {
		in := []string{"path/to/in1", "-verify_arch", "x86_64", "arm64", "arm64e", "x86_64h", "arm"}
		f, g := fset(t, in)
		eq(t, g.Name, "verify_arch")
		equal(t, []string{"path/to/in1"}, f.Args())
		equal(t, []string{"x86_64", "arm64", "arm64e", "x86_64h", "arm"}, verifyArch)
	})

	// TODO
	// t.Run("usage", func(t *testing.T) {
	// 	f := fset(t, []string{})
	// 	f.Usage()
	// })

	// not actual input
	t.Run("flex string stop after flag(-output)", func(t *testing.T) {
		args := []string{
			"-verify_arch", "x86_64", "arm64",
			"-output", "stop-archs", "-input1",
		}

		f := sflag.NewFlagSet("test")
		out := ""
		arches := []string{}
		f.String(&out, "output", "-output <file>")
		f.FlexStrings(&arches, "verify_arch", "verify_arch <arch>")
		if err := f.Parse(args); err != nil {
			t.Fatal(err)
		}
		equal(t, []string{"stop-archs"}, []string{out})
		equal(t, []string{"x86_64", "arm64"}, arches)
		equal(t, []string{"-input1"}, f.Args())
	})
}

func TestFlagSet_ParseError(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name: "-replace without args",
			args: []string{
				"path/to/in1",
				"-output", "output1",
				"-replace", "x86_64",
			},
			errMsg: "more values are required",
		},
		{
			name: "-replace without args",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64",
				"-output", "output2",
			},
			errMsg: "more values are required",
		},
		{
			name: "-replace without args",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64", "target1",
				"-output",
			},
			errMsg: "value is not specified",
		},
		{
			name: "dup flag",
			args: []string{
				"path/to/in1",
				"-output", "out1",
				"-output", "out2",
			},
			errMsg: "more than one -output option specified",
		},
		{
			name: "multiple flag group",
			args: []string{
				"path/to/in1",
				"-output", "out1",
				"-create",
				"-archs",
			},
			errMsg: "found no flag group",
		},
		{
			name: "no flag group",
			args: []string{
				"path/to/in1",
				"-output", "out1",
			},
			errMsg: "found no flag group",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, groups := register()
			err := f.Parse(tt.args)
			if err != nil {
				if err.Error() != tt.errMsg {
					t.Fatalf("want: %v, got: %v\n", tt.errMsg, err.Error())
				}
			}
			if err == nil {
				if _, err := sflag.LookupGroup(groups...); err != nil {
					if err.Error() != tt.errMsg {
						t.Fatalf("want: %v, got: %v\n", tt.errMsg, err.Error())
					}
				}
			}
		})
	}
}

func shuffle(dataSet [][]string) [][]string {
	patterns := [][]string{}
	permutation(dataSet, func(ds [][]string) {
		ptn := []string{}
		for _, iv := range ds {
			ptn = append(ptn, iv...)
		}
		patterns = append(patterns, ptn)
	})
	return patterns
}

func permutation(a [][]string, f func([][]string)) {
	perm(a, f, 0)
}

func perm(a [][]string, f func([][]string), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}

func equal(t *testing.T, want []string, got []string) {
	t.Helper()

	if len(want) != len(got) {
		t.Fatalf("want: %v, got: %v\n", want, got)
	}

	seen := map[string]struct{}{}
	for _, v := range want {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
		}
	}

	for _, v := range got {
		if _, ok := seen[v]; !ok {
			t.Errorf("got: %v\n", v)
		}
	}
}

func eq(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("want: %v got: %v\n", want, got)
	}
}
