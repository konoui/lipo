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
	// fat binary test
	arches := cpuNames()
	p := testlipo.Setup(t, arches...)

	if p.Skip() {
		t.Skip("skip lipo binary tests")
	}

	l := lipo.New(lipo.WithInputs(p.FatBin))
	gotArches, err := l.Archs()
	if err != nil {
		t.Fatal(err)
	}

	got := strings.Join(gotArches, " ")
	want := p.Archs(t, p.FatBin)
	if want != got {
		t.Errorf("fat bin want %v\ngot %v\n", want, got)
	}

	// single binary test
	tg := p.Bin(t, inAmd64)
	l = lipo.New(lipo.WithInputs(tg))
	gotArches, err = l.Archs()
	if err != nil {
		t.Fatal(err)
	}
	got = strings.Join(gotArches, " ")
	want = p.Archs(t, tg)
	if want != got {
		t.Errorf("thin bin want %v\ngot %v\n", want, got)
	}
}

func TestLipo_ArchsToLocalFiles(t *testing.T) {
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

		if lipoBin.Skip() {
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
