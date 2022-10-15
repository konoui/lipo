package lipo_test

import (
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
	"github.com/konoui/lipo/pkg/util"
)

func TestLipo_DetailedInfo(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		addThin []string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs...)
			fat1 := p.FatBin
			fat2 := p.FatBin
			args := append([]string{}, fat1, fat2)
			args = append(args, util.Map(tt.addThin, func(v string) string { return p.Bin(t, v) })...)
			l := lipo.New(lipo.WithInputs(args...))
			got, err := l.DetailedInfo()
			if err != nil {
				t.Fatal(err)
			}

			want := p.DetailedInfo(t, args...)
			if want != got {
				t.Errorf("want:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}
