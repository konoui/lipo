package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Replace(t *testing.T) {
	tests := []struct {
		name   string
		arch   string
		arches []string
	}{
		{
			name:   "-replace x86_64",
			arch:   "x86_64",
			arches: []string{inArm64, inAmd64},
		},
		{
			name:   "-replace arm64e",
			arch:   "arm64e",
			arches: []string{inArm64, inAmd64, "arm64e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches...)
			arch := tt.arch
			to := p.Bin(t, arch)

			got := filepath.Join(p.Dir, randName())
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
			ri := []lipo.ReplaceInput{{Arch: arch, Bin: to}}
			if err := l.Replace(ri); err != nil {
				t.Fatalf("replace error: %v\n", err)
			}

			verifyArches(t, got, tt.arches...)

			want := filepath.Join(p.Dir, randName())
			p.Replace(t, want, p.FatBin, arch, to)
			diffSha256(t, want, got)
		})
	}
}
