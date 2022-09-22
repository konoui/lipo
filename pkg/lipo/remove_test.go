package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Remove(t *testing.T) {
	t.Run("remove", func(t *testing.T) {
		p := setup(t)

		got := filepath.Join(p.dir, "got-fat-arm64")
		arch := "x86_64"
		l := lipo.New(lipo.WithInputs(p.lipoFatBin), lipo.WithOutput(got))
		if err := l.Remove(arch); err != nil {
			t.Errorf("remove error %v\n", err)
		}

		if p.skip() {
			t.Skip("skip lipo binary tests")
		}

		want := filepath.Join(p.dir, "want-fat-arm64")
		p.remove(t, want, p.lipoFatBin, arch)
		diffSha256(t, want, got)

		// next test
		// FIXME table tests
		got = filepath.Join(p.dir, "got-fat-amd64")
		arch = "arm64"
		l = lipo.New(lipo.WithInputs(p.lipoFatBin), lipo.WithOutput(got))
		if err := l.Remove(arch); err != nil {
			t.Errorf("remove error %v\n", err)
		}

		if p.skip() {
			t.Skip("skip lipo binary tests")
		}

		want = filepath.Join(p.dir, "want-fat-amd64")
		p.remove(t, want, p.lipoFatBin, arch)
		diffSha256(t, want, got)
	})
}
