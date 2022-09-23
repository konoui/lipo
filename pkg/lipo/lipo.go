package lipo

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const (
	alignBitAmd64 = 13
	alignBitArm64 = 14
)

type Lipo struct {
	in  []string
	out string
}

type Option func(l *Lipo)

func WithInputs(in ...string) Option {
	return func(l *Lipo) {
		l.in = in
	}
}

func WithOutput(out string) Option {
	return func(l *Lipo) {
		l.out = out
	}
}

func New(opts ...Option) *Lipo {
	l := &Lipo{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(l)
	}
	return l
}

// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type fatHeader struct {
	magic uint32
	narch uint32
}

func (h *fatHeader) size() uint32 {
	// sizeof(fatHeader) = uint32 * 2
	sizeofFatHdr := uint32(4 * 2)
	// sizeof(macho.FatArchHeader) = uint32 * 5
	sizeofFatArchHdr := uint32(4 * 5)
	size := sizeofFatHdr + sizeofFatArchHdr*h.narch
	return size
}

var _ io.ReadCloser = &fatArch{}

// fatArch consist of FatArchHeader and io.Reader for binary
type fatArch struct {
	macho.FatArchHeader
	r io.Reader
	c io.Closer
}

func (fa *fatArch) Read(p []byte) (int, error) {
	return fa.r.Read(p)
}

func (fa *fatArch) Close() error {
	if fa == nil || fa.c == nil {
		return nil
	}

	err := fa.c.Close()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}
	return err
}

func close(closers []*fatArch) error {
	msg := ""
	for _, closer := range closers {
		err := closer.Close()
		if err != nil {
			msg += err.Error()
		}
	}
	if msg != "" {
		return fmt.Errorf("close errors: %s", msg)
	}
	return nil
}

func sortByArch(fatArches []*fatArch) ([]*fatArch, error) {
	sort.Slice(fatArches, func(i, j int) bool {
		icpu := fatArches[i].Cpu
		isub := fatArches[i].SubCpu
		v1 := (uint64(icpu) << 32) | uint64(isub)
		jcpu := fatArches[j].Cpu
		jsub := fatArches[j].SubCpu
		v2 := (uint64(jcpu) << 32) | uint64(jsub)
		return v1 < v2
	})

	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(fatArches)),
	}

	// update offset
	offset := int64(fatHeader.size())
	for _, hdr := range fatArches {
		offset = align(int64(offset), 1<<int64(hdr.Align))
		if validateFatSize(offset) {
			return nil, fmt.Errorf("exceeds maximum fat32 size")
		}
		hdr.Offset = uint32(offset)
		offset += int64(hdr.Size)
	}
	return fatArches, nil
}

func alignBit(cpu macho.Cpu, sub uint32) uint32 {
	if CpuString(cpu, sub) == "x86_64" {
		return alignBitAmd64
	}
	return alignBitArm64
}

func align(offset, v int64) int64 {
	return (offset + v - 1) / v * v
}

func validateFatSize(s int64) bool {
	return s >= 1<<32
}

func outputFatBinary(p string, perm os.FileMode, fatArches []*fatArch) (err error) {
	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Chmod(perm); cerr != nil && err == nil {
			err = cerr
			return
		}
		if ferr := out.Close(); ferr != nil && err == nil {
			err = ferr
			return
		}
	}()

	return createFatBinary(out, fatArches)
}

func createFatBinary(out io.Writer, fatArches []*fatArch) error {
	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(fatArches)),
	}

	// sort by offset by asc for effective writing binary data
	sort.Slice(fatArches, func(i, j int) bool {
		return fatArches[i].Offset < fatArches[j].Offset
	})

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("failed to write fat header: %w", err)
	}

	// write headers
	for _, hdr := range fatArches {
		if err := binary.Write(out, binary.BigEndian, hdr.FatArchHeader); err != nil {
			return fmt.Errorf("failed to write arch headers: %w", err)
		}
	}

	off := fatHeader.size()
	for _, fatArch := range fatArches {
		if off < fatArch.Offset {
			// write empty data for alignment
			empty := make([]byte, fatArch.Offset-off)
			if _, err := out.Write(empty); err != nil {
				return err
			}
			off = fatArch.Offset
		}

		// write binary data
		if _, err := io.CopyN(out, fatArch.r, int64(fatArch.Size)); err != nil {
			return fmt.Errorf("failed to write binary data: %w", err)
		}
		off += fatArch.Size
	}

	return nil
}

func contain(tg string, l []string) bool {
	for _, s := range l {
		if tg == s {
			return true
		}
	}
	return false
}
