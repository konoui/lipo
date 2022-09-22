package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

func TestLipo_Thin(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		arch   string
	}{
		{
			name:   "-thin x86_64",
			arch:   "x86_64",
			inputs: []string{inAmd64, inArm64},
		},
		{
			name:   "-thin arm64",
			arch:   "arm64",
			inputs: []string{inAmd64, inArm64},
		},
		{
			name:   "-thin arm64",
			arch:   "arm64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := setup(t, tt.inputs...)

			got := filepath.Join(p.dir, randName())
			arch := tt.arch
			l := lipo.New(lipo.WithInputs(p.fatBin), lipo.WithOutput(got))
			if err := l.Thin(arch); err != nil {
				t.Errorf("thin error %v\n", err)
			}

			if p.skip() {
				t.Skip("skip lipo binary tests")
			}

			want := filepath.Join(p.dir, randName())
			p.thin(t, want, p.fatBin, tt.arch)
			diffSha256(t, want, got)
		})
	}
}
