package lipo

import (
	"debug/macho"
	"errors"
	"os"
	"path/filepath"
)

func (l *Lipo) Extract(arches ...string) error {
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

	fatArches, err := fatArchesFromFatBin(abs, func(c macho.Cpu) bool {
		return contain(lipoCpu(c.String()), arches)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	return outputFatBinary(l.out, perm, fatArches)
}
