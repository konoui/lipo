package lipo_test

import (
	"debug/macho"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Create(t *testing.T) {
	tests := []struct {
		name   string
		arches []string
	}{
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64, "arm64e"},
		},
		{
			name:   "-create",
			arches: mcpu.CpuNames(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches...)

			got := filepath.Join(p.Dir, randName())
			createFatBin(t, got, p.Bins()...)
			// tests for fat bin is expected
			verifyArches(t, got, tt.arches...)

			if p.Skip() {
				t.Skip("skip lipo binary test")
			}

			diffSha256(t, p.FatBin, got)
		})
	}
}

func createFatBin(t *testing.T, out string, inputs ...string) {
	t.Helper()

	l := lipo.New(lipo.WithInputs(inputs...), lipo.WithOutput(out))
	if err := l.Create(); err != nil {
		t.Fatalf("failed to create fat bin %v", err)
	}

	f, err := macho.OpenFat(out)
	if err != nil {
		t.Fatalf("invalid fat file: %v\n", err)
	}
	defer f.Close()
}
