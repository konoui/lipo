package lipo_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
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
			inputs: []string{"x86_64", "arm64"},
		},
		{
			name:   "-thin arm64",
			arch:   "arm64",
			inputs: []string{"x86_64", "arm64"},
		},
		{
			name:   "-thin arm64",
			arch:   "arm64",
			inputs: []string{"x86_64", "arm64", "arm64e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs)

			got := filepath.Join(p.Dir, gotName(t))
			arch := tt.arch
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
			if err := l.Thin(arch); err != nil {
				t.Errorf("thin error %v\n", err)
			}

			if p.Skip() {
				t.Skip("skip lipo binary tests")
			}

			want := filepath.Join(p.Dir, wantName(t))
			p.Thin(t, want, p.FatBin, tt.arch)
			diffSha256(t, want, got)
		})
	}
}

func TestLipo_ThinWithOverwriteInput(t *testing.T) {
	t.Run("overwrite-input", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"x86_64", "arm64"})
		// input and output are same path
		got := p.FatBin
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Thin("x86_64")
		if err != nil {
			t.Fatal(err)
		}
		verifyArches(t, got, "x86_64")
	})
}

func TestLipo_ThinError(t *testing.T) {
	t.Run("not-match-arch", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"arm64", "x86_64"})

		got := filepath.Join(p.Dir, gotName(t))
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Thin("arm64e")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("fat input file (%s) does not contain the specified architecture (%s) to thin it to", p.FatBin, "arm64e")
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
