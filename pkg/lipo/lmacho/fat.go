package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type FatHeader struct {
	Magic uint32
	NArch uint32
}

const (
	// sizeof(fatHeader) = uint32 * 2
	fatHeaderSize = uint32(4 * 2)
	// sizeof(macho.FatArchHeader) = uint32 * 5
	fatArchHeaderSize = uint32(4 * 5)
)

type FatFile struct {
	Magic        uint32
	Arches       []FatArch
	HiddenArches []FatArch
}

func OpenFat(name string) (*FatFile, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0766)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ff, err := newFatFile(f, name)
	if err != nil {
		return nil, err
	}
	return ff, nil
}

func NewFatFileFromArch(farches []FatArch, hideArm64 bool) *FatFile {
	if !hideArm64 {
		return &FatFile{
			Magic:  macho.MagicFat,
			Arches: farches,
		}
	}

	var found bool
	for _, fatArch := range farches {
		if fatArch.Cpu == CpuTypeArm {
			found = true
			break
		}
	}

	if !found {
		return &FatFile{
			Magic:  macho.MagicFat,
			Arches: farches,
		}
	}

	ff := FatFile{
		Arches:       make([]FatArch, 0, len(farches)),
		HiddenArches: make([]FatArch, 0),
		Magic:        macho.MagicFat,
	}
	for i := range farches {
		if farches[i].Cpu == CpuTypeArm64 {
			ff.HiddenArches = append(ff.HiddenArches, farches[i])
		} else {
			ff.Arches = append(ff.Arches, farches[i])
		}
	}
	return &ff
}

func (f *FatFile) AllArches() []FatArch {
	fa := make([]FatArch, 0, len(f.Arches)+len(f.HiddenArches))
	fa = append(fa, f.Arches...)
	fa = append(fa, f.HiddenArches...)
	return fa
}

func (f *FatFile) FatHeader() *FatHeader {
	return &FatHeader{
		NArch: uint32(len(f.Arches)),
		Magic: f.Magic,
	}
}

func (f *FatFile) Create(out io.Writer) error {
	fatHeader := f.FatHeader()

	arches := f.AllArches()

	if err := ValidateFatArches(arches); err != nil {
		return err
	}

	// sort and update offset
	arches, err := SortBy(arches)
	if err != nil {
		return err
	}

	// sort by offset by asc for effective writing binary data
	sort.Slice(arches, func(i, j int) bool {
		return arches[i].Offset < arches[j].Offset
	})

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("error write fat header: %w", err)
	}

	// write headers
	for _, hdr := range arches {
		if err := binary.Write(out, binary.BigEndian, hdr.FatArchHeader); err != nil {
			return fmt.Errorf("error write arch headers: %w", err)
		}
	}

	// calculate offset with raw narch
	off := fatHeaderSize + fatArchHeaderSize*uint32(len(arches))
	for _, fatArch := range arches {
		if off < fatArch.Offset {
			// write empty data for alignment
			empty := make([]byte, fatArch.Offset-off)
			if _, err := out.Write(empty); err != nil {
				return fmt.Errorf("error alignment: %w", err)
			}
			off = fatArch.Offset
		}

		r, err := fatArch.Open()
		if err != nil {
			return err
		}
		defer r.Close()

		// write binary data
		if _, err := io.CopyN(out, r, int64(fatArch.Size)); err != nil {
			return fmt.Errorf("error write binary data: %w", err)
		}
		off += fatArch.Size
	}

	return nil
}

type FatArch struct {
	macho.FatArchHeader
	FileHeader *macho.FileHeader
	Name       string
	fileOffset uint64
}

