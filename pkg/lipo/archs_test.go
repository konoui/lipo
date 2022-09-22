package lipo_test

import (
	"bytes"
	"os"
	"path/filepath"
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
			got := archs(t, in)
			if got != want {
				t.Errorf("want %s, got %s\n", want, got)
			}
		}
	})
}

func archs(t *testing.T, in string) string {
	stdout := &bytes.Buffer{}
	l := lipo.New(
		lipo.WithInputs(in),
		lipo.WithStdout(stdout),
	)
	err := l.Archs()
	if err != nil {
		t.Fatal(err)
	}
	return stdout.String()
}
