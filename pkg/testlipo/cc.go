package testlipo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var cdata = `
#include <stdio.h>

int main()
{
  printf("Hello World\n");
}`

func NewObject(t *testing.T, path string) string {
	mainfile := filepath.Join(filepath.Dir(path), "main.c")
	err := os.WriteFile(mainfile, []byte(cdata), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("gcc", "-g", "-o", path, "-O", "-c", mainfile)
	execute(t, cmd, true)
	return mainfile
}
