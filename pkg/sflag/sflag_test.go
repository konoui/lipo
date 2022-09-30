package sflag_test

import (
	"testing"

	"github.com/konoui/lipo/pkg/sflag"
)

var (
	output                      = ""
	create, archs               = false, false
	extract, remove, verifyArch = []string{}, []string{}, []string{}
	replace                     = [][]string{make([]string, 2)}
)

func register() *sflag.FlagSet {
	fs := sflag.NewFlagSet("lipo")
	// init
	output = ""
	create, archs = false, false
	replace = [][]string{make([]string, 2)}
	extract, remove, verifyArch = []string{}, []string{}, []string{}

	fs.String(&output, "output", "-output <file>")
	fs.Bool(&create, "create", "-create")
	fs.MultipleFlagFixedStrings(&replace, "replace", "-replace <arch> <file>")
	fs.MultipleFlagString(&extract, "extract", "-extract <arch>")
	fs.MultipleFlagString(&remove, "remove", "-remove <arch>")
	fs.Bool(&archs, "archs", "-archs <arch> ...")
	fs.FlexStrings(&verifyArch, "verify_arch", "verify_arch <arch>")
	return fs
}

func fset(t *testing.T, in []string) *sflag.FlagSet {
	fs := register()
	if err := fs.Parse(in); err != nil {
		t.Fatal(err)
	}
	return fs
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
			fs := fset(t, in)
			equal(t, gotInput, fs.Args())
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
			fs := fset(t, in)
			equal(t, []string{"path/to/in1"}, fs.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			got1 := replace[0]
			got2 := replace[1]

			if len(replace) != 2 {
				t.Fatalf("len() is not equal. want: %v, got: %v", dataSet[2:], replace)
			}

			if replace[0][0] == "arm64" {
				got1, got2 = got2, got1
			}
			equal(t, []string{"x86_64", "path/to/target1"}, got1)
			equal(t, []string{"arm64", "path/to/target2"}, got2)
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
			fs := fset(t, in)
			equal(t, []string{"path/to/in1"}, fs.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"arm64", "arm64e"}, extract)
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
			fs := fset(t, in)
			equal(t, []string{"path/to/in1"}, fs.Args())
			equal(t, []string{"path/to/out"}, []string{output})
			equal(t, []string{"x86_64h", "x86_64"}, remove)
		}
	})
	t.Run("archs", func(t *testing.T) {
		dataSet := [][]string{
			{"path/to/in1"},
			{"-archs"},
		}
		for _, in := range shuffle(dataSet) {
			fs := fset(t, in)
			equal(t, []string{"path/to/in1"}, fs.Args())
			if !archs {
				t.Errorf("archs is false")
			}
		}
	})
	t.Run("verify_arch", func(t *testing.T) {
		in := []string{"path/to/in1", "-verify_arch", "x86_64", "arm64", "arm64e", "x86_64h", "arm"}
		fs := fset(t, in)
		equal(t, []string{"path/to/in1"}, fs.Args())
		equal(t, []string{"x86_64", "arm64", "arm64e", "x86_64h", "arm"}, verifyArch)
	})

	// unexpected input tests
	t.Run("flex string stop after flag(-output)", func(t *testing.T) {
		args := []string{
			"-verify_arch", "x86_64", "arm64",
			"-output", "stop-archs",
		}
		_ = fset(t, args)
		equal(t, []string{"stop-archs"}, []string{output})
		equal(t, []string{"x86_64", "arm64"}, verifyArch)
	})
}

func TestFlagSet_ParseError(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		errMsg   string
		addCheck func(t *testing.T)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := register()
			err := fs.Parse(tt.args)
			if err != nil {
				if err.Error() != tt.errMsg {
					t.Errorf("want: %v, got: %v", err.Error(), tt.errMsg)
				}
			}
			if tt.addCheck != nil {
				tt.addCheck(t)
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
