package lipo

import (
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	alignBitAmd64 = 12
	alignBitArm64 = 14
)

type Lipo struct {
	in  []string
	out string
}

func New(out string, in ...string) *Lipo {
	return &Lipo{
		out: out,
		in:  in,
	}
}

func (l *Lipo) Create() error {
	inputs, err := newInputs(l.in...)
	if err != nil {
		return err
	}
	out, err := newOutput(l.out, inputs)
	if err != nil {
		return err
	}
	return out.create()
}

type input struct {
	path string
	hdr  *macho.FileHeader
	size int64
	perm fs.FileMode
}

func (i *input) alignBit() uint32 {
	if i.hdr.Cpu == macho.CpuAmd64 {
		return alignBitAmd64
	}
	return alignBitArm64
}

func newInputs(paths ...string) ([]*input, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no inputs")
	}

	inputs := make([]*input, len(paths))
	for idx, path := range paths {
		in, err := newInput(path)
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

func newInput(bin string) (*input, error) {
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

	i := &input{
		path: path,
		hdr:  &f.FileHeader,
		size: size,
		perm: perm,
	}
	return i, nil
}

// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type fatHeader struct {
	magic uint32
	narch uint32
}

type output struct {
	path       string
	fatHdr     fatHeader
	inputPaths []string
	fatArches  map[string]macho.FatArchHeader
	perm       fs.FileMode
}

func (h *fatHeader) size() uint32 {
	// sizeof(fatHeader) = uint32 * 2
	sizeofFatHdr := uint32(4 * 2)
	// sizeof(macho.FatArchHeader) = uint32 * 5
	sizeofFatArchHdr := uint32(4 * 5)
	size := sizeofFatHdr + sizeofFatArchHdr*h.narch
	return size
}

func align(offset, v int64) int64 {
	return (offset + v - 1) / v * v
}

func validateFatSize(s int64) bool {
	return s >= 1<<32
}

func newOutput(path string, inputs []*input) (*output, error) {
	fatHdr := fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(inputs)),
	}

	fatArches := make(map[string]macho.FatArchHeader)
	paths := make([]string, len(inputs))
	offset := int64(fatHdr.size())
	for idx, in := range inputs {
		offset = align(offset, 1<<in.alignBit())

		// validate adressing boundary since size and offset of fat32 are uint32
		if validateFatSize(offset) || validateFatSize(in.size) {
			return nil, fmt.Errorf("exceeds maximum fat32 size at %s", path)
		}

		hdrOffset := uint32(offset)
		hdrSize := uint32(in.size)
		hdr := macho.FatArchHeader{
			Cpu:    in.hdr.Cpu,
			SubCpu: in.hdr.SubCpu,
			Offset: hdrOffset,
			Size:   hdrSize,
			Align:  in.alignBit(),
		}

		fatArches[in.path] = hdr
		paths[idx] = in.path

		offset += int64(hdr.Size)
	}

	var perm fs.FileMode
	for _, in := range inputs {
		if in.perm > perm {
			perm = in.perm
		}
	}

	o := &output{
		path:       path,
		fatHdr:     fatHdr,
		fatArches:  fatArches,
		inputPaths: paths,
		perm:       perm,
	}

	return o, nil
}

func (o *output) create() (err error) {
	out, err := os.Create(o.path)
	if err != nil {
		return err
	}
	defer func() {
		if ferr := out.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, o.fatHdr); err != nil {
		return fmt.Errorf("failed to wirte fat header: %w", err)
	}

	// write headers
	for _, key := range o.inputPaths {
		hdr := o.fatArches[key]
		if err := binary.Write(out, binary.BigEndian, hdr); err != nil {
			return fmt.Errorf("failed to write arch headers: %w", err)
		}
	}

	off := o.fatHdr.size()
	for _, path := range o.inputPaths {
		hdr := o.fatArches[path]
		if off < hdr.Offset {
			// write empty data for alignment
			empty := make([]byte, hdr.Offset-off)
			if _, err = out.Write(empty); err != nil {
				return err
			}
			off = hdr.Offset
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// write binary data
		if _, err := io.CopyN(out, f, int64(hdr.Size)); err != nil {
			return fmt.Errorf("failed to write binary data: %w", err)
		}
		off += hdr.Size
	}

	if err := out.Chmod(o.perm); err != nil {
		return err
	}

	return nil
}
