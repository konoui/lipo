package lmacho

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func CreateFat[T Object](w io.Writer, objects []T, fat64 bool, hideARM64 bool) error {
	if len(objects) == 0 {
		return errors.New("file contains no images")
	}

	if err := validateHideARM64Objects(objects, hideARM64); err != nil {
		return err
	}

	magic := MagicFat
	if fat64 {
		magic = MagicFat64
	}

	fatArches := newFatArches(objects)
	hdr := makeFatHeader(fatArches, magic, hideARM64)
	if err := sortAndUpdateArches(fatArches, hdr.Magic); err != nil {
		return err
	}

	if err := writeHeaders(w, hdr, fatArches); err != nil {
		return err
	}

	if err := writeArches(w, fatArches, hdr.Magic); err != nil {
		return err
	}

	return nil
}

func newFatArches[T Object](objects []T) []*FatArch {
	arches := make([]*FatArch, len(objects))
	for i, obj := range objects {
		fa := &FatArch{
			sr:  io.NewSectionReader(obj, 0, int64(obj.Size())),
			typ: obj.Type(),
			faHdr: FatArchHeader{
				Cpu:    obj.CPU(),
				SubCpu: obj.SubCPU(),
				Size:   obj.Size(),
				Offset: 0, // will be filled
				Align:  obj.Align(),
			},
		}
		arches[i] = fa
	}
	return arches
}

// writeHEaders validates inputs and write data to destination
func writeHeaders(w io.Writer, hdr FatHeader, arches []*FatArch) error {
	if err := hasDuplicatesErr(arches); err != nil {
		return err
	}

	if err := checkMaxAlignBit(arches); err != nil {
		return err
	}

	// write a fat header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(w, binary.BigEndian, hdr); err != nil {
		return fmt.Errorf("error write fat_header: %w", err)
	}

	// write fat arch headers
	for _, arch := range arches {
		if err := writeFatArchHeader(w, arch.faHdr, hdr.Magic); err != nil {
			return err
		}
	}
	return nil
}

func writeArches(w io.Writer, arches []*FatArch, magic uint32) error {
	firstObjectOffset := FatHeaderSize() + FatArchHeaderSize(magic)*uint64(len(arches))
	offset := firstObjectOffset
	for _, fatArch := range arches {
		if offset < fatArch.faHdr.Offset {
			// write empty data for alignment
			empty := make([]byte, fatArch.faHdr.Offset-offset)
			if _, err := w.Write(empty); err != nil {
				return fmt.Errorf("error alignment: %w", err)
			}
			offset = fatArch.faHdr.Offset
		}

		// write binary data
		if _, err := io.CopyN(w, fatArch, int64(fatArch.Size())); err != nil {
			return fmt.Errorf("error write binary data: %w", err)
		}
		offset += fatArch.Size()
	}

	return nil
}

func writeFatArchHeader(out io.Writer, hdr FatArchHeader, magic uint32) error {
	if magic == MagicFat64 {
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

func makeFatHeader[T Object](objects []T, magic uint32, hideARM64 bool) FatHeader {
	var found bool
	for _, fatArch := range objects {
		if fatArch.CPU() == TypeArm {
			found = true
			break
		}
	}

	if !(hideARM64 && found) {
		return FatHeader{
			Magic: magic,
			NArch: uint32(len(objects)),
		}
	}

	narch := uint32(0)
	for i := range objects {
		if objects[i].CPU() == TypeArm64 {
			continue
		}
		narch++
	}

	return FatHeader{
		Magic: magic,
		NArch: narch,
	}
}
