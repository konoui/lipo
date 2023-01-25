package lipo

import (
	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

func (l *Lipo) Create() error {
	l.arches = append(l.arches, util.Map(l.in, func(v string) *ArchInput { return &ArchInput{Bin: v} })...)
	archInputs := l.arches
	if len(archInputs) == 0 {
		return errNoInput
	}

	fatArches, err := newFatArches(archInputs...)
	if err != nil {
		return err
	}

	// apple lipo will use a last file permission
	// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#L1124
	perm, err := perm(fatArches[len(fatArches)-1].Name)
	if err != nil {
		return err
	}

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	if l.hideArm64 {
		if err := hideArmObjectErr(fatArches); err != nil {
			return err
		}
	}

	return fatArches.createFatBinary(l.out, perm, &lmacho.FatFileConfig{
		HideArm64: l.hideArm64,
		Fat64:     l.fat64,
	})
}
