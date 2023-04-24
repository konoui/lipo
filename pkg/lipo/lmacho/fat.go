package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/konoui/go-qsort"
)

// FatHeader presets a header for a fat 32 bit and fat 64 bit
// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type FatHeader struct {
	Magic uint32
	NArch uint32
}

// FatArchHeader presets an architecture header for a Macho-0 32 bit and 64 bit
type FatArchHeader struct {
	Cpu    macho.Cpu
	SubCpu uint32
	Offset uint64
	Size   uint64
	Align  uint32
}

// FatArch has architecture information for a fat file.
type FatArch struct {
	FatArchHeader
	FileHeader *macho.FileHeader
	Name       string
	fileOffset uint64
}

// FatFile presets an universal file
type FatFile struct {
	Magic        uint32
	Arches       []FatArch
	HiddenArches []FatArch
}

// offset returns an offset of first macho header
func (f *FatFile) offset() uint64 {
	return f.fatHeaderSize() + f.fatArchHeaderSize()*uint64(len(f.Arches)+len(f.HiddenArches))
}

func (f *FatFile) fatHeaderSize() uint64 {
	// sizeof(FatHeader) = uint32 * 2
	return uint64(4 * 2)
}

func (f *FatFile) fatArchHeaderSize() uint64 {
	if f.Magic == MagicFat64 {
		// sizeof(Fat64ArchHeader) = uint32 * 4 + uint64 * 2
		return uint64(4*4 + 8*2)
	}
	// sizeof(macho.FatArchHeader) = uint32 * 5
	return uint64(4 * 5)
}

func (f *FatFile) fatHeader() *FatHeader {
	return &FatHeader{
		NArch: uint32(len(f.Arches)),
		Magic: f.Magic,
	}
}

func (f *FatFile) readFatArchHeader(r io.Reader) (*FatArchHeader, error) {
	if f.Magic == MagicFat64 {
		var fatHdr fatArch64Header
		err := binary.Read(r, binary.BigEndian, &fatHdr)
		if err != nil {
			return nil, errors.New("invalid fat arch64 header")
		}

		return &FatArchHeader{
			Cpu:    fatHdr.Cpu,
			SubCpu: fatHdr.SubCpu,
			Align:  fatHdr.Align,
			Size:   fatHdr.Size,
			Offset: fatHdr.Offset,
		}, nil
	}

	var fatHdr macho.FatArchHeader
	err := binary.Read(r, binary.BigEndian, &fatHdr)
	if err != nil {
		return nil, errors.New("invalid fat arch header")
	}
	return &FatArchHeader{
		Cpu:    fatHdr.Cpu,
		SubCpu: fatHdr.SubCpu,
		Align:  fatHdr.Align,
		Size:   uint64(fatHdr.Size),
		Offset: uint64(fatHdr.Offset),
	}, nil
}

