package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Extract(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		arches    []string
		segAligns []*lipo.SegAlignInput
	}{
		{
			name:   "-extract arm64 -extract arm64e",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64", "arm64e"},
		},
		{
			name:   "-extract arm64 -extract arm64e -extract x86_64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64", "arm64e", "x86_64"},
		},
		{
			name:      "-extract x86_64 -segalign x86_64 2 -segalign arm64e 2",
			inputs:    []string{inAmd64, inArm64, "arm64e"},
			arches:    []string{"x86_64", "arm64e"},
			segAligns: []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "2"}, {Arch: "arm64e", AlignHex: "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs...)

			got := filepath.Join(p.Dir, randName())
			arches := tt.arches
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got), lipo.WithSegAlign(tt.segAligns))
			if err := l.Extract(arches...); err != nil {
				t.Errorf("extract error %v\n", err)
			}
			// tests for fat bin is expected
			verifyArches(t, got, arches...)

			if p.Skip() {
				t.Skip("skip lipo binary tests")
			}

			// set segalign for next Extract
			for _, segAlign := range tt.segAligns {
				p.AddSegAlign(segAlign.Arch, segAlign.AlignHex)
			}

			want := filepath.Join(p.Dir, randName())
			p.Extract(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}
