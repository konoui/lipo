package lipo

import (
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
)

func (l *Lipo) Remove(arches ...string) (err error) {
	if err := l.validateOneInput(); err != nil {
		return err
	}

	if err := validateInputArches(arches); err != nil {
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

	if l.hideArm64 {
		if err := hideArmObjectErr(all); err != nil {
			return err
		}
	}

	fatArches := all.remove(arches...)
	if (len(all) - len(fatArches)) != len(arches) {
		diffArch := remove(all.arches(), arches)
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm, &lmacho.FatFileConfig{
		HideArm64: l.hideArm64,
		Fat64:     l.fat64,
	})
}
