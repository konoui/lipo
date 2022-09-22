package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (l *Lipo) Extract(arches ...string) error {
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
		return contain(s, arches)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	return outputFatBinary(l.out, perm, fatArches)
}
