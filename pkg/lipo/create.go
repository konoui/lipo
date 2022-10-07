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
	defer func() { _ = close(fatArches) }()

	if err := updateAlignBit(fatArches, l.segAligns); err != nil {
		return err
	}

	return outputFatBinary(l.out, perm, fatArches)
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
		return nil, fmt.Errorf("no inputs")
	}

	inputs := make([]*createInput, len(paths))
	for idx, path := range paths {
		in, err := newCreateInput(path)
		if err != nil {
			return nil, fmt.Errorf("%v for %s", err, path)
		}
		inputs[idx] = in
	}

	// validate inputs
	seenArches := make(map[string]bool, len(inputs))
	for _, i := range inputs {
		seenArch := mcpu.ToString(i.hdr.Cpu, i.hdr.SubCpu)
		if o, k := seenArches[seenArch]; o || k {
			return nil, fmt.Errorf("duplicate architecture %s", seenArch)
		}
		seenArches[seenArch] = true
	}

	return inputs, nil
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

	if f.Type != macho.TypeExec {
		return nil, fmt.Errorf("not supported non TypeExec %s", bin)
	}

	align := segmentAlignBit(f)

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

func fatArchesFromCreateInputs(inputs []*createInput) ([]*fatArch, error) {
	fatHdr := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(inputs)),
	}

	fatArches := make([]*fatArch, 0, len(inputs))

	offset := int64(fatHdr.size())
	for _, in := range inputs {
		offset = align(offset, 1<<in.align)

		// validate addressing boundary since size and offset of fat32 are uint32
		if !(boundaryOK(offset) && boundaryOK(in.size)) {
			return nil, fmt.Errorf("exceeds maximum fat32 size at %s", in.path)
		}

		hdrOffset := uint32(offset)
		hdrSize := uint32(in.size)
		hdr := macho.FatArchHeader{
			Cpu:    in.hdr.Cpu,
			SubCpu: in.hdr.SubCpu,
			Offset: hdrOffset,
			Size:   hdrSize,
			Align:  in.align,
		}

		offset += int64(hdr.Size)

		f, err := os.Open(in.path)
		if err != nil {
			return nil, err
		}
		fatArches = append(fatArches, &fatArch{
			FatArchHeader: hdr,
			r:             f,
			c:             f,
		})
	}

	return sortByArch(fatArches)
}
