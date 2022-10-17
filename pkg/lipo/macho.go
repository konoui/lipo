package lipo

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

type FatFile struct {
	Magic        uint32
	Arches       []macho.FatArch
	HiddenArches []macho.FatArch
	closer       io.Closer
}

const fatArchHeaderSize = 5 * 4

func OpenFat(name string) (*FatFile, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	ff, err := NewFatFile(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	ff.closer = f
	return ff, nil
}

func NewFatFile(r io.ReaderAt) (*FatFile, error) {
	var ff FatFile
	sr := io.NewSectionReader(r, 0, 1<<63-1)

	// Read the fat_header struct, which is always in big endian.
	// Start with the magic number.
	err := binary.Read(sr, binary.BigEndian, &ff.Magic)
	if err != nil {
		return nil, errors.New("error reading magic number")
	} else if ff.Magic != macho.MagicFat {
		// See if this is a Mach-O file via its magic number. The magic
		// must be converted to little endian first though.
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

	seenArches := make(map[uint64]bool, narch)

	ff.Arches = make([]macho.FatArch, narch)
	for i := uint32(0); i < narch; i++ {
		fa := &ff.Arches[i]
		err = binary.Read(sr, binary.BigEndian, &fa.FatArchHeader)
		if err != nil {
			return nil, errors.New("invalid fat_arch header")
		}

		fr := io.NewSectionReader(sr, int64(fa.Offset), int64(fa.Size))
		fa.File, err = macho.NewFile(fr)
		if err != nil {
			return nil, err
		}

		seenArch := (uint64(fa.Cpu) << 32) | uint64(fa.SubCpu)
		if o, k := seenArches[seenArch]; o || k {
			return nil, fmt.Errorf("duplicate architecture cpu=%v, subcpu=%#x", fa.Cpu, fa.SubCpu)
		}
		seenArches[seenArch] = true
	}

	// handling hidden arm64
	ff.HiddenArches = []macho.FatArch{}
	fatHeader := &fatHeader{narch: narch}
	nextHdrOffset := fatHeader.size()
	firstOffset := ff.Arches[0].Offset
	for start := nextHdrOffset + fatArchHeaderSize; start <= firstOffset; start += fatArchHeaderSize {
		var fahdr macho.FatArchHeader
		err := binary.Read(sr, binary.BigEndian, &fahdr)
		if err == nil {
			fr := io.NewSectionReader(sr, int64(fahdr.Offset), int64(fahdr.Size))
			if fahdr.Cpu == mcpu.TypeArm64 {
				f, err := macho.NewFile(fr)
				if err != nil {
					return nil, fmt.Errorf("hideARM64: %w", err)
				}
				ff.HiddenArches = append(ff.HiddenArches, macho.FatArch{File: f, FatArchHeader: fahdr})
			}
		}
	}

	return &ff, nil
}

func (ff *FatFile) Close() error {
	var err error
	if ff.closer != nil {
		err = ff.closer.Close()
		ff.closer = nil
	}
	return err
}
