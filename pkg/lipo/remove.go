package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (l *Lipo) Remove(arches ...string) (err error) {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	for _, arch := range arches {
		if !isSupportedArch(arch) {
			return fmt.Errorf("unsupported architecture %s", arch)
		}
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

	fatArches, err := fatArchesFromFatBin(abs, func(hdr *macho.FatArchHeader) bool {
		s := cpuString(hdr.Cpu, hdr.SubCpu)
		return !contain(s, arches)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	return outputFatBinary(l.out, perm, fatArches)
}

// atArchesFromFatBin gathers fatArches from fat binary header if `cond` returns true
func fatArchesFromFatBin(path string, cond func(hdr *macho.FatArchHeader) bool) ([]*fatArch, error) {
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
		if cond(&hdr.FatArchHeader) {
			fatArches = append(fatArches, &fatArch{
				FatArchHeader: hdr.FatArchHeader,
				r:             io.NewSectionReader(f, int64(hdr.Offset), int64(hdr.Size)),
				c:             f,
			})
		}
	}

	if len(fatArches) == len(fat.Arches) {
		return nil, errors.New("found no architecture")
	}

	if len(fatArches) == 0 {
		return nil, errors.New("result arch is zero")
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
