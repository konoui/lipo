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
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	fatArches, err := fatArchesFromFatBin(fatBin)
	if err != nil {
		return err
	}
	defer fatArches.close()

	fatArches = fatArches.extract(arch)
	if len(fatArches) == 0 {
		return fmt.Errorf("fat input file (%s) does not contain the specified architecture (%s) to thin it to", fatBin, arch)
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
