package lipo_test

import (
	"debug/macho"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Create(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
	}{
		{
			name:   "-create",
			inputs: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			inputs: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			inputs: []string{inAmd64, inArm64, "arm64e"},
		},
		{
			name:   "-create",
			inputs: []string{inAmd64, inArm64, "x86_64h", "arm64e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := setup(t, tt.inputs...)

			gotBin := filepath.Join(p.dir, randName())
			createFatBin(t, gotBin, p.bins()...)

			if p.skip() {
				t.Skip("skip lipo binary test")
			}

			p.detailedInfo(t, gotBin)
			p.detailedInfo(t, p.fatBin)
			diffSha256(t, p.fatBin, gotBin)
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
