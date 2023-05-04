package sflag_test

import (
	"testing"

	"github.com/konoui/lipo/pkg/sflag"
)

type flagRefs struct {
	create        *sflag.FlagRef[bool]
	archs         *sflag.FlagRef[bool]
	output        *sflag.FlagRef[string]
	thin          *sflag.FlagRef[string]
	remove        *sflag.FlagRef[[]string]
	extract       *sflag.FlagRef[[]string]
	extractFamily *sflag.FlagRef[[]string]
	verifyArch    *sflag.FlagRef[[]string]
	replace       *sflag.FlagRef[[][2]string]
	segAligns     *sflag.FlagRef[[][2]string]
}

func register() (*sflag.FlagSet, []*sflag.Group, *flagRefs) {
	f := sflag.NewFlagSet("lipo")
	refs := new(flagRefs)

	refs.output = f.String("output", "-output <file>", sflag.WithShortName("o"))
	refs.segAligns = f.FixedStringFlags("segalign", "<arch_type> <alignment>", sflag.WithShortName("s"))
	refs.create = f.Bool("create", "-create", sflag.WithShortName("c"))
	refs.thin = f.String("thin", "thin <arch>", sflag.WithShortName("t"))
	refs.replace = f.FixedStringFlags("replace", "-replace <arch> <file>", sflag.WithShortName("rep"))
	refs.extract = f.StringFlags("extract", "-extract <arch>", sflag.WithShortName("e"))
	refs.extractFamily = f.StringFlags("extract_family", "-extract_family <arch>")
	refs.remove = f.StringFlags("remove", "-remove <arch>", sflag.WithShortName("rem"))
	refs.archs = f.Bool("archs", "-archs <arch> ...")
	refs.verifyArch = f.Strings("verify_arch", "verify_arch <arch>")

	createGroup := f.NewGroup("create").
		AddRequired(refs.create).
		AddRequired(refs.output).
		AddOptional(refs.segAligns)
	thinGroup := f.NewGroup("thin").
		AddRequired(refs.thin).
		AddRequired(refs.output)
	extractGroup := f.NewGroup("extract").
		AddRequired(refs.extract).
		AddRequired(refs.output).
		AddOptional(refs.segAligns)
	extractFamilyGroup := f.NewGroup("extract_family").
		AddRequired(refs.extractFamily).
		AddRequired(refs.output).
		AddOptional(refs.extract). // if specified, apple lipo regard values as family
		AddOptional(refs.segAligns)
	removeGroup := f.NewGroup("remove").
		AddRequired(refs.remove).
		AddRequired(refs.output).
		AddOptional(refs.segAligns)
	replaceGroup := f.NewGroup("replace").
		AddRequired(refs.replace).
		AddRequired(refs.output).
		AddOptional(refs.segAligns)
	archsGroup := f.NewGroup("archs").
		AddRequired(refs.archs)
	verifyArchGroup := f.NewGroup("verify_arch").
		AddRequired(refs.verifyArch)
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
			eq(t, "path/to/out", refs.output.Get())
			if !refs.create.Get() {
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
			equal(t, []string{"path/to/out"}, []string{refs.output.Get()})
			replaces := refs.replace.Get()
			got1 := replaces[0]
			got2 := replaces[1]

			if len(replaces) != 2 {
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
			eq(t, "path/to/out", refs.output.Get())
			equal(t, []string{"arm64", "arm64e"}, refs.extract.Get())
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
			eq(t, "path/to/out", refs.output.Get())
			equal(t, []string{"arm64e"}, refs.extractFamily.Get())
			equal(t, []string{"x86_64"}, refs.extract.Get())
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
			eq(t, "path/to/out", refs.output.Get())
			equal(t, []string{"x86_64h", "x86_64"}, refs.remove.Get())
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
			eq(t, "path/to/out", refs.output.Get())
			eq(t, "x86_64", refs.thin.Get())
		}
	})
	t.Run("thin-with-short-flag", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-o", "path/to/out"},
			{"-t", "x86_64"},
		}
		for _, in := range shuffle(dataSet) {
			f, g, refs := fset(t, in)
			eq(t, g.Name, "thin")
			equal(t, []string{"path/to/in1"}, f.Args())
			eq(t, "path/to/out", refs.output.Get())
			eq(t, "x86_64", refs.thin.Get())
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
			if !refs.archs.Get() {
				t.Errorf("archs is false")
			}
		}
	})
	t.Run("verify_arch", func(t *testing.T) {
		in := []string{"path/to/in1", "-verify_arch", "x86_64", "arm64", "arm64e", "x86_64h", "arm"}
		f, g, refs := fset(t, in)
		eq(t, g.Name, "verify_arch")
		equal(t, []string{"path/to/in1"}, f.Args())
		equal(t, []string{"x86_64", "arm64", "arm64e", "x86_64h", "arm"}, refs.verifyArch.Get())
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
		eq(t, "stop-archs", out.Get())
		equal(t, []string{"x86_64", "arm64"}, archs.Get())
		equal(t, []string{"-input1"}, f.Args())
	})

	t.Run("must call parse before group lookup", func(t *testing.T) {
		f := sflag.NewFlagSet("test")
		g := f.NewGroup("test")
		_, err := sflag.LookupGroup(g)
		if err == nil {
			t.Error("error should occur")
			return
		}
		want := "must call FlagSet.Parse() before LookupGroup()"
		got := err.Error()
		if want != got {
			t.Errorf("want: %s, got: %s", want, got)
		}
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
			errMsg: "the -replace flag requires 2 values at least",
		},
		{
			name: "-replace without args",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64",
				"-output", "output2",
			},
			errMsg: "the -replace flag requires 2 values at least",
		},
		{
			name: "-output without an arg",
			args: []string{
				"path/to/in1",
				"-replace", "x86_64", "target1",
				"-output",
			},
			errMsg: "the -output flag requires one value",
		},
		{
			name: "-thin without arg. -output should not regard as a value for -thin",
			args: []string{
				"path/to/in1",
				"-thin",
				"-output", "out",
			},
			errMsg: "the -thin flag requires one value",
		},
		{
			name: "-verify_arch without arg",
			args: []string{
				"path/to/in1",
				"-verify_arch",
			},
			errMsg: "the -verify_arch flag requires one value at least",
		},
		{
			name: "-verify_arch without arg",
			args: []string{
				"path/to/in1",
				"-verify_arch",
			},
			errMsg: "the -verify_arch flag requires one value at least",
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
			name: "no flag group",
			args: []string{
				"path/to/in1",
				"-create",
			},
			errMsg: `found no flag group
a required flag output in the group create is not specified
a required flag thin in the group thin is not specified
a required flag extract in the group extract is not specified
a required flag extract_family in the group extract_family is not specified
a required flag remove in the group remove is not specified
a required flag replace in the group replace is not specified
a required flag archs in the group archs is not specified
a required flag verify_arch in the group verify_arch is not specified`,
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
						t.Fatalf("want:\n%v, got:\n%v\n", tt.errMsg, err.Error())
					}
				} else {
					t.Error("error should occur")
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
