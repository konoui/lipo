package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Extract(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		arches []string
	}{
		{
			name:   "-extract x86_64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"x86_64"},
		},
		{
			name:   "-extract arm64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64"},
		},
		{
			name:   "-extract arm64 -extract arm64e",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64", "arm64e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs...)

			got := filepath.Join(p.Dir, randName())
			arches := tt.arches
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
			if err := l.Extract(arches...); err != nil {
				t.Errorf("extract error %v\n", err)
			}
			// tests for fat bin is expected
			verifyArches(t, got, arches...)

			if p.Skip() {
				t.Skip("skip lipo binary tests")
			}

			want := filepath.Join(p.Dir, randName())
			p.Extract(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}
