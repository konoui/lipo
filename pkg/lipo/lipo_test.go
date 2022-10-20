package lipo_test

import (
	"math/rand"
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
	diffSha256 = testlipo.DiffSha256
	cpuNames   = func() []string {
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
