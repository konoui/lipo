package lipo_test

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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
		t.Helper()
		diffPerm(t, wantBin, gotBin)
		patchFat64Reserved(t, wantBin)
		testlipo.DiffSha256(t, wantBin, gotBin)
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
	patchFat64Reserved = func(t *testing.T, p string) {
		ff, err := lmacho.OpenFat(p)
		if err != nil {
			if errors.Is(err, macho.ErrNotFat) {
				return
			}
			t.Fatal(err)
		}

		if ff.Magic != lmacho.MagicFat64 {
			return
		}

		f, err := os.OpenFile(p, os.O_RDWR, 0777)
		fatalIf(t, err)

		// seek fatHeader
		_, err = f.Seek(4*2, io.SeekStart)
		fatalIf(t, err)

		for _, fa := range ff.AllArches() {
			off := binary.Size(fa.FatArchHeader)
			_, err = f.Seek(int64(off), io.SeekCurrent)
			fatalIf(t, err)
			reserved := uint32(0)
			err = binary.Write(f, binary.BigEndian, &reserved)
			fatalIf(t, err)
		}

		err = f.Close()
		fatalIf(t, err)
	}
	diffPerm = func(t *testing.T, wantBin, gotBin string) {
		wantInfo, err := os.Stat(wantBin)
		fatalIf(t, err)

		gotInfo, err := os.Stat(gotBin)
		fatalIf(t, err)

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

func fatalIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