func NewFatArch(name string) (*FatArch, error) {
	f, err := macho.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	size := info.Size()
	if size > 1<<32 {
		return nil, fmt.Errorf("%s(%d) exceeds maximum 32 bit size", name, size)
	}

	align := SegmentAlignBit(f)
	if f.Type == macho.TypeObj {
		align = GuessAlignBit(uint64(os.Getpagesize()), alignBitMin, alignBitMax)
	}

	fa := &FatArch{
		Name:       name,
		fileOffset: 0,
		FileHeader: &f.FileHeader,
		FatArchHeader: macho.FatArchHeader{
			Cpu:    f.Cpu,
			SubCpu: f.SubCpu,
			Size:   uint32(size),
			Align:  align,
			// offset will be updated
			Offset: 0,
		},
	}
	return fa, nil
}

func (fa *FatArch) Open() (*File, error) {
	f, err := os.OpenFile(fa.Name, os.O_RDONLY, 0766)
	if err != nil {
		return nil, err
	}
	sr := io.NewSectionReader(f, int64(fa.fileOffset), int64(fa.Size))
	return &File{sr: sr, c: f}, nil
}

func newFatFile(r io.ReaderAt, name string) (*FatFile, error) {
	ff := FatFile{}
	sr := io.NewSectionReader(r, 0, 1<<63-1)

	err := binary.Read(sr, binary.BigEndian, &ff.Magic)
	if err != nil {
		return nil, errors.New("error reading magic number")
	} else if ff.Magic != macho.MagicFat {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], ff.Magic)
		leMagic := binary.LittleEndian.Uint32(buf[:])
		if leMagic == macho.Magic32 || leMagic == macho.Magic64 {
			return nil, macho.ErrNotFat
		} else {
			return nil, errors.New("invalid magic number")
		}
	}

	var narch uint32
	err = binary.Read(sr, binary.BigEndian, &narch)
	if err != nil {
		return nil, errors.New("invalid fat_header")
	}

	if narch < 1 {
		return nil, errors.New("file contains no images")
	}

	ff.Arches = make([]FatArch, narch)
	for i := uint32(0); i < narch; i++ {
		fa := &ff.Arches[i]
		err = binary.Read(sr, binary.BigEndian, &fa.FatArchHeader)
		if err != nil {
			return nil, errors.New("invalid fat_arch header")
		}
		fa.Name = name
		fa.fileOffset = uint64(fa.Offset)

		fr := io.NewSectionReader(sr, int64(fa.Offset), int64(fa.Size))
		f, err := macho.NewFile(fr)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fa.FileHeader = &f.FileHeader
	}

	// handling hidden arm64
	ff.HiddenArches = []FatArch{}
	nextHdrOffset := fatHeaderSize + fatArchHeaderSize*narch
	firstOffset := ff.Arches[0].Offset
	for start := nextHdrOffset + fatArchHeaderSize; start <= firstOffset; start += fatArchHeaderSize {
		var fahdr macho.FatArchHeader
		err := binary.Read(sr, binary.BigEndian, &fahdr)
		if err != nil {
			break
		}

		fr := io.NewSectionReader(sr, int64(fahdr.Offset), int64(fahdr.Size))
		if fahdr.Cpu != CpuTypeArm64 {
			break
		}
		f, err := macho.NewFile(fr)
		if err != nil {
			return nil, fmt.Errorf("hideARM64: %w", err)
		}
		defer f.Close()
		ff.HiddenArches = append(ff.HiddenArches, FatArch{
			FatArchHeader: fahdr,
			FileHeader:    &f.FileHeader,
			Name:          name,
			fileOffset:    uint64(fahdr.Offset),
		})
	}

	if err := ValidateFatArches(ff.AllArches()); err != nil {
		return nil, err
	}

	return &ff, nil
}

func ValidateFatArches(arches []FatArch) error {
	seenArches := make(map[uint64]bool, len(arches))
	for _, fa := range arches {
		seenArch := (uint64(fa.Cpu) << 32) | uint64(fa.SubCpu)
		if o, k := seenArches[seenArch]; o || k {
			return fmt.Errorf("duplicate architecture %s", ToCpuString(fa.Cpu, fa.SubCpu))
		}
		seenArches[seenArch] = true
	}

	return nil
}
