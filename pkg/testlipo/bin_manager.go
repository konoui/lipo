package testlipo

import (
	"debug/macho"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type BinManager struct {
	mu       sync.Mutex
	archBins map[string]string
	Dir      string
	FatBin   string
	arm64Bin string
	mainFile string
}

func NewBinManager(dir string) *BinManager {
	return &BinManager{
		mu:       sync.Mutex{},
		archBins: map[string]string{},
		Dir:      dir,
	}
}

func (bm *BinManager) add(t *testing.T, arches ...string) {
	t.Helper()

	for _, arch := range arches {
		bm.singleAdd(t, arch)
	}
}

func (bm *BinManager) getBinPaths(t *testing.T, arches []string) (paths []string) {
	t.Helper()

	bins := make([]string, len(arches))
	for i, a := range arches {
		bins[i] = bm.getBinPath(t, a)
	}
	return bins
}

func (bm *BinManager) getBinPath(t *testing.T, arch string) (path string) {
	t.Helper()

	bm.mu.Lock()
	defer bm.mu.Unlock()

	b, ok := bm.archBins[arch]
	if !ok {
		t.Fatalf("found no arch: %s", arch)
	}
	return b
}

func (bm *BinManager) singleAdd(t *testing.T, arch string) (path string) {
	t.Helper()

	// arm64 is a base file so create it first.
	if arch != "arm64" && bm.arm64Bin == "" {
		bm.singleAdd(t, "arm64")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// if arch is seen before, return it
	bin, ok := bm.archBins[arch]
	if ok {
		return bin
	}

	archBin := filepath.Join(bm.Dir, arch)
	defer func() {
		if path == "" {
			panic(fmt.Sprintf("empty result for %s %s", arch, archBin))
		}
	}()

	// from file cache
	// _, err := macho.Open(archBin)
	// if err == nil {
	// 	bm.archBins[arch] = archBin
	// 	return archBin
	// }

	// generate a new binary
	if arch == "x86_64" {
		bm.writeMainFile(t)
		compile(t, bm.mainFile, archBin, "amd64")
		bm.archBins[arch] = archBin
		return archBin
	}

	if arch == "arm64" {
		bm.writeMainFile(t)
		compile(t, bm.mainFile, archBin, "arm64")
		bm.arm64Bin = archBin
		bm.archBins[arch] = archBin
		return archBin
	}

	if strings.HasPrefix(arch, "obj_") {
		copyAndManipulate(t, bm.arm64Bin, archBin, arch[4:], macho.TypeObj)
		bm.archBins[arch] = archBin
		return archBin
	}

	copyAndManipulate(t, bm.arm64Bin, archBin, arch, macho.TypeExec)
	bm.archBins[arch] = archBin
	return archBin
}

func (bm *BinManager) writeMainFile(t *testing.T) {
	t.Helper()

	if bm.mainFile != "" {
		return
	}
	mainfile := filepath.Join(bm.Dir, "main.go")
	err := os.WriteFile(mainfile, []byte(godata), os.ModePerm)
	fatalIf(t, err)

	bm.mainFile = mainfile
}
