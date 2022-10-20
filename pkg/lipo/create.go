package lipo

import (
	"github.com/konoui/lipo/pkg/util"
)

func (l *Lipo) Create() error {
	archInputs := append(l.arches, util.Map(l.in, func(v string) *ArchInput { return &ArchInput{Bin: v} })...)
	if len(archInputs) == 0 {
		return errNoInput
	}

	fatArches, err := newFatArches(archInputs...)
	if err != nil {
		return err
	}

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	if l.hideArm64 {
		if err := hideARmObjectErr(fatArches); err != nil {
			return err
		}
	}

	return fatArches.createFatBinary(l.out, 0731, l.hideArm64)
}
