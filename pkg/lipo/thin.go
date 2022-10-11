package lipo

import (
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Thin(arch string) error {
	if err := l.validateOneInput(); err != nil {
		return err
	}

	if !mcpu.IsSupported(arch) {
		return fmt.Errorf("unsupported architecture %s", arch)
	}

	fatBin := l.in[0]

	info, err := os.Stat(l.in[0])
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	all, err := fatArchesFromFatBin(fatBin)
	if err != nil {
		return err
	}
	defer all.close()

	fatArches := all.extract(arch)
	if len(fatArches) == 0 {
		return fmt.Errorf("fat input file (%s) does not contain the specified architecture (%s) to thin it to", fatBin, arch)
	}

	fatArch := fatArches[0]
	return l.thin(perm, fatArch)
}

func (l *Lipo) thin(perm os.FileMode, fatArch *fatArch) error {
	out, err := os.Create(l.out)
	if err != nil {
		return err
	}

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
