package lipo

import (
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
)

func (l *Lipo) Extract(arches ...string) error {
	if err := l.validateOneInput(); err != nil {
		return err
	}

	// check1 duplicate arches
	if err := validateInputArches(arches); err != nil {
		return err
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

	fatArches := all.extract(arches...)

	if len(fatArches) != len(arches) {
		diffArch := remove(fatArches.arches(), arches)
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm, &lmacho.FatFileConfig{
		HideArm64: false,
		Fat64:     l.fat64,
	})
}
