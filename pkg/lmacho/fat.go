package lmacho

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"
)

const MagicFat = macho.MagicFat

// FatHeader presents a header for a fat 32 bit and fat 64 bit
// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type FatHeader struct {
	Magic uint32
	NArch uint32
}

// FatArchHeader presents an architecture header for a Macho-0 32 bit and 64 bit
type FatArchHeader struct {
	Cpu    Cpu
	SubCpu SubCpu
	Offset uint64
	Size   uint64
	Align  uint32
}

type Object interface {
	CPU() Cpu
	SubCPU() SubCpu
	Size() uint64
	Align() uint32
	Type() macho.Type
	CPUString() string
	io.Reader
	io.ReaderAt
}

var (
	_ Object = &FatArch{}
	_ Object = &Arch{}
)

// FatArch presents an object of fat file
type FatArch struct {
	faHdr  FatArchHeader
	Hidden bool
	typ    macho.Type
	sr     *io.SectionReader
}

func (fa *FatArch) CPU() Cpu {
	return fa.faHdr.Cpu
}

func (fa *FatArch) SubCPU() SubCpu {
	return fa.faHdr.SubCpu
}

func (fa *FatArch) Size() uint64 {
	return fa.faHdr.Size
}

func (fa *FatArch) Align() uint32 {
	return fa.faHdr.Align
}

func (fa *FatArch) Type() macho.Type {
	return fa.typ
}

func (fa *FatArch) Offset() uint64 {
	return fa.faHdr.Offset
}

func (fa *FatArch) CPUString() string {
	return ToCpuString(fa.CPU(), fa.SubCPU())
}

func (fa *FatArch) Read(p []byte) (int, error) {
	return fa.sr.Read(p)
}

func (fa *FatArch) ReadAt(p []byte, off int64) (int, error) {
	return fa.sr.ReadAt(p, off)
}

func (fa *FatArch) Seek(offset int64, whence int) (int64, error) {
	return fa.sr.Seek(offset, whence)
}

// Arch presents an object of thin file
type Arch struct {
	cpu    Cpu
	subCpu SubCpu
	align  uint32
	typ    macho.Type
	sr     *io.SectionReader
}

func (a *Arch) CPU() Cpu {
	return a.cpu
}

func (a *Arch) SubCPU() SubCpu {
	return a.subCpu
}

func (a *Arch) Size() uint64 {
	return uint64(a.sr.Size())
}

func (a *Arch) Align() uint32 {
	return a.align
}

func (a *Arch) Type() macho.Type {
	return a.typ
}

func (a *Arch) CPUString() string {
	return ToCpuString(a.CPU(), a.SubCPU())
}

func (a *Arch) Read(p []byte) (int, error) {
	return a.sr.Read(p)
}

func (a *Arch) ReadAt(p []byte, off int64) (int, error) {
	return a.sr.ReadAt(p, off)
}

func (a *Arch) Seek(offset int64, whence int) (int64, error) {
	return a.sr.Seek(offset, whence)
}

type FormatError struct {
	Err error
}

func (e *FormatError) Error() string {
	return fmt.Sprintf("invalid file format %s", e.Err.Error())
}

type FatFile struct {
	FatHeader
	Arches []*FatArch
}

// NewFatFile is wrapper for Fat NewFatIter
func NewFatFile(ra io.ReaderAt) (*FatFile, error) {
	r, err := NewFatIter(ra)
	if err != nil {
		return nil, err
	}

	fa := &FatFile{
		Arches:    make([]*FatArch, 0),
		FatHeader: r.FatHeader,
	}

	for a, err := range r.Next() {
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		fa.Arches = append(fa.Arches, a)
	}
	return fa, nil
}

func NewArch(sr *io.SectionReader) (*Arch, error) {
	mf, err := macho.NewFile(sr)
	if err != nil {
		fe := &macho.FormatError{}
		if errors.As(err, &fe) {
			return nil, &FormatError{Err: err}
		}
		return nil, err
	}

	if _, err := sr.Seek(0, io.SeekStart); err != nil {
		return nil, err // TODO detail error
	}

	align := SegmentAlignBit(mf)
	if mf.Type == macho.TypeObj {
		alignBitMin := AlignBitMin64
		if mf.Magic == macho.Magic32 {
			alignBitMin = AlignBitMin32
		}
		align = GuessAlignBit(uint64(os.Getpagesize()), alignBitMin, AlignBitMax)
	}

	return &Arch{
		cpu:    mf.Cpu,
		subCpu: mf.SubCpu,
		typ:    mf.Type,
		align:  align,
		sr:     sr,
	}, nil
}

func FatHeaderSize() uint64 {
	// sizeof(FatHeader) = uint32 * 2
	return uint64(4 * 2)
}

func FatArchHeaderSize(magic uint32) uint64 {
	if magic == MagicFat64 {
		// sizeof(Fat64ArchHeader) = uint32 * 4 + uint64 * 2
		return uint64(4*4 + 8*2)
	}
	// sizeof(macho.FatArchHeader) = uint32 * 5
	return uint64(4 * 5)
}
