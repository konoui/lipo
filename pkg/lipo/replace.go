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
		return fmt.Errorf("%s specified but fat file: %s does not contain that architecture", diffArch, fatBin)
	}

	fatArches := all.replace(fatInputs)
	if err := fatArches.updateAlignBit(l.segAligns); err != nil {
		return err
	}

	return fatArches.createFatBinary(l.out, perm)
}

func createInputsFromReplaceInputs(ins []*ReplaceInput) ([]*createInput, error) {
	creates := make([]*createInput, 0, len(ins))
	for _, r := range ins {
		if !mcpu.IsSupported(r.Arch) {
			return nil, fmt.Errorf("unsupported architecture %s", r.Arch)
		}

		i, err := newCreateInput(r.Bin)
		if err != nil {
			return nil, err
		}

		if arch := mcpu.ToString(i.hdr.Cpu, i.hdr.SubCpu); arch != r.Arch {
			return nil, fmt.Errorf("specified architecture: %s for replacement file: %s does not match the file's architecture", r.Arch, r.Bin)
		}
		creates = append(creates, i)
	}

	if err := validateCreateInputs(creates); err != nil {
		return nil, err
	}
	return creates, nil
}
