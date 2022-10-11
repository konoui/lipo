package lipo

import (
	"fmt"
	"os"
)

func (l *Lipo) ExtractFamily(arches ...string) error {
	if err := l.validateOneInput(); err != nil {
		return err
	}

	if err := validateInputArches(arches); err != nil {
		return err
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	all, err := fatArchesFromFatBin(fatBin)
	if err != nil {
		return err
	}

	fatArches := all.extractFamily(arches...)
	if len(fatArches) == 0 {
		return fmt.Errorf(noMatchFmt, arches[0], fatBin)
	}

	if len(fatArches) == 1 {
		return l.thin(perm, fatArches[0])
	}

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm)
}
