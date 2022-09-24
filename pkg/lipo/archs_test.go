package lipo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func TestLipo_Archs(t *testing.T) {
	arches := mcpu.CpuNames()
	p := setup(t, arches...)

	if p.skip() {
		t.Skip("skip lipo binary tests")
	}

	l := lipo.New(lipo.WithInputs(p.fatBin))
	gotArches, err := l.Archs()
	if err != nil {
		t.Fatal(err)
	}

	got := strings.Join(gotArches, " ") + "\n"
	want := p.archs(t, p.fatBin)
	if want != got {
		t.Errorf("want %v\ngot %v\n", want, got)
	}
}

func TestLipo_ArchsForLocationFiles(t *testing.T) {
	t.Run("archs", func(t *testing.T) {
		lipoBin := newLipoBin(t)
		dir := "/bin/"
		ents, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}

		if len(ents) == 0 {
			t.Skip("found no files")
		}

		if lipoBin.skip() {
			t.Skip("skip lipo binary tests")
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
			got := strings.Join(gotArches, " ") + "\n"

			want := lipoBin.archs(t, bin)
			if want != got {
				t.Errorf("want %v\ngot %v\n", want, got)
			}
		}
	})
}

func verifyArches(t *testing.T, bin string, arches ...string) {
	want := arches
	got, err := lipo.New(lipo.WithInputs(bin)).Archs()
	if err != nil {
		t.Fatalf("archs error: %v", err)
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
		t.Fatalf("want %v, got %v\n", want, got)
	}
}
