package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	r                 *io.SectionReader
	FatHeader         FatHeader
	firstObjectOffset uint64
	nextNArch         uint32
}

func NewReader(r io.ReaderAt) (*Reader, error) {
	sr := io.NewSectionReader(r, 0, 1<<63-1)

	var ff FatHeader
	err := binary.Read(sr, binary.BigEndian, &ff.Magic)
	if err != nil {
		return nil, &FormatError{errors.New("error reading magic number")}
	}

	if ff.Magic != macho.MagicFat && ff.Magic != MagicFat64 {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], ff.Magic)
		leMagic := binary.LittleEndian.Uint32(buf[:])
		if leMagic == macho.Magic32 || leMagic == macho.Magic64 {
			return nil, ErrThin
		}
		return nil, &FormatError{errors.New("invalid magic number")}
	}

	err = binary.Read(sr, binary.BigEndian, &ff.NArch)
	if err != nil {
		return nil, &FormatError{errors.New("invalid fat_header")}
	}

	if ff.NArch < 1 {
		return nil, &FormatError{errors.New("file contains no images")}
	}

	return &Reader{r: sr, FatHeader: ff, nextNArch: 1}, nil
}

func (r *Reader) Next() (*FatArch, error) {
	defer func() {
		r.nextNArch++
	}()
	magic := r.FatHeader.Magic
	if r.nextNArch <= r.FatHeader.NArch {
		faHdr, err := readFatArchHeader(r.r, magic)
		if err != nil {
			return nil, &FormatError{err}
		}

		fa := &FatArch{
			sr:     io.NewSectionReader(r.r, int64(faHdr.Offset), int64(faHdr.Size)),
			faHdr:  *faHdr,
			Hidden: false,
		}

		if r.firstObjectOffset == 0 {
			r.firstObjectOffset = faHdr.Offset
		}

		return fa, nil
	}

	// hidden arches
	nextObjectOffset := FatHeaderSize() + FatArchHeaderSize(magic)*uint64(r.nextNArch-1)
	// require to add fatArchHeaderSize, to read the header
	if nextObjectOffset+FatArchHeaderSize(magic) > r.firstObjectOffset {
		return nil, io.EOF
	}

	hr := io.NewSectionReader(r.r, int64(nextObjectOffset), int64(FatArchHeaderSize(magic)))
	faHdr, err := readFatArchHeader(hr, magic)
	if err != nil {
		return nil, &FormatError{fmt.Errorf("hideARM64: %w", err)}
	}

	if faHdr.Cpu != TypeArm64 {
		// TODO handle error
		return nil, io.EOF
	}

	return &FatArch{
		sr:     io.NewSectionReader(r.r, int64(faHdr.Offset), int64(faHdr.Size)),
		faHdr:  *faHdr,
		Hidden: true,
	}, nil
}

func readFatArchHeader(r io.Reader, magic uint32) (*FatArchHeader, error) {
	if magic == MagicFat64 {
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
