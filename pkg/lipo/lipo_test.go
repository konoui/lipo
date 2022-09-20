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
	"sync"
	"testing"

	"github.com/konoui/lipo/pkg/lipo"
)

var data = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

func TestLipo_Create(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		dir := t.TempDir()
		mainfile := filepath.Join(dir, "main.go")
		err := os.WriteFile(mainfile, []byte(data), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		amd64 := "amd64"
		arm64 := "arm64"
		amd64Bin := filepath.Join(dir, amd64)
		arm64Bin := filepath.Join(dir, arm64)
		wg := &sync.WaitGroup{}
		for _, cmd := range []*exec.Cmd{
			compileCmd(mainfile, amd64Bin, amd64),
			compileCmd(mainfile, arm64Bin, arm64),
		} {
			cmd := cmd
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := cmd.Run(); err != nil {
					t.Errorf(err.Error())
				}
			}()
		}

		wg.Wait()

		// check fat file format
		gotBin := filepath.Join(dir, "out-amd64-arm64-binary")
		createFatBin(t, gotBin, amd64Bin, arm64Bin)
		if _, err := macho.OpenFat(gotBin); err != nil {
			t.Errorf("invalid fat file: %v\n", err)
		}

		revBin := filepath.Join(dir, "out-arm64-amd64-binary")
		createFatBin(t, revBin, arm64Bin, amd64Bin)
		if _, err := macho.OpenFat(revBin); err != nil {
			t.Errorf("invalid fat file: %v\n", err)
		}

		// if lipo bin exists, execute lipo tests
		lipoBin := lookupLipoBin(t)
		if lipoBin == "" {
			t.Skip("lipo binary does not exist")
		}

		// lipo inspect
		testLipoDetail(t, lipoBin, gotBin)
		testLipoDetail(t, lipoBin, revBin)

		// compare my lipo with original lipo
		wantBin := filepath.Join(dir, "want-binary")
		createFatBinWithLipo(t, lipoBin, wantBin, amd64Bin, arm64Bin)
		testSha256Diff(t, wantBin, gotBin)
	})
}

func testLipoDetail(t *testing.T, lipoBin, gotBin string) {
	t.Helper()
	cmd := exec.Command(lipoBin, "-detailed_info", gotBin)
	if err := cmd.Run(); err != nil {
		t.Errorf("Error lipo -detailed_info: %v\n", err)
	}
}

func testSha256Diff(t *testing.T, wantBin, gotBin string) {
	t.Helper()

	got := calcBinSha256(t, gotBin)
	want := calcBinSha256(t, wantBin)
	if want != got {
		t.Errorf("want %s got %s", want, got)
	}
}

func compileCmd(mainfile, binpath, arch string) *exec.Cmd {
	args := []string{"build", "-o"}
	args = append(args, binpath, mainfile)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH="+arch)
	return cmd
}

func createFatBin(t *testing.T, gotBin, input1, input2 string) {
	t.Helper()
	l := lipo.New(lipo.WithInputs(input1, input2), lipo.WithOutput(gotBin))
	if err := l.Create(); err != nil {
		t.Fatalf("failed to create fat bin %v", err)
	}
}

func createFatBinWithLipo(t *testing.T, lipoBin, wantBin, input1, input2 string) {
	t.Helper()
	// specify 2^14(0x2000) alignment for X86_64 to remove platform dependency.
	cmd := exec.Command(lipoBin, "-segalign", "x86_64", "2000", "-create", input1, input2, "-output", wantBin)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create original fat binary: %v\n %s", err, cmd.String())
	}
}

func calcBinSha256(t *testing.T, p string) string {
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

func lookupLipoBin(t *testing.T) string {
	t.Helper()
	lipoBin, err := exec.LookPath("lipo")
	if errors.Is(err, exec.ErrNotFound) {
		return ""
	}
	if err != nil {
		t.Fatalf("could not find lipo command %v\n", err)
	}
	return lipoBin
}
