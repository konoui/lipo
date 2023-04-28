package lipo

import (
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
)

func (l *Lipo) ExtractFamily(arches ...string) error {
	if err := validateOneInput(l.in); err != nil {
		return err
	}

	if err := validateInputArches(arches); err != nil {
		return err
	}

	fatBin := l.in[0]
	perm, err := perm(fatBin)
	if err != nil {
		return err
	}

	ff, err := lmacho.NewFatFile(fatBin)
	if err != nil {
		return err
	}
	all := fatArches(ff.AllArches())

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

	return fatArches.createFatBinary(l.out, perm, &lmacho.FatFileConfig{
		HideArm64: false,
		Fat64:     l.fat64,
	})
}
