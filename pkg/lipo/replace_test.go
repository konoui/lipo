package lipo_test

import (
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Replace(t *testing.T) {
	tests := []struct {
		name          string
		replaceArches []string
		arches        []string
		segAligns     []*lipo.SegAlignInput
	}{
		{
			name:          "-replace x86_64",
			replaceArches: []string{"x86_64"},
			arches:        []string{inArm64, inAmd64},
		},
		{
			name:          "-replace arm64e",
			replaceArches: []string{"arm64e"},
			arches:        []string{inArm64, inAmd64, "arm64e"},
		},
		{
			name:          "-replace arm64 -replace x86_64",
			replaceArches: []string{"arm64", "x86_64", "arm64e"},
			arches:        []string{inArm64, inAmd64, "arm64e"},
		},
		{
			name:          "-replace x86_64 from x86_64 fat binary",
			replaceArches: []string{"x86_64"},
			arches:        []string{"x86_64"},
		},
		{
			name:          "-replace x86_64 <file> -segalign x86_64 2 -segalign arm64 2",
			replaceArches: []string{"x86_64"},
			arches:        []string{"x86_64", "arm64"},
			segAligns:     []*lipo.SegAlignInput{{Arch: "x86_64", AlignHex: "2"}, {Arch: "arm64", AlignHex: "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.arches...)
			ri := []*lipo.ReplaceInput{}
			for _, a := range tt.replaceArches {
				ri = append(ri, &lipo.ReplaceInput{
					Arch: a,
					Bin:  p.Bin(t, a),
				})
			}

			got := filepath.Join(p.Dir, randName())
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got), lipo.WithSegAlign(tt.segAligns))
			if err := l.Replace(ri); err != nil {
				t.Fatalf("replace error: %v\n", err)
			}

			verifyArches(t, got, tt.arches...)

			// set segalign for next Replace
			for _, segAlign := range tt.segAligns {
				p.AddSegAlign(segAlign.Arch, segAlign.AlignHex)
			}

			want := filepath.Join(p.Dir, randName())
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
	t.Run("duplicate arch", func(t *testing.T) {
		p := testlipo.Setup(t, "arm64", "x86_64")

		to := p.Bin(t, "arm64")
		got := filepath.Join(p.Dir, randName())
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))

		ri := []*lipo.ReplaceInput{{Arch: "arm64", Bin: to}, {Arch: "arm64", Bin: to}}
		err := l.Replace(ri)
		wantErrMsg := "replace inputs: want 1 but got 2"
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}

		ri = []*lipo.ReplaceInput{{Arch: "x86_64", Bin: p.Bin(t, "arm64")}}
		err = l.Replace(ri)
		wantErrMsg = "specified architecture: x86_64 for replacement file: arm64 does not match the file's architecture"
		if err.Error() != wantErrMsg {
			t.Errorf("want: %s, got: %s", wantErrMsg, err.Error())
		}
	})
}
