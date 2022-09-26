package lipo_test

import (
	"crypto/sha256"
	"debug/macho"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/konoui/lipo/pkg/lipo"
	"github.com/konoui/lipo/pkg/lipo/cgo_qsort"
	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func init() {
	// using apple lipo sorter
	lipo.SortFunc = cgo_qsort.Slice
}

var godata = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

const (
	inArm64 = "arm64"
	inAmd64 = "x86_64"
)

type lipoBin struct {
	bin   string
	exist bool
}

type testLipo struct {
	archBins map[string]string
	// for reserving bins() order
	arches []string
	dir    string
	fatBin string
	lipoBin
}

func setup(t *testing.T, arches ...string) *testLipo {
	t.Helper()

	dir := t.TempDir()

	mainfile := filepath.Join(dir, "main.go")
	err := os.WriteFile(mainfile, []byte(godata), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// base binaries
	amd64Bin := filepath.Join(dir, inAmd64)
	arm64Bin := filepath.Join(dir, inArm64)
	compile(t, mainfile, amd64Bin, "amd64")
	compile(t, mainfile, arm64Bin, "arm64")

	archBins := map[string]string{}
	for _, arch := range arches {
		// create base binary first,
		if arch == inAmd64 {
			archBins[inAmd64] = amd64Bin
		} else if arch == inArm64 {
			archBins[inArm64] = arm64Bin
		}
	}

	for _, arch := range arches {
		if !(arch == inAmd64 || arch == inArm64) {
			archBin := filepath.Join(dir, arch)
			copyAndManipulate(t, arm64Bin, archBin, arch)
			archBins[arch] = archBin
		}
	}

	lipoBin := newLipoBin(t)
	fatBin := filepath.Join(dir, randName())
	if !lipoBin.skip() && len(archBins) > 0 {
		// create fat bit for inputs
		inputs := make([]string, 0, len(archBins))
		for _, in := range arches {
			inputs = append(inputs, archBins[in])
		}
		lipoBin.create(t, fatBin, inputs...)
	}

	return &testLipo{
		dir:      dir,
		archBins: archBins,
		arches:   arches,
		fatBin:   fatBin,
		lipoBin:  lipoBin,
	}
}

func (l *testLipo) bin(t *testing.T, arch string) string {
	bin, ok := l.archBins[arch]
	if !ok {
		t.Fatalf("found no arch %s\n", arch)
	}
	return bin
}

func (l *testLipo) bins() []string {
	bins := make([]string, 0, len(l.archBins))
	for _, a := range l.arches {
		bins = append(bins, l.archBins[a])
	}
	return bins
}

func (l *lipoBin) skip() bool {
	return !l.exist
}

func (l *lipoBin) detailedInfo(t *testing.T, bin string) string {
	t.Helper()

	cmd := exec.Command(l.bin, "-detailed_info", bin)
	return execute(t, cmd, true)
}

func (l *lipoBin) create(t *testing.T, out string, inputs ...string) {
	t.Helper()
	args := []string{"-create", "-output", out}
	for _, f := range inputs {
		if "x86_64" == filepath.Base(f) {
			// specify 2^14(0x2000) alignment for X86_64 to remove platform dependency.
			args = append(args, "-segalign", "x86_64", "2000")
		}
	}
	args = append(args, inputs...)
	cmd := exec.Command(l.bin, args...)
	execute(t, cmd, true)
}

func (l *lipoBin) remove(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-remove", arches)
	args = append([]string{in, "-output", out}, args...)
	cmd := exec.Command(l.bin, args...)
	execute(t, cmd, true)
}

func (l *lipoBin) extract(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-extract", arches)
	args = append([]string{in, "-output", out}, args...)
	cmd := exec.Command(l.bin, args...)
	execute(t, cmd, true)
}

func (l *lipoBin) thin(t *testing.T, out, in, arch string) {
	t.Helper()
	cmd := exec.Command(l.bin, in, "-thin", arch, "-output", out)
	execute(t, cmd, true)
}

func (l *lipoBin) replace(t *testing.T, out, in, arch, replace string) {
	t.Helper()
	args := []string{in, "-replace", arch, replace, "-output", out}
	if arch == "x86_64" {
		args = append(args, "-segalign", "x86_64", "2000")
	}
	cmd := exec.Command(l.bin, args...)
	execute(t, cmd, true)
}

func (l *lipoBin) archs(t *testing.T, in string) string {
	t.Helper()
	cmd := exec.Command(l.bin, in, "-archs")
	return execute(t, cmd, false)
}

func execute(t *testing.T, cmd *exec.Cmd, combine bool) string {
	t.Helper()

	var out []byte
	var err error
	if combine {
		out, err = cmd.CombinedOutput()
	} else {
		out, err = cmd.Output()
	}
	if err != nil {
		t.Log("CMD:", cmd.String())
		t.Log("OUTPUT:", string(out))
		t.Fatalf("Error: %v", err)
	}
	return string(out)
}

func appendCmd(cmd string, args []string) []string {
	ret := []string{}
	for _, a := range args {
		ret = append(ret, cmd, a)
	}
	return ret
}

func compile(t *testing.T, mainfile, binPath, arch string) {
	t.Helper()

	args := []string{"build", "-o"}
	args = append(args, binPath, mainfile)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH="+arch)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func diffSha256(t *testing.T, wantBin, gotBin string) {
	t.Helper()

	want := calcSha256(t, wantBin)
	got := calcSha256(t, gotBin)
	if want != got {
		t.Errorf("want %s got %s", want, got)
		t.Log("dumping detail")
		b := newLipoBin(t)
		if b.skip() {
			return
		}
		t.Logf("want:\n%s\n", b.detailedInfo(t, wantBin))
		t.Logf("got:\n%s\n", b.detailedInfo(t, gotBin))
	}
}

func calcSha256(t *testing.T, p string) string {
	t.Helper()
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func newLipoBin(t *testing.T) lipoBin {
	t.Helper()
	bin, err := exec.LookPath("lipo")
	if errors.Is(err, exec.ErrNotFound) {
		return lipoBin{exist: false}
	}

	if err != nil {
		t.Fatalf("could not find lipo command %v\n", err)
	}
	return lipoBin{exist: true, bin: bin}
}

func copyAndManipulate(t *testing.T, src, dst string, arch string) {
	t.Helper()

	cpu, sub, ok := mcpu.ToCpu(arch)
	if !ok {
		t.Fatalf("unsupported arch: %s\n", arch)
	}

	f, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	totalSize := info.Size()

	mo, err := macho.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	hdr := mo.FileHeader
	wantHdrSize := binary.Size(hdr)
	hdr.Cpu = cpu
	hdr.SubCpu = sub
	hdrSize := binary.Size(hdr)

	if hdrSize != wantHdrSize {
		t.Fatalf("unexpected header size want: %d, got: %d\n", wantHdrSize, hdrSize)
	}

	if _, err := f.Seek(int64(hdrSize), io.SeekCurrent); err != nil {
		t.Fatal(err)
	}

	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}

	if err := binary.Write(out, binary.LittleEndian, hdr); err != nil {
		t.Fatal(err)
	}

	n, err := io.Copy(out, f)
	if err != nil {
		t.Fatal(err)
	}

	if err := out.Close(); err != nil {
		t.Fatal(err)
	}

	if wantN := totalSize - int64(hdrSize); n != wantN {
		t.Fatalf("wrote body size. want: %d, got: %d\n", n, wantN)
	}
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randName() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
