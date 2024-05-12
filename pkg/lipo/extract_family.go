package lipo

import (
	"fmt"
)

func (l *Lipo) ExtractFamily(arches ...string) error {
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

	extracted := extractFamily(ff.Arches, arches...)
	if len(extracted) == 0 {
		return fmt.Errorf(noMatchFmt, arches[0], fatBin)
	}

	if len(extracted) == 1 {
		return l.thin(perm, extracted[0])
	}

	if err := updateAlignBit(ff.Arches, l.segAligns); err != nil {
		return err
	}

	return createFatBinary(l.out, extracted, perm, l.fat64, l.hideArm64)
}
