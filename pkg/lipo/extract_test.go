package lipo_test

import (
	"fmt"
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
		fat64     bool
	}{
		{
			name:   "-extract arm64 -extract arm64e",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"arm64", "arm64e"},
		},
		{
			name:   "-extract arm64 -extract arm64e -extract x86_64",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"arm64", "arm64e", "x86_64"},
		},
		{
			name:   "-extract x86_64 -segalign x86_64 2 -segalign arm64e 2",
			inputs: []string{"x86_64", "arm64", "arm64e"},
			arches: []string{"x86_64", "arm64e"},
			segAligns: []*lipo.SegAlignInput{
				{Arch: "x86_64", AlignHex: "2"},
				{Arch: "arm64e", AlignHex: "1"},
			},
		},
		{
			name:   "-extract -fat64",
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
				lipo.WithSegAlign(tt.segAligns...),
			}
			if tt.fat64 {
				opts = append(opts, lipo.WithFat64())
			}
			l := lipo.New(opts...)
			if err := l.Extract(arches...); err != nil {
				t.Errorf("extract error %v\n", err)
			}
			// tests for fat bin is expected
			verifyArches(t, got, arches...)

			want := filepath.Join(p.Dir, wantName(t))
			p.Extract(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}

func TestLipo_ExtractWithOverwriteInput(t *testing.T) {
	t.Run("overwrite-input", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"x86_64", "arm64"})
		// input and output are same path
		got := p.FatBin
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Extract("x86_64")
		if err != nil {
			t.Fatal(err)
		}
		verifyArches(t, got, "x86_64")
	})
}

func TestLipo_ExtractError(t *testing.T) {
	p := testlipo.Setup(t, []string{"arm64", "x86_64"})
	got := filepath.Join(p.Dir, gotName(t))
	l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))

	t.Run("not-match-arch1", func(t *testing.T) {
		err := l.Extract("arm64e")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", "arm64e", p.FatBin)
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
	t.Run("not-match-arch2", func(t *testing.T) {
		err := l.Extract("arm64e", "arm64")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", "arm64e", p.FatBin)
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
