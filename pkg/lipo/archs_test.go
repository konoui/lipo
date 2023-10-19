package lipo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
	"github.com/konoui/lipo/pkg/util"
)

func TestLipo_Archs(t *testing.T) {
	tests := []struct {
		name     string
		setupper func(t *testing.T) (p string, want string)
	}{
		{
			name: "fat",
			setupper: func(t *testing.T) (string, string) {
				arches := cpuNames()
				p := testlipo.Setup(t, bm, arches)
				return p.FatBin, p.Archs(t, p.FatBin)
			},
		},
		{
			name: "thin",
			setupper: func(t *testing.T) (string, string) {
				arches := cpuNames()
				p := testlipo.Setup(t, bm, arches)
				tg := p.Bin(t, "x86_64")
				return tg, p.Archs(t, tg)
			},
		},
		{
			name: "archive",
			setupper: func(t *testing.T) (string, string) {
				l := testlipo.NewLipoBin(t)
				p := filepath.Join("./ar/testdata/arm64-func123.a")
				return p, l.Archs(t, p)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, want := tt.setupper(t)
			l := lipo.New(lipo.WithInputs(p))
			gotArches, err := l.Archs()
			if err != nil {
				t.Fatal(err)
			}

			got := strings.Join(gotArches, " ")
			if want != got {
				t.Errorf("bin want %v\ngot %v\n", want, got)
			}
		})
	}
}

func TestLipo_ArchsWithError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := lipo.New(lipo.WithInputs("not-found")).Archs()
		if err == nil {
			t.Error("should occur error")
			return
		}
		want := "open not-found: no such file or directory"
		got := err.Error()
		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
	t.Run("not binary", func(t *testing.T) {
		f, err := os.Create("not-binary")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		input := f.Name()
		_, err = lipo.New(lipo.WithInputs(input)).Archs()
		if err == nil {
			t.Error("should occur error")
			return
		}
		tl := testlipo.NewLipoBin(t, testlipo.IgnoreErr(true))
		want := "can't figure out the architecture type of: not-binary"
		got1 := tl.Archs(t, input)
		got2 := err.Error()
		if !strings.Contains(got1, want) {
			t.Errorf("want: %s, got1: %s", want, got1)
		}
		if !strings.Contains(got2, want) {
			t.Errorf("want: %s, got2: %s", want, got2)
		}
	})

	// TODO invalid archive
}

func TestLipo_ArchsForLocalFiles(t *testing.T) {
	t.Run("archs", func(t *testing.T) {
		lipoBin := testlipo.NewLipoBin(t)
		dir := "/bin/"
		ents, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}

		if len(ents) == 0 {
			t.Skip("found no files")
		}

		for _, ent := range ents {
			if ent.IsDir() {
				continue
			}

			bin := filepath.Join(dir, ent.Name())
			gotArches, err := lipo.New(lipo.WithInputs(bin)).Archs()
			if err != nil {
				t.Fatalf("archs error: %v", err)
			}
			got := strings.Join(gotArches, " ")

			want := lipoBin.Archs(t, bin)
			if want != got {
				t.Errorf("want %v\ngot %v\n", want, got)
			}
		}
	})
}

func verifyArches(t *testing.T, bin string, arches ...string) {
	t.Helper()

	// trim object
	arches = util.Map(arches, func(v string) string {
		return strings.TrimPrefix(v, "obj_")
	})
	want := arches
	got, err := lipo.New(lipo.WithInputs(bin)).Archs()
	if err != nil {
		t.Fatalf("verifyArches: archs error: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("want %v, got %v\n", want, got)
	}

	mb := make(map[string]struct{}, len(want))
	for _, x := range want {
		mb[x] = struct{}{}
	}

	var diff []string
	for _, x := range got {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	if len(diff) != 0 {
		t.Fatalf("verifyArches: want %v, got %v\n", want, got)
	}
}
