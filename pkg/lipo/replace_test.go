package lipo_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
	"github.com/konoui/lipo/pkg/util"
)

func TestLipo_Replace(t *testing.T) {
	tests := []struct {
		name          string
		replaceArches []string
		arches        []string
		segAligns     []*lipo.SegAlignInput
		hideArm64     bool
	}{
		{
			name:          "-replace x86_64",
			replaceArches: []string{"x86_64"},
			arches:        []string{"arm64", "x86_64"},
		},
		{
			name:          "-replace arm64e",
			replaceArches: []string{"arm64e"},
			arches:        []string{"arm64", "x86_64", "arm64e"},
		},
		{
			name:          "-replace arm64 -replace x86_64",
			replaceArches: []string{"arm64", "x86_64", "arm64e"},
			arches:        []string{"arm64", "x86_64", "arm64e"},
		},
		{
			name:          "-replace x86_64 from x86_64 fat binary",
			replaceArches: []string{"x86_64"},
			arches:        []string{"x86_64"},
		},
		{
			name:          "-replace x86_64 <file> -segalign x86_64 1 -segalign arm64 2",
			replaceArches: []string{"x86_64"},
			arches:        []string{"x86_64", "arm64"},
			segAligns:     []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "1"}, {Arch: "arm64", AlignHex: "2"}},
		},
		{
			name:          "-replace amd64 hideARM64",
			arches:        []string{"armv7k", "arm64"},
			replaceArches: []string{"arm64"},
			hideArm64:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches,
				testSegAlignOpt(tt.segAligns),
				testlipo.WithHideArm64(tt.hideArm64))
			ri := util.Map(tt.replaceArches, func(v string) *lipo.ReplaceInput {
				return &lipo.ReplaceInput{Arch: v, Bin: p.Bin(t, v)}
			})
			got := filepath.Join(p.Dir, gotName(t))
			opts := []lipo.Option{
				lipo.WithInputs(p.FatBin),
				lipo.WithOutput(got),
				lipo.WithSegAlign(tt.segAligns),
			}
			if tt.hideArm64 {
				opts = append(opts, lipo.WithHideArm64())
			}
			l := lipo.New(opts...)
			if err := l.Replace(ri); err != nil {
				t.Fatalf("replace error: %v\n", err)
			}

			verifyArches(t, got, tt.arches...)

			want := filepath.Join(p.Dir, wantName(t))
			p.Replace(t, want, p.FatBin, rapReplaceInputs(ri))
			diffSha256(t, want, got)
		})
	}
}

func rapReplaceInputs(ri []*lipo.ReplaceInput) [][2]string {
	ret := [][2]string{}
	for _, i := range ri {
		ret = append(ret, [2]string{i.Arch, i.Bin})
	}
	return ret
}

func TestLipo_ReplaceError(t *testing.T) {
	p := testlipo.Setup(t, []string{"arm64", "x86_64"})
	got := filepath.Join(p.Dir, gotName(t))
	l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))

	t.Run("duplicate arch", func(t *testing.T) {
		to := p.Bin(t, "arm64")
		ri := []*lipo.ReplaceInput{{Arch: "arm64", Bin: to}, {Arch: "arm64", Bin: to}}
		err := l.Replace(ri)
		wantErrMsg := "duplicate architecture arm64"
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}
	})

	t.Run("not-match-arch-and-bin", func(t *testing.T) {
		ri := []*lipo.ReplaceInput{{Arch: "x86_64", Bin: p.Bin(t, "arm64")}}
		err := l.Replace(ri)
		wantErrMsg := fmt.Sprintf("specified architecture: %s for input file: %s does not match the file's architecture", ri[0].Arch, ri[0].Bin)
		if err == nil {
			t.Fatal("no error")
		}
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}
	})

	t.Run("fat-bin-not-have-input-arch", func(t *testing.T) {
		ri := []*lipo.ReplaceInput{{Arch: "arm64e", Bin: p.NewArchBin(t, "arm64e")}}
		err := l.Replace(ri)
		wantErrMsg := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", ri[0].Arch, p.FatBin)
		if err == nil {
			t.Fatal("no error")
		}
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}
	})

	t.Run("-hideARM64-with-obj", func(t *testing.T) {
		p := testlipo.Setup(t, []string{"arm64", "armv7k"})
		objArm64 := p.NewArchObj(t, "arm64")
		l := lipo.New(
			lipo.WithInputs(p.FatBin),
			lipo.WithOutput(got),
			lipo.WithHideArm64())
		err := l.Replace([]*lipo.ReplaceInput{{Arch: "arm64", Bin: objArm64}})
		wantErrMsg := fmt.Sprintf("hideARM64 specified but thin file %s is not of type MH_EXECUTE", objArm64)
		if err == nil {
			t.Fatal("no error")
		}
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}
	})
}
