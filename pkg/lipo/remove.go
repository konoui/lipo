package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Remove(arches ...string) (err error) {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	for _, arch := range arches {
		if !mcpu.IsSupported(arch) {
			return fmt.Errorf("unsupported architecture %s", arch)
		}
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	num := 0
	fatArches, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		num++
		s := mcpu.ToString(hdr.Cpu, hdr.SubCpu)
		return !contain(s, arches)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	// TODO replace <arch_file> with actual value
	if (num - len(fatArches)) != len(arches) {
		return fmt.Errorf(noMatchFmt, "-remove", fatBin)
	}

	if err := updateAlignBit(fatArches, l.segAligns); err != nil {
		return err
	}

	return outputFatBinary(l.out, perm, fatArches)
}

var (
	errFoundNoFatArch = errors.New("result arch is zero")
)

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

	if len(fatArches) == 0 {
		return nil, errFoundNoFatArch
	}

	return sortByArch(fatArches)
}
