package lipo_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Remove(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		arches    []string
		segAligns []*lipo.SegAlignInput
		hideArm64 bool
		fat64     bool
	}{
		{
			name:   "-remove x86_64",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"x86_64"},
		},
		{
			name:   "-remove arm64",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"arm64"},
		},
		{
			name:   "-remove arm64 -remove arm64e",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"arm64", "arm64e"},
		},
		{
			name:      "-remove x86_64 -segalign arm64 2",
			inputs:    []string{"x86_64", "arm64"},
			arches:    []string{"x86_64"},
			segAligns: []*lipo.SegAlignInput{{Arch: "arm64", AlignHex: "2"}},
		},
		{
			name:      "-remove x86_64 -hideARM64",
			inputs:    []string{"x86_64", "armv7k", "arm64"},
			arches:    []string{"x86_64"},
			hideArm64: true,
		},
		{
			name:   "-remove -fat64",
			inputs: []string{"armv7k", "arm64", "arm64e"},
			arches: []string{"arm64"},
			fat64:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, bm, tt.inputs,
				testSegAlignOpt(tt.segAligns),
				testlipo.WithHideArm64(tt.hideArm64),
				testlipo.WithFat64(tt.fat64),
			)

			got := filepath.Join(p.Dir, gotName(t))
			arches := tt.arches
			opts := []lipo.Option{
				lipo.WithInputs(p.FatBin),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns...),
			}
			if tt.hideArm64 {
				opts = append(opts, lipo.WithHideArm64())
			}
			if tt.fat64 {
				opts = append(opts, lipo.WithFat64())
			}
			l := lipo.New(opts...)
			if err := l.Remove(arches...); err != nil {
				t.Errorf("remove error %v\n", err)
			}

			wantArches := []string{}
			for _, i := range tt.inputs {
				if !contain(i, tt.arches) {
					wantArches = append(wantArches, i)
				}
			}

			verifyArches(t, got, wantArches...)

			want := filepath.Join(p.Dir, wantName(t))
			p.Remove(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}

func TestLipo_RemoveError(t *testing.T) {
	t.Run("not-match-arch", func(t *testing.T) {
		p := testlipo.Setup(t, bm, []string{"arm64", "x86_64"})

		got := filepath.Join(p.Dir, wantName(t))
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Remove("arm64e")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", "arm64e", p.FatBin)
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})

	t.Run("result-in-empty-inputs", func(t *testing.T) {
		p := testlipo.Setup(t, bm, []string{"arm64"})
		got := filepath.Join(p.Dir, wantName(t))
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Remove("arm64")
		if err == nil {
			t.Errorf("error does not occur")
		}
		want := "no inputs would result in an empty fat file"
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
