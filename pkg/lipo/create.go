package lipo

import (
	"debug/macho"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/util"
)

func (l *Lipo) Create() error {
	archInputs := append(l.arches, util.Map(l.in, func(v string) *ArchInput { return &ArchInput{Bin: v} })...)
	inputs, err := newCreateInputs(archInputs...)
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

func newCreateInputs(in ...*ArchInput) ([]*createInput, error) {
	if len(in) == 0 {
		return nil, errNoInput
	}

	inputs := make([]*createInput, len(in))
	for idx, path := range in {
		in, err := newCreateInput(path)
		if err != nil {
			return nil, err
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

func newCreateInput(in *ArchInput) (*createInput, error) {
	path, err := filepath.Abs(in.Bin)
	if err != nil {
		return nil, err
	}

	f, err := macho.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if in.Arch != "" {
		if f.Cpu == 0 && f.SubCpu == 0 {
			cpu, sub, ok := mcpu.ToCpu(in.Arch)
			if !ok {
				return nil, fmt.Errorf(unsupportedArchFmt, in.Arch)
			}
			f.Cpu = cpu
			f.SubCpu = sub
		} else if mcpu.ToString(f.Cpu, f.SubCpu) != in.Arch {
			return nil, fmt.Errorf("specified architecture: %s for input file: %s does not match the file's architecture", in.Arch, in.Bin)
		}
	}

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
