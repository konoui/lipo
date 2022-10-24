package lipo

import (
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

type ReplaceInput struct {
	Arch string
	Bin  string
}

func (l *Lipo) Replace(inputs []*ReplaceInput) error {
	if err := l.validateOneInput(); err != nil {
		return err
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	ff, err := lmacho.OpenFat(fatBin)
	if err != nil {
		return err
	}
	all := fatArches(ff.AllArches())

	// check1 replace input arch equals to replace input bin
	archInputs := util.Map(inputs, func(i *ReplaceInput) *ArchInput {
		return &ArchInput{
			Arch: i.Arch,
			Bin:  i.Bin,
		}
	})
	fatInputs, err := newFatArches(archInputs...)
	if err != nil {
		return err
	}

	if l.hideArm64 {
		if err := hideArmObjectErr(fatInputs); err != nil {
			return err
		}
	}

	// check2 fat bin contains all replace inputs
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
