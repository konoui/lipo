package lipo_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Create(t *testing.T) {
	tests := []struct {
		name      string
		arches    []string
		segAligns []*lipo.SegAlignInput
		hideArm64 bool
	}{
		{
			name:   "-create with single thin",
			arches: []string{"x86_64"},
		},
		{
			name:   "-create 2 files",
			arches: []string{"x86_64", "arm64"},
		},
		{
			name:   "-create 3 files",
			arches: []string{"arm64", "x86_64", "arm64e"},
		},
		{
			name:   "-create many files",
			arches: mcpu.CpuNames(),
		},
		{
			name:   "-create object files",
			arches: []string{"obj_" + currentArch(), "arm64e", "x86_64h"},
		},
		{
			name:      "-create -segalign x86_64 10 (2^4)",
			arches:    []string{"x86_64", "arm64"},
			segAligns: []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "10"}, {Arch: "arm64", AlignHex: "1"}},
		},
		{
			name:      "-create hideARM64",
			arches:    []string{"armv7k", "arm64"},
			hideArm64: true,
		},
		{
			name:      "-create hideARM64",
			arches:    []string{"armv7k", "arm64", "arm64e"},
			hideArm64: true,
		},
		{
			name:      "-create hideARM64 -segalign armv7k 2 -segalign arm64 2",
			arches:    []string{"armv7k", "arm64"},
			segAligns: []*lipo.SegAlignInput{{Arch: "armv7k", AlignHex: "1"}, {Arch: "arm64", AlignHex: "2"}},
			hideArm64: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches,
				testSegAlignOpt(tt.segAligns),
				testlipo.WithHideArm64(tt.hideArm64))

			got := filepath.Join(p.Dir, gotName(t))
			opts := []lipo.Option{
				lipo.WithInputs(p.Bins()...),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns),
			}
			if tt.hideArm64 {
				opts = append(opts, lipo.WithHideArm64())
			}

			if err := lipo.New(opts...).Create(); err != nil {
				t.Fatalf("failed to create fat bin %v", err)
			}

			// tests for fat bin is expected
			verifyArches(t, got, tt.arches...)

			if p.Skip() {
				t.Skip("skip lipo binary test")
			}

			diffSha256(t, p.FatBin, got)
		})
	}
}

func TestLipo_CreateError(t *testing.T) {
	tests := []struct {
		name       string
		segAligns  []*lipo.SegAlignInput
		wantErrMsg string
	}{
		{
			name:       "-create -segalign x86_64 0",
			segAligns:  []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "0"}},
			wantErrMsg: "segalign 0 (hex) must be a non-zero power of two",
		},
		{
			name:       "-create -segalign x86_64 10 (2^4) -segalign x86_64 (1^2)",
			segAligns:  []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "10"}, {Arch: "x86_64", AlignHex: "2"}},
			wantErrMsg: "segalign x86_64 specified multiple times",
		},
		{
			name:       "-create -segalign arm64e 10 (2^4)",
			segAligns:  []*lipo.SegAlignInput{{Arch: "arm64e", AlignHex: "10"}},
			wantErrMsg: "segalign arm64e specified but resulting fat file does not contain that architecture",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, []string{"x86_64", "arm64"})
			got := filepath.Join(p.Dir, gotName(t))
			l := lipo.New(
				lipo.WithInputs(p.Bins()...),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns))
			err := l.Create()
			if err == nil {
				t.Fatal("error not occur")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("want: %s, got: %s", tt.wantErrMsg, err.Error())
			}
		})
	}
}
