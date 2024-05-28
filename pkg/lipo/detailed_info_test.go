package lipo_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
	"github.com/konoui/lipo/pkg/util"
)

func TestLipo_DetailedInfo(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		addThin   []string
		hideArm64 bool
	}{
		{
			name:   "two",
			inputs: []string{"arm64", "x86_64"},
		},
		{
			name:    "fat and thin",
			inputs:  []string{"arm64", "x86_64"},
			addThin: []string{"arm64"},
		},
		{
			name:   "all arches",
			inputs: cpuNames(),
		},
		{
			name:      "hideARM64",
			inputs:    []string{"arm64", "armv7k"},
			hideArm64: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, bm, tt.inputs, testlipo.WithHideArm64(tt.hideArm64))
			fat1 := p.FatBin
			fat2 := p.FatBin
			args := append([]string{}, fat1, fat2)
			args = append(args, util.Map(tt.addThin, func(v string) string { return p.Bin(t, v) })...)
			l := lipo.New(lipo.WithInputs(args...))

			got := &bytes.Buffer{}
			l.DetailedInfo(got, got)

			want := p.DetailedInfo(t, args...)
			if want != got.String() {
				t.Errorf("want:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func TestLipo_DetailedInfoWithError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		stderr := &bytes.Buffer{}
		lipo.New(lipo.WithInputs("not-found")).DetailedInfo(io.Discard, stderr)

		got := stderr.String()

		want := "open not-found: no such file or directory"
		if !strings.Contains(got, want) {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
	t.Run("not binary", func(t *testing.T) {
		f, err := os.Create(filepath.Join(bm.Dir, "empty-file"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		input := f.Name()

		stderr := &bytes.Buffer{}
		lipo.New(lipo.WithInputs(input)).DetailedInfo(io.Discard, stderr)

		tl := testlipo.NewLipoBin(t, testlipo.WithIgnoreErr(true))
		want := fmt.Sprintf("can't figure out the architecture type of: %s", f.Name())
		got1 := tl.DetailedInfo(t, input)
		got2 := stderr.String()
		if !strings.Contains(got1, want) {
			t.Errorf("want: %s, got1: %s", want, got1)
		}
		if !strings.Contains(got2, want) {
			t.Errorf("want: %s, got2: %s", want, got2)
		}
	})
}
