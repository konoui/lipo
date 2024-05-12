package lipo

import (
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lmacho"
)

func (l *Lipo) Thin(arch string) error {
	if err := validateOneInput(l.in); err != nil {
		return err
	}

	if !lmacho.IsSupportedCpu(arch) {
		return fmt.Errorf(unsupportedArchFmt, arch)
	}

	fatBin := l.in[0]
	perm, err := perm(fatBin)
	if err != nil {
		return err
	}

	ff, err := OpenFatFile(fatBin)
	if err != nil {
		return err
	}
	defer ff.Close()

	extracted := extract(ff.Arches, arch)
	if len(extracted) == 0 {
		return fmt.Errorf("fat input file (%s) does not contain the specified architecture (%s) to thin it to", fatBin, arch)
	}

	return l.thin(perm, extracted[0])
}

func (l *Lipo) thin(perm os.FileMode, arch Arch) error {
	out, err := createTemp(l.out)
	if err != nil {
		return err
	}

	if _, err := io.CopyN(out, arch, int64(arch.Size())); err != nil {
		return fmt.Errorf("error write binary data: %w", err)
	}

	if err := out.Chmod(perm); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	// close before rename
	if err := out.Close(); err != nil {
		return err
	}

	return os.Rename(out.Name(), l.out)
}
