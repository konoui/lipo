package lipo_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/testlipo"
)

func TestLipo_Remove(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		arches    []string
		segAligns []*lipo.SegAlignInput
	}{
		{
			name:   "-remove x86_64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"x86_64"},
		},
		{
			name:   "-remove arm64",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64"},
		},
		{
			name:   "-remove arm64 -remove arm64e",
			inputs: []string{inAmd64, inArm64, "arm64e"},
			arches: []string{"arm64", "arm64e"},
		},
		{
			name:      "-remove x86_64 -segalign arm64 2",
			inputs:    []string{"x86_64", "arm64"},
			arches:    []string{"x86_64"},
			segAligns: []*lipo.SegAlignInput{{Arch: "arm64", AlignHex: "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testlipo.Setup(t, tt.inputs...)

			got := filepath.Join(p.Dir, gotName(t))
			arches := tt.arches
			l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got), lipo.WithSegAlign(tt.segAligns))
			if err := l.Remove(arches...); err != nil {
				t.Errorf("remove error %v\n", err)
			}

			wantArches := []string{}
			for _, i := range tt.inputs {
				if !contain(i, tt.arches) {
					wantArches = append(wantArches, i)
				}
			}
			verifyArches(t, got, wantArches...)

			if p.Skip() {
				t.Skip("skip lipo binary tests")
			}

			// set segalign for next Remove
			for _, segAlign := range tt.segAligns {
				p.AddSegAlign(segAlign.Arch, segAlign.AlignHex)
			}

			want := filepath.Join(p.Dir, wantName(t))
			p.Remove(t, want, p.FatBin, arches)
			diffSha256(t, want, got)
		})
	}
}

func TestLipo_RemoveError(t *testing.T) {
	t.Run("not-match-arch", func(t *testing.T) {
		p := testlipo.Setup(t, "arm64", "x86_64")

		got := filepath.Join(p.Dir, wantName(t))
		l := lipo.New(lipo.WithInputs(p.FatBin), lipo.WithOutput(got))
		err := l.Remove("arm64e")
		if err == nil {
			t.Errorf("error does not occur")
		}

		want := fmt.Sprintf("%s specified but fat file: %s does not contain that architecture", "arm64e", p.FatBin)
		if got := err.Error(); got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
