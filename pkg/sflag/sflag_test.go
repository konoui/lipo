package sflag_test

import (
	"testing"

	"github.com/konoui/lipo/pkg/sflag"
)

type flagRefs struct {
	create        sflag.FlagRef[bool]
	archs         sflag.FlagRef[bool]
	output        sflag.FlagRef[string]
	thin          sflag.FlagRef[string]
	remove        sflag.FlagRef[[]string]
	extract       sflag.FlagRef[[]string]
	extractFamily sflag.FlagRef[[]string]
	verifyArch    sflag.FlagRef[[]string]
	replace       sflag.FlagRef[[][2]string]
	segAligns     sflag.FlagRef[[][2]string]
}

func register() (*sflag.FlagSet, []*sflag.Group, *flagRefs) {
	f := sflag.NewFlagSet("lipo")
	refs := new(flagRefs)

	refs.output = f.String("output", "-output <file>")
	refs.segAligns = f.FixedStringFlags("segalign", "<arch_type> <alignment>")
	refs.create = f.Bool("create", "-create")
	refs.thin = f.String("thin", "thin <arch>")
	refs.replace = f.FixedStringFlags("replace", "-replace <arch> <file>")
	refs.extract = f.StringFlags("extract", "-extract <arch>")
	refs.extractFamily = f.StringFlags("extract_family", "-extract_family <arch>")
	refs.remove = f.StringFlags("remove", "-remove <arch>")
	refs.archs = f.Bool("archs", "-archs <arch> ...")
	refs.verifyArch = f.Strings("verify_arch", "verify_arch <arch>")

	createGroup := f.NewGroup("create").
		AddRequired(refs.create.Flag()).
		AddRequired(refs.output.Flag()).
		AddOptional(refs.segAligns.Flag())
	thinGroup := f.NewGroup("thin").
		AddRequired(refs.thin.Flag()).
		AddRequired(refs.output.Flag())
	extractGroup := f.NewGroup("extract").
		AddRequired(refs.extract.Flag()).
		AddRequired(refs.output.Flag()).
		AddOptional(refs.segAligns.Flag())
	extractFamilyGroup := f.NewGroup("extract_family").
		AddRequired(refs.extractFamily.Flag()).
		AddRequired(refs.output.Flag()).
		AddOptional(refs.extract.Flag()). // if specified, apple lipo regard values as family
		AddOptional(refs.segAligns.Flag())
	removeGroup := f.NewGroup("remove").
		AddRequired(refs.remove.Flag()).
		AddRequired(refs.output.Flag()).
		AddOptional(refs.segAligns.Flag())
	replaceGroup := f.NewGroup("replace").
		AddRequired(refs.replace.Flag()).
		AddRequired(refs.output.Flag()).
		AddOptional(refs.segAligns.Flag())
	archsGroup := f.NewGroup("archs").
		AddRequired(refs.archs.Flag())
	verifyArchGroup := f.NewGroup("verify_arch").
		AddRequired(refs.verifyArch.Flag())
	return f, []*sflag.Group{
		createGroup, thinGroup, extractGroup, extractFamilyGroup,
		removeGroup, replaceGroup, archsGroup,
		verifyArchGroup}, refs
}

func fset(t *testing.T, in []string) (*sflag.FlagSet, *sflag.Group, *flagRefs) {
	f, groups, refs := register()
	if err := f.Parse(in); err != nil {
		t.Fatal(err)
	}

	group, err := sflag.LookupGroup(groups...)
	if err != nil {
		t.Fatal(err)
	}
	return f, group, refs
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
			f, g, refs := fset(t, in)
			eq(t, g.Name, "create")
			equal(t, gotInput, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			if !refs.create.Value() {
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
			f, g, refs := fset(t, in)
			eq(t, g.Name, "replace")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			replaces := refs.replace.Value()
			got1 := replaces[0]
			got2 := replaces[1]

			if replaces := refs.replace.Value(); len(replaces) != 2 {
				t.Fatalf("len() is not equal. want: %v, got: %v", dataSet[2:], replaces)
			}

			if replaces[0][0] == "arm64" {
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
			f, g, refs := fset(t, in)
			eq(t, g.Name, "extract")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			equal(t, []string{"arm64", "arm64e"}, refs.extract.Value())
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
			f, g, refs := fset(t, in)
			eq(t, g.Name, "extract_family")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			equal(t, []string{"arm64e"}, refs.extractFamily.Value())
			equal(t, []string{"x86_64"}, refs.extract.Value())
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
			f, g, refs := fset(t, in)
			eq(t, g.Name, "remove")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			equal(t, []string{"x86_64h", "x86_64"}, refs.remove.Value())
		}
	})
	t.Run("thin", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-output", "path/to/out"},
			{"-thin", "x86_64"},
		}
		for _, in := range shuffle(dataSet) {
			f, g, refs := fset(t, in)
			eq(t, g.Name, "thin")
			equal(t, []string{"path/to/in1"}, f.Args())
			equal(t, []string{"path/to/out"}, []string{refs.output.Value()})
			equal(t, []string{"x86_64"}, []string{refs.thin.Value()})
		}
	})
	t.Run("archs", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-archs"},
		}
		for _, in := range shuffle(dataSet) {
			f, g, refs := fset(t, in)
			eq(t, g.Name, "archs")
			equal(t, []string{"path/to/in1"}, f.Args())
			if !refs.archs.Value() {
				t.Errorf("archs is false")
			}
		}
	})
	t.Run("verify_arch", func(t *testing.T) {
		in := []string{"path/to/in1", "-verify_arch", "x86_64", "arm64", "arm64e", "x86_64h", "arm"}
		f, g, refs := fset(t, in)
		eq(t, g.Name, "verify_arch")
		equal(t, []string{"path/to/in1"}, f.Args())
		equal(t, []string{"x86_64", "arm64", "arm64e", "x86_64h", "arm"}, refs.verifyArch.Value())
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
		out := f.String("output", "-output <file>")
		archs := f.Strings("verify_arch", "verify_arch <arch>")
		if err := f.Parse(args); err != nil {
			t.Fatal(err)
		}
		equal(t, []string{"stop-archs"}, []string{out.Value()})
		equal(t, []string{"x86_64", "arm64"}, archs.Value())
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
			errMsg: "-replace flag: more values are required",
		},
		{
			name: "-replace without args",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64",
				"-output", "output2",
			},
			errMsg: "-replace flag: more values are required",
		},
		{
			name: "-output without an arg",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64", "target1",
				"-output",
			},
			errMsg: "-output flag: value is not specified",
		},
		{
			name: "dup flag",
			args: []string{
				"path/to/in1",
				"-output", "out1",
				"-output", "out2",
			},
			errMsg: "duplication: more than one -output flag specified",
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
			f, groups, _ := register()
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
