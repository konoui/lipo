package lipo

import (
	"fmt"
)

func (l *Lipo) Replace(inputs []*ReplaceInput) error {
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

	arches, err := OpenArches(inputs)
	if err != nil {
		return err
	}
	defer close(arches...)

	// check fat bin contains all arches in replace inputs
	if !contains(ff.Arches, arches...) {
		diffArch := diff(cpuStrings(ff.Arches), cpuStrings(arches))
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	newArches := replace(ff.Arches, arches)
	if err := updateAlignBit(newArches, l.segAligns); err != nil {
		return err
	}

	return createFatBinary(l.out, newArches, perm, l.fat64, l.hideArm64)
}
