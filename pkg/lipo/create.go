package lipo

import (
	"debug/macho"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Create() error {
	inputs, err := newCreateInputs(l.in...)
	if err != nil {
		return err
	}

	var perm fs.FileMode
	for _, in := range inputs {
		if in.perm > perm {
			perm = in.perm
		}
	}

	fatArches, err := fatArchesFromCreateInputs(inputs)
	if err != nil {
		return err
	}
	defer fatArches.close()

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm)
}

type createInput struct {
	path  string
	align uint32
	hdr   *macho.FileHeader
	size  int64
	perm  fs.FileMode
}

func newCreateInputs(paths ...string) ([]*createInput, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no input files specified")
	}

	inputs := make([]*createInput, len(paths))
	for idx, path := range paths {
		in, err := newCreateInput(path)
		if err != nil {
			return nil, fmt.Errorf("%v for %s", err, path)
		}
		inputs[idx] = in
	}

	if err := validateCreateInputs(inputs); err != nil {
		return nil, err
	}

	return inputs, nil
}

func validateCreateInputs(inputs []*createInput) error {
	// validate inputs
	seenArches := make(map[string]bool, len(inputs))
	for _, i := range inputs {
		seenArch := mcpu.ToString(i.hdr.Cpu, i.hdr.SubCpu)
		if o, k := seenArches[seenArch]; o || k {
			return fmt.Errorf("duplicate architecture %s", seenArch)
		}
		seenArches[seenArch] = true
	}
	return nil
}

func newCreateInput(bin string) (*createInput, error) {
	path, err := filepath.Abs(bin)
	if err != nil {
		return nil, err
	}

	f, err := macho.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var align uint32
	if f.Type == macho.TypeObj {
		align = guessAlignBit(uint64(os.Getpagesize()), alignBitMin, alignBitMax)
	} else {
		align = segmentAlignBit(f)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	size := info.Size()
	perm := info.Mode().Perm()

	i := &createInput{
		path:  path,
		align: align,
		hdr:   &f.FileHeader,
		size:  size,
		perm:  perm,
	}
	return i, nil
}
