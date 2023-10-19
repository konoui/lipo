package lipo_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Info(t *testing.T) {

	tests := []struct {
		name     string
		setupper func(t *testing.T) (inputs []string, want string)
	}{
		{
			name: "fat",
			setupper: func(t *testing.T) ([]string, string) {
				p := testlipo.Setup(t, bm, []string{"arm64", "arm64e", "x86_64"})
				ins := []string{p.Bin(t, "arm64"), p.FatBin, p.Bin(t, "arm64e")}
				return ins, p.Info(t, ins...)
			},
		},
		{
			name: "thin",
			setupper: func(t *testing.T) ([]string, string) {
				p := testlipo.Setup(t, bm, []string{"arm64"})
				ins := []string{p.FatBin}
				return ins, p.Info(t, ins...)
			},
		},
		{
			name: "archive",
			setupper: func(t *testing.T) ([]string, string) {
				l := testlipo.NewLipoBin(t)
				ins := []string{filepath.Join("./ar/testdata/arm64-func123.a")}
				return ins, l.Info(t, ins...)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs, want := tt.setupper(t)
			l := lipo.New(lipo.WithInputs(inputs...))

			info, err := l.Info()
			if err != nil {
				t.Fatal(err)
			}

			got := strings.Join(info, "\n")

			if want != got {
				t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

// TODO failed test
