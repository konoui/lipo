package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Thin(arch string) error {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	fatArches, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		s := mcpu.ToString(hdr.Cpu, hdr.SubCpu)
		return s == arch
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(fatArches) }()

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
