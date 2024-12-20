package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
)

type FatIter struct {
	r         *io.SectionReader
	FatHeader FatHeader
}

func NewFatIter(r io.ReaderAt) (*FatIter, error) {
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

	return &FatIter{r: sr, FatHeader: ff}, nil
}

func (r *FatIter) Next() iter.Seq2[*FatArch, error] {
	return func(yield func(*FatArch, error) bool) {
		nextNArch := uint32(1)
		firstObjectOffset := uint64(0)
		for {
			fa, err := r.next(nextNArch, firstObjectOffset)
			if errors.Is(err, io.EOF) {
				return
			}

			if nextNArch == 1 && err == nil {
				firstObjectOffset = fa.Offset()
			}
			nextNArch++

			if !yield(fa, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

func (r *FatIter) next(nextNArch uint32, firstObjectOffset uint64) (*FatArch, error) {
	magic := r.FatHeader.Magic

	nextFatArchHdrOffset := FatHeaderSize() +
		FatArchHeaderSize(magic)*uint64(nextNArch-1)
	hr := io.NewSectionReader(r.r,
		int64(nextFatArchHdrOffset), int64(FatArchHeaderSize(magic)))
	if nextNArch <= r.FatHeader.NArch {
		fa, err := load(magic, r.r, hr, false)
		if err != nil {
			return nil, err
		}

		return fa, nil
	}

	// for hidden
	if nextFatArchHdrOffset+FatArchHeaderSize(magic) > firstObjectOffset {
		return nil, io.EOF
	}

	fa, err := load(magic, r.r, hr, true)
	if err != nil {
		return nil, fmt.Errorf("hideARM64: %w", err)
	}

	if fa.CPU() != TypeArm64 {
		// TODO handle error
		return nil, io.EOF
	}

	return fa, nil
}

func load(magic uint32, body io.ReaderAt, header io.Reader, hidden bool) (*FatArch, error) {
	hdr, err := readFatArchHeader(header, magic)
	if err != nil {
		return nil, &FormatError{err}
	}

	fa := &FatArch{
		sr: io.NewSectionReader(body,
			int64(hdr.Offset), int64(hdr.Size)),
		faHdr:  *hdr,
		Hidden: hidden,
	}

	return fa, nil

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
