package lipo

import (
	"debug/macho"
	"errors"
	"os"
	"path/filepath"
)

func (l *Lipo) Extract(arch string) error {
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
		return arch == cpu(c.String())
	}

	fatArches, err := fatArchesFromFatBin(abs, cond)
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

	return outputFatBinary(l.out, perm, fatArches)
}
