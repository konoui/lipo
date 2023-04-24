package lipo

import (
	"fmt"
	"io"
	"os"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
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

	ff, err := lmacho.OpenFat(fatBin)
	if err != nil {
		return err
	}
	all := fatArches(ff.AllArches())

	fatArches := all.extract(arch)
	if len(fatArches) == 0 {
		return fmt.Errorf("fat input file (%s) does not contain the specified architecture (%s) to thin it to", fatBin, arch)
	}

	fatArch := fatArches[0]
	return l.thin(perm, fatArch)
}

func (l *Lipo) thin(perm os.FileMode, fatArch lmacho.FatArch) error {
	out, err := createTemp(l.out)
	if err != nil {
		return err
	}

	r, err := fatArch.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := io.CopyN(out, r, int64(fatArch.Size)); err != nil {
		return fmt.Errorf("error write binary data: %w", err)
	}

	if err := out.Chmod(perm); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Rename(out.Name(), l.out)
}
