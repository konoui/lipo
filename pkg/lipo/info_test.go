package lipo_test

import (
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Info(t *testing.T) {
	t.Run("multiple-inputs", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"arm64", "arm64e", "x86_64"})
		ins := []string{p.Bin(t, "arm64"), p.FatBin, p.Bin(t, "arm64e")}
		l := lipo.New(lipo.WithInputs(ins...))
		info, err := l.Info()
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Join(info, "\n")

		got := p.Info(t, ins...)
		if want != got {
			t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
		}
	})

	t.Run("fat-file-of-single-arch", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"arm64"})
		ins := []string{p.FatBin}
		l := lipo.New(lipo.WithInputs(ins...))
		info, err := l.Info()
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Join(info, "\n")

		got := p.Info(t, ins...)
		if want != got {
			t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
		}
	})
}
