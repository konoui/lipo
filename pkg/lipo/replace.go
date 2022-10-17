package lipo

import (
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
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

	all, err := fatArchesFromFatBin(fatBin)
	if err != nil {
		return err
	}
	defer all.close()

	// check1 replace input arch equals to replace input bin
	createInputs, err := createInputsFromReplaceInputs(inputs)
	if err != nil {
		return err
	}
	fatInputs, err := fatArchesFromCreateInputs(createInputs)
	if err != nil {
		return err
	}
	defer fatInputs.close()

	// check2 fat bin contains all replace inputs
	if !all.contains(fatInputs) {
		diffArch := remove(all.arches(), fatInputs.arches())
		return fmt.Errorf(noMatchFmt, diffArch, fatBin)
	}

	fatArches := all.replace(fatInputs)
	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm, l.hideArm64)
}

func createInputsFromReplaceInputs(ins []*ReplaceInput) ([]*createInput, error) {
	archInputs := make([]*ArchInput, len(ins))
	for i, r := range ins {
		if !mcpu.IsSupported(r.Arch) {
			return nil, fmt.Errorf(unsupportedArchFmt, r.Arch)
		}
		archInputs[i] = &ArchInput{Bin: r.Bin, Arch: r.Arch}
	}

	creates, err := newCreateInputs(archInputs...)
	if err != nil {
		return nil, err
	}

	return creates, nil
}
