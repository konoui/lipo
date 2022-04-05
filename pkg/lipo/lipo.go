package lipo

import (
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"unsafe"
)

const (
	align = 1 << 14
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
	out := newOutput(l.out, inputs)
	return out.create()
}

type input struct {
	path string
	hdr  *macho.FileHeader
	size int64
	perm fs.FileMode
}

func newInputs(paths ...string) ([]*input, error) {
	inputs := make([]*input, 0, len(paths))
	for _, bin := range paths {
		i, err := newInput(bin)
		if err != nil {
			return nil, fmt.Errorf("%v for %s", err, bin)
		}
		inputs = append(inputs, i)
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
	path           string
	fatHeader      fatHeader
	inputPaths     []string
	fatArchHeaders map[string]macho.FatArchHeader
	perm           fs.FileMode
}

func newOutput(path string, inputs []*input) *output {
	fatHeader := fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(inputs)),
	}

	fatArchHeaders := make(map[string]macho.FatArchHeader)
	paths := make([]string, 0, len(inputs))
	offset := int64(align)
	for _, i := range inputs {
		hdr := macho.FatArchHeader{
			Cpu:    i.hdr.Cpu,
			SubCpu: i.hdr.SubCpu,
			Offset: uint32(offset),
			Size:   uint32(i.size),
			Align:  align,
		}

		fatArchHeaders[i.path] = hdr
		paths = append(paths, i.path)

		offset += i.size
		offset = (offset + align - 1) / align * align
	}

	var perm fs.FileMode
	for _, i := range inputs {
		if i.perm > perm {
			perm = i.perm
		}
	}

	o := &output{
		path:           path,
		fatHeader:      fatHeader,
		fatArchHeaders: fatArchHeaders,
		inputPaths:     paths,
		perm:           perm,
	}

	return o
}

func (o *output) create() (err error) {
	out, err := os.Create(o.path)
	if err != nil {
		return err
	}
	defer func() {
		if ferr := out.Close(); ferr != nil {
			if ferr != nil && err == nil {
				err = ferr
			}
		}
	}()

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, o.fatHeader); err != nil {
		return fmt.Errorf("failed to wirte fat header: %w", err)
	}

	// write headers
	for _, key := range o.inputPaths {
		hdr := o.fatArchHeaders[key]
		if err := binary.Write(out, binary.BigEndian, hdr); err != nil {
			return fmt.Errorf("failed to write arch headers: %w", err)
		}
	}

	// TODO
	var fhdr fatHeader
	var fahdr macho.FatArchHeader
	size := unsafe.Sizeof(fhdr) + unsafe.Sizeof(fahdr)*uintptr(o.fatHeader.narch)
	off := uint32(size)
	for _, path := range o.inputPaths {
		hdr := o.fatArchHeaders[path]
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
		if _, err := io.Copy(out, f); err != nil {
			return fmt.Errorf("failed to write binary data: %w", err)
		}
		off += hdr.Size
	}

	if err := out.Chmod(o.perm); err != nil {
		return err
	}

	return nil
}
