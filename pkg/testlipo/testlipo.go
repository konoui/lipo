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
	"sync"
	"testing"

	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

var (
	godata = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

	once sync.Once
)

var TestDir = func() string {
	const dirname = "testlipo-output"
	// Try to find git repository root
	if gitRoot := findGitRoot(); gitRoot != "" {
		return filepath.Join(gitRoot, dirname)
	}
	// Fallback to temp directory
	return filepath.Join(os.TempDir(), dirname)
}()

func findGitRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for range 4 {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

type LipoBin struct {
	Bin       string
	segAligns []string
	hideArm64 bool
	fat64     bool
	ignoreErr bool
}

type TestLipo struct {
	*BinManager
	arches []string
	FatBin string
	LipoBin
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

func WithFat64(v bool) Opt {
	return func(l *LipoBin) {
		l.fat64 = v
	}
}

// WithIgnoreErr will ignore a lipo command error not to call t.Fatal()
func WithIgnoreErr(v bool) Opt {
	return func(l *LipoBin) {
		l.ignoreErr = v
	}
}

func Setup(t *testing.T, bm *BinManager, arches []string, opts ...Opt) *TestLipo {
	t.Helper()

	once.Do(func() {
		t.Logf("test dir %s", TestDir)
	})

	if len(arches) == 0 {
		t.Fatal("input arches are zero")
	}

	err := os.MkdirAll(bm.Dir, 0740)
	fatalIf(t, err)
	dir := bm.Dir

	bm.add(t, arches...)

	lipoBin := NewLipoBin(t, opts...)
	fatBin := filepath.Join(dir, "fat-"+strings.Join(arches, "-"))
	if lipoBin.hideArm64 {
		fatBin = fatBin + "-hideARM64"
	}
	if lipoBin.fat64 {
		fatBin = fatBin + "-fat64"
	}
	tmp := lipoBin.ignoreErr
	lipoBin.ignoreErr = false
	lipoBin.Create(t, fatBin, bm.getBinPaths(t, arches)...)
	lipoBin.ignoreErr = tmp

	return &TestLipo{
		arches:     arches,
		BinManager: bm,
		FatBin:     fatBin,
		LipoBin:    lipoBin,
	}
}

func (l *TestLipo) Bin(t *testing.T, arch string) (path string) {
	bin := l.getBinPath(t, arch)
	return bin
}

func (l *TestLipo) Bins(t *testing.T) (paths []string) {
	return l.getBinPaths(t, l.arches)
}

func (l *TestLipo) NewArchBin(t *testing.T, arch string) (path string) {
	t.Helper()
	return l.singleAdd(t, arch)
}

func (l *TestLipo) NewArchObj(t *testing.T, arch string) (path string) {
	t.Helper()
	return l.singleAdd(t, "obj_"+arch)
}

func NewLipoBin(t *testing.T, opts ...Opt) LipoBin {
	t.Helper()
	bin, err := exec.LookPath("lipo")
	if err != nil {
		t.Fatalf("could not find lipo command %v\n", err)
	}

	l := LipoBin{Bin: bin, segAligns: []string{}}
	for _, opt := range opts {
		if opt != nil {
			opt(&l)
		}
	}
	return l
}

func (l *LipoBin) DetailedInfo(t *testing.T, bins ...string) string {
	t.Helper()
	args := append([]string{"-detailed_info"}, bins...)
	cmd := exec.Command(l.Bin, args...)
	return execute(t, cmd, l.ignoreErr)
}

func (l *LipoBin) Info(t *testing.T, bins ...string) string {
	t.Helper()
	args := append([]string{"-info"}, bins...)
	cmd := exec.Command(l.Bin, args...)
	v := execute(t, cmd, l.ignoreErr)
	// Note arrange the output
	// if no fat case, suffix has `/n`
	// if fat case, suffix has `a space` and `/n`
	vs := strings.SplitN(v, "\n", len(bins))
	vvs := util.Map(vs, func(s string) string {
		return strings.TrimSuffix(strings.TrimSuffix(s, "\n"), " ")
	})
	return strings.Join(vvs, "\n") + "\n"
}

func (l *LipoBin) Create(t *testing.T, out string, inputs ...string) {
	t.Helper()
	args := []string{"-create", "-output", out}
	args = append(args, inputs...)
	args = append(args, l.segAligns...)
	if l.hideArm64 {
		args = append(args, "-hideARM64")
	}
	if l.fat64 {
		args = append(args, "-fat64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, l.ignoreErr)
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
	execute(t, cmd, l.ignoreErr)
}

func (l *LipoBin) Extract(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-extract", arches)
	args = append([]string{in, "-output", out}, args...)
	args = append(args, l.segAligns...)
	if l.fat64 {
		args = append(args, "-fat64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, l.ignoreErr)
}

func (l *LipoBin) ExtractFamily(t *testing.T, out, in string, arches []string) {
	t.Helper()
	args := appendCmd("-extract_family", arches)
	args = append([]string{in, "-output", out}, args...)
	args = append(args, l.segAligns...)
	if l.fat64 {
		args = append(args, "-fat64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, l.ignoreErr)
}

func (l *LipoBin) Thin(t *testing.T, out, in, arch string) {
	t.Helper()
	cmd := exec.Command(l.Bin, in, "-thin", arch, "-output", out)
	execute(t, cmd, l.ignoreErr)
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
	if l.fat64 {
		args = append(args, "-fat64")
	}
	cmd := exec.Command(l.Bin, args...)
	execute(t, cmd, l.ignoreErr)
}

func (l *LipoBin) Archs(t *testing.T, in string) string {
	t.Helper()
	cmd := exec.Command(l.Bin, in, "-archs")
	v := execute(t, cmd, l.ignoreErr)
	// Note arrange the output
	v = strings.TrimSuffix(v, "\n")
	return v
}

func execute(t *testing.T, cmd *exec.Cmd, ignoreErr bool) string {
	t.Helper()

	out, err := cmd.CombinedOutput()
	if err != nil && !ignoreErr {
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

func PatchFat64Reserved(t *testing.T, p string) {
	f, err := os.OpenFile(p, os.O_RDWR, 0777)
	fatalIf(t, err)

	ff, err := lmacho.NewFatFile(f)
	if err != nil {
		if errors.Is(err, lmacho.ErrThin) {
			return
		}
		fatalIf(t, err)
	}

	if ff.Magic != lmacho.MagicFat64 {
		return
	}

	// seek fatHeader
	_, err = f.Seek(int64(lmacho.FatHeaderSize()), io.SeekStart)
	fatalIf(t, err)

	for range ff.Arches {
		// seek an offset of fat64 reserved field
		off, err := f.Seek(int64(lmacho.FatArchHeaderSize(lmacho.MagicFat64)-4), io.SeekCurrent)
		fatalIf(t, err)

		// get current value of reserved field
		cur := uint32(0)
		err = binary.Read(f, binary.BigEndian, &cur)
		fatalIf(t, err)
		if cur == 0 {
			continue
		}

		t.Logf("[WARNING] fa64 reserved field is not zero: %d(0x%x) at %d, patching it with zero\n: %s", cur, cur, off, p)
		// reset the offset to patch reserved field
		_, err = f.Seek(int64(off), io.SeekStart)
		fatalIf(t, err)

		reserved := uint32(0)
		err = binary.Write(f, binary.BigEndian, &reserved)
		fatalIf(t, err)
	}

	err = f.Close()
	fatalIf(t, err)
}

func DiffPerm(t *testing.T, wantBin, gotBin string) {
	wantInfo, err := os.Stat(wantBin)
	fatalIf(t, err)

	gotInfo, err := os.Stat(gotBin)
	fatalIf(t, err)

	want, got := wantInfo.Mode().Perm(), gotInfo.Mode().Perm()
	if want != got {
		t.Errorf("want %s got %s", want, got)
	}
}

func DiffSha256(t *testing.T, wantBin, gotBin string) {
	t.Helper()

	want := calcSha256(t, wantBin)
	got := calcSha256(t, gotBin)
	if want != got {
		t.Errorf("want %s got %s", want, got)
		t.Log("dumping detail")
		b := NewLipoBin(t)

		t.Logf("want:\n%s\n", b.DetailedInfo(t, wantBin))
		printStat(t, wantBin)
		t.Logf("got:\n%s\n", b.DetailedInfo(t, gotBin))
		printStat(t, gotBin)
	}
}

func printStat(t *testing.T, bin string) {
	info, err := os.Stat(bin)
	fatalIf(t, err)
	t.Logf("size: %d\n", info.Size())
}

func compile(t *testing.T, mainfile, binPath, arch string) {
	args := []string{"build", "-o"}
	args = append(args, binPath, mainfile)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH="+arch)
	err := cmd.Run()
	fatalIf(t, err)
}

func calcSha256(t *testing.T, p string) string {
	f, err := os.Open(p)
	fatalIf(t, err)
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	fatalIf(t, err)

	return hex.EncodeToString(h.Sum(nil))
}

func copyAndManipulate(t *testing.T, src, dst string, arch string, typ macho.Type) {
	cpu, sub, ok := lmacho.ToCpu(arch)
	if !ok {
		t.Fatalf("copyAndManipulate: unsupported arch: %s\n", arch)
	}

	f, err := os.Open(src)
	fatalIf(t, err)
	defer f.Close()

	info, err := f.Stat()
	fatalIf(t, err)
	totalSize := info.Size()

	mo, err := macho.Open(src)
	fatalIf(t, err)
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

	_, err = f.Seek(int64(hdrSize), io.SeekCurrent)
	fatalIf(t, err)

	out, err := os.Create(dst)
	fatalIf(t, err)

	err = binary.Write(out, binary.LittleEndian, hdr)
	fatalIf(t, err)

	n, err := io.Copy(out, f)
	fatalIf(t, err)

	fatalIf(t, out.Close())

	if wantN := totalSize - int64(hdrSize); n != wantN {
		t.Fatalf("wrote body size. want: %d, got: %d\n", n, wantN)
	}
}

func fatalIf(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
