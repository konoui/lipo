package lmacho_test

import (
	"debug/macho"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/konoui/lipo/pkg/ar"
	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/testlipo"
)

var bm = testlipo.NewBinManager(os.TempDir())

func TestNewReader(t *testing.T) {
	tests := []struct {
		name      string
		setupper  func(t *testing.T) (string, int)
		validator func(*testing.T, io.ReaderAt)
	}{
		{
			name: "fat",
			setupper: func(t *testing.T) (string, int) {
				p := testlipo.Setup(t, bm, []string{"arm64", "x86_64"})
				return p.FatBin, 2
			},
			validator: func(t *testing.T, ra io.ReaderAt) {
				_, err := macho.NewFile(ra)
				if err != nil {
					t.Errorf("macho.NewFile failed: %v", err)
				}
			},
		},
		{
			name: "hidden fat",
			setupper: func(t *testing.T) (string, int) {
				p := testlipo.Setup(t, bm,
					[]string{"armv7k", "arm64", "arm64e"},
					testlipo.WithHideArm64(true))
				return p.FatBin, 3
			},
			validator: func(t *testing.T, ra io.ReaderAt) {
				_, err := macho.NewFile(ra)
				if err != nil {
					t.Errorf("macho.NewFile failed: %v", err)
				}
			},
		},
		{
			name: "archive fat",
			setupper: func(t *testing.T) (string, int) {
				return "./../ar/testdata/fat-arm64-amd64-func1", 2
			},
			validator: func(t *testing.T, ra io.ReaderAt) {
				_, err := ar.NewReader(ra)
				if err != nil {
					t.Errorf("ar.NewReader failed: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, got := tt.setupper(t)
			//			t.Log("fat file", p)
			f, err := os.Open(p)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			reader, err := lmacho.NewReader(f)
			if err != nil {
				t.Errorf("NewReader() error = %v", err)
				return
			}

			narch := 0
			for {
				obj, err := reader.Next()
				if errors.Is(err, io.EOF) {
					break
				}

				if err != nil {
					t.Fatal(err)
				}

				tt.validator(t, obj)

				narch++
			}

			if narch != got {
				t.Errorf("want: %d, got: %d", narch, got)
			}
		})
	}
}
