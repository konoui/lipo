package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (l *Lipo) Thin(arch string) error {
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

	fatArches, err := fatArchesFromFatBin(abs, func(hdr *macho.FatArchHeader) bool {
		s := cpuString(hdr.Cpu, hdr.SubCpu)
		return s == arch
	})
	defer func() { _ = close(fatArches) }()

	if err != nil {
		return err
	}

	out, err := os.Create(l.out)
	if err != nil {
		return err
	}

	fatArch := fatArches[0]
	if _, err := io.CopyN(out, fatArch, int64(fatArch.Size)); err != nil {
		return fmt.Errorf("failed to write binary data: %w", err)
	}

	if err := out.Chmod(perm); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}
	return nil
}
