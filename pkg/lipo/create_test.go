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
	}{
		{
			name:   "-create with single thin",
			arches: []string{inAmd64},
		},
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
			arches: mcpu.CpuNames(),
		},
		{
			name:      "-create -segalign x86_64 10 (2^4)",
			arches:    []string{inAmd64, inArm64},
			segAligns: []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "10"}, {Arch: "arm64", AlignHex: "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches...)

			got := filepath.Join(p.Dir, randName())
			l := lipo.New(
				lipo.WithInputs(p.Bins()...),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns))
			if err := l.Create(); err != nil {
				t.Fatalf("failed to create fat bin %v", err)
			}

			// tests for fat bin is expected
			verifyArches(t, got, tt.arches...)

			if p.Skip() {
				t.Skip("skip lipo binary test")
			}

			// re-create fat binary with seg align
			for _, segAlign := range tt.segAligns {
				p.AddSegAlign(segAlign.Arch, segAlign.AlignHex)
			}
			if len(tt.segAligns) != 0 {
				p.Create(t, p.FatBin, p.Bins()...)
			}
			diffSha256(t, p.FatBin, got)
		})
	}
}

func TestLipo_CreateError(t *testing.T) {
	tests := []struct {
		name       string
		arches     []string
		segAligns  []*lipo.SegAlignInput
		wantErrMsg string
	}{
		{
			name:       "-create -segalign x86_64 1",
			arches:     []string{inAmd64, inArm64},
			segAligns:  []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "1"}},
			wantErrMsg: "argument to -segalign <arch_type> 1 (hex) must be a non-zero power of two",
		},
		{
			name:       "-create -segalign x86_64 10 (2^4) -segalign x86_64 (1^2)",
			arches:     []string{inAmd64, inArm64},
			segAligns:  []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "10"}, {Arch: "x86_64", AlignHex: "2"}},
			wantErrMsg: "-segalign x86_64 <value> specified multiple times",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches...)

			got := filepath.Join(p.Dir, randName())
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
