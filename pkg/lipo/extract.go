package lipo

import (
	"fmt"
)

func (l *Lipo) Extract(arches ...string) error {
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

	extracted := extract(ff.Arches, arches...)
	if len(extracted) != len(arches) {
		diffArch := diff(cpuStrings(ff.Arches), arches)
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	if err := updateAlignBit(ff.Arches, l.segAligns); err != nil {
		return err
	}

	return createFatBinary(l.out, extracted, perm, l.fat64, l.hideArm64)
}
