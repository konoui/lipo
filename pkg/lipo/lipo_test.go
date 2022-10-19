package lipo_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/cgo_qsort"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/testlipo"
)

func init() {
	// using apple lipo sorter
	lipo.SortFunc = cgo_qsort.Slice
	tempDir := filepath.Join(os.TempDir(), "testlipo-output")
	fmt.Println("using testlipo-output", tempDir)
	err := os.MkdirAll(tempDir, 0740)
	if err != nil {
		panic(err)
	}
	testlipo.TempDir = tempDir
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
		for _, v := range mcpu.CpuNames() {
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
