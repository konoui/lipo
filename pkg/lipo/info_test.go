package lipo_test

import (
	"bytes"
	"io"
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

		stdout := &bytes.Buffer{}
		l.Info(stdout, stdout)
		got := stdout.String()

		want := p.Info(t, ins...)
		if want != got {
			t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
		}
	})

	t.Run("fat-file-of-single-arch", func(t *testing.T) {
		p := testlipo.Setup(t, bm, []string{"arm64"})
		ins := []string{p.FatBin}
		l := lipo.New(lipo.WithInputs(ins...))

		stdout := &bytes.Buffer{}
		l.Info(stdout, stdout)
		got := stdout.String()

		want := p.Info(t, ins...)
		if want != got {
			t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
		}
	})

	t.Run("archive-object", func(t *testing.T) {
		in := "../ar/testdata/arm64-func12.a"
		l := lipo.New(lipo.WithInputs(in))
		tl := testlipo.NewLipoBin(t)

		stdout := &bytes.Buffer{}
		l.Info(stdout, stdout)
		got := stdout.String()

		want := tl.Info(t, in)

		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("invalid-archive-object", func(t *testing.T) {
		in := "../ar/testdata/arm64-amd64-func12.a"
		l := lipo.New(lipo.WithInputs(in))
		stderr := &bytes.Buffer{}
		l.Info(io.Discard, stderr)

		got := stderr.String()

		tl := testlipo.NewLipoBin(t, testlipo.WithIgnoreErr(true))
		want := tl.Info(t, in)

		if !strings.Contains(want, got) {
			t.Errorf("\nwant: %s\ngot: %s", want, got)
		}
	})
}
