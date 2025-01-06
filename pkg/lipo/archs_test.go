package lipo_test

import (
	"fmt"
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
	p := testlipo.Setup(t, bm, arches)

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
	tg := p.Bin(t, "x86_64")
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
	t.Run("non binary", func(t *testing.T) {
		f, err := os.Create(filepath.Join(bm.Dir, "empty-file"))
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
		tl := testlipo.NewLipoBin(t, testlipo.WithIgnoreErr(true))
		want := fmt.Sprintf("can't figure out the architecture type of: %s", f.Name())
		got1 := tl.Archs(t, input)
		got2 := err.Error()
		if !strings.Contains(got1, want) {
			t.Errorf("want: %s\ngot1: %s", want, got1)
		}
		if !strings.Contains(got2, want) {
			t.Errorf("want: %s, got2: %s", want, got2)
		}
	})
}

func verifyArches(t *testing.T, bin string, arches ...string) {
	t.Helper()

	// trim object prefix
	arches = util.Map(arches, func(v string) string {
		return strings.TrimPrefix(v, "obj_")
	})

	want := arches

	got, err := lipo.New(lipo.WithInputs(bin)).Archs()
	if err != nil {
		t.Errorf("verifyArches: archs error: %v", err)
	}

	if len(got) != len(want) {
		t.Errorf("verifyArches: want %v, got %v\n", want, got)
	}

	m := util.ExistenceMap(want, func(a string) string { return a })
	for _, x := range got {
		if _, found := m[x]; !found {
			t.Errorf("verifyArches: encountered an unexpected arch: %s in fat bin", x)
		}
	}
}
