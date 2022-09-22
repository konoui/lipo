package lipo_test

import (
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

var godata = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

type lipoBin struct {
	bin   string
	exist bool
}

type testLipo struct {
	amd64Bin string
	arm64Bin string
	dir      string
	lipoBin
	lipoFatBin string
}

func (l *lipoBin) skip() bool {
	return !l.exist
}

func (l *lipoBin) lipoDetail(t *testing.T, bin string) string {
	t.Helper()

	cmd := exec.Command(l.bin, "-detailed_info", bin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error lipo -detailed_info: %v\n%s\n", err, string(out))
	}
	return string(out)
}

func (l *lipoBin) create(t *testing.T, out, input1, input2 string) {
	t.Helper()
	// specify 2^14(0x2000) alignment for X86_64 to remove platform dependency.
	cmd := exec.Command(l.bin, "-segalign", "x86_64", "2000", "-create", input1, input2, "-output", out)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create original fat binary: %v\n %s", err, cmd.String())
	}
}

func (l *lipoBin) remove(t *testing.T, out, in, arch string) {
	t.Helper()
	cmd := exec.Command(l.bin, in, "-remove", arch, "-output", out)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to remove from original fat binary: %v\n %s", err, cmd.String())
	}
}

func (l *lipoBin) extract(t *testing.T, out, in, arch string) {
	t.Helper()
	cmd := exec.Command(l.bin, in, "-extract", arch, "-output", out)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to extract from original fat binary: %v\n %s", err, cmd.String())
	}
}

func setup(t *testing.T) *testLipo {
	t.Helper()

	dir := t.TempDir()

	mainfile := filepath.Join(dir, "main.go")
	err := os.WriteFile(mainfile, []byte(godata), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	amd64Bin := filepath.Join(dir, "amd64")
	arm64Bin := filepath.Join(dir, "arm64")
	compile(t, mainfile, amd64Bin, "amd64")
	compile(t, mainfile, arm64Bin, "arm64")

	lipoFatBin := ""
	lipoBin := newLipoBin(t)
	if !lipoBin.skip() {
		lipoFatBin = filepath.Join(dir, "lipo-fat-bin")
		lipoBin.create(t, lipoFatBin, amd64Bin, arm64Bin)
	}

	return &testLipo{
		amd64Bin:   amd64Bin,
		arm64Bin:   arm64Bin,
		dir:        dir,
		lipoBin:    lipoBin,
		lipoFatBin: lipoFatBin,
	}
}

func createFatBin(t *testing.T, out, input1, input2 string) {
	t.Helper()
	l := lipo.New(lipo.WithInputs(input1, input2), lipo.WithOutput(out))
	if err := l.Create(); err != nil {
		t.Fatalf("failed to create fat bin %v", err)
	}

	f, err := macho.OpenFat(out)
	if err != nil {
		t.Fatalf("invalid fat file: %v\n", err)
	}
	defer f.Close()
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

func diffSha256(t *testing.T, input1, input2 string) {
	t.Helper()

	got := calcSha256(t, input1)
	want := calcSha256(t, input2)
	if want != got {
		t.Errorf("want %s got %s", want, got)
		t.Log("dumping detail")
		b := newLipoBin(t)
		if b.skip() {
			return
		}
		t.Logf("want:\n%s\n", b.lipoDetail(t, input1))
		t.Logf("got:\n%s\n", b.lipoDetail(t, input2))
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
