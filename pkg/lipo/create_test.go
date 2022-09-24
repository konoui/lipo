package lipo_test

import (
	"debug/macho"
	"path/filepath"
	"sort"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func TestLipo_Create(t *testing.T) {
	tests := []struct {
		name   string
		arches []string
	}{
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64},
		},
		{
			name:   "-create",
			arches: []string{inAmd64, inArm64, "arm64e"},
		},
		{
			name:   "-create",
			arches: testSort(mcpu.CpuNames()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := setup(t, tt.arches...)

			got := filepath.Join(p.dir, randName())
			createFatBin(t, got, p.bins()...)
			// tests for fat bin is expected
			verifyArches(t, got, tt.arches...)

			if p.skip() {
				t.Skip("skip lipo binary test")
			}

			p.detailedInfo(t, got)
			p.detailedInfo(t, p.fatBin)
			diffSha256(t, p.fatBin, got)
		})
	}
}

func createFatBin(t *testing.T, out string, inputs ...string) {
	t.Helper()
	l := lipo.New(lipo.WithInputs(inputs...), lipo.WithOutput(out))
	if err := l.Create(); err != nil {
		t.Fatalf("failed to create fat bin %v", err)
	}

	f, err := macho.OpenFat(out)
	if err != nil {
		t.Fatalf("invalid fat file: %v\n", err)
	}
	defer f.Close()
}

// Note temporary fix
func testSort(cpus []string) []string {
	values := []string{}
	for _, c := range cpus {
		if c == "x86_64" || c == "i386" || c == "x86_64h" {
			continue
		}
		values = append(values, c)
	}

	sort.Slice(values, func(i, j int) bool {
		icpu, isub, _ := mcpu.ToCpu(values[i])
		v1 := (uint64(icpu) << 32) | uint64(isub)
		jcpu, jsub, _ := mcpu.ToCpu(values[j])
		v2 := (uint64(jcpu) << 32) | uint64(jsub)
		return v1 < v2
	})
	return values
}
