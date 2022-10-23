package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_ExtractFamily(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		arches    []string
		segAligns []*lipo.SegAlignInput
		fat64     bool
	}{
		{
			name:   "-extract_family arm64e",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"arm64e"},
		},
		{
			name:   "-extract_family x86_64",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"x86_64"},
		},
		{
			name:   "-extract_family x86_64 -extract_family arm64e",
			inputs: []string{"x86_64", "arm64", "arm64v8"},
			arches: []string{"x86_64", "arm64e"},
		},
		{
			name:   "-extract_family x86_64 -segalign x86_64 2 -segalign arm64e 1",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"x86_64", "arm64e"},
			segAligns: []*lipo.SegAlignInput{
				{Arch: "x86_64", AlignHex: "2"},
				{Arch: "arm64e", AlignHex: "1"},
			},
		},
		{
			name:   "-extract_family -fat64",
			inputs: []string{"armv7k", "arm64", "arm64e"},
			arches: []string{"arm64"},
			fat64:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs,
				testSegAlignOpt(tt.segAligns),
				testlipo.WithFat64(tt.fat64))

			got := filepath.Join(p.Dir, gotName(t))
			arches := tt.arches
			opts := []lipo.Option{
				lipo.WithInputs(p.FatBin),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns),
			}
			if tt.fat64 {
				opts = append(opts, lipo.WithFat64())
			}
			l := lipo.New(opts...)
			if err := l.ExtractFamily(arches...); err != nil {
				t.Errorf("extract error %v\n", err)
			}

			if p.Skip() {
				t.Skip("skip lipo binary tests")
			}

			want := filepath.Join(p.Dir, wantName(t))
			p.ExtractFamily(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}
