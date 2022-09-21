package lipo

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type newFatArchHeader struct {
	macho.FatArchHeader
	oldOffset uint32
}

func (l *Lipo) Remove(arch string) error {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	abs, err := filepath.Abs(l.in[0])
	if err != nil {
		return nil
	}

	fat, err := macho.OpenFat(abs)
	if err != nil {
		return err
	}

	if len(fat.Arches) < 1 {
		return errors.New("less than 2 arches")
	}

	fatArchHeaders := []*newFatArchHeader{}
	for _, hdr := range fat.Arches {
		if arch == cpu(hdr.Cpu.String()) {
			continue
		}
		fatArchHeaders = append(fatArchHeaders, &newFatArchHeader{
			FatArchHeader: hdr.FatArchHeader,
			oldOffset:     hdr.FatArchHeader.Offset,
		})
	}

	if len(fatArchHeaders) == len(fat.Arches) {
		return fmt.Errorf("found no arch %s", arch)
	}

	fatHeader := fatHeader{
		magic: fat.Magic,
		narch: uint32(len(fat.Arches) - 1),
	}

	offset := int64(fatHeader.size())
	for _, hdr := range fatArchHeaders {
		offset = align(int64(offset), 1<<int64(hdr.Align))
		// update offset for remove
		hdr.Offset = uint32(offset)
	}

	out, err := os.Create(l.out)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("failed to wirte fat header: %w", err)
	}

	for _, hdr := range fatArchHeaders {
		if err := binary.Write(out, binary.BigEndian, hdr.FatArchHeader); err != nil {
			return fmt.Errorf("failed to write arch headers: %w", err)
		}
	}

	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()

	off := fatHeader.size()
	for _, hdr := range fatArchHeaders {
		if off < hdr.Offset {
			// write empty data for alignment
			empty := make([]byte, hdr.Offset-off)
			if _, err = out.Write(empty); err != nil {
				return err
			}
			off = hdr.Offset
		}

		if _, err := f.Seek(int64(hdr.oldOffset), io.SeekStart); err != nil {
			return err
		}

		// write binary data
		if _, err := io.CopyN(out, f, int64(hdr.Size)); err != nil {
			return fmt.Errorf("failed to write binary data: %w", err)
		}
		off += hdr.Size
	}

	return nil
}

func cpu(s string) string {
	switch s {
	case "CpuArm64":
		return "arm64"
	case "CpuAmd64":
		return "x86_64"
	}
	panic(s)
}
