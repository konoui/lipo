package lmacho_test

import (
	"debug/macho"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestNewWriter(t *testing.T) {

	tests := []struct {
		name      string
		setupper  func(t *testing.T) []string
		validator func(t *testing.T, in string)
	}{
		{
			name: "fat",
			setupper: func(t *testing.T) []string {
				p := testlipo.Setup(t, bm, []string{"arm64", "x86_64"})
				return []string{p.Bin(t, "x86_64"), p.Bin(t, "arm64")}
			},
			validator: func(t *testing.T, in string) {
				_, err := macho.OpenFat(in)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bins := tt.setupper(t)

			outPath := filepath.Join(os.TempDir(), "fat")
			out, err := os.Create(outPath)
			// t.Log("out path", outPath)
			// t.Log("num of arches", len(bins))
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()

			arches := []lmacho.Object{}
			for _, bin := range bins {
				f, err := os.Open(bin)
				if err != nil {
					t.Fatal(err)
				}

				info, err := os.Stat(bin)
				if err != nil {
					t.Fatal(err)
				}

				sr := io.NewSectionReader(f, 0, info.Size())
				a, err := lmacho.NewArch(sr)
				if err != nil {
					t.Fatal(err)
				}
				arches = append(arches, a)
			}

			if err := lmacho.Create(out, arches, false, false); err != nil {
				t.Fatal(err)
			}

			tt.validator(t, outPath)
		})
	}
}
