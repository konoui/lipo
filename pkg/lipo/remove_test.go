package lipo_test

import (
	"debug/macho"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Remove(t *testing.T) {
	t.Run("remove", func(t *testing.T) {
		p := setup(t)

		fatBin := filepath.Join(p.dir, "out-amd64-arm64-binary")
		createFatBin(t, fatBin, p.amd64Bin, p.arm64Bin)

		got := filepath.Join(p.dir, "got-arm64")
		arch := "x86_64"
		l := lipo.New(lipo.WithInputs(fatBin), lipo.WithOutput(got))
		if err := l.Remove(arch); err != nil {
			t.Errorf("remove error %v\n", err)
		}

		if p.skip() {
			t.Skip("skip lipo binary test")
		}

		want := filepath.Join(p.dir, "want-arm64")
		p.removeFatBin(t, p.lipoFatBin, want, arch)
		diffSha256(t, want, got)
	})
}

func createFatBin(t *testing.T, out, input1, input2 string) {
	t.Helper()
	l := lipo.New(lipo.WithInputs(input1, input2), lipo.WithOutput(out))
	if err := l.Create(); err != nil {
		t.Fatalf("failed to create fat bin %v", err)
	}

	f, err := macho.OpenFat(out)
	if err != nil {
		t.Errorf("invalid fat file: %v\n", err)
	}
	defer f.Close()
}
