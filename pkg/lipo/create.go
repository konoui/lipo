package lipo

import (
	"debug/macho"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

	return outputFatBinary(l.out, perm, fatArches)
}

type createInput struct {
	path string
	hdr  *macho.FileHeader
	size int64
	perm fs.FileMode
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
	seenArches := make(map[uint64]bool, len(inputs))
	for _, i := range inputs {
		seenArch := (uint64(i.hdr.Cpu) << 32) | uint64(i.hdr.SubCpu)
		if o, k := seenArches[seenArch]; o || k {
			return nil, fmt.Errorf("duplicate architecture cpu=%v, subcpu=%#x", i.hdr.Cpu, i.hdr.SubCpu)
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

	// Note Magic32 is not tested
	if f.Magic != macho.Magic64 {
		return nil, fmt.Errorf("unsupported magic %#x", f.Magic)
	}

	// Note CpuPpc64 is not tested
	if f.Cpu != macho.CpuAmd64 && f.Cpu != macho.CpuArm64 {
		return nil, fmt.Errorf("unsupported cpu %s", f.Cpu)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	size := info.Size()
	perm := info.Mode().Perm()

	i := &createInput{
		path: path,
		hdr:  &f.FileHeader,
		size: size,
		perm: perm,
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
		offset = align(offset, 1<<alignBit(in.hdr.Cpu))

		// validate addressing boundary since size and offset of fat32 are uint32
		if validateFatSize(offset) || validateFatSize(in.size) {
			return nil, fmt.Errorf("exceeds maximum fat32 size at %s", in.path)
		}

		hdrOffset := uint32(offset)
		hdrSize := uint32(in.size)
		hdr := macho.FatArchHeader{
			Cpu:    in.hdr.Cpu,
			SubCpu: in.hdr.SubCpu,
			Offset: hdrOffset,
			Size:   hdrSize,
			Align:  alignBit(in.hdr.Cpu),
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
	return fatArches, nil
}
