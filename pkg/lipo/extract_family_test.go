package lipo_test

import (
	"fmt"
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
			inputs: cpuNames(),
			arches: []string{"arm"},
			fat64:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, bm, tt.inputs,
				testSegAlignOpt(tt.segAligns),
				testlipo.WithFat64(tt.fat64))

			got := filepath.Join(p.Dir, gotName(t))
			arches := tt.arches
			opts := []lipo.Option{
				lipo.WithInputs(p.FatBin),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns...),
			}
			if tt.fat64 {
				opts = append(opts, lipo.WithFat64())
			}
			l := lipo.New(opts...)
			if err := l.ExtractFamily(arches...); err != nil {
				t.Errorf("extract error %v\n", err)
			}

			want := filepath.Join(p.Dir, wantName(t))
			p.ExtractFamily(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}

func TestLipo_ExtracFamilyError(t *testing.T) {
	p := testlipo.Setup(t, bm, []string{"x86_64"})
	got := filepath.Join(p.Dir, gotName(t))
	l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))

	t.Run("not-match-arch", func(t *testing.T) {
		err := l.ExtractFamily("arm64e")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", "arm64e", p.FatBin)
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
