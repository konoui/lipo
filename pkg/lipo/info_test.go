package lipo_test

import (
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Info(t *testing.T) {
	p := testlipo.Setup(t, "arm64", "arm64e", "x86_64")
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
}
