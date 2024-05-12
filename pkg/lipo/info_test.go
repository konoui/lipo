package lipo_test

import (
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Info(t *testing.T) {
	t.Run("multiple-inputs", func(t *testing.T) {
		p := testlipo.Setup(t, bm, []string{"arm64", "arm64e", "x86_64"})
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
		p := testlipo.Setup(t, bm, []string{"arm64"})
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

	t.Run("archive-object", func(t *testing.T) {
		in := "../ar/testdata/arm64-func12.a"
		l := lipo.New(lipo.WithInputs(in))
		info, err := l.Info()
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Join(info, "\n")

		tl := testlipo.NewLipoBin(t)
		got := tl.Info(t, in)

		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("invalid-archive-object", func(t *testing.T) {
		in := "../ar/testdata/arm64-amd64-func12.a"
		l := lipo.New(lipo.WithInputs(in))
		_, err := l.Info()
		if err == nil {
			t.Fatal("error is nil")
		}

		want := err.Error()

		tl := testlipo.NewLipoBin(t, testlipo.IgnoreErr(true))
		got := tl.Info(t, in)

		if !strings.HasSuffix(got, want) {
			t.Errorf("\nwant: %s\ngot: %s", want, got)
		}
	})
}
