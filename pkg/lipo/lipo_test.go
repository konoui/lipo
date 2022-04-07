package lipo

import (
	"debug/macho"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var data = `
package main

import "fmt"

func main() {
        fmt.Println("Hello World")
}
`

func TestLipo_Create(t *testing.T) {
	t.Run("generate", func(t *testing.T) {
		dir := t.TempDir()
		mainfile := filepath.Join(dir, "main.go")
		err := os.WriteFile(mainfile, []byte(data), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		arm64 := "arm64"
		amd64 := "amd64"
		armbin := filepath.Join(dir, arm64)
		amdbin := filepath.Join(dir, amd64)
		armCmd := compileCmd(mainfile, armbin, arm64)
		amdCmd := compileCmd(mainfile, amdbin, amd64)
		wg := &sync.WaitGroup{}
		for _, cmd := range []*exec.Cmd{
			armCmd,
			amdCmd,
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

		out := filepath.Join(dir, "lipo-binary")
		l := New(out, armbin, amdbin)
		if err := l.Create(); err != nil {
			t.Fatal(err)
		}

		// check fat file format
		if _, err := macho.OpenFat(out); err != nil {
			t.Errorf("invalid fat file: %v\n", err)
		}
	})
}

func compileCmd(mainfile, binpath, arch string) *exec.Cmd {
	args := []string{"build", "-o"}
	args = append(args, binpath, mainfile)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH="+arch)
	return cmd
}
