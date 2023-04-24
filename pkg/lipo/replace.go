package lipo

import (
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

type ReplaceInput struct {
	Arch string
	Bin  string
}

func (l *Lipo) Replace(inputs []*ReplaceInput) error {
	if err := validateOneInput(l.in); err != nil {
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

	archInputs := util.Map(inputs, func(i *ReplaceInput) *ArchInput {
		return &ArchInput{
			Arch: i.Arch,
			Bin:  i.Bin,
		}
	})
	// check an ReplaceInput.Arch is equal to an arch in ReplaceInput.Bin
	// check no duplication arches
	fatInputs, err := newFatArches(archInputs...)
	if err != nil {
		return err
	}

	if l.hideArm64 {
		if err := hideArmObjectErr(fatInputs); err != nil {
			return err
		}
	}

	// check fat bin contains all arches in replace inputs
	if !all.contains(fatInputs) {
		diffArch := remove(all.arches(), fatInputs.arches())
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	fatArches := all.replace(fatInputs)
	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm, &lmacho.FatFileConfig{
		HideArm64: l.hideArm64,
		Fat64:     l.fat64,
	})
}
