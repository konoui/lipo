package lipo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Archs(t *testing.T) {
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

		for _, ent := range ents {
			if ent.IsDir() {
				continue
			}

			in := filepath.Join(dir, ent.Name())
			want := lipoBin.archs(t, in)
			arches, err := lipo.New(lipo.WithInputs(in)).Archs()
			if err != nil {
				t.Fatalf("archs error: %v", err)
			}
			got := strings.Join(arches, " ") + "\n"
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
