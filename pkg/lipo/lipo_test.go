package lipo_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/testlipo"
	"github.com/konoui/lipo/pkg/util"
)

func testSegAlignOpt(inputs []*lipo.SegAlignInput) testlipo.Opt {
	ain := []string{}
	for _, v := range inputs {
		ain = append(ain, "-segalign", v.Arch, v.AlignHex)
	}
	return testlipo.WithSegAlign(ain)
}

var (
	diffSha256 = func(t *testing.T, wantBin, gotBin string) {
		t.Helper()
		testlipo.DiffPerm(t, wantBin, gotBin)
		testlipo.PatchFat64Reserved(t, wantBin)
		testlipo.DiffSha256(t, wantBin, gotBin)
	}
	cpuNames = func() []string {
		return util.Filter(lmacho.CpuNames(), func(v string) bool {
			return v != "armv8m" && v != "arm64_32"
		})
	}
)

func currentArch() string {
	a := runtime.GOARCH
	if a == "amd64" {
		return "x86_64"
	}
	return a
}

func gotName(t *testing.T) string {
	return "got_" + filepath.Base(t.Name())
}

func wantName(t *testing.T) string {
	return "want_" + filepath.Base(t.Name())
}

func contain(tg string, l []string) bool {
	for _, s := range l {
		if tg == s {
			return true
		}
	}
	return false
}
