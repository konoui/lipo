package lipo_test

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/cgo_qsort"
	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/testlipo"
)

func init() {
	// using apple lipo sorter
	lmacho.SortFunc = cgo_qsort.Slice
}

func testSegAlignOpt(inputs []*lipo.SegAlignInput) testlipo.Opt {
	ain := []string{}
	for _, v := range inputs {
		ain = append(ain, "-segalign", v.Arch, v.AlignHex)
	}
	return testlipo.WithSegAlign(ain)
}

var (
	diffSha256 = func(t *testing.T, wantBin, gotBin string) {
		testlipo.DiffSha256(t, wantBin, gotBin)
		diffPerm(t, wantBin, gotBin)
	}
	cpuNames = func() []string {
		ret := []string{}
		for _, v := range lmacho.CpuNames() {
			// apple lipo does not support them
			if v == "armv8m" || v == "arm64_32" {
				continue
			}
			ret = append(ret, v)
		}
		return ret
	}
	diffPerm = func(t *testing.T, wantBin, gotBin string) {
		wantInfo, err := os.Stat(wantBin)
		if err != nil {
			t.Fatal(err)
		}
		gotInfo, err := os.Stat(gotBin)
		if err != nil {
			t.Fatal(err)
		}
		want, got := wantInfo.Mode().Perm(), gotInfo.Mode().Perm()
		if want != got {
			t.Errorf("want %s got %s", want, got)
		}
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

func init() {
	rand.Seed(time.Now().UnixNano())
}
