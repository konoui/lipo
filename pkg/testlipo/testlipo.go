package testlipo

import (
	"crypto/sha256"
	"debug/macho"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/util"
)

var godata = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

type LipoBin struct {
	Bin       string
	segAligns []string
	hideArm64 bool
	exist     bool
}

type TestLipo struct {
	archBins map[string]string
	// for reserving bins() order
	arches []string
	Dir    string
	FatBin string
	LipoBin
	arm64Bin string
}

type Opt func(l *LipoBin)

func WithSegAlign(sa []string) Opt {
	return func(l *LipoBin) {
		l.segAligns = sa
	}
}

func WithHideArm64(v bool) Opt {
	return func(l *LipoBin) {
		l.hideArm64 = v
	}
}

var TempDir string

func Setup(t *testing.T, arches []string, opts ...Opt) *TestLipo {
	t.Helper()

	dir := TempDir
	if TempDir == "" {
		dir = t.TempDir()
	}

	mainfile := filepath.Join(dir, "main.go")
	err := os.WriteFile(mainfile, []byte(godata), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// base binaries
	amd64Bin := filepath.Join(dir, "x86_64")
	arm64Bin := filepath.Join(dir, "arm64")
	compile(t, mainfile, amd64Bin, "amd64")
	compile(t, mainfile, arm64Bin, "arm64")

	archBins := map[string]string{}
	for _, arch := range arches {
		// create base binary first,
		if arch == "x86_64" {
			archBins[arch] = amd64Bin
		} else if arch == "arm64" {
			archBins[arch] = arm64Bin
		} else if strings.HasPrefix(arch, "obj_") {
			archBin := filepath.Join(dir, arch)
			copyAndManipulate(t, arm64Bin, archBin, arch[4:], macho.TypeObj)
			archBins[arch] = archBin
		} else {
			archBin := filepath.Join(dir, arch)
			copyAndManipulate(t, arm64Bin, archBin, arch, macho.TypeExec)
			archBins[arch] = archBin
		}
	}

	lipoBin := NewLipoBin(t, opts...)
	fatBin := filepath.Join(dir, "fat-"+strings.Join(arches, "-"))
	if !lipoBin.Skip() && len(archBins) > 0 {
		// create fat bit for inputs
		inputs := make([]string, 0, len(archBins))
		for _, in := range arches {
			inputs = append(inputs, archBins[in])
		}
		lipoBin.Create(t, fatBin, inputs...)
	}

	return &TestLipo{
		Dir:      dir,
		archBins: archBins,
		arches:   arches,
		FatBin:   fatBin,
		LipoBin:  lipoBin,
		arm64Bin: arm64Bin,
	}
}

func (l *TestLipo) Bin(t *testing.T, arch string) string {
	bin, ok := l.archBins[arch]
	if !ok {
		t.Fatalf("found no arch %s\n", arch)
	}
	return bin
}

func (l *TestLipo) Bins() []string {
	bins := make([]string, 0, len(l.archBins))
	for _, a := range l.arches {
		bins = append(bins, l.archBins[a])
	}
	return bins
}

func (l *TestLipo) NewArchBin(t *testing.T, arch string) (path string) {
	t.Helper()
	archBin := filepath.Join(l.Dir, "new-arch-bin-"+arch)
	copyAndManipulate(t, l.arm64Bin, archBin, arch, macho.TypeExec)
	return archBin
}

func (l *TestLipo) NewArchObj(t *testing.T, arch string) (path string) {
	t.Helper()
	archBin := filepath.Join(l.Dir, "new-arch-obj-"+arch)
	copyAndManipulate(t, l.arm64Bin, archBin, arch, macho.TypeObj)
	return archBin
}

func NewLipoBin(t *testing.T, opts ...Opt) LipoBin {
	t.Helper()
	bin, err := exec.LookPath("lipo")
	if errors.Is(err, exec.ErrNotFound) {
		return LipoBin{exist: false}
	}

	l := LipoBin{exist: true, Bin: bin, segAligns: []string{}}
	for _, opt := range opts {
		if opt != nil {
			opt(&l)
		}
	}

	if err != nil {
		t.Fatalf("could not find lipo command %v\n", err)
	}
	return l
}

func (l *LipoBin) Skip() bool {
	return !l.exist
}

func (l *LipoBin) DetailedInfo(t *testing.T, bins ...string) string {
	t.Helper()
	args := append([]string{"-detailed_info"}, bins...)
	cmd := exec.Command(l.Bin, args...)
	return execute(t, cmd, true)
}

func (l *LipoBin) Info(t *testing.T, bins ...string) string {
	t.Helper()
	args := append([]string{"-info"}, bins...)
	cmd := exec.Command(l.Bin, args...)
	// if no fat case, suffix has /n
	// if fat case, suffix has space and /n
	v := execute(t, cmd, true)
	vs := strings.SplitN(v, "\n", len(bins))
	vvs := util.Map(vs, func(s string) string {
		return strings.TrimSuffix(strings.TrimSuffix(s, "\n"), " ")
	})
	return strings.Join(vvs, "\n")
}

func (l *LipoBin) Create(t *testing.T, out string, inputs ...string) {
	t.Helper()
	args := []string{"-create", "-output", out}
	args = append(args, inputs...)
	args = append(args, l.segAligns...)
	if l.hideArm64 {
		args = append(args, "-hideARM64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, true)
}

func (l *LipoBin) Remove(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-remove", arches)
	args = append([]string{in, "-output", out}, args...)
	args = append(args, l.segAligns...)
	if l.hideArm64 {
		args = append(args, "-hideARM64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, true)
}

func (l *LipoBin) Extract(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-extract", arches)
	args = append([]string{in, "-output", out}, args...)
	args = append(args, l.segAligns...)
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, true)
}

func (l *LipoBin) ExtractFamily(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-extract_family", arches)
	args = append([]string{in, "-output", out}, args...)
	args = append(args, l.segAligns...)
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, true)
}

func (l *LipoBin) Thin(t *testing.T, out, in, arch string) {
	t.Helper()
	cmd := exec.Command(l.Bin, in, "-thin", arch, "-output", out)
	execute(t, cmd, true)
}

func (l *LipoBin) Replace(t *testing.T, out, in string, archBins [][2]string) {
	t.Helper()

	archBinArgs := []string{}
	for _, archBin := range archBins {
		archBinArgs = append(archBinArgs, "-replace", archBin[0], archBin[1])
	}
	args := append([]string{in, "-output", out}, archBinArgs...)
	args = append(args, l.segAligns...)
	if l.hideArm64 {
		args = append(args, "-hideARM64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, true)
}

func (l *LipoBin) Archs(t *testing.T, in string) string {
	t.Helper()
	cmd := exec.Command(l.Bin, in, "-archs")
	v := execute(t, cmd, false)
	v = strings.TrimSuffix(v, "\n")
	return v
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

func DiffSha256(t *testing.T, wantBin, gotBin string) {
	t.Helper()

	want := calcSha256(t, wantBin)
	got := calcSha256(t, gotBin)
	if want != got {
		t.Errorf("want %s got %s", want, got)
		t.Log("dumping detail")
		b := NewLipoBin(t)
		if b.Skip() {
			return
		}
		t.Logf("want:\n%s\n", b.DetailedInfo(t, wantBin))
		t.Logf("got:\n%s\n", b.DetailedInfo(t, gotBin))
	}
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

func copyAndManipulate(t *testing.T, src, dst string, arch string, typ macho.Type) {
	t.Helper()
	cpu, sub, ok := mcpu.ToCpu(arch)
	if !ok {
		t.Fatalf("copyAndManipulate: unsupported arch: %s\n", arch)
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
	hdr.Type = typ
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
