package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Extract(arches ...string) error {
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

	fatArches, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		s := mcpu.ToString(hdr.Cpu, hdr.SubCpu)
		return contain(s, arches)
	})
	if err != nil {
		if errors.Is(err, errFoundNoFatArch) {
			return fmt.Errorf(noMatchFmt, "-extract", fatBin)
		}
		return err
	}
	defer func() { _ = close(fatArches) }()

	// TODO replace <arch_file> with actual value
	if len(fatArches) != len(arches) {
		return fmt.Errorf(noMatchFmt, "-extract", fatBin)
	}

	if err := updateAlignBit(fatArches, l.segAligns); err != nil {
		return err
	}

	return outputFatBinary(l.out, perm, fatArches)
}
