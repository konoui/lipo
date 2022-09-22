package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Extract(t *testing.T) {
	t.Run("extract", func(t *testing.T) {
		p := setup(t)

		got := filepath.Join(p.dir, "got-amd64")
		arch := "x86_64"
		l := lipo.New(lipo.WithInputs(p.lipoFatBin), lipo.WithOutput(got))
		if err := l.Extract(arch); err != nil {
			t.Errorf("extract error %v\n", err)
		}

		if p.skip() {
			t.Skip("skip lipo binary tests")
		}

		want := filepath.Join(p.dir, "want-amd64")
		p.extract(t, want, p.lipoFatBin, arch)
		diffSha256(t, want, got)

		// next test
		// FIXME table tests
		got = filepath.Join(p.dir, "got-arm64")
		arch = "arm64"
		l = lipo.New(lipo.WithInputs(p.lipoFatBin), lipo.WithOutput(got))
		if err := l.Extract(arch); err != nil {
			t.Errorf("remove error %v\n", err)
		}

		if p.skip() {
			t.Skip("skip lipo binary tests")
		}

		want = filepath.Join(p.dir, "want-arm64")
		p.extract(t, want, p.lipoFatBin, arch)
		diffSha256(t, want, got)
	})
}
