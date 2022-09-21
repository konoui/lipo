package lipo_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
		lipoBin.createFatBin(t, lipoFatBin, amd64Bin, arm64Bin)
	}

	return &testLipo{
		amd64Bin:   amd64Bin,
		arm64Bin:   arm64Bin,
		dir:        dir,
		lipoBin:    lipoBin,
		lipoFatBin: lipoFatBin,
	}
}

func TestLipo_Create(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		p := setup(t)

		// check fat file format
		gotBin := filepath.Join(p.dir, "out-amd64-arm64-binary")
		createFatBin(t, gotBin, p.amd64Bin, p.arm64Bin)

		fatBin := filepath.Join(p.dir, "out-arm64-amd64-binary")
		createFatBin(t, fatBin, p.arm64Bin, p.amd64Bin)
		if p.skip() {
			t.Skip("skip lipo binary test")
		}

		p.lipoDetail(t, gotBin)
		p.lipoDetail(t, fatBin)
		diffSha256(t, p.lipoFatBin, gotBin)
	})
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

func (l *lipoBin) skip() bool {
	return !l.exist
}

func (l *lipoBin) lipoDetail(t *testing.T, bin string) string {
	t.Helper()

	cmd := exec.Command(l.bin, "-detailed_info", bin)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error lipo -detailed_info: %v\n", err)
	}
	return string(out)
}

func (l *lipoBin) createFatBin(t *testing.T, out, input1, input2 string) {
	t.Helper()
	// specify 2^14(0x2000) alignment for X86_64 to remove platform dependency.
	cmd := exec.Command(l.bin, "-segalign", "x86_64", "2000", "-create", input1, input2, "-output", out)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create original fat binary: %v\n %s", err, cmd.String())
	}
}

func (l *lipoBin) removeFatBin(t *testing.T, in, out, arch string) {
	t.Helper()
	cmd := exec.Command(l.bin, in, "-remove", arch, "-output", out)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to remove from original fat binary: %v\n %s", err, cmd.String())
	}
}