func (f *FatFile) writeFatArchHeader(out io.Writer, hdr FatArchHeader) error {
	if f.Magic == MagicFat64 {
		fatArchHdr := fatArch64Header{FatArchHeader: hdr, Reserved: 0}
		if err := binary.Write(out, binary.BigEndian, fatArchHdr); err != nil {
			return fmt.Errorf("error write fat_arch64 header: %w", err)
		}
		return nil
	}

	fatArchHdr := macho.FatArchHeader{
		Cpu:    hdr.Cpu,
		SubCpu: hdr.SubCpu,
		Offset: uint32(hdr.Offset),
		Size:   uint32(hdr.Size),
		Align:  hdr.Align,
	}
	if err := binary.Write(out, binary.BigEndian, fatArchHdr); err != nil {
		return fmt.Errorf("error write fat_arch header: %w", err)
	}
	return nil
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

type FatFileConfig struct {
	HideArm64 bool
	Fat64     bool
}

func NewFatFileFromArches(farches []FatArch, cfg *FatFileConfig) *FatFile {
	if cfg == nil {
		cfg = &FatFileConfig{}
	}

	magic := macho.MagicFat
	if cfg.Fat64 {
		magic = MagicFat64
	}

	if !cfg.HideArm64 {
		return &FatFile{
			Magic:  magic,
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
			Magic:  magic,
			Arches: farches,
		}
	}

	ff := FatFile{
		Magic:        magic,
		Arches:       make([]FatArch, 0, len(farches)),
		HiddenArches: make([]FatArch, 0),
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

func (f *FatFile) sortedArches() ([]FatArch, error) {
	arches := f.AllArches()
	qsort.Slice(arches, CmpArchFunc)

	// update offset
	offset := f.offset()
	for i := range arches {
		offset = align(offset, 1<<arches[i].Align)
		arches[i].Offset = offset
		offset += arches[i].Size
		if f.Magic == macho.MagicFat && !boundaryOK(offset) {
			return nil, fmt.Errorf("exceeds maximum 32 bit size at %s. please handle it as fat64", arches[i].Name)
		}
	}

	return arches, nil
}

func (f *FatFile) Create(out io.Writer) error {
	fatHeader := f.fatHeader()

	// sort and update offset
	arches, err := f.sortedArches()
	if err != nil {
		return err
	}

	if err := hasDuplicatesErr(arches); err != nil {
		return err
	}

	// write a fat header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("error write fat_header: %w", err)
	}

	// write architecture headers
	for _, hdr := range arches {
		if err := f.writeFatArchHeader(out, hdr.FatArchHeader); err != nil {
			return err
		}
	}

	off := f.offset()
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
	align := SegmentAlignBit(f)
	if f.Type == macho.TypeObj {
		alignBitMin := alignBitMin64
		if f.Magic == macho.Magic32 {
			alignBitMin = alignBitMin32
		}
		align = GuessAlignBit(uint64(os.Getpagesize()), alignBitMin, alignBitMax)
	}

	fa := &FatArch{
		Name:       name,
		fileOffset: 0,
		FileHeader: &f.FileHeader,
		FatArchHeader: FatArchHeader{
			Cpu:    f.Cpu,
			SubCpu: f.SubCpu,
			Size:   uint64(size),
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
	}

	if ff.Magic != macho.MagicFat && ff.Magic != MagicFat64 {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], ff.Magic)
		leMagic := binary.LittleEndian.Uint32(buf[:])
		if leMagic == macho.Magic32 || leMagic == macho.Magic64 {
			return nil, macho.ErrNotFat
		}
		return nil, errors.New("invalid magic number")
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
		fatHdr, err := ff.readFatArchHeader(sr)
		if err != nil {
			return nil, err
		}

		fr := io.NewSectionReader(sr, int64(fatHdr.Offset), int64(fatHdr.Size))
		f, err := macho.NewFile(fr)
		if err != nil {
			return nil, fmt.Errorf("invalid macho-file: %w", err)
		}
		defer f.Close()

		fa := &ff.Arches[i]
		fa.FatArchHeader = *fatHdr
		fa.FileHeader = &f.FileHeader
		fa.Name = name
		fa.fileOffset = uint64(fatHdr.Offset)
	}

	// handling hidden arm64
	ff.HiddenArches = []FatArch{}
	nextHdrOffset := ff.offset()
	firstOffset := ff.Arches[0].Offset
	for {
		if nextHdrOffset+ff.fatArchHeaderSize() > firstOffset {
			break
		}

		hr := io.NewSectionReader(sr, int64(nextHdrOffset), int64(ff.fatArchHeaderSize()))
		fatHdr, err := ff.readFatArchHeader(hr)
		if err != nil {
			return nil, err
		}

		fr := io.NewSectionReader(sr, int64(fatHdr.Offset), int64(fatHdr.Size))
		if fatHdr.Cpu != CpuTypeArm64 {
			break
		}
		f, err := macho.NewFile(fr)
		if err != nil {
			return nil, fmt.Errorf("hideARM64: %w", err)
		}
		defer f.Close()
		ff.HiddenArches = append(ff.HiddenArches, FatArch{
			FatArchHeader: *fatHdr,
			FileHeader:    &f.FileHeader,
			Name:          name,
			fileOffset:    uint64(fatHdr.Offset),
		})
		nextHdrOffset += ff.fatArchHeaderSize()
	}

	if err := hasDuplicatesErr(ff.AllArches()); err != nil {
		return nil, err
	}

	return &ff, nil
}

func hasDuplicatesErr(arches []FatArch) error {
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
