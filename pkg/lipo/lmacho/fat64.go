package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// sizeof(fatHeader) = uint32 * 2
	fat64HeaderSize = uint64(4 * 2)
	// sizeof(Fat64ArchHeader) = uint32 * 4 + uint64 * 2
	fat64ArchHeaderSize = uint64(4*4 + 8*2)
)

const MagicFat64 = macho.MagicFat + 1

type Fat64ArchHeader struct {
	Cpu      macho.Cpu
	SubCpu   uint32
	Offset   uint64
	Size     uint64
	Align    uint32
	Reserved uint32
}

type Fat64Arch struct {
	Fat64ArchHeader
	FileHeader *macho.FileHeader
	Name       string
	fileOffset uint64
}

type Fat64File struct {
	Magic        uint32
	Arches       []Fat64Arch
	HiddenArches []Fat64Arch
}

func OpenFat64(name string) (*Fat64File, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0766)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ff, err := newFat64File(f, name)
	if err != nil {
		return nil, err
	}
	return ff, nil
}

func (f *Fat64File) AllArches() []Fat64Arch {
	fa := make([]Fat64Arch, 0, len(f.Arches)+len(f.HiddenArches))
	fa = append(fa, f.Arches...)
	fa = append(fa, f.HiddenArches...)
	return fa
}

func newFat64File(r io.ReaderAt, name string) (*Fat64File, error) {
	ff := Fat64File{}
	sr := io.NewSectionReader(r, 0, 1<<63-1)

	err := binary.Read(sr, binary.BigEndian, &ff.Magic)
	if err != nil {
		return nil, errors.New("error reading magic number")
	} else if ff.Magic != MagicFat64 {
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

	ff.Arches = make([]Fat64Arch, narch)
	for i := uint32(0); i < narch; i++ {
		fa := &ff.Arches[i]
		err = binary.Read(sr, binary.BigEndian, &fa.Fat64ArchHeader)
		if err != nil {
			return nil, errors.New("invalid fat_arch header")
		}
		fa.Name = name
		fa.fileOffset = fa.Offset

		fr := io.NewSectionReader(sr, int64(fa.Offset), int64(fa.Size))
		f, err := macho.NewFile(fr)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fa.FileHeader = &f.FileHeader
	}

	// handling hidden arm64
	ff.HiddenArches = []Fat64Arch{}
	nextHdrOffset := fat64HeaderSize + fat64ArchHeaderSize*uint64(narch)
	firstOffset := ff.Arches[0].Offset
	for start := nextHdrOffset + fat64ArchHeaderSize; start <= firstOffset; start += fat64ArchHeaderSize {
		var fahdr Fat64ArchHeader
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
		ff.HiddenArches = append(ff.HiddenArches, Fat64Arch{
			Fat64ArchHeader: fahdr,
			FileHeader:      &f.FileHeader,
			Name:            name,
			fileOffset:      fahdr.Offset,
		})
	}

	if err := ValidateFat64Arches(ff.AllArches()); err != nil {
		return nil, err
	}

	return &ff, nil
}

func ValidateFat64Arches(arches []Fat64Arch) error {
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
