package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

func (l *Lipo) Create() error {
	l.arches = append(l.arches, util.Map(l.in, func(v string) *ArchInput { return &ArchInput{Bin: v} })...)
	archInputs := l.arches
	if len(archInputs) == 0 {
		return errNoInput
	}

	arches, err := OpenArches(archInputs)
	if err != nil {
		return err
	}
	defer close(arches...)

	if err := updateAlignBit(arches, l.segAligns); err != nil {
		return err
	}

	// apple lipo will use a last file permission
	// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#L1124
	perm, err := perm(arches[len(arches)-1].Name())
	if err != nil {
		return err
	}

	return createFatBinary(l.out, arches, perm, l.fat64, l.hideArm64)
}

func createFatBinary[T Arch](path string, arches []T, perm os.FileMode, fat64 bool, hideARM64 bool) error {
	if len(arches) == 0 {
		return errors.New("no inputs would result in an empty fat file")
	}

	if hideARM64 {
		for _, obj := range arches {
			if obj.Type() == macho.TypeObj {
				return fmt.Errorf("hideARM64 specified but thin file %s is not of type MH_EXECUTE", obj.Name())
			}
		}
	}

	out, err := createTemp(path)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := lmacho.CreateFat(out, arches, fat64, hideARM64); err != nil {
		return err
	}

	if err := out.Chmod(perm); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	// close before rename
	if err := out.Close(); err != nil {
		return err
	}

	// atomic operation
	return os.Rename(out.Name(), path)
}
