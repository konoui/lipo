package lipo

import (
	"fmt"
)

func (l *Lipo) Remove(arches ...string) (err error) {
	if err := validateOneInput(l.in); err != nil {
		return err
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

	removed := remove(ff.Arches, arches...)
	if (len(ff.Arches) - len(removed)) != len(arches) {
		diffArch := diff(cpuStrings(ff.Arches), arches)
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	if err := updateAlignBit(removed, l.segAligns); err != nil {
		return err
	}

	return createFatBinary(l.out, removed, perm, l.fat64, l.hideArm64)

}
