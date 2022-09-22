package lipo

import (
	"debug/macho"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func (l *Lipo) Remove(arch string) (err error) {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	abs, err := filepath.Abs(l.in[0])
	if err != nil {
		return nil
	}

	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	cond := func(c macho.Cpu) bool {
		return arch != cpu(c.String())
	}

	fatArches, err := fatArchesFromFatBin(abs, cond)
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	return outputFatBinary(l.out, perm, fatArches)
}

// atArchesFromFatBin returns fatArches from fat binary if `cond` returns true
func fatArchesFromFatBin(path string, cond func(cpu macho.Cpu) bool) ([]*fatArch, error) {
	fat, err := macho.OpenFat(path)
	if err != nil {
		return nil, err
	}
	defer fat.Close()

	if len(fat.Arches) < 1 {
		return nil, errors.New("number of arches must be greater than 1")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fatArches := []*fatArch{}
	for _, hdr := range fat.Arches {
		if !cond(hdr.Cpu) {
			continue
		}
		fatArches = append(fatArches, &fatArch{
			FatArchHeader: hdr.FatArchHeader,
			r:             io.NewSectionReader(f, int64(hdr.Offset), int64(hdr.Size)),
			c:             f,
		})
	}

	if len(fatArches) == len(fat.Arches) {
		return nil, errors.New("found no arch")
	}

	fatHeader := &fatHeader{
		magic: fat.Magic,
		narch: uint32(len(fat.Arches) - 1),
	}

	offset := int64(fatHeader.size())
	for _, hdr := range fatArches {
		offset = align(int64(offset), 1<<int64(hdr.Align))
		// update offset for remove
		hdr.Offset = uint32(offset)
	}

	return fatArches, nil
}
